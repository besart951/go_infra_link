<script lang="ts">
  import { CircleCheck } from '@lucide/svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import type { ExcelReadSession } from '$lib/domain/excel/index.js';

  interface Props {
    session: ExcelReadSession;
    duplicateSoftwareIds: Set<string>;
    duplicateCheckDone: boolean;
    preparing: boolean;
    creating: boolean;
    onRunDuplicateSoftwareCheck: () => void;
    onPrepareCreatePayloads: () => void | Promise<void>;
    onCreateAllPreparedSequentially: () => void | Promise<void>;
  }

  let {
    session,
    duplicateSoftwareIds,
    duplicateCheckDone,
    preparing,
    creating,
    onRunDuplicateSoftwareCheck,
    onPrepareCreatePayloads,
    onCreateAllPreparedSequentially
  }: Props = $props();
</script>

<div class="mb-3 flex items-center gap-2">
  <CircleCheck class="size-4 text-primary" />
  <h2 class="text-sm font-semibold">Objektdaten geladen</h2>
</div>
<div class="mb-4 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
  <span>{session.fileName}</span>
  <span>{session.objectDataExcel.length} Objektdaten</span>
  <Button type="button" size="sm" variant="outline" onclick={onRunDuplicateSoftwareCheck}>
    Doppelte Software-IDs prüfen
  </Button>
  <Button
    type="button"
    size="sm"
    variant="outline"
    onclick={onPrepareCreatePayloads}
    disabled={preparing}
  >
    {preparing ? 'Vorbereitung läuft...' : 'Erstellung vorbereiten'}
  </Button>
  <Button
    type="button"
    size="sm"
    variant="outline"
    onclick={onCreateAllPreparedSequentially}
    disabled={creating || preparing}
  >
    {creating ? 'Erstellung läuft...' : 'Alles nacheinander erstellen'}
  </Button>
  {#if duplicateCheckDone}
    <span>
      {duplicateSoftwareIds.size > 0
        ? `${duplicateSoftwareIds.size} nicht eindeutige BACnet-Zeilen markiert`
        : 'Alle BACnet-Software-IDs sind eindeutig'}
    </span>
  {/if}
</div>
