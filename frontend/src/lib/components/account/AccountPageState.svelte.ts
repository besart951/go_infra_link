import { GetNotificationPreferenceUseCase } from '$lib/application/useCases/notification/getNotificationPreferenceUseCase.js';
import { SaveNotificationPreferenceUseCase } from '$lib/application/useCases/notification/saveNotificationPreferenceUseCase.js';
import { SendNotificationEmailVerificationUseCase } from '$lib/application/useCases/notification/sendNotificationEmailVerificationUseCase.js';
import { VerifyNotificationEmailUseCase } from '$lib/application/useCases/notification/verifyNotificationEmailUseCase.js';
import { getErrorMessage } from '$lib/api/client.js';
import { addToast } from '$lib/components/toast.svelte';
import {
  createNotificationPreferenceFormValues,
  hasVerifiedNotificationEmail,
  normalizeNotificationPreferenceInput,
  type NotificationPreferenceFormValues,
  type UpsertUserNotificationPreferenceRequest,
  type UserNotificationPreference
} from '$lib/domain/notification/index.js';
import { t as translate } from '$lib/i18n/index.js';
import { notificationPreferenceRepository } from '$lib/infrastructure/api/notificationPreferenceRepository.js';
import { teamRepository } from '$lib/infrastructure/api/teamRepository.js';
import {
  userRepository,
  type UpdateUserRequest,
  type User
} from '$lib/infrastructure/api/userRepository.js';
import type { AccountTab } from './types.js';

const getNotificationPreference = new GetNotificationPreferenceUseCase(
  notificationPreferenceRepository
);
const saveNotificationPreference = new SaveNotificationPreferenceUseCase(
  notificationPreferenceRepository
);
const sendNotificationEmailVerification = new SendNotificationEmailVerificationUseCase(
  notificationPreferenceRepository
);
const verifyNotificationEmail = new VerifyNotificationEmailUseCase(
  notificationPreferenceRepository
);

export class AccountPageState {
  activeTab = $state<AccountTab>('information');
  currentUser = $state<User | null>(null);
  firstName = $state('');
  lastName = $state('');
  email = $state('');
  newPassword = $state('');
  confirmPassword = $state('');
  isSavingProfile = $state(false);
  isSavingPassword = $state(false);
  isSavingNotifications = $state(false);
  isSendingEmailVerification = $state(false);
  isVerifyingNotificationEmail = $state(false);
  isLoadingNotifications = $state(false);
  isLoading = $state(true);
  userTeams = $state<string[]>([]);
  teamsError = $state<string | null>(null);
  notificationPreference = $state<UserNotificationPreference | null>(null);
  notificationDraft = $state<NotificationPreferenceFormValues>(
    createNotificationPreferenceFormValues(null)
  );
  verificationCode = $state('');
  notificationsError = $state<string | null>(null);

  readonly permissions = $derived(this.currentUser?.permissions ?? []);
  readonly notificationEmailVerified = $derived(
    hasVerifiedNotificationEmail(this.notificationPreference)
  );
  readonly notificationIsDirty = $derived.by(() => {
    const initial = normalizeNotificationPreferenceInput(
      createNotificationPreferenceFormValues(this.notificationPreference)
    );
    const current = normalizeNotificationPreferenceInput(this.notificationDraft);
    return JSON.stringify(initial) !== JSON.stringify(current);
  });
  readonly canSendEmailVerification = $derived(
    Boolean(this.notificationPreference?.notification_email) &&
      !this.notificationIsDirty &&
      !this.isSendingEmailVerification
  );

  applyUserToForm(user: User): void {
    this.firstName = user.first_name;
    this.lastName = user.last_name;
    this.email = user.email;
  }

  async loadUserTeams(userId: string): Promise<void> {
    this.teamsError = null;
    this.userTeams = [];

    try {
      const teamsResponse = await teamRepository.list({ page: 1, limit: 100, search: '' });
      const memberLists = await Promise.all(
        teamsResponse.items.map(async (team) => ({
          team,
          members: await teamRepository.listMembers(team.id, { page: 1, limit: 1000 })
        }))
      );

      this.userTeams = memberLists
        .filter((entry) => entry.members.items.some((member) => member.user_id === userId))
        .map((entry) => entry.team.name);
    } catch (error) {
      this.teamsError = getErrorMessage(error);
    }
  }

