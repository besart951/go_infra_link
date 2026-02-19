<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Trash2 } from '@lucide/svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import { getErrorMessage } from '$lib/api/client.js';
	import { alarmUnitRepository } from '$lib/infrastructure/api/alarmUnitRepository.js';
	import { alarmFieldRepository } from '$lib/infrastructure/api/alarmFieldRepository.js';
	import { alarmTypeRepository } from '$lib/infrastructure/api/alarmTypeRepository.js';
	import type {
		AlarmField,
		AlarmType,
		AlarmTypeField,
		CreateAlarmTypeFieldRequest,
		Unit
	} from '$lib/domain/facility/index.js';

	let units = $state<Unit[]>([]);
	let fields = $state<AlarmField[]>([]);
	let types = $state<AlarmType[]>([]);
	let selectedTypeId = $state('');
	let typeFields = $state<AlarmTypeField[]>([]);
	let loading = $state(false);

	let unitForm = $state({ code: '', symbol: '', name: '' });
	let fieldForm = $state({ key: '', label: '', data_type: 'string', default_unit_code: '' });
	let typeForm = $state({ code: '', name: '' });
	let mapForm = $state<CreateAlarmTypeFieldRequest>({
		alarm_field_id: '',
		display_order: 0,
		is_required: false,
		is_user_editable: true,
		ui_group: '',
		default_unit_id: ''
	});

	const dataTypeOptions = ['number', 'integer', 'boolean', 'string', 'enum', 'duration', 'state_map', 'json'];
	const selectClass =
		'h-9 w-full rounded-md border border-input bg-background px-3 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

	async function loadAll() {
		loading = true;
		try {
			const [unitsRes, fieldsRes, typesRes] = await Promise.all([
				alarmUnitRepository.list({ pagination: { page: 1, pageSize: 200 }, search: { text: '' } }),
				alarmFieldRepository.list({ pagination: { page: 1, pageSize: 200 }, search: { text: '' } }),
				alarmTypeRepository.list({ page: 1, pageSize: 200 })
			]);
			units = unitsRes.items;
			fields = fieldsRes.items;
			types = typesRes.items;
			if (!selectedTypeId && types.length > 0) selectedTypeId = types[0].id;
			if (selectedTypeId) {
				await loadTypeFields(selectedTypeId);
			}
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		} finally {
			loading = false;
		}
	}

	async function loadTypeFields(typeId: string) {
		if (!typeId) {
			typeFields = [];
			return;
		}
		try {
			const detail = await alarmTypeRepository.getWithFields(typeId);
			typeFields = detail.fields ?? [];
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		}
	}

	async function createUnit() {
		try {
			await alarmUnitRepository.create(unitForm);
			unitForm = { code: '', symbol: '', name: '' };
			await loadAll();
			addToast('Einheit erstellt', 'success');
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		}
	}

	async function deleteUnit(id: string) {
		try {
			await alarmUnitRepository.delete(id);
			await loadAll();
			addToast('Einheit gelöscht', 'success');
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		}
	}

	async function createField() {
		try {
			await alarmFieldRepository.create({
				key: fieldForm.key,
				label: fieldForm.label,
				data_type: fieldForm.data_type as AlarmField['data_type'],
				default_unit_code: fieldForm.default_unit_code || undefined
			});
			fieldForm = { key: '', label: '', data_type: 'string', default_unit_code: '' };
			await loadAll();
			addToast('Alarmfeld erstellt', 'success');
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		}
	}

	async function deleteField(id: string) {
		try {
			await alarmFieldRepository.delete(id);
			await loadAll();
			addToast('Alarmfeld gelöscht', 'success');
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		}
	}

	async function createType() {
		try {
			const created = await alarmTypeRepository.create(typeForm);
			typeForm = { code: '', name: '' };
			selectedTypeId = created.id;
			await loadAll();
			addToast('Alarmtyp erstellt', 'success');
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		}
	}

	async function deleteType(id: string) {
		try {
			await alarmTypeRepository.delete(id);
			if (selectedTypeId === id) selectedTypeId = '';
			await loadAll();
			addToast('Alarmtyp gelöscht', 'success');
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		}
	}

	async function createMapping() {
		if (!selectedTypeId) return;
		try {
			await alarmTypeRepository.createField(selectedTypeId, {
				...mapForm,
				default_unit_id: mapForm.default_unit_id || undefined,
				ui_group: mapForm.ui_group || undefined
			});
			mapForm = {
				alarm_field_id: '',
				display_order: 0,
				is_required: false,
				is_user_editable: true,
				ui_group: '',
				default_unit_id: ''
			};
			await loadTypeFields(selectedTypeId);
			addToast('Zuordnung erstellt', 'success');
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		}
	}

	async function deleteMapping(id: string) {
		try {
			await alarmTypeRepository.deleteField(id);
			await loadTypeFields(selectedTypeId);
			addToast('Zuordnung gelöscht', 'success');
		} catch (err) {
			addToast(getErrorMessage(err), 'error');
		}
	}

	onMount(loadAll);
