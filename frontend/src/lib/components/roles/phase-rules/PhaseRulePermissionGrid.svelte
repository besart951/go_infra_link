<script lang="ts">
  import type { PhasePermission } from '$lib/domain/phase/index.js';
  import type { UserRole } from '$lib/domain/user/index.js';
  import PhaseRulePermissionResource from './PhaseRulePermissionResource.svelte';
  import { usePhasePermissionRulesState } from './state/context.svelte.js';

  interface Props {
    phaseID: string;
    role: UserRole;
    rule: PhasePermission;
    disabled: boolean;
  }

  let { phaseID, role, rule, disabled }: Props = $props();

  const state = usePhasePermissionRulesState();
</script>

<div class="grid min-w-150 gap-3 xl:grid-cols-2">
  {#each state.resourceNames as resource (resource)}
    <PhaseRulePermissionResource {phaseID} {role} {rule} {resource} {disabled} />
  {/each}
</div>
