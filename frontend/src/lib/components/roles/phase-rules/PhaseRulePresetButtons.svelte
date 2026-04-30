<script lang="ts">
  import type { PhasePermission } from '$lib/domain/phase/index.js';
  import type { UserRole } from '$lib/domain/user/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { t as translate } from '$lib/i18n/index.js';
  import { Ban, Eye, Pencil, Plus, RotateCcw, ShieldCheck } from '@lucide/svelte';
  import { usePhasePermissionRulesState } from './state/context.svelte.js';

  interface Props {
    phaseID: string;
    role: UserRole;
    rule?: PhasePermission;
    disabled: boolean;
  }

  let { phaseID, role, rule = undefined, disabled }: Props = $props();

  const state = usePhasePermissionRulesState();
</script>

<div class="flex flex-wrap gap-2">
  <Button
    size="sm"
    variant="outline"
    {disabled}
    onclick={() => state.applyPreset(phaseID, role, 'read')}
  >
    <Eye class="mr-2 size-4" />
    {translate('roles.phase_rules.presets_read')}
  </Button>
  <Button
    size="sm"
    variant="outline"
    {disabled}
    onclick={() => state.applyPreset(phaseID, role, 'edit')}
  >
    <Pencil class="mr-2 size-4" />
    {translate('roles.phase_rules.presets_edit')}
  </Button>
  <Button
    size="sm"
    variant="outline"
    {disabled}
    onclick={() => state.applyPreset(phaseID, role, 'full')}
  >
    <ShieldCheck class="mr-2 size-4" />
    {translate('roles.phase_rules.presets_full')}
  </Button>
  <Button
    size="sm"
    variant="outline"
    {disabled}
    onclick={() => state.applyPreset(phaseID, role, 'none')}
  >
    <Ban class="mr-2 size-4" />
    {translate('roles.phase_rules.presets_none')}
  </Button>
  {#if rule}
    <Button size="sm" variant="ghost" {disabled} onclick={() => state.removeRule(rule)}>
      <RotateCcw class="mr-2 size-4" />
      {translate('roles.phase_rules.use_default')}
    </Button>
  {:else}
    <Button
      size="sm"
      variant="ghost"
      {disabled}
      onclick={() => state.createEmptyRule(phaseID, role)}
    >
      <Plus class="mr-2 size-4" />
      {translate('roles.phase_rules.add_empty')}
    </Button>
  {/if}
</div>
