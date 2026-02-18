<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import type { StateText } from '$lib/domain/facility/index.js';
	import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';
	import { stateTextRepository } from '$lib/infrastructure/api/stateTextRepository.js';
	const manageStateText = new ManageEntityUseCase(stateTextRepository);
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface StateTextFormProps {
		initialData?: StateText;
		onSuccess?: (stateText: StateText) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: StateTextFormProps = $props();

	const t = createTranslator();

	let ref_number = $state(0);
	let state_text1 = $state('');
	let state_text2 = $state('');
	let state_text3 = $state('');
	let state_text4 = $state('');
	let state_text5 = $state('');
	let state_text6 = $state('');
	let state_text7 = $state('');
	let state_text8 = $state('');
	let state_text9 = $state('');
	let state_text10 = $state('');
	let state_text11 = $state('');
	let state_text12 = $state('');
	let state_text13 = $state('');
	let state_text14 = $state('');
	let state_text15 = $state('');
	let state_text16 = $state('');

	$effect(() => {
		if (initialData) {
			ref_number = initialData.ref_number;
			state_text1 = initialData.state_text1 ?? '';
			state_text2 = initialData.state_text2 ?? '';
			state_text3 = initialData.state_text3 ?? '';
			state_text4 = initialData.state_text4 ?? '';
			state_text5 = initialData.state_text5 ?? '';
			state_text6 = initialData.state_text6 ?? '';
			state_text7 = initialData.state_text7 ?? '';
			state_text8 = initialData.state_text8 ?? '';
			state_text9 = initialData.state_text9 ?? '';
			state_text10 = initialData.state_text10 ?? '';
			state_text11 = initialData.state_text11 ?? '';
			state_text12 = initialData.state_text12 ?? '';
			state_text13 = initialData.state_text13 ?? '';
			state_text14 = initialData.state_text14 ?? '';
			state_text15 = initialData.state_text15 ?? '';
			state_text16 = initialData.state_text16 ?? '';
		}
	});

	const formState = useFormState({
		onSuccess: (result: StateText) => {
			onSuccess?.(result);
		}
	});

	async function handleSubmit() {
		await formState.handleSubmit(async () => {
			if (initialData) {
				return await manageStateText.update(initialData.id, {
					ref_number,
					state_text1: state_text1 || undefined,
					state_text2: state_text2 || undefined,
					state_text3: state_text3 || undefined,
					state_text4: state_text4 || undefined,
					state_text5: state_text5 || undefined,
					state_text6: state_text6 || undefined,
					state_text7: state_text7 || undefined,
					state_text8: state_text8 || undefined,
					state_text9: state_text9 || undefined,
					state_text10: state_text10 || undefined,
					state_text11: state_text11 || undefined,
					state_text12: state_text12 || undefined,
					state_text13: state_text13 || undefined,
					state_text14: state_text14 || undefined,
					state_text15: state_text15 || undefined,
					state_text16: state_text16 || undefined
				});
			} else {
				return await manageStateText.create({
					ref_number,
					state_text1: state_text1 || undefined,
					state_text2: state_text2 || undefined,
					state_text3: state_text3 || undefined,
					state_text4: state_text4 || undefined,
					state_text5: state_text5 || undefined,
					state_text6: state_text6 || undefined,
					state_text7: state_text7 || undefined,
					state_text8: state_text8 || undefined,
					state_text9: state_text9 || undefined,
					state_text10: state_text10 || undefined,
					state_text11: state_text11 || undefined,
					state_text12: state_text12 || undefined,
					state_text13: state_text13 || undefined,
					state_text14: state_text14 || undefined,
					state_text15: state_text15 || undefined,
					state_text16: state_text16 || undefined
				});
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
			{initialData
				? $t('facility.forms.state_text.title_edit')
				: $t('facility.forms.state_text.title_new')}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="state_ref">{$t('facility.forms.state_text.ref_label')}</Label>
			<Input id="state_ref" type="number" bind:value={ref_number} required />
			{#if formState.getFieldError('ref_number', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('ref_number', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text1">{$t('facility.forms.state_text.state_text', { index: 1 })}</Label>
			<Input id="state_text1" bind:value={state_text1} />
			{#if formState.getFieldError('state_text1', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text1', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text2">{$t('facility.forms.state_text.state_text', { index: 2 })}</Label>
			<Input id="state_text2" bind:value={state_text2} />
			{#if formState.getFieldError('state_text2', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text2', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text3">{$t('facility.forms.state_text.state_text', { index: 3 })}</Label>
			<Input id="state_text3" bind:value={state_text3} />
			{#if formState.getFieldError('state_text3', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text3', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text4">{$t('facility.forms.state_text.state_text', { index: 4 })}</Label>
			<Input id="state_text4" bind:value={state_text4} />
			{#if formState.getFieldError('state_text4', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text4', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text5">{$t('facility.forms.state_text.state_text', { index: 5 })}</Label>
			<Input id="state_text5" bind:value={state_text5} />
			{#if formState.getFieldError('state_text5', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text5', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text6">{$t('facility.forms.state_text.state_text', { index: 6 })}</Label>
			<Input id="state_text6" bind:value={state_text6} />
			{#if formState.getFieldError('state_text6', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text6', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text7">{$t('facility.forms.state_text.state_text', { index: 7 })}</Label>
			<Input id="state_text7" bind:value={state_text7} />
			{#if formState.getFieldError('state_text7', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text7', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text8">{$t('facility.forms.state_text.state_text', { index: 8 })}</Label>
			<Input id="state_text8" bind:value={state_text8} />
			{#if formState.getFieldError('state_text8', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text8', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text9">{$t('facility.forms.state_text.state_text', { index: 9 })}</Label>
			<Input id="state_text9" bind:value={state_text9} />
			{#if formState.getFieldError('state_text9', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text9', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text10">{$t('facility.forms.state_text.state_text', { index: 10 })}</Label>
			<Input id="state_text10" bind:value={state_text10} />
			{#if formState.getFieldError('state_text10', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text10', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text11">{$t('facility.forms.state_text.state_text', { index: 11 })}</Label>
			<Input id="state_text11" bind:value={state_text11} />
			{#if formState.getFieldError('state_text11', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text11', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text12">{$t('facility.forms.state_text.state_text', { index: 12 })}</Label>
			<Input id="state_text12" bind:value={state_text12} />
			{#if formState.getFieldError('state_text12', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text12', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text13">{$t('facility.forms.state_text.state_text', { index: 13 })}</Label>
			<Input id="state_text13" bind:value={state_text13} />
			{#if formState.getFieldError('state_text13', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text13', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text14">{$t('facility.forms.state_text.state_text', { index: 14 })}</Label>
			<Input id="state_text14" bind:value={state_text14} />
			{#if formState.getFieldError('state_text14', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text14', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text15">{$t('facility.forms.state_text.state_text', { index: 15 })}</Label>
			<Input id="state_text15" bind:value={state_text15} />
			{#if formState.getFieldError('state_text15', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text15', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text16">{$t('facility.forms.state_text.state_text', { index: 16 })}</Label>
			<Input id="state_text16" bind:value={state_text16} />
			{#if formState.getFieldError('state_text16', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text16', ['statetext'])}</p>
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
