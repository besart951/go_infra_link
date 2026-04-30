<script lang="ts">
  import type { ExcelReadSession } from '$lib/domain/excel/index.js';

  interface Props {
    session: ExcelReadSession;
    duplicateSoftwareIds: Set<string>;
    rowIdentifier: (objectDataId: string, bacnetObjectId: string) => string;
  }

  let { session, duplicateSoftwareIds, rowIdentifier }: Props = $props();
</script>

{#if session.objectDataExcel.length === 0}
  <div class="rounded-md border border-dashed p-4 text-sm text-muted-foreground">
    Im Excel-Blatt wurden keine Objektdaten-Einträge gefunden.
  </div>
{:else}
  <div class="overflow-x-auto rounded-md border">
    <div class="min-w-[840px] text-xs">
      <div
        class="grid grid-cols-[32px_240px_1fr_96px_88px] border-b bg-muted/30 font-medium text-muted-foreground"
      >
        <div class="px-2 py-1"></div>
        <div class="px-2 py-1">Objektdaten-ID</div>
        <div class="px-2 py-1">Beschreibung</div>
        <div class="px-2 py-1">BACnet</div>
        <div class="px-2 py-1">Optional</div>
      </div>

      {#each session.objectDataExcel as objectData}
        <details class="group border-b last:border-b-0">
          <summary class="cursor-pointer list-none">
            <div class="grid grid-cols-[32px_240px_1fr_96px_88px] items-center hover:bg-muted/20">
              <div class="px-2 py-1.5 text-muted-foreground">
                <span class="group-open:hidden">▸</span><span class="hidden group-open:inline"
                  >▾</span
                >
              </div>
              <div class="truncate px-2 py-1.5 font-medium">{objectData.id}</div>
              <div class="truncate px-2 py-1.5 text-muted-foreground">
                {objectData.description || '-'}
              </div>
              <div class="px-2 py-1.5 text-muted-foreground">
                {objectData.bacnet_objects.length}
              </div>
              <div class="px-2 py-1.5 text-muted-foreground">
                {objectData.is_optional_anchor ? 'Ja' : 'Nein'}
              </div>
            </div>
          </summary>

          <div class="border-t bg-muted/10 px-2 py-2">
            {#if objectData.bacnet_objects.length === 0}
              <p class="px-2 py-1 text-muted-foreground">
                Keine BACnet-Objekte für diesen Eintrag vorhanden.
              </p>
            {:else}
              <div class="overflow-x-auto rounded-sm border bg-background">
                <div class="min-w-245 text-[11px]">
                  <div
                    class="grid grid-cols-[180px_220px_70px_70px_180px_80px_90px_110px_140px_150px_150px_140px] border-b bg-muted/30 font-medium text-muted-foreground"
                  >
                    <div class="px-1 py-1">Text fix</div>
                    <div class="px-1 py-1">Beschreibung</div>
                    <div class="px-1 py-1">Sichtbar</div>
                    <div class="px-1 py-1">Optional</div>
                    <div class="px-1 py-1">Text individuell</div>
                    <div class="px-1 py-1">Typ</div>
                    <div class="px-1 py-1">Nummer</div>
                    <div class="px-1 py-1">Hardware</div>
                    <div class="px-1 py-1">Software-Ref.</div>
                    <div class="px-1 py-1">Statustext</div>
                    <div class="px-1 py-1">Benachrichtigungsklasse</div>
                    <div class="px-1 py-1">Alarmdefinition</div>
                    <div class="px-1 py-1">Apparat</div>
                  </div>

                  {#each objectData.bacnet_objects as bacnetObject}
                    <div
                      class={`grid grid-cols-[180px_220px_70px_70px_180px_80px_90px_110px_140px_150px_150px_140px] border-b last:border-b-0 ${duplicateSoftwareIds.has(rowIdentifier(objectData.id, bacnetObject.id)) ? 'bg-destructive/10' : ''}`}
                    >
                      <div class="truncate px-1 py-1">{bacnetObject.text_fix || '-'}</div>
                      <div class="truncate px-1 py-1 text-muted-foreground">
                        {bacnetObject.description || '-'}
                      </div>
                      <div class="px-1 py-1">{bacnetObject.gms_visible ? 'Ja' : 'Nein'}</div>
                      <div class="px-1 py-1">{bacnetObject.is_optional ? 'Ja' : 'Nein'}</div>
                      <div class="truncate px-1 py-1">{bacnetObject.text_individual || '-'}</div>
                      <div class="px-1 py-1">{bacnetObject.software_type || '-'}</div>
                      <div class="px-1 py-1">{bacnetObject.software_number || '-'}</div>
                      <div class="px-1 py-1">{bacnetObject.hardware_label || '-'}</div>
                      <div class="truncate px-1 py-1">
                        {bacnetObject.software_reference_label || '-'}
                      </div>
                      <div class="truncate px-1 py-1">{bacnetObject.state_text_label || '-'}</div>
                      <div class="truncate px-1 py-1">
                        {bacnetObject.notification_class_label || '-'}
                      </div>
                      <div class="truncate px-1 py-1">
                        {bacnetObject.alarm_definition_label || '-'}
                      </div>
                      <div class="truncate px-1 py-1">{bacnetObject.apparat_label || '-'}</div>
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
