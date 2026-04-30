import { GetSMTPSettingsUseCase } from '$lib/application/useCases/notification/getSMTPSettingsUseCase.js';
import { SaveSMTPSettingsUseCase } from '$lib/application/useCases/notification/saveSMTPSettingsUseCase.js';
import { SendSMTPTestEmailUseCase } from '$lib/application/useCases/notification/sendSMTPTestEmailUseCase.js';
import { getErrorMessage, getFieldErrors } from '$lib/api/client.js';
import { addToast } from '$lib/components/toast.svelte';
import type {
  SendSMTPTestEmailRequest,
  SMTPSettings,
  UpsertSMTPSettingsRequest
} from '$lib/domain/notification/index.js';
import { t as translate } from '$lib/i18n/index.js';
import { notificationSettingsRepository } from '$lib/infrastructure/api/notificationSettingsRepository.js';

const getSMTPSettings = new GetSMTPSettingsUseCase(notificationSettingsRepository);
const saveSMTPSettings = new SaveSMTPSettingsUseCase(notificationSettingsRepository);
const sendSMTPTestEmail = new SendSMTPTestEmailUseCase(notificationSettingsRepository);

export class SMTPManagementPageState {
  settings = $state<SMTPSettings | null>(null);
  isLoading = $state(true);
  isRefreshing = $state(false);
  loadError = $state<string | null>(null);
  lastLoadedAt = $state<string | null>(null);
  isSaving = $state(false);
  saveError = $state<string | null>(null);
  saveFieldErrors = $state<Record<string, string>>({});
  isSendingTest = $state(false);
  testError = $state<string | null>(null);
  testFieldErrors = $state<Record<string, string>>({});

  readonly serviceStatus = $derived.by(() => {
    if (!this.settings) return translate('notifications.status.not_configured');
    return translate(
      this.settings.enabled ? 'notifications.status.enabled' : 'notifications.status.disabled'
    );
  });

  readonly connectionLabel = $derived(
    this.settings
      ? `${this.settings.host}:${this.settings.port}`
      : translate('notifications.status.not_configured')
  );

  statusVariant(): 'secondary' | 'outline' | 'success' | 'warning' {
    if (this.isLoading) return 'secondary';
    if (!this.settings) return 'outline';
    return this.settings.enabled ? 'success' : 'warning';
  }

  async loadSettings(mode: 'initial' | 'refresh' = 'initial'): Promise<void> {
    if (mode === 'initial') {
      this.isLoading = true;
    } else {
      this.isRefreshing = true;
    }

    this.loadError = null;

    try {
      this.settings = await getSMTPSettings.execute();
      this.lastLoadedAt = new Date().toISOString();

      if (mode === 'refresh') {
        addToast(translate('notifications.toasts.refreshed'), 'success');
      }
    } catch (error) {
      this.loadError = getErrorMessage(error);

      if (mode === 'refresh') {
        addToast(this.loadError, 'error');
      }
    } finally {
      this.isLoading = false;
      this.isRefreshing = false;
    }
  }

  async handleSave(payload: UpsertSMTPSettingsRequest): Promise<void> {
    this.isSaving = true;
    this.saveError = null;
    this.saveFieldErrors = {};

    try {
      this.settings = await saveSMTPSettings.execute(payload);
      this.lastLoadedAt = new Date().toISOString();
      this.saveError = null;
      this.saveFieldErrors = {};
      addToast(translate('notifications.toasts.saved'), 'success');
    } catch (error) {
      this.saveFieldErrors = getFieldErrors(error);
      this.saveError =
        Object.keys(this.saveFieldErrors).length === 0 ? getErrorMessage(error) : null;

      if (this.saveError) {
        addToast(this.saveError, 'error');
      }
    } finally {
      this.isSaving = false;
    }
  }

  async handleSendTest(payload: SendSMTPTestEmailRequest): Promise<void> {
    this.isSendingTest = true;
    this.testError = null;
    this.testFieldErrors = {};

    try {
      await sendSMTPTestEmail.execute(payload);
      this.testError = null;
      this.testFieldErrors = {};
      addToast(translate('notifications.toasts.test_sent'), 'success');
    } catch (error) {
      this.testFieldErrors = getFieldErrors(error);
      this.testError =
        Object.keys(this.testFieldErrors).length === 0 ? getErrorMessage(error) : null;

      if (this.testError) {
        addToast(this.testError, 'error');
      }
    } finally {
      this.isSendingTest = false;
    }
  }

  formatDateTime(value: string | null): string {
    if (!value) return translate('common.not_available');
    return new Intl.DateTimeFormat('de-CH', {
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(new Date(value));
  }
}
