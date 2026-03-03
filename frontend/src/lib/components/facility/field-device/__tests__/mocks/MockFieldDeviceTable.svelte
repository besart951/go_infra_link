<script lang="ts">
	import type { FieldDevice } from '$lib/domain/facility/index.js';

	interface Props {
		items: FieldDevice[];
		onDelete: (device: FieldDevice) => void;
		onToggleSelect: (id: string) => void;
	}

	let { items, onDelete, onToggleSelect }: Props = $props();
</script>

<div data-testid="table-count">{items.length}</div>
{#each items as item (item.id)}
	<button data-testid={`select-${item.id}`} onclick={() => onToggleSelect(item.id)}>
		select
	</button>
	<button data-testid={`delete-${item.id}`} onclick={() => onDelete(item)}>delete</button>
{/each}
