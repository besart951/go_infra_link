<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import RefreshCw from '@lucide/svelte/icons/refresh-cw';
  import { ApiException } from '$lib/api/client.js';
  import {
    loadFacilityDetail,
    patchBuildingDetail,
    patchFacilityDetail,
    type FacilityDetailResponse
  } from '$lib/api/facilityDetails.js';
  import {
    facilityDetailCache,
    type FacilityDetailKind,
    type FacilityDetailScope
  } from '$lib/services/facilityDetailCache.js';
  import { facilityReferenceDataCache } from '$lib/services/facilityReferenceDataCache.js';
  import {
    useProjectSyncCoordinator,
    type ProjectChange
  } from '$lib/services/projectCollaboration.svelte.js';
  import DetailRealtimeStatus, { type DetailSaveStatus } from './DetailRealtimeStatus.svelte';
  import InlineAutosaveField from './InlineAutosaveField.svelte';
  import DetailRelations from './DetailRelations.svelte';

  interface Entity {
    id?: string;
    version?: number;
    iws_code?: string;
    building_group?: number;
    control_cabinet_nr?: string | null;
    building_id?: string;
    device_name?: string;
    ga_device?: string | null;
    device_description?: string | null;
    device_location?: string | null;
    ip_address?: string | null;
    subnet?: string | null;
    gateway?: string | null;
    vlan?: string | null;
    number?: number | null;
    document_name?: string | null;
    bmk?: string | null;
    description?: string | null;
    text_fix?: string | null;
    apparat_nr?: number | null;
    apparat_id?: string;
    system_part_id?: string | null;
    sps_controller_system_type_id?: string;
  }

  type FieldConfig = {
    key: keyof Entity;
    label: string;
    type?: 'text' | 'number';
    description?: string;
    readOnly?: boolean;
    minLength?: number;
    maxLength?: number;
    min?: number;
    max?: number;
    transform?: (value: string) => string;
  };

  let {
    kind,
    id,
    projectId
  }: {
    kind: FacilityDetailKind;
    id: string;
    projectId?: string;
  } = $props();

  let detail = $state.raw<FacilityDetailResponse | null>(null);
  let loading = $state(true);
  let loadError = $state<string | null>(null);
  let status = $state<DetailSaveStatus>('saved');
  let remoteChangePending = $state(false);
  let pendingConflict = $state<{ key: keyof Entity; value: string | number } | null>(null);
  let saveQueue: Promise<void> = Promise.resolve();
  let relationPage = $state(1);

  const scope = $derived<FacilityDetailScope>(
    projectId ? { projectId, relationPage } : { relationPage }
  );
  const collaboration = $derived.by(() => (projectId ? useProjectSyncCoordinator() : null));
  const entity = $derived.by<Entity>(() => entityFromDetail(kind, detail));
  const canUpdate = $derived(
    detail?.capabilities?.can_update === true && !(projectId && kind === 'buildings')
  );
  const relations = $derived(detail?.relations ?? []);
  const title = $derived.by(() => detailTitle(kind, entity));
  const description = $derived(
    projectId ? 'Projektbezogene Hierarchieansicht' : 'Globale Anlagenhierarchie'
  );
  const fields = $derived.by(() => detailFields(kind));
  const backHref = $derived(projectId ? `/projects/${projectId}` : `/facility/${kind}`);

  async function reload(force = false, liveUpdate = false): Promise<void> {
    loading = true;
    loadError = null;
    try {
      detail = await loadFacilityDetail(kind, id, scope, { force });
      remoteChangePending = false;
      status = liveUpdate ? 'updated' : 'saved';
    } catch (error) {
      loadError =
        error instanceof Error ? error.message : 'Detaildaten konnten nicht geladen werden.';
    } finally {
      loading = false;
    }
  }

  function saveField(key: keyof Entity, value: string | number): Promise<void> {
    const field = fields.find((item) => item.key === key);
    const normalizedValue =
      typeof value === 'string' && field?.transform ? field.transform(value) : value;
    return enqueueSave(async () => {
      try {
        await persistField(entity, key, normalizedValue);
        pendingConflict = null;
        await reload(true);
      } catch (error) {
        if (error instanceof ApiException && error.status === 409) {
          pendingConflict = { key, value };
          remoteChangePending = true;
        }
        throw error;
      }
    });
  }

  async function persistField(
    current: Entity,
    key: keyof Entity,
    value: string | number
  ): Promise<void> {
    const version = current.version;
    if (!version) throw new Error('Die aktuelle Version konnte nicht geladen werden.');

    if (kind === 'buildings') {
      if (projectId) throw new Error('Gebäude sind im Projekt schreibgeschützt.');
      await patchBuildingDetail(id, {
        base_version: version,
        [key]: value
      });
    } else {
      const patch = detailPatch(kind, current, key, value);
      await patchFacilityDetail(kind, id, patch, scope);
    }
  }

  async function retryConflictSave(): Promise<void> {
    const conflict = pendingConflict;
    if (!conflict) return;

    await enqueueSave(async () => {
      status = 'saving';
      try {
        const latest = await loadFacilityDetail(kind, id, scope, { force: true });
        await persistField(entityFromDetail(kind, latest), conflict.key, conflict.value);
        pendingConflict = null;
        remoteChangePending = false;
        await reload(true);
      } catch {
        status = 'conflict';
      }
    });
  }

  function enqueueSave(operation: () => Promise<void>): Promise<void> {
    const next = saveQueue.then(operation, operation);
    saveQueue = next.then(
      () => undefined,
      () => undefined
    );
    return next;
  }

  function handleRealtimeChange(): void {
    facilityDetailCache.invalidateForFacilityChange({
      type: 'facility.changed',
      resource: realtimeResource(kind),
      action: 'updated',
      ids: [id],
      at: new Date().toISOString()
    });
    handleExternalChange();
  }

  onMount(() => {
    void reload();
    if (collaboration) {
      return collaboration.subscribeProjectChanges((change: ProjectChange) => {
        if (isRelevantProjectChange(change)) handleExternalChange();
      });
    }

    return facilityReferenceDataCache.subscribeFacilityChanges((event) => {
      facilityDetailCache.invalidateForFacilityChange(event);
      if (
        event.resource === 'all' ||
        event.resource === realtimeResource(kind) ||
        affectsHierarchy(event.resource, kind)
      ) {
        handleRealtimeChange();
      }
    });
  });

  function handleExternalChange(): void {
    facilityDetailCache.invalidate(kind, id, scope);
    if (status === 'editing' || status === 'saving') {
      remoteChangePending = true;
      status = 'conflict';
      return;
    }
    void reload(true, true);
  }

  function updateStatus(next: DetailSaveStatus): void {
    if (remoteChangePending && next === 'saved') return;
    status = next;
  }

  function changeRelationPage(nextPage: number): void {
    if (nextPage < 1 || nextPage === relationPage) return;
    relationPage = nextPage;
    void reload();
  }

  function entityFromDetail(
    kind: FacilityDetailKind,
    response: FacilityDetailResponse | null
  ): Entity {
    if (!response) return {};
    const detail = response as unknown as Record<string, Entity | undefined>;
    if (kind === 'buildings') return detail.building ?? {};
    if (kind === 'control-cabinets') return detail.control_cabinet ?? {};
    if (kind === 'sps-controllers') return detail.sps_controller ?? {};
    if (kind === 'sps-controller-system-types') return detail.sps_controller_system_type ?? {};
    return detail.field_device ?? {};
  }

  function detailPatch(
    kind: Exclude<FacilityDetailKind, 'buildings'>,
    current: Entity,
    key: keyof Entity,
    value: string | number
  ) {
    const base = { base_version: requiredVersion(current.version) };
    switch (kind) {
      case 'control-cabinets':
        return { ...base, [key]: value };
      case 'sps-controllers':
        return { ...base, [key]: value };
      case 'sps-controller-system-types':
        return { ...base, [key]: value };
      case 'field-devices':
        if (!current.system_part_id)
          throw new Error('Systemteil fehlt und kann nicht gespeichert werden.');
        return { ...base, system_part_id: current.system_part_id, [key]: value };
    }
  }

  function requiredVersion(version: number | undefined): number {
    if (!version) throw new Error('Die aktuelle Version konnte nicht geladen werden.');
    return version;
  }

  function detailFields(kind: FacilityDetailKind): FieldConfig[] {
    switch (kind) {
      case 'buildings':
        return [
          { key: 'iws_code', label: 'IWS-Code', minLength: 4, maxLength: 4 },
          { key: 'building_group', label: 'Gebäudegruppe', type: 'number' }
        ];
      case 'control-cabinets':
        return [{ key: 'control_cabinet_nr', label: 'Schaltschranknummer', maxLength: 11 }];
      case 'sps-controllers':
        return [
          {
            key: 'device_name',
            label: 'Name',
            readOnly: true,
            description: 'Wird automatisch aus IWS-Code, Schaltschranknummer und GA-Gerät gebildet.'
          },
          {
            key: 'ga_device',
            label: 'GA-Gerät',
            minLength: 3,
            maxLength: 3,
            transform: (value) => value.toUpperCase(),
            description: 'Genau drei Grossbuchstaben (A–Z).'
          },
          { key: 'device_description', label: 'Beschreibung', maxLength: 250 },
          { key: 'device_location', label: 'Standort', maxLength: 250 },
          { key: 'ip_address', label: 'IP-Adresse', maxLength: 50 },
          { key: 'subnet', label: 'Subnetz', maxLength: 50 },
          { key: 'gateway', label: 'Gateway', maxLength: 50 },
          { key: 'vlan', label: 'VLAN', maxLength: 50 }
        ];
      case 'sps-controller-system-types':
        return [
          { key: 'number', label: 'Nummer', type: 'number', min: 1 },
          { key: 'document_name', label: 'Dokumentname', maxLength: 250 }
        ];
      case 'field-devices':
        return [
          { key: 'bmk', label: 'BMK', maxLength: 10 },
          { key: 'description', label: 'Beschreibung', maxLength: 250 },
          { key: 'text_fix', label: 'Individueller Text', maxLength: 250 },
          { key: 'apparat_nr', label: 'Apparatnummer', type: 'number', min: 1, max: 99 }
        ];
    }
  }

  function detailTitle(kind: FacilityDetailKind, current: Entity): string {
    switch (kind) {
      case 'buildings':
        return current.iws_code || 'Gebäude';
      case 'control-cabinets':
        return current.control_cabinet_nr || 'Schaltschrank';
      case 'sps-controllers':
        return current.device_name || 'SPS-Regler';
      case 'sps-controller-system-types':
        return current.document_name || 'SPS-Systemtyp';
      case 'field-devices':
        return current.bmk || current.description || 'Feldgerät';
    }
  }

  function realtimeResource(kind: FacilityDetailKind) {
    return {
      buildings: 'buildings',
      'control-cabinets': 'control_cabinets',
      'sps-controllers': 'sps_controllers',
      'sps-controller-system-types': 'sps_controller_system_types',
      'field-devices': 'field_devices'
    }[kind] as
      | 'buildings'
      | 'control_cabinets'
      | 'sps_controllers'
      | 'sps_controller_system_types'
      | 'field_devices';
  }

  function affectsHierarchy(resource: string, kind: FacilityDetailKind): boolean {
    const hierarchy = {
      buildings: [
        'control_cabinets',
        'sps_controllers',
        'sps_controller_system_types',
        'field_devices'
      ],
      'control-cabinets': [
        'buildings',
        'sps_controllers',
        'sps_controller_system_types',
        'field_devices'
      ],
      'sps-controllers': [
        'buildings',
        'control_cabinets',
        'sps_controller_system_types',
        'field_devices'
      ],
      'sps-controller-system-types': [
        'buildings',
        'control_cabinets',
        'sps_controllers',
        'field_devices'
      ],
      'field-devices': [
        'buildings',
        'control_cabinets',
        'sps_controllers',
        'sps_controller_system_types',
        'apparats',
        'system_parts',
        'bacnet_objects'
      ]
    };
    return hierarchy[kind].includes(resource);
  }

  function isRelevantProjectChange(change: ProjectChange): boolean {
    const aggregateByKind: Record<FacilityDetailKind, string[]> = {
      buildings: [
        'building',
        'control_cabinet',
        'sps_controller',
        'sps_controller_system_type',
        'field_device'
      ],
      'control-cabinets': [
        'control_cabinet',
        'sps_controller',
        'sps_controller_system_type',
        'field_device'
      ],
      'sps-controllers': [
        'control_cabinet',
        'sps_controller',
        'sps_controller_system_type',
        'field_device'
      ],
      'sps-controller-system-types': [
        'sps_controller',
        'sps_controller_system_type',
        'field_device'
      ],
      'field-devices': [
        'control_cabinet',
        'sps_controller',
        'sps_controller_system_type',
        'field_device',
        'bacnet_object'
      ]
    };
    return aggregateByKind[kind].includes(change.aggregate_type);
  }
