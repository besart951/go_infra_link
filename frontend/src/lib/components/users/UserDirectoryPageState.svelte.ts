import { goto } from '$app/navigation';
import { getErrorMessage } from '$lib/api/client.js';
import { addToast } from '$lib/components/toast.svelte';
import { t as translate } from '$lib/i18n/index.js';
import {
  userRepository,
  type UserDirectoryPageCapabilities,
  type UserDirectoryRoleFilter,
  type UserDirectoryTeamFilter,
  type UserDirectoryUser,
  type UserRole
} from '$lib/infrastructure/api/userRepository.js';
import { PaginatedSearchState } from '$lib/state/PaginatedSearchState.svelte.js';
import { auth, getAllowedRolesForCreation } from '$lib/stores/auth.svelte.js';
import { confirm } from '$lib/stores/confirm-dialog.js';

export class UserDirectoryPageState {
  readonly query = new PaginatedSearchState({
    pageSize: 10,
    initialTotalPages: 1,
    initialLoading: true
  });

  users = $state<UserDirectoryUser[]>([]);
  showDeletedUsers = $state(false);
  selectedTeamId = $state('all');
  selectedRole = $state<UserRole | 'all'>('all');
  teamFilters = $state<UserDirectoryTeamFilter[]>([]);
  roleFilters = $state<UserDirectoryRoleFilter[]>([]);
  pageCapabilities = $state<UserDirectoryPageCapabilities>({
    can_create_user: false,
    can_read_deleted: false
  });
  createDialogOpen = $state(false);
  resendClock = $state(Date.now());

  get total(): number {
    return this.query.total;
  }

  set total(total: number) {
    this.query.total = total;
  }

  get page(): number {
    return this.query.page;
  }

  set page(page: number) {
    this.query.goToPage(page);
  }

  get totalPages(): number {
    return this.query.totalPages;
  }

  set totalPages(totalPages: number) {
    this.query.totalPages = totalPages;
  }

  get searchText(): string {
    return this.query.searchText;
  }

  set searchText(searchText: string) {
    this.query.setSearchText(searchText);
  }

  get isLoading(): boolean {
    return this.query.loading;
  }

  set isLoading(isLoading: boolean) {
    this.query.setLoading(isLoading);
  }

  get error(): string | null {
    return this.query.error;
  }

  set error(error: string | null) {
    this.query.setError(error);
  }

  async initialize(canAccessDirectory = auth.canAccessUserDirectory): Promise<void> {
    if (!canAccessDirectory) {
      await goto('/');
      return;
    }
    await this.loadDirectory();
  }

  startResendClock(): () => void {
    const interval = window.setInterval(() => {
      this.resendClock = Date.now();
    }, 1000);

    return () => window.clearInterval(interval);
  }

  openCreateDialog(): void {
    this.createDialogOpen = true;
  }

  closeCreateDialog(): void {
    this.createDialogOpen = false;
  }

  async refreshDirectory(): Promise<void> {
    await this.loadDirectory(1, this.query.searchText, this.selectedTeamId, this.selectedRole);
  }

  async goToPage(nextPage: number): Promise<void> {
    await this.loadDirectory(
      nextPage,
      this.query.searchText,
      this.selectedTeamId,
      this.selectedRole
    );
  }

  async setTeamFilter(teamId: string): Promise<void> {
    this.selectedTeamId = teamId;
    await this.loadDirectory(1, this.query.searchText, teamId, this.selectedRole);
  }

  async setRoleFilter(role: UserRole | 'all'): Promise<void> {
    this.selectedRole = role;
    await this.loadDirectory(1, this.query.searchText, this.selectedTeamId, role);
  }

  async setShowDeletedUsers(showDeletedUsers: boolean): Promise<void> {
    this.showDeletedUsers = showDeletedUsers;
    await this.loadDirectory(1, this.query.searchText, this.selectedTeamId, this.selectedRole);
  }

  async loadDirectory(
    nextPage = this.query.page,
    nextSearch = this.query.searchText,
    nextTeamId = this.selectedTeamId,
    nextRole = this.selectedRole
  ): Promise<void> {
    this.query.setLoading(true);
    this.query.clearError();
    try {
      const result = await userRepository.listDirectory({
        page: nextPage,
        limit: this.query.pageSize,
        search: nextSearch || undefined,
        team_id: nextTeamId === 'all' ? undefined : nextTeamId,
        role: nextRole === 'all' ? undefined : nextRole,
        include_deleted: this.pageCapabilities.can_read_deleted && this.showDeletedUsers
      });
      this.users = result.items;
      this.query.applyResult({
        total: result.total,
        page: result.page,
        totalPages: result.total_pages,
        searchText: nextSearch
      });
      this.teamFilters = result.teams ?? [];
      this.roleFilters = result.roles ?? [];
      this.pageCapabilities = result.capabilities;
      if (!this.pageCapabilities.can_read_deleted) {
        this.showDeletedUsers = false;
      }
    } catch (error) {
      this.query.setError(getErrorMessage(error));
    } finally {
      this.query.setLoading(false);
    }
  }