</script>

<svelte:head>
	<title>Alarm-Katalog | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-start justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Alarm-Katalog</h1>
			<p class="text-sm text-muted-foreground">
				Verwaltung von Alarmtypen, Alarmfeldern, Einheiten und Type-Field-Zuordnungen.
			</p>
		</div>
		{#if loading}
			<span class="text-sm text-muted-foreground">Lade Daten…</span>
		{/if}
	</div>

	<div class="grid gap-6 xl:grid-cols-2">
		<Card.Root>
			<Card.Header class="border-b">
				<Card.Title>Einheiten</Card.Title>
				<Card.Description>Verfügbare Einheiten für Alarmfelder und Zuordnungen.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 md:grid-cols-3">
					<div class="space-y-2">
						<Label for="unit-code">Code</Label>
						<Input id="unit-code" bind:value={unitForm.code} />
					</div>
					<div class="space-y-2">
						<Label for="unit-symbol">Symbol</Label>
						<Input id="unit-symbol" bind:value={unitForm.symbol} />
					</div>
					<div class="space-y-2">
						<Label for="unit-name">Name</Label>
						<Input id="unit-name" bind:value={unitForm.name} />
					</div>
				</div>
				<div class="flex justify-end">
					<Button onclick={createUnit} disabled={!unitForm.code || !unitForm.symbol || !unitForm.name}>
						Einheit erstellen
					</Button>
				</div>
				<div class="overflow-hidden rounded-md border">
					<div class="max-h-72 overflow-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Code</Table.Head>
									<Table.Head>Symbol</Table.Head>
									<Table.Head>Name</Table.Head>
									<Table.Head class="w-24 text-right">Aktion</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#if units.length === 0}
									<Table.Row>
										<Table.Cell colspan={4} class="py-8 text-center text-sm text-muted-foreground">
											Keine Einheiten vorhanden.
										</Table.Cell>
									</Table.Row>
								{:else}
									{#each units as u}
										<Table.Row>
											<Table.Cell class="font-medium">{u.code}</Table.Cell>
											<Table.Cell>{u.symbol}</Table.Cell>
											<Table.Cell>{u.name}</Table.Cell>
											<Table.Cell class="text-right">
												<Button
													size="icon-sm"
													variant="ghost"
													class="text-destructive hover:text-destructive"
													onclick={() => deleteUnit(u.id)}
													aria-label="Einheit löschen"
													title="Einheit löschen"
												>
													<Trash2 class="size-4" />
												</Button>
											</Table.Cell>
										</Table.Row>
									{/each}
								{/if}
							</Table.Body>
						</Table.Root>
					</div>
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header class="border-b">
				<Card.Title>Alarmfelder</Card.Title>
				<Card.Description>Definition von Key, Datentyp und optionaler Standard-Einheit.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 md:grid-cols-2">
					<div class="space-y-2">
						<Label for="field-key">Key</Label>
						<Input id="field-key" bind:value={fieldForm.key} />
					</div>
					<div class="space-y-2">
						<Label for="field-label">Label</Label>
						<Input id="field-label" bind:value={fieldForm.label} />
					</div>
					<div class="space-y-2">
						<Label for="field-datatype">Datentyp</Label>
						<select id="field-datatype" class={selectClass} bind:value={fieldForm.data_type}>
							{#each dataTypeOptions as dt}
								<option value={dt}>{dt}</option>
							{/each}
						</select>
					</div>
					<div class="space-y-2">
						<Label for="field-unit">Default Unit Code</Label>
						<select id="field-unit" class={selectClass} bind:value={fieldForm.default_unit_code}>
							<option value="">-</option>
							{#each units as u}
								<option value={u.code}>{u.code}</option>
							{/each}
						</select>
					</div>
				</div>
				<div class="flex justify-end">
					<Button onclick={createField} disabled={!fieldForm.key || !fieldForm.label}>
						Alarmfeld erstellen
					</Button>
				</div>
				<div class="overflow-hidden rounded-md border">
					<div class="max-h-72 overflow-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Key</Table.Head>
									<Table.Head>Label</Table.Head>
									<Table.Head>Typ</Table.Head>
									<Table.Head>Unit</Table.Head>
									<Table.Head class="w-24 text-right">Aktion</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#if fields.length === 0}
									<Table.Row>
										<Table.Cell colspan={5} class="py-8 text-center text-sm text-muted-foreground">
											Keine Alarmfelder vorhanden.
										</Table.Cell>
									</Table.Row>
								{:else}
									{#each fields as f}
										<Table.Row>
											<Table.Cell class="font-medium">{f.key}</Table.Cell>
											<Table.Cell>{f.label}</Table.Cell>
											<Table.Cell>{f.data_type}</Table.Cell>
											<Table.Cell>{f.default_unit_code ?? '-'}</Table.Cell>
											<Table.Cell class="text-right">
												<Button
													size="icon-sm"
													variant="ghost"
													class="text-destructive hover:text-destructive"
													onclick={() => deleteField(f.id)}
													aria-label="Alarmfeld löschen"
													title="Alarmfeld löschen"
												>
													<Trash2 class="size-4" />
												</Button>
											</Table.Cell>
										</Table.Row>
									{/each}
								{/if}
							</Table.Body>
						</Table.Root>
					</div>
				</div>
			</Card.Content>
		</Card.Root>
	</div>

	<div class="grid gap-6 xl:grid-cols-2">
		<Card.Root>
			<Card.Header class="border-b">
				<Card.Title>Alarmtypen</Card.Title>
				<Card.Description>Alarmtypen mit fachlichem Code und Anzeigenamen.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="grid gap-3 md:grid-cols-2">
					<div class="space-y-2">
						<Label for="type-code">Code</Label>
						<Input id="type-code" bind:value={typeForm.code} />
					</div>
					<div class="space-y-2">
						<Label for="type-name">Name</Label>
						<Input id="type-name" bind:value={typeForm.name} />
					</div>
				</div>
				<div class="flex justify-end">
					<Button onclick={createType} disabled={!typeForm.code || !typeForm.name}>
						Alarmtyp erstellen
					</Button>
				</div>
				<div class="overflow-hidden rounded-md border">
					<div class="max-h-72 overflow-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Code</Table.Head>
									<Table.Head>Name</Table.Head>
									<Table.Head class="w-24 text-right">Aktion</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#if types.length === 0}
									<Table.Row>
										<Table.Cell colspan={3} class="py-8 text-center text-sm text-muted-foreground">
											Keine Alarmtypen vorhanden.
										</Table.Cell>
									</Table.Row>
								{:else}
									{#each types as t}
										<Table.Row>
											<Table.Cell class="font-medium">{t.code}</Table.Cell>
											<Table.Cell>{t.name}</Table.Cell>
											<Table.Cell class="text-right">
												<Button
													size="icon-sm"
													variant="ghost"
													class="text-destructive hover:text-destructive"
													onclick={() => deleteType(t.id)}
													aria-label="Alarmtyp löschen"
													title="Alarmtyp löschen"
												>
													<Trash2 class="size-4" />
												</Button>
											</Table.Cell>
										</Table.Row>
									{/each}
								{/if}
							</Table.Body>
						</Table.Root>
					</div>
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header class="border-b">
				<Card.Title>Type-Field-Zuordnungen</Card.Title>
				<Card.Description>Zuordnung von Feldern zu Alarmtypen inkl. Reihenfolge und Regeln.</Card.Description>
			</Card.Header>
			<Card.Content class="space-y-4">
				<div class="space-y-2">
					<Label for="mapping-type">Alarmtyp</Label>
					<select
						id="mapping-type"
						class={selectClass}
						bind:value={selectedTypeId}
						onchange={() => loadTypeFields(selectedTypeId)}
					>
						<option value="">Bitte wählen</option>
						{#each types as t}
							<option value={t.id}>{t.name} ({t.code})</option>
						{/each}
					</select>
				</div>

				<div class="grid gap-3 md:grid-cols-2">
					<div class="space-y-2">
						<Label for="mapping-field">Alarmfeld</Label>
						<select id="mapping-field" class={selectClass} bind:value={mapForm.alarm_field_id}>
							<option value="">Bitte wählen</option>
							{#each fields as f}
								<option value={f.id}>{f.label} ({f.key})</option>
							{/each}
						</select>
					</div>
					<div class="space-y-2">
						<Label for="mapping-order">Display Order</Label>
						<Input id="mapping-order" type="number" bind:value={mapForm.display_order} />
					</div>
					<div class="space-y-2">
						<Label for="mapping-group">UI Group</Label>
						<Input id="mapping-group" bind:value={mapForm.ui_group} />
					</div>
					<div class="space-y-2">
						<Label for="mapping-unit">Default Unit</Label>
						<select id="mapping-unit" class={selectClass} bind:value={mapForm.default_unit_id}>
							<option value="">-</option>
							{#each units as u}
								<option value={u.id}>{u.code}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="flex flex-wrap gap-6">
					<label class="flex items-center gap-2 text-sm text-foreground">
						<Checkbox bind:checked={mapForm.is_required} />
						Pflicht
					</label>
					<label class="flex items-center gap-2 text-sm text-foreground">
						<Checkbox bind:checked={mapForm.is_user_editable} />
						Editierbar
					</label>
				</div>

				<div class="flex justify-end">
					<Button onclick={createMapping} disabled={!selectedTypeId || !mapForm.alarm_field_id}>
						Zuordnung erstellen
					</Button>
				</div>

				<div class="overflow-hidden rounded-md border">
					<div class="max-h-72 overflow-auto">
						<Table.Root>
							<Table.Header>
								<Table.Row>
									<Table.Head>Feld</Table.Head>
									<Table.Head>Group</Table.Head>
									<Table.Head>Pflicht</Table.Head>
									<Table.Head>Order</Table.Head>
									<Table.Head class="w-24 text-right">Aktion</Table.Head>
								</Table.Row>
							</Table.Header>
							<Table.Body>
								{#if !selectedTypeId}
									<Table.Row>
										<Table.Cell colspan={5} class="py-8 text-center text-sm text-muted-foreground">
											Bitte zuerst einen Alarmtyp auswählen.
										</Table.Cell>
									</Table.Row>
								{:else if typeFields.length === 0}
									<Table.Row>
										<Table.Cell colspan={5} class="py-8 text-center text-sm text-muted-foreground">
											Keine Zuordnungen für diesen Alarmtyp vorhanden.
										</Table.Cell>
									</Table.Row>
								{:else}
									{#each typeFields as tf}
										<Table.Row>
											<Table.Cell class="font-medium">{tf.alarm_field?.label ?? tf.alarm_field_id}</Table.Cell>
											<Table.Cell>{tf.ui_group ?? '-'}</Table.Cell>
											<Table.Cell>{tf.is_required ? 'Ja' : 'Nein'}</Table.Cell>
											<Table.Cell>{tf.display_order}</Table.Cell>
											<Table.Cell class="text-right">
												<Button
													size="icon-sm"
													variant="ghost"
													class="text-destructive hover:text-destructive"
													onclick={() => deleteMapping(tf.id)}
													aria-label="Zuordnung löschen"
													title="Zuordnung löschen"
												>
													<Trash2 class="size-4" />
												</Button>
											</Table.Cell>
										</Table.Row>
									{/each}
								{/if}
							</Table.Body>
						</Table.Root>
					</div>
				</div>
			</Card.Content>
		</Card.Root>
	</div>
</div>
