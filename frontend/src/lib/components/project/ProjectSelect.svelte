<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { getProject, listProjects } from '$lib/infrastructure/api/project.adapter.js';
	import type { Project } from '$lib/domain/project/index.js';

	type Props = {
		value?: string;
		width?: string;
	};

	let { value = $bindable(''), width = 'w-[250px]' }: Props = $props();

	async function fetcher(search: string): Promise<Project[]> {
		const res = await listProjects({ search, limit: 20 });
		return res.items || [];
	}

	async function fetchById(id: string): Promise<Project> {
		return getProject(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="name"
	placeholder="Select Project..."
	{width}
/>
