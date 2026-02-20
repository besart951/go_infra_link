<script lang="ts">
	import { Label } from '$lib/components/ui/label/index.js';
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import FieldDevicePreselection from '../FieldDevicePreselection.svelte';
	import ObjectDataBacnetPreview from './ObjectDataBacnetPreview.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	import type { ObjectData, SPSControllerSystemType } from '$lib/domain/facility/index.js';
	import type { MultiCreateSelection } from '$lib/domain/facility/fieldDeviceMultiCreate.js';
	import type { FieldDevicePreselection as PreselectionType } from '$lib/domain/facility/preselectionFilter.js';

	type Props = {
		projectId?: string;
		selection: MultiCreateSelection;
		preselectionValue: PreselectionType;
		submitting: boolean;
		availableNumbersCount: number;
		loadingAvailableNumbers: boolean;
		showStatus: boolean;
		selectedObjectData: ObjectData | null;
		loadingObjectDataPreview: boolean;
		objectDataPreviewError?: string;
		onSpsSystemTypeChange: (value: string) => void;
		onPreselectionChange: (value: PreselectionType) => void;
		fetchSpsControllerSystemTypes: (search: string) => Promise<SPSControllerSystemType[]>;
		fetchSpsControllerSystemTypeById: (id: string) => Promise<SPSControllerSystemType | null>;
		formatSpsControllerSystemTypeLabel: (item: SPSControllerSystemType) => string;
	};

	let {
		projectId,
		selection,
		preselectionValue,
		submitting,
		availableNumbersCount,
		loadingAvailableNumbers,
		showStatus,
		selectedObjectData,
		loadingObjectDataPreview,
		objectDataPreviewError,
		onSpsSystemTypeChange,
		onPreselectionChange,
		fetchSpsControllerSystemTypes,
		fetchSpsControllerSystemTypeById,
		formatSpsControllerSystemTypeLabel
	}: Props = $props();

	const t = createTranslator();
</script>

<Card.Root class="p-6">
	<h3 class="mb-4 text-lg font-semibold">{$t('field_device.multi_create.selection.title')}</h3>
	<p class="mb-4 text-sm text-muted-foreground">
		{$t('field_device.multi_create.selection.description')}
	</p>

	<div class="grid gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="sps-system-type"
				>{$t('field_device.multi_create.selection.sps_system_type')}</Label
			>
			<AsyncCombobox
				id="sps-system-type"
				placeholder={$t('field_device.multi_create.selection.sps_system_type_placeholder')}
				searchPlaceholder={$t('field_device.multi_create.selection.sps_system_type_search')}
				emptyText={$t('field_device.multi_create.selection.sps_system_type_empty')}
				fetcher={fetchSpsControllerSystemTypes}
				fetchById={fetchSpsControllerSystemTypeById}
				labelKey="system_type_name"
				labelFormatter={formatSpsControllerSystemTypeLabel}
				width="w-full"
				value={selection.spsControllerSystemTypeId}
				onValueChange={onSpsSystemTypeChange}
				clearable
				clearText={$t('field_device.multi_create.selection.sps_system_type_clear')}
				disabled={submitting}
			/>
		</div>
	</div>

	<div class="mt-4">
		<FieldDevicePreselection
			value={preselectionValue}
			onChange={onPreselectionChange}
			{projectId}
			disabled={!selection.spsControllerSystemTypeId || submitting}
			className="grid grid-cols-1 gap-4 md:grid-cols-3"
		/>
	</div>

	<ObjectDataBacnetPreview
		objectData={selectedObjectData}
		loading={loadingObjectDataPreview}
		error={objectDataPreviewError}
	/>

	{#if showStatus}
		<div class="mt-4">
			<Alert.Root>
				<Alert.Description>
					<div class="text-sm">
						<p class="font-medium">{$t('field_device.multi_create.status.title')}</p>
						<ul class="mt-2 space-y-1 text-muted-foreground">
							<li>
								• {$t('field_device.multi_create.status.available', {
									count: availableNumbersCount
								})}
								{#if loadingAvailableNumbers}
									{$t('field_device.multi_create.status.loading_suffix')}
								{/if}
							</li>
							{#if availableNumbersCount === 0 && !loadingAvailableNumbers}
								<li class="text-destructive">
									• {$t('field_device.multi_create.status.none_available')}
								</li>
							{/if}
						</ul>
					</div>
				</Alert.Description>
			</Alert.Root>
		</div>
	{/if}
</Card.Root>