  async loadNotificationPreference(): Promise<void> {
    this.isLoadingNotifications = true;
    this.notificationsError = null;

    try {
      this.notificationPreference = await getNotificationPreference.execute();
      this.notificationDraft = createNotificationPreferenceFormValues(this.notificationPreference);
    } catch (error) {
      this.notificationsError = getErrorMessage(error);
    } finally {
      this.isLoadingNotifications = false;
    }
  }

  async loadAccount(): Promise<void> {
    this.isLoading = true;
    try {
      const user = await userRepository.getCurrent();
      this.currentUser = user;
      this.applyUserToForm(user);
      await Promise.all([this.loadUserTeams(user.id), this.loadNotificationPreference()]);
    } catch (error) {
      addToast(getErrorMessage(error), 'error');
    } finally {
      this.isLoading = false;
    }
  }

  async handleInformationSubmit(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    if (!this.currentUser) return;

    this.isSavingProfile = true;
    try {
      const payload: UpdateUserRequest = {
        first_name: this.firstName,
        last_name: this.lastName,
        email: this.email
      };
      const updated = await userRepository.updateCurrent(this.currentUser.id, payload);
      this.currentUser = updated;
      this.applyUserToForm(updated);
      addToast(translate('messages.account_info_saved'), 'success');
    } catch (error) {
      addToast(getErrorMessage(error), 'error');
    } finally {
      this.isSavingProfile = false;
    }
  }

  async handlePasswordSubmit(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    if (!this.currentUser) return;

    if (this.newPassword.length < 8) {
      addToast(
        translate('validation.password_too_short', {
          field: translate('auth.password'),
          min: 8
        }),
        'error'
      );
      return;
    }

    if (this.newPassword !== this.confirmPassword) {
      addToast(
        translate('validation.must_match', {
          field1: translate('auth.new_password'),
          field2: translate('auth.confirm_password')
        }),
        'error'
      );
      return;
    }

    this.isSavingPassword = true;
    try {
      await userRepository.updateCurrentPassword(this.currentUser.id, this.newPassword);
      this.newPassword = '';
      this.confirmPassword = '';
      addToast(translate('messages.account_password_saved'), 'success');
    } catch (error) {
      addToast(getErrorMessage(error), 'error');
    } finally {
      this.isSavingPassword = false;
    }
  }

  async handleNotificationSubmit(event: SubmitEvent): Promise<void> {
    event.preventDefault();

    this.isSavingNotifications = true;
    this.notificationsError = null;
    try {
      const payload: UpsertUserNotificationPreferenceRequest = normalizeNotificationPreferenceInput(
        this.notificationDraft
      );
      this.notificationPreference = await saveNotificationPreference.execute(payload);
      this.notificationDraft = createNotificationPreferenceFormValues(this.notificationPreference);
      this.verificationCode = '';
      addToast(translate('messages.account_notifications_saved'), 'success');
    } catch (error) {
      this.notificationsError = getErrorMessage(error);
      addToast(this.notificationsError, 'error');
    } finally {
      this.isSavingNotifications = false;
    }
  }

  resetNotificationDraft(): void {
    this.notificationDraft = createNotificationPreferenceFormValues(this.notificationPreference);
    this.notificationsError = null;
  }

  async handleSendEmailVerification(): Promise<void> {
    this.isSendingEmailVerification = true;
    this.notificationsError = null;
    try {
      this.notificationPreference = await sendNotificationEmailVerification.execute();
      this.notificationDraft = createNotificationPreferenceFormValues(this.notificationPreference);
      this.verificationCode = '';
      addToast(translate('messages.account_notification_email_code_sent'), 'success');
    } catch (error) {
      this.notificationsError = getErrorMessage(error);
      addToast(this.notificationsError, 'error');
    } finally {
      this.isSendingEmailVerification = false;
    }
  }

  async handleVerifyNotificationEmail(): Promise<void> {
    if (this.verificationCode.trim().length !== 6) {
      addToast(translate('notifications.preferences.email.code_invalid'), 'error');
      return;
    }

    this.isVerifyingNotificationEmail = true;
    this.notificationsError = null;
    try {
      this.notificationPreference = await verifyNotificationEmail.execute({
        code: this.verificationCode.trim()
      });
      this.notificationDraft = createNotificationPreferenceFormValues(this.notificationPreference);
      this.verificationCode = '';
      addToast(translate('messages.account_notification_email_verified'), 'success');
    } catch (error) {
      this.notificationsError = getErrorMessage(error);
      addToast(this.notificationsError, 'error');
    } finally {
      this.isVerifyingNotificationEmail = false;
    }
  }
}