</script>

<div class="mx-auto w-full max-w-6xl space-y-6">
  <EntityListHeader {title} {description} {backHref} backLabel="Zurück">
    <DetailRealtimeStatus {status} />
  </EntityListHeader>

  {#if remoteChangePending || status === 'conflict'}
    <div
      class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-warning/40 bg-warning/10 px-4 py-3 text-sm text-warning-foreground"
    >
      <span>
        {pendingConflict
          ? 'Deine Änderung kollidiert mit einer neueren Version. Dein lokaler Wert bleibt erhalten.'
          : 'Diese Ansicht wurde im Hintergrund geändert. Dein lokaler Entwurf bleibt unverändert.'}
      </span>
      <div class="flex items-center gap-2">
        <Button size="sm" variant="outline" onclick={() => reload(true, true)}>Neu laden</Button>
        {#if pendingConflict}
          <Button size="sm" onclick={() => void retryConflictSave()}>Erneut speichern</Button>
        {/if}
      </div>
    </div>
  {/if}

  {#if loading && !detail}
    <Card
      ><CardContent class="py-12 text-sm text-muted-foreground"
        >Detaildaten werden geladen …</CardContent
      ></Card
    >
  {:else if loadError}
    <Card class="border-destructive/50"
      ><CardContent class="space-y-3 py-6 text-sm text-destructive"
        ><p>{loadError}</p>
        <Button variant="outline" size="sm" onclick={() => reload(true)}
          ><RefreshCw class="size-4" /> Erneut versuchen</Button
        ></CardContent
      ></Card
    >
  {:else}
    <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,0.72fr)]">
      <Card>
        <CardHeader class="flex-row items-center justify-between gap-3 space-y-0">
          <CardTitle>Details</CardTitle>
          {#if canUpdate}
            <span class="text-xs text-muted-foreground"
              >Speichert automatisch beim Verlassen des Feldes</span
            >
          {/if}
        </CardHeader>
        <CardContent class="grid gap-x-6 gap-y-5 sm:grid-cols-2">
          {#each fields as field (field.key)}
            <InlineAutosaveField
              label={field.label}
              type={field.type}
              value={entity[field.key] as string | number | null | undefined}
              disabled={!canUpdate || field.readOnly}
              description={field.description}
              minLength={field.minLength}
              maxLength={field.maxLength}
              min={field.min}
              max={field.max}
              transform={field.transform}
              onSave={(value) => saveField(field.key, value)}
              onStatusChange={updateStatus}
            />
          {/each}
        </CardContent>
      </Card>
      <DetailRelations {relations} {scope} onPageChange={changeRelationPage} />
    </div>
  {/if}
</div>
