import { goto } from '$app/navigation';
import { addToast } from '$lib/components/toast.svelte';
import type { Team } from '$lib/domain/entities/team.js';
import { t as translate } from '$lib/i18n/index.js';
import { teamRepository } from '$lib/infrastructure/api/teamRepository.js';
import { teamsStore } from '$lib/stores/list/entityStores.js';
import { confirm } from '$lib/stores/confirm-dialog.js';

type CreateTeamForm = {
  name: string;
  description: string;
};

export class TeamListPageState {
  createOpen = $state(false);
  createBusy = $state(false);
  form = $state<CreateTeamForm>({ name: '', description: '' });
  memberCounts = $state<Map<string, number>>(new Map());

  initialize(): void {
    teamsStore.load();
  }

  canSubmitCreate(): boolean {
    return this.form.name.trim().length > 0 && !this.createBusy;
  }

  async submitCreate(): Promise<void> {
    if (!this.canSubmitCreate()) return;
    this.createBusy = true;
    try {
      const team = await teamRepository.create({
        name: this.form.name.trim(),
        description: this.form.description.trim() ? this.form.description.trim() : null
      });
      addToast(translate('messages.team_created'), 'success');
      this.form = { name: '', description: '' };
      this.createOpen = false;
      teamsStore.reload();
      await goto(`/teams/${team.id}`);
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('errors.create_team_failed'),
        'error'
      );
    } finally {
      this.createBusy = false;
    }
  }

  async handleDeleteTeam(team: Team): Promise<void> {
    const confirmed = await confirm({
      title: translate('common.delete_team'),
      message: translate('messages.delete_team_confirm', { name: team.name }),
      confirmText: translate('common.delete'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });

    if (!confirmed) return;

    try {
      await teamRepository.delete(team.id);
      teamsStore.reload();
      addToast(translate('messages.team_deleted_success'), 'success');
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('errors.delete_team_failed'),
        'error'
      );
    }
  }

  async loadMemberCounts(teams: Team[]): Promise<void> {
    const counts = new Map<string, number>();
    await Promise.all(
      teams.map(async (team) => {
        try {
          const res = await teamRepository.listMembers(team.id, { page: 1, limit: 1 });
          counts.set(team.id, res.total);
        } catch {
          counts.set(team.id, 0);
        }
      })
    );
    this.memberCounts = counts;
  }
}
