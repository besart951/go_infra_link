<script lang="ts">
	import type { Permission } from '$lib/domain/role/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { cn } from '$lib/utils.js';

	interface Props {
		permission: Permission;
		isSelected: boolean;
		disabled?: boolean;
		onToggle: () => void;
	}

	let { permission, isSelected, disabled = false, onToggle }: Props = $props();
</script>

<label
	class={cn(
		'flex items-center gap-2 rounded px-2 py-1.5 transition-colors',
		disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer hover:bg-accent/50'
	)}
>
	<Checkbox checked={isSelected} {disabled} onCheckedChange={onToggle} />
	<div class="flex flex-1 items-center gap-2">
		<Badge variant="outline" class="text-xs">
			{permission.action}
		</Badge>
		<span class="text-xs text-muted-foreground">{permission.description}</span>
	</div>
</label>
