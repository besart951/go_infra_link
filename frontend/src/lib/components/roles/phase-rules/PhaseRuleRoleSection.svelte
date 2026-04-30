<script lang="ts">
  import type { Role } from '$lib/domain/role/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import RoleBadge from '$lib/components/role-badge.svelte';
  import { t as translate } from '$lib/i18n/index.js';
  import PhaseRulePermissionGrid from './PhaseRulePermissionGrid.svelte';
  import PhaseRulePresetButtons from './PhaseRulePresetButtons.svelte';
  import { usePhasePermissionRulesState } from './state/context.svelte.js';

  interface Props {
    role: Role;
  }

  let { role }: Props = $props();

  const state = usePhasePermissionRulesState();
</script>

<section class="overflow-hidden rounded-lg border bg-background">
  <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
    <div class="flex items-center gap-3">
      <RoleBadge role={role.name} showIcon={false} />
      <div class="text-sm text-muted-foreground">
        {state.getRoleSummary(role.name)}
      </div>
    </div>
    <Badge variant="outline">{translate('roles.phase_rules.restrictive_badge')}</Badge>
  </div>

  <div class="overflow-x-auto">
    <Table.Root>
      <Table.Header>
        <Table.Row>
          <Table.Head class="min-w-36">{translate('roles.phase_rules.phase')}</Table.Head>
          <Table.Head class="min-w-90">{translate('roles.phase_rules.presets')}</Table.Head>
          <Table.Head>{translate('roles.phase_rules.permissions')}</Table.Head>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#each state.phases as phase (phase.id)}
          {@const rule = state.getRule(phase.id, role.name)}
          {@const disabled = !state.canManage || state.isSaving(phase.id, role.name)}
          <Table.Row>
            <Table.Cell class="align-top">
              <div class="font-medium">{phase.name}</div>
              <div class="mt-1">
                {#if rule}
                  <Badge variant="secondary">{state.getRuleBadgeLabel(rule)}</Badge>
                {:else}
                  <Badge variant="outline">{state.getRuleBadgeLabel(rule)}</Badge>
                {/if}
              </div>
            </Table.Cell>
            <Table.Cell class="align-top">
              <PhaseRulePresetButtons phaseID={phase.id} role={role.name} {rule} {disabled} />
            </Table.Cell>
            <Table.Cell class="align-top">
              {#if rule}
                <PhaseRulePermissionGrid phaseID={phase.id} role={role.name} {rule} {disabled} />
              {:else}
                <div class="text-sm text-muted-foreground">
                  {translate('roles.phase_rules.default_hint')}
                </div>
              {/if}
            </Table.Cell>
          </Table.Row>
        {/each}
      </Table.Body>
    </Table.Root>
  </div>
</section>
