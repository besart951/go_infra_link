<script lang="ts">
  import type { CreateObjectDataRequest } from '$lib/domain/facility/object-data.js';

  interface PreparedObjectData {
    objectDataId: string;
    request: CreateObjectDataRequest;
    plannedAlarmDefinitions: Array<{
      bacnetIndex: number;
      bacnetSoftwareId: string;
      name: string;
      alarmTypeId?: string;
      alarmTypeCode?: string;
    }>;
    plannedSoftwareReferenceLinks: Array<{ fromSoftwareId: string; toSoftwareId: string }>;
    issues: {
      missingApparatLabels: string[];
      missingStateTextLabels: string[];
      missingNotificationClassLabels: string[];
      missingSoftwareReferences: string[];
      missingHardwareEntries: string[];
      missingSoftwareNumberEntries: string[];
      missingHardwareCount: number;
      missingSoftwareNumberCount: number;
    };
  }

  interface Props {
    preparedPayloads: PreparedObjectData[] | null;
    filteredPreparedPayloads: PreparedObjectData[];
  }

  let { preparedPayloads, filteredPreparedPayloads }: Props = $props();
</script>

{#if preparedPayloads}
  <div class="mb-4 space-y-2">
    {#if filteredPreparedPayloads.length === 0}
      <div class="rounded-md border border-dashed p-3 text-xs text-muted-foreground">
        Keine Einträge für den ausgewählten Filter gefunden.
      </div>
    {/if}
    {#each filteredPreparedPayloads as preparedItem}
      <details class="rounded-md border bg-background p-3 text-xs">
        <summary class="cursor-pointer font-medium">
          {preparedItem.objectDataId} - fehlende Details
        </summary>
        <div class="mt-2 space-y-2 text-muted-foreground">
          {#if preparedItem.issues.missingApparatLabels.length > 0}
            <div>
              <strong class="text-foreground">Fehlende Apparat-Bezeichnungen:</strong>
              <p>{preparedItem.issues.missingApparatLabels.join(', ')}</p>
            </div>
          {/if}
          {#if preparedItem.issues.missingStateTextLabels.length > 0}
            <div>
              <strong class="text-foreground">Fehlende Statustext-Bezeichnungen:</strong>
              <p>{preparedItem.issues.missingStateTextLabels.join(', ')}</p>
            </div>
          {/if}
          {#if preparedItem.issues.missingNotificationClassLabels.length > 0}
            <div>
              <strong class="text-foreground"
                >Fehlende Bezeichnungen der Benachrichtigungsklassen:</strong
              >
              <p>{preparedItem.issues.missingNotificationClassLabels.join(', ')}</p>
            </div>
          {/if}
          {#if preparedItem.plannedAlarmDefinitions.length > 0}
            <div>
              <strong class="text-foreground">Geplante Alarmtyp-Zuordnungen:</strong>
              <p>
                {preparedItem.plannedAlarmDefinitions
                  .map(
                    (entry) =>
                      `${entry.bacnetSoftwareId} -> ${entry.name}${entry.alarmTypeCode ? ` [${entry.alarmTypeCode}]` : ''}`
                  )
                  .join(' | ')}
              </p>
            </div>
          {/if}
          {#if preparedItem.plannedSoftwareReferenceLinks.length > 0}
            <div>
              <strong class="text-foreground">Geplante Software-Referenzverknüpfungen:</strong>
              <p>
                {preparedItem.plannedSoftwareReferenceLinks
                  .map((entry) => `${entry.fromSoftwareId} -> ${entry.toSoftwareId}`)
                  .join(' | ')}
              </p>
            </div>
          {/if}
          {#if preparedItem.issues.missingSoftwareReferences.length > 0}
            <div>
              <strong class="text-foreground">Fehlende Software-Referenzen:</strong>
              <p>{preparedItem.issues.missingSoftwareReferences.join(', ')}</p>
            </div>
          {/if}
          {#if preparedItem.issues.missingHardwareEntries.length > 0}
            <div>
              <strong class="text-foreground">Ungültige oder fehlende Hardware-Zeilen:</strong>
              <p>{preparedItem.issues.missingHardwareEntries.join(' | ')}</p>
            </div>
          {/if}
          {#if preparedItem.issues.missingSoftwareNumberEntries.length > 0}
            <div>
              <strong class="text-foreground">Ungültige Softwarenummern-Zeilen:</strong>
              <p>{preparedItem.issues.missingSoftwareNumberEntries.join(' | ')}</p>
            </div>
          {/if}
        </div>
      </details>
    {/each}
  </div>
{/if}
