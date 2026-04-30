import { goto } from '$app/navigation';
import { getErrorMessage } from '$lib/api/client.js';
import { addToast } from '$lib/components/toast.svelte';
import { t as translate } from '$lib/i18n/index.js';
import {
  userRepository,
  type UserDirectoryPageCapabilities,
  type UserDirectoryTeamFilter,
  type UserDirectoryUser,
  type UserRole
} from '$lib/infrastructure/api/userRepository.js';
import { auth, getAllowedRolesForCreation } from '$lib/stores/auth.svelte.js';
import { confirm } from '$lib/stores/confirm-dialog.js';

export class UserDirectoryPageState {
  users = $state<UserDirectoryUser[]>([]);
  total = $state(0);
  page = $state(1);
  totalPages = $state(1);
  searchText = $state('');
  selectedTeamId = $state('all');
  teamFilters = $state<UserDirectoryTeamFilter[]>([]);
  pageCapabilities = $state<UserDirectoryPageCapabilities>({ can_create_user: false });
  isLoading = $state(true);
  error = $state<string | null>(null);
  createDialogOpen = $state(false);

  async initialize(): Promise<void> {
    if (!auth.canAccessUserDirectory) {
      await goto('/');
      return;
    }
    await this.loadDirectory();
  }

  async loadDirectory(
    nextPage = this.page,
    nextSearch = this.searchText,
    nextTeamId = this.selectedTeamId
  ): Promise<void> {
    this.isLoading = true;
    this.error = null;
    try {
      const result = await userRepository.listDirectory({
        page: nextPage,
        limit: 10,
        search: nextSearch || undefined,
        team_id: nextTeamId === 'all' ? undefined : nextTeamId
      });
      this.users = result.items;
      this.total = result.total;
      this.page = result.page;
      this.totalPages = result.total_pages;
      this.teamFilters = result.teams;
      this.pageCapabilities = result.capabilities;
    } catch (error) {
      this.error = getErrorMessage(error);
    } finally {
      this.isLoading = false;
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

  authVerified(user: UserDirectoryUser): boolean {
    return Boolean(user.is_active && !user.disabled_at);
  }

  twoFactorEnabled(_user: UserDirectoryUser): boolean {
    return false;
  }

  roleOptionsFor(user: UserDirectoryUser) {
    return getAllowedRolesForCreation().filter((roleObj) => roleObj.role !== user.role);
  }

  async handleUserCreated(): Promise<void> {
    this.createDialogOpen = false;
    await this.loadDirectory();
    addToast(translate('messages.user_created_success'), 'success');
  }
}
