<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import type { Phase } from '$lib/domain/phase/index.js';
	import { createPhase, updatePhase } from '$lib/infrastructure/api/phase.adapter.js';
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';
	import { createTranslator } from '$lib/i18n/translator.js';
	import { t as translate } from '$lib/i18n/index.js';

	interface PhaseFormProps {
		initialData?: Phase;
		onSuccess?: (phase: Phase) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: PhaseFormProps = $props();

	const t = createTranslator();

	let name = $state('');

	// Watch for changes to initialData
	$effect(() => {
		if (initialData) {
			name = initialData.name ?? '';
		}
	});

	const formState = useFormState({
		onSuccess: (result: Phase) => {
			onSuccess?.(result);
		},
		showSuccessToast: true,
		// svelte-ignore state_referenced_locally
		successMessage: initialData
			? translate('phases.toasts.updated')
			: translate('phases.toasts.created')
	});

	async function handleSubmit() {
		if (!name.trim()) {
			formState.resetErrors();
			// We can't set error directly on formState, so we'll handle validation inline
			return;
		}

		await formState.handleSubmit(async () => {
			if (initialData) {
				return await updatePhase(initialData.id, { name: name.trim() });
			} else {
				return await createPhase({ name: name.trim() });
			}
		});
	}
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		handleSubmit();
	}}
	class="space-y-4 rounded-md border bg-muted/20 p-4"
>
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">
			{initialData ? $t('phases.form.title_edit') : $t('phases.form.title_new')}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="phase_name">{$t('phases.form.name_label')}</Label>
			<Input
				id="phase_name"
				bind:value={name}
				required
				placeholder={$t('phases.form.name_placeholder')}
			/>
			{#if formState.getFieldError('name')}
				<p class="text-sm text-red-500">{formState.getFieldError('name')}</p>
			{/if}
		</div>
	</div>

	{#if formState.error}
		<p class="text-sm text-red-500">{formState.error}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button type="button" variant="ghost" onclick={onCancel}>{$t('common.cancel')}</Button>
		<Button type="submit" disabled={formState.loading}>
			{initialData ? $t('common.update') : $t('common.create')}
		</Button>
	</div>
</form>
