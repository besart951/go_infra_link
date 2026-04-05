<script lang="ts">
  import type { Permission } from '$lib/domain/role/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { cn } from '$lib/utils.js';
  import { useRolePermissionEditorState } from './state/context.svelte.js';

  interface Props {
    permission: Permission;
  }

  let { permission }: Props = $props();

  const state = useRolePermissionEditorState();

  const isSelected = $derived(state.isPermissionSelected(permission.name));

  function handleToggle(): void {
    state.togglePermission(permission.name);
  }
</script>

<label
  class={cn(
    'flex items-center gap-2 rounded px-2 py-1.5 transition-colors',
    'cursor-pointer hover:bg-accent/50'
  )}
>
  <Checkbox checked={isSelected} onCheckedChange={handleToggle} />
  <div class="flex flex-1 items-center gap-2">
    <Badge variant="outline" class="text-xs">
      {permission.action}
    </Badge>
    <span class="text-xs text-muted-foreground">{permission.description}</span>
  </div>
</label>
