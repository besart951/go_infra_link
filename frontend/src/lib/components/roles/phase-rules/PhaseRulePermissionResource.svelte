<script lang="ts">
  import type { PhasePermission } from '$lib/domain/phase/index.js';
  import type { UserRole } from '$lib/domain/user/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { usePhasePermissionRulesState } from './state/context.svelte.js';

  interface Props {
    phaseID: string;
    role: UserRole;
    rule: PhasePermission;
    resource: string;
    disabled: boolean;
  }

  let { phaseID, role, rule, resource, disabled }: Props = $props();

  const state = usePhasePermissionRulesState();
  const permissions = $derived(state.getPermissionsForResource(resource));
</script>

<div class="rounded-md border p-3">
  <div class="mb-2 text-sm font-medium">{state.getResourceLabel(resource)}</div>
  <div class="flex flex-wrap gap-3">
    {#each permissions as permission (permission.id)}
      <label class="flex items-center gap-2 text-sm">
        <Checkbox
          checked={state.isPermissionSelected(rule, permission.name)}
          {disabled}
          onCheckedChange={() => state.togglePermission(phaseID, role, permission.name)}
        />
        <span>{state.getActionLabel(permission.action)}</span>
      </label>
    {/each}
  </div>
</div>
