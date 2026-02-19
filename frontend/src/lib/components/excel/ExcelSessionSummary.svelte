<script lang="ts">
	import { CircleCheck } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import type { ExcelReadSession } from '$lib/domain/excel/index.js';

	interface Props {
		session: ExcelReadSession;
	}

	let { session }: Props = $props();

	let duplicateDescriptionIds = $state<Set<string>>(new Set());
	let duplicateCheckDone = $state(false);

	function normalizeDescription(value: string): string {
		return value.trim().toLowerCase();
	}

	function rowIdentifier(objectDataId: string, bacnetObjectId: string): string {
		return `${objectDataId}::${bacnetObjectId}`;
	}

	function runDuplicateDescriptionCheck(): void {
		const descriptionCounts = new Map<string, number>();
		const rows: Array<{ objectDataId: string; bacnetObjectId: string; description: string }> = [];

		for (const objectData of session.objectDataExcel) {
			for (const bacnetObject of objectData.bacnet_objects) {
				const normalized = normalizeDescription(bacnetObject.description || '');
				if (normalized.length === 0) continue;

				descriptionCounts.set(normalized, (descriptionCounts.get(normalized) ?? 0) + 1);
				rows.push({
					objectDataId: objectData.id,
					bacnetObjectId: bacnetObject.id,
					description: normalized
				});
			}
		}

		const duplicates = new Set<string>();
		for (const row of rows) {
			if ((descriptionCounts.get(row.description) ?? 0) > 1) {
				duplicates.add(rowIdentifier(row.objectDataId, row.bacnetObjectId));
			}
		}

		duplicateDescriptionIds = duplicates;
		duplicateCheckDone = true;
	}

	$effect(() => {
		session;
		duplicateDescriptionIds = new Set();
		duplicateCheckDone = false;
	});
</script>

<div class="rounded-lg border bg-background p-4">
	<div class="mb-3 flex items-center gap-2">
		<CircleCheck class="size-4 text-primary" />
		<h2 class="text-sm font-semibold">Object data loaded</h2>
	</div>
	<div class="mb-4 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
		<span>{session.fileName}</span>
		<span>{session.objectDataExcel.length} object data</span>
		<Button type="button" size="sm" variant="outline" onclick={runDuplicateDescriptionCheck}>
			Check duplicate descriptions
		</Button>
		{#if duplicateCheckDone}
			<span>
				{duplicateDescriptionIds.size > 0
					? `${duplicateDescriptionIds.size} non-unique BACnet rows marked`
					: 'All BACnet descriptions are unique'}
			</span>
		{/if}
	</div>

	{#if session.objectDataExcel.length === 0}
		<div class="rounded-md border border-dashed p-4 text-sm text-muted-foreground">
			No ObjectData entries were found in the Excel sheet.
		</div>
	{:else}
		<div class="overflow-x-auto rounded-md border">
			<div class="min-w-[840px] text-xs">
				<div class="grid grid-cols-[32px_240px_1fr_96px_88px] border-b bg-muted/30 font-medium text-muted-foreground">
					<div class="px-2 py-1"></div>
					<div class="px-2 py-1">ObjectData ID</div>
					<div class="px-2 py-1">Description</div>
					<div class="px-2 py-1">BACnet</div>
					<div class="px-2 py-1">Optional</div>
				</div>

				{#each session.objectDataExcel as objectData}
					<details class="group border-b last:border-b-0">
						<summary class="list-none cursor-pointer">
							<div
								class="grid grid-cols-[32px_240px_1fr_96px_88px] items-center hover:bg-muted/20"
							>
								<div class="px-2 py-1.5 text-muted-foreground">
									<span class="group-open:hidden">▸</span><span class="hidden group-open:inline">▾</span>
								</div>
								<div class="truncate px-2 py-1.5 font-medium">{objectData.id}</div>
								<div class="truncate px-2 py-1.5 text-muted-foreground">{objectData.description || '-'}</div>
								<div class="px-2 py-1.5 text-muted-foreground">{objectData.bacnet_objects.length}</div>
								<div class="px-2 py-1.5 text-muted-foreground">{objectData.is_optional_anchor ? 'Yes' : 'No'}</div>
							</div>
						</summary>

						<div class="border-t bg-muted/10 px-2 py-2">
							{#if objectData.bacnet_objects.length === 0}
								<p class="px-2 py-1 text-muted-foreground">No BACnet objects for this entry.</p>
							{:else}
								<div class="overflow-x-auto rounded-sm border bg-background">
									<div class="min-w-[980px] text-[11px]">
										<div class="grid grid-cols-[180px_220px_70px_70px_180px_80px_90px_110px_140px_150px_150px_140px] border-b bg-muted/30 font-medium text-muted-foreground">
											<div class="px-2 py-1">Text fix</div>
											<div class="px-2 py-1">Description</div>
											<div class="px-2 py-1">Visible</div>
											<div class="px-2 py-1">Optional</div>
											<div class="px-2 py-1">Text individual</div>
											<div class="px-2 py-1">Type</div>
											<div class="px-2 py-1">Number</div>
											<div class="px-2 py-1">Hardware</div>
											<div class="px-2 py-1">Software ref</div>
											<div class="px-2 py-1">State text</div>
											<div class="px-2 py-1">Notification class</div>
											<div class="px-2 py-1">Alarm definition</div>
											<div class="px-2 py-1">Apparat</div>
										</div>

										{#each objectData.bacnet_objects as bacnetObject}
											<div
												class={`grid grid-cols-[180px_220px_70px_70px_180px_80px_90px_110px_140px_150px_150px_140px] border-b last:border-b-0 ${duplicateDescriptionIds.has(rowIdentifier(objectData.id, bacnetObject.id)) ? 'bg-destructive/10' : ''}`}
											>
												<div class="truncate px-2 py-1">{bacnetObject.text_fix || '-'}</div>
												<div class="truncate px-2 py-1 text-muted-foreground">{bacnetObject.description || '-'}</div>
												<div class="px-2 py-1">{bacnetObject.gms_visible ? 'Yes' : 'No'}</div>
												<div class="px-2 py-1">{bacnetObject.is_optional ? 'Yes' : 'No'}</div>
												<div class="truncate px-2 py-1">{bacnetObject.text_individual || '-'}</div>
												<div class="px-2 py-1">{bacnetObject.software_type || '-'}</div>
												<div class="px-2 py-1">{bacnetObject.software_number || '-'}</div>
												<div class="px-2 py-1">{bacnetObject.hardware_label || '-'}</div>
												<div class="truncate px-2 py-1">{bacnetObject.software_reference_label || '-'}</div>
												<div class="truncate px-2 py-1">{bacnetObject.state_text_label || '-'}</div>
												<div class="truncate px-2 py-1">{bacnetObject.notification_class_label || '-'}</div>
												<div class="truncate px-2 py-1">{bacnetObject.alarm_definition_label || '-'}</div>
												<div class="truncate px-2 py-1">{bacnetObject.apparat_label || '-'}</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}
						</div>
					</details>
				{/each}
			</div>
		</div>
	{/if}
</div>
