<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import Toasts, { addToast } from '$lib/components/toast.svelte';
	import ProjectPhaseSelect from '$lib/components/project/ProjectPhaseSelect.svelte';
	import { createProject } from '$lib/infrastructure/api/project.adapter.js';
	import type { ProjectStatus, CreateProjectRequest } from '$lib/domain/project/index.js';
	import { ArrowLeft } from '@lucide/svelte';

	type CreateProjectForm = {
		name: string;
		description: string;
		status: ProjectStatus;
		start_date: string;
		phase_id: string;
	};

	function todayInputValue(): string {
		const now = new Date();
		const yyyy = now.getFullYear();
		const mm = String(now.getMonth() + 1).padStart(2, '0');
		const dd = String(now.getDate()).padStart(2, '0');
		return `${yyyy}-${mm}-${dd}`;
	}

	const statusOptions: Array<{ value: ProjectStatus; label: string }> = [
		{ value: 'planned', label: 'Planned' },
		{ value: 'ongoing', label: 'Ongoing' },
		{ value: 'completed', label: 'Completed' }
	];

	let busy = $state(false);
	let form = $state<CreateProjectForm>({
		name: '',
		description: '',
		status: 'planned',
		start_date: todayInputValue(),
		phase_id: ''
	});

	function canSubmit(): boolean {
		return form.name.trim().length > 0 && !busy;
	}

	async function submit() {
		if (!canSubmit()) return;
		busy = true;
		try {
			const payload: CreateProjectRequest = {
				name: form.name.trim(),
				description: form.description.trim() || undefined,
				status: form.status,
				start_date: form.start_date || undefined,
				phase_id: form.phase_id || undefined
			};

			const project = await createProject(payload);
			addToast('Project created', 'success');
			goto(`/projects/${project.id}`);
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to create project', 'error');
		} finally {
			busy = false;
		}
	}
</script>

<Toasts />

<div class="flex flex-col gap-6">
	<div class="flex items-start gap-3">
		<Button variant="outline" onclick={() => goto('/projects')}
			><ArrowLeft class="mr-2 h-4 w-4" />Back</Button
		>
		<div>
			<h1 class="text-3xl font-bold tracking-tight">Create Project</h1>
			<p class="mt-1 text-muted-foreground">Start a new infrastructure project.</p>
		</div>
	</div>

	<div class="rounded-lg border bg-background p-6">
		<div class="grid gap-4 md:grid-cols-2">
			<div class="flex flex-col gap-2">
				<label class="text-sm font-medium" for="project_name">Name</label>
				<Input
					id="project_name"
					placeholder="Project name"
					bind:value={form.name}
					disabled={busy}
				/>
			</div>

			<div class="flex flex-col gap-2">
				<label class="text-sm font-medium" for="project_status">Status</label>
				<select
					id="project_status"
					class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
					bind:value={form.status}
					disabled={busy}
				>
					{#each statusOptions as opt}
						<option value={opt.value}>{opt.label}</option>
					{/each}
				</select>
			</div>

			<div class="flex flex-col gap-2">
				<label class="text-sm font-medium" for="project_start">Start date</label>
				<Input
					id="project_start"
					type="date"
					bind:value={form.start_date}
					disabled={busy}
				/>
			</div>

			<div class="flex flex-col gap-2">
				<label class="text-sm font-medium">Phase</label>
				<ProjectPhaseSelect bind:value={form.phase_id} width="w-full" />
				<p class="text-xs text-muted-foreground">Select an existing phase ID.</p>
			</div>

			<div class="flex flex-col gap-2 md:col-span-2">
				<label class="text-sm font-medium" for="project_desc">Description</label>
				<Textarea
					id="project_desc"
					placeholder="Describe the project goals"
					rows={4}
					bind:value={form.description}
					disabled={busy}
				/>
			</div>
		</div>

		<div class="mt-6 flex items-center justify-end gap-2">
			<Button variant="outline" onclick={() => goto('/projects')} disabled={busy}
				>Cancel</Button
			>
			<Button onclick={submit} disabled={!canSubmit()}>Create project</Button>
		</div>
	</div>
</div>
