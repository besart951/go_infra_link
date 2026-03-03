<script lang="ts">
	import type { FieldDevice } from '$lib/domain/facility/index.js';

	interface Props {
		onSuccess: (createdDevices: FieldDevice[]) => void;
		onCancel: () => void;
	}

	let { onSuccess, onCancel }: Props = $props();
</script>

<button
	data-testid="multi-create-success"
	onclick={() =>
		onSuccess([
			{
				id: 'created-1',
				bmk: 'NEW-1',
				description: 'created',
				text_fix: 'txt',
				apparat_nr: '1',
				sps_controller_system_type_id: 'sps-1',
				system_part_id: 'sp-1',
				apparat_id: 'app-1',
				created_at: '2026-01-01T00:00:00Z',
				updated_at: '2026-01-01T00:00:00Z'
			}
		])}
>
	complete multi-create
</button>
<button data-testid="multi-create-cancel" onclick={onCancel}>cancel</button>
