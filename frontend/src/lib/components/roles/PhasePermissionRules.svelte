<script lang="ts">
  import type { Phase, PhasePermission } from '$lib/domain/phase/index.js';
  import type { Permission, Role } from '$lib/domain/role/index.js';
  import PhaseRuleRoleSection from './phase-rules/PhaseRuleRoleSection.svelte';
  import { providePhasePermissionRulesState } from './phase-rules/state/context.svelte.js';

  interface Props {
    roles: Role[];
    phases: Phase[];
    permissions: Permission[];
    rules: PhasePermission[];
    canManage: boolean;
    onRulesChange?: () => Promise<void> | void;
  }

  let { roles, phases, permissions, rules, canManage, onRulesChange }: Props = $props();

  const state = providePhasePermissionRulesState({
    roles: function (): Role[] {
      return roles;
    },
    phases: function (): Phase[] {
      return phases;
    },
    permissions: function (): Permission[] {
      return permissions;
    },
    rules: function (): PhasePermission[] {
      return rules;
    },
    canManage: function (): boolean {
      return canManage;
    },
    onRulesChange: function (): Promise<void> | void {
      return onRulesChange?.();
    }
  });
</script>

<div class="space-y-6">
  {#if !state.hasPhases || !state.hasPhaseRulePermissions}
    <div class="rounded-md border bg-muted/20 p-6 text-sm text-muted-foreground">
      {state.emptyMessage}
    </div>
  {:else}
    {#each state.sortedRoles as role (role.id)}
      <PhaseRuleRoleSection {role} />
    {/each}
  {/if}
</div>
