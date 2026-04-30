<script lang="ts">
  type PrepareFilterKey =
    | 'all'
    | 'missingApparats'
    | 'missingStateTexts'
    | 'missingNotificationClasses'
    | 'missingSoftwareReferences'
    | 'missingHardware'
    | 'missingSoftwareNumbers'
    | 'plannedAlarmDefinitions'
    | 'plannedSoftwareLinks';

  interface PreparedSummary {
    objectDataCount: number;
    bacnetCount: number;
    missingApparats: number;
    missingStateTexts: number;
    missingNotificationClasses: number;
    missingSoftwareReferences: number;
    missingHardware: number;
    missingSoftwareNumbers: number;
    plannedAlarmDefinitionCreates: number;
    plannedSoftwareReferenceLinks: number;
  }

  interface Props {
    preparedSummary: PreparedSummary | null;
    activePrepareFilter: PrepareFilterKey;
    onSetPrepareFilter: (filter: PrepareFilterKey) => void;
  }

  let { preparedSummary, activePrepareFilter, onSetPrepareFilter }: Props = $props();

  function filterClass(filter: PrepareFilterKey, prefix = '') {
    return `${prefix}cursor-pointer ${activePrepareFilter === filter ? 'font-semibold text-foreground underline' : ''}`;
  }
</script>

{#if preparedSummary}
  <div class="mb-4 rounded-md border bg-muted/20 p-3 text-xs text-muted-foreground">
    <button type="button" onclick={() => onSetPrepareFilter('all')} class={filterClass('all')}>
      {preparedSummary.objectDataCount} Objektdaten vorbereitet
    </button>
    <button type="button" onclick={() => onSetPrepareFilter('all')} class="ml-3 cursor-pointer">
      {preparedSummary.bacnetCount} BACnet-Objekte
    </button>
    <button
      type="button"
      onclick={() => onSetPrepareFilter('missingApparats')}
      class={filterClass('missingApparats', 'ml-3 ')}
    >
      Fehlende Apparate: {preparedSummary.missingApparats}
    </button>
    <button
      type="button"
      onclick={() => onSetPrepareFilter('missingStateTexts')}
      class={filterClass('missingStateTexts', 'ml-3 ')}
    >
      Fehlende Statustexte: {preparedSummary.missingStateTexts}
    </button>
    <button
      type="button"
      onclick={() => onSetPrepareFilter('missingNotificationClasses')}
      class={filterClass('missingNotificationClasses', 'ml-3 ')}
    >
      Fehlende Benachrichtigungsklassen: {preparedSummary.missingNotificationClasses}
    </button>
    <button
      type="button"
      onclick={() => onSetPrepareFilter('missingSoftwareReferences')}
      class={filterClass('missingSoftwareReferences', 'ml-3 ')}
    >
      Fehlende Software-Referenzen: {preparedSummary.missingSoftwareReferences}
    </button>
    <button
      type="button"
      onclick={() => onSetPrepareFilter('missingHardware')}
      class={filterClass('missingHardware', 'ml-3 ')}
    >
      Fehlende Hardware: {preparedSummary.missingHardware}
    </button>
    <button
      type="button"
      onclick={() => onSetPrepareFilter('missingSoftwareNumbers')}
      class={filterClass('missingSoftwareNumbers', 'ml-3 ')}
    >
      Fehlende Softwarenummern: {preparedSummary.missingSoftwareNumbers}
    </button>
    <button
      type="button"
      onclick={() => onSetPrepareFilter('plannedAlarmDefinitions')}
      class={filterClass('plannedAlarmDefinitions', 'ml-3 ')}
    >
      Geplante Alarmtyp-Zuordnungen: {preparedSummary.plannedAlarmDefinitionCreates}
    </button>
    <button
      type="button"
      onclick={() => onSetPrepareFilter('plannedSoftwareLinks')}
      class={filterClass('plannedSoftwareLinks', 'ml-3 ')}
    >
      Geplante Software-Verknüpfungen: {preparedSummary.plannedSoftwareReferenceLinks}
    </button>
  </div>
{/if}
