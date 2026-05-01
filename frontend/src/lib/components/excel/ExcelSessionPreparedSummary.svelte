<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';

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
    <Button
      variant="link"
      class={filterClass('all', 'h-auto p-0 text-xs ')}
      onclick={() => onSetPrepareFilter('all')}
    >
      {preparedSummary.objectDataCount} Objektdaten vorbereitet
    </Button>
    <Button
      variant="link"
      class="ml-3 h-auto p-0 text-xs text-muted-foreground"
      onclick={() => onSetPrepareFilter('all')}
    >
      {preparedSummary.bacnetCount} BACnet-Objekte
    </Button>
    <Button
      variant="link"
      onclick={() => onSetPrepareFilter('missingApparats')}
      class={filterClass('missingApparats', 'ml-3 h-auto p-0 text-xs text-muted-foreground ')}
    >
      Fehlende Apparate: {preparedSummary.missingApparats}
    </Button>
    <Button
      variant="link"
      onclick={() => onSetPrepareFilter('missingStateTexts')}
      class={filterClass('missingStateTexts', 'ml-3 h-auto p-0 text-xs text-muted-foreground ')}
    >
      Fehlende Statustexte: {preparedSummary.missingStateTexts}
    </Button>
    <Button
      variant="link"
      onclick={() => onSetPrepareFilter('missingNotificationClasses')}
      class={filterClass(
        'missingNotificationClasses',
        'ml-3 h-auto p-0 text-xs text-muted-foreground '
      )}
    >
      Fehlende Benachrichtigungsklassen: {preparedSummary.missingNotificationClasses}
    </Button>
    <Button
      variant="link"
      onclick={() => onSetPrepareFilter('missingSoftwareReferences')}
      class={filterClass(
        'missingSoftwareReferences',
        'ml-3 h-auto p-0 text-xs text-muted-foreground '
      )}
    >
      Fehlende Software-Referenzen: {preparedSummary.missingSoftwareReferences}
    </Button>
    <Button
      variant="link"
      onclick={() => onSetPrepareFilter('missingHardware')}
      class={filterClass('missingHardware', 'ml-3 h-auto p-0 text-xs text-muted-foreground ')}
    >
      Fehlende Hardware: {preparedSummary.missingHardware}
    </Button>
    <Button
      variant="link"
      onclick={() => onSetPrepareFilter('missingSoftwareNumbers')}
      class={filterClass(
        'missingSoftwareNumbers',
        'ml-3 h-auto p-0 text-xs text-muted-foreground '
      )}
    >
      Fehlende Softwarenummern: {preparedSummary.missingSoftwareNumbers}
    </Button>
    <Button
      variant="link"
      onclick={() => onSetPrepareFilter('plannedAlarmDefinitions')}
      class={filterClass(
        'plannedAlarmDefinitions',
        'ml-3 h-auto p-0 text-xs text-muted-foreground '
      )}
    >
      Geplante Alarmtyp-Zuordnungen: {preparedSummary.plannedAlarmDefinitionCreates}
    </Button>
    <Button
      variant="link"
      onclick={() => onSetPrepareFilter('plannedSoftwareLinks')}
      class={filterClass('plannedSoftwareLinks', 'ml-3 h-auto p-0 text-xs text-muted-foreground ')}
    >
      Geplante Software-Verknüpfungen: {preparedSummary.plannedSoftwareReferenceLinks}
    </Button>
  </div>
{/if}