  async handleRoleChange(userId: string, newRole: UserRole): Promise<void> {
    try {
      await userRepository.setRole(userId, newRole);
      await this.loadDirectory();
      addToast(translate('messages.role_updated_success'), 'success');
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('errors.change_role_failed'),
        'error'
      );
    }
  }

  async handleToggleActive(userId: string, isActive: boolean): Promise<void> {
    try {
      if (isActive) {
        await userRepository.disable(userId);
        addToast(translate('messages.user_disabled_success'), 'success');
      } else {
        await userRepository.enable(userId);
        addToast(translate('messages.user_enabled_success'), 'success');
      }
      await this.loadDirectory();
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('errors.toggle_user_status_failed'),
        'error'
      );
    }
  }

  async handleDeleteUser(userId: string, userName: string): Promise<void> {
    const confirmed = await confirm({
      title: translate('common.delete_user'),
      message: translate('messages.delete_user_confirm', { name: userName }),
      confirmText: translate('common.delete'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });

    if (!confirmed) return;

    try {
      await userRepository.delete(userId);
      await this.loadDirectory();
      addToast(translate('messages.user_deleted_success'), 'success');
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('errors.delete_user_failed'),
        'error'
      );
    }
  }

  async handleRestoreUser(userId: string, userName: string): Promise<void> {
    const confirmed = await confirm({
      title: translate('actions.restore_user'),
      message: translate('messages.restore_user_confirm', { name: userName }),
      confirmText: translate('history.restore'),
      cancelText: translate('common.cancel')
    });

    if (!confirmed) return;

    try {
      await userRepository.restore(userId);
      await this.loadDirectory();
      addToast(translate('messages.user_restored_success'), 'success');
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('errors.restore_user_failed'),
        'error'
      );
    }
  }

  async handleResendInvitation(userId: string): Promise<void> {
    try {
      await userRepository.resendRegistration(userId);
      await this.loadDirectory();
      addToast(translate('user.invitation_resent'), 'success');
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('user.invitation_resend_failed'),
        'error'
      );
    }
  }

  formatDate(dateString: string | null | undefined): string {
    if (!dateString) return translate('messages.never');
    const date = new Date(dateString);
    const now = new Date();
    const diffInMs = now.getTime() - date.getTime();
    const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));

    if (diffInDays === 0) return translate('messages.today');
    if (diffInDays === 1) return translate('messages.yesterday');
    if (diffInDays < 7)
      return translate('messages.days_ago').replace('{count}', String(diffInDays));
    if (diffInDays < 30)
      return translate('messages.weeks_ago').replace('{count}', String(Math.floor(diffInDays / 7)));
    if (diffInDays < 365)
      return translate('messages.months_ago').replace(
        '{count}',
        String(Math.floor(diffInDays / 30))
      );
    return translate('messages.years_ago').replace('{count}', String(Math.floor(diffInDays / 365)));
  }

  hasInvitationResendAction(user: UserDirectoryUser): boolean {
    const process = user.registration_process;
    if (!this.pageCapabilities.can_create_user || !process) return false;
    if (process.email_status === 'not_applicable') return false;
    return (
      process.status === 'pending' ||
      process.status === 'email_failed' ||
      process.status === 'expired'
    );
  }

  canResendInvitation(user: UserDirectoryUser, now = Date.now()): boolean {
    const process = user.registration_process;
    if (!process) return false;
    if (process.can_resend) return true;

    const availableAt = timestampMs(process.resend_available_at);
    return availableAt !== null && now >= availableAt;
  }

  invitationResendDisabledReason(user: UserDirectoryUser, now = Date.now()): string | null {
    const process = user.registration_process;
    if (!process || this.canResendInvitation(user, now)) return null;

    const availableAt = timestampMs(process.resend_available_at);
    if (availableAt !== null && availableAt > now) {
      return translate('user.invitation_resend_wait', {
        duration: formatWaitDuration(availableAt - now)
      });
    }

    return translate('user.invitation_resend_unavailable');
  }

  roleOptionsFor(user: UserDirectoryUser) {
    return getAllowedRolesForCreation().filter((roleObj) => roleObj.role !== user.role);
  }

  async handleUserCreated(): Promise<void> {
    this.closeCreateDialog();
    await this.loadDirectory();
    addToast(translate('user.invitation_created_and_sent'), 'success');
  }
}

function timestampMs(value: string | null | undefined): number | null {
  if (!value) return null;
  const timestamp = new Date(value).getTime();
  return Number.isNaN(timestamp) ? null : timestamp;
}

function formatWaitDuration(milliseconds: number): string {
  const totalSeconds = Math.max(1, Math.ceil(milliseconds / 1000));
  if (totalSeconds < 60) {
    return translate(totalSeconds === 1 ? 'user.duration_second' : 'user.duration_seconds', {
      count: totalSeconds
    });
  }

  const totalMinutes = Math.ceil(totalSeconds / 60);
  return translate(totalMinutes === 1 ? 'user.duration_minute' : 'user.duration_minutes', {
    count: totalMinutes
  });
}
