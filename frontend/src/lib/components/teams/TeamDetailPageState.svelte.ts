import { addToast } from '$lib/components/toast.svelte';
import { t as translate } from '$lib/i18n/index.js';
import {
  teamRepository,
  type Team,
  type TeamMember
} from '$lib/infrastructure/api/teamRepository.js';
import { userRepository, type User } from '$lib/infrastructure/api/userRepository.js';
import { confirm } from '$lib/stores/confirm-dialog.js';

export class TeamDetailPageState {
  team = $state<Team | null>(null);
  members = $state<TeamMember[]>([]);
  users = $state<User[]>([]);
  loading = $state(true);
  error = $state<string | null>(null);
  busy = $state(false);
  addMemberOpen = $state(false);
  addMemberSearch = $state('');
  addMemberResults = $state<User[]>([]);
  addMemberLoading = $state(false);

  private debounceTimer: ReturnType<typeof setTimeout> | undefined;

  constructor(private readonly resolveTeamId: () => string) {}

  get teamId(): string {
    return this.resolveTeamId();
  }

  userById(id: string): User | undefined {
    return this.users.find((user) => user.id === id);
  }

  memberUserIds(): Set<string> {
    return new Set(this.members.map((member) => member.user_id));
  }

  async load(): Promise<void> {
    if (!this.teamId) {
      this.error = translate('teams.errors.missing_id');
      this.loading = false;
      return;
    }

    this.loading = true;
    this.error = null;
    try {
      const [team, members, users] = await Promise.all([
        teamRepository.get(this.teamId),
        teamRepository.listMembers(this.teamId, { page: 1, limit: 100 }),
        userRepository.list({ page: 1, limit: 100, search: '' })
      ]);
      this.team = team;
      this.members = members.items;
      this.users = users.items;
    } catch (error) {
      this.error = error instanceof Error ? error.message : translate('teams.errors.load_failed');
    } finally {
      this.loading = false;
    }
  }

  async searchUsers(query: string): Promise<void> {
    this.addMemberLoading = true;
    try {
      const res = await userRepository.list({ page: 1, limit: 20, search: query });
      const existingIds = this.memberUserIds();
      this.addMemberResults = res.items.filter((user) => !existingIds.has(user.id));
    } catch {
      this.addMemberResults = [];
    } finally {
      this.addMemberLoading = false;
    }
  }

  scheduleUserSearch(): void {
    const query = this.addMemberSearch;
    clearTimeout(this.debounceTimer);
    this.debounceTimer = setTimeout(() => {
      void this.searchUsers(query);
    }, 300);
  }

  async handleAddMember(userId: string): Promise<void> {
    if (!this.teamId) return;
    this.busy = true;
    try {
      await teamRepository.addMember(this.teamId, { user_id: userId, role: 'member' });
      addToast(translate('teams.toasts.member_added'), 'success');
      this.addMemberOpen = false;
      this.addMemberSearch = '';
      this.addMemberResults = [];
      await this.load();
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('teams.toasts.member_add_failed'),
        'error'
      );
    } finally {
      this.busy = false;
    }
  }

  async changeRole(userId: string, role: 'member' | 'manager' | 'owner'): Promise<void> {
    if (!this.teamId) return;
    this.busy = true;
    try {
      await teamRepository.addMember(this.teamId, { user_id: userId, role });
      addToast(translate('teams.toasts.role_updated'), 'success');
      await this.load();
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('teams.toasts.role_update_failed'),
        'error'
      );
    } finally {
      this.busy = false;
    }
  }

  async remove(userId: string): Promise<void> {
    if (!this.teamId) return;
    const user = this.userById(userId);
    const ok = await confirm({
      title: translate('teams.confirm.remove_title'),
      message: translate('teams.confirm.remove_message', {
        name: user
          ? `${user.first_name} ${user.last_name}`
          : translate('teams.confirm.user_fallback')
      }),
      confirmText: translate('teams.confirm.remove_confirm'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });
    if (!ok) return;

    this.busy = true;
    try {
      await teamRepository.removeMember(this.teamId, userId);
      addToast(translate('teams.toasts.member_removed'), 'success');
      await this.load();
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('teams.toasts.member_remove_failed'),
        'error'
      );
    } finally {
      this.busy = false;
    }
  }
}
