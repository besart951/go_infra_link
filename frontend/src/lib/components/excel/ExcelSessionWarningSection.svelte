<script lang="ts">
  interface CreateExecutionReport {
    total: number;
    success: number;
    failed: Array<{ objectDataId: string; reason: string }>;
    unresolvedSoftwareLinks: Array<{ objectDataId: string; from: string; to: string }>;
  }

  interface Props {
    prepareError: string | null;
    createError: string | null;
    createReport: CreateExecutionReport | null;
  }

  let { prepareError, createError, createReport }: Props = $props();
</script>

{#if prepareError}
  <div
    class="mb-4 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-xs text-destructive"
  >
    {prepareError}
  </div>
{/if}
{#if createError}
  <div
    class="mb-4 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-xs text-destructive"
  >
    {createError}
  </div>
{/if}
{#if createReport}
  <div class="mb-4 rounded-md border bg-muted/20 p-3 text-xs text-muted-foreground">
    <span>Erstellt: {createReport.success}/{createReport.total}</span>
    <span class="ml-3">Fehlgeschlagen: {createReport.failed.length}</span>
    <span class="ml-3"
      >Nicht aufgelöste Software-Verknüpfungen: {createReport.unresolvedSoftwareLinks.length}</span
    >
    {#if createReport.failed.length > 0}
      <div class="mt-2">
        <strong class="text-foreground">Fehlgeschlagene Objektdaten:</strong>
        <p>
          {createReport.failed.map((item) => `${item.objectDataId} (${item.reason})`).join(' | ')}
        </p>
      </div>
    {/if}
    {#if createReport.unresolvedSoftwareLinks.length > 0}
      <div class="mt-2">
        <strong class="text-foreground">Nicht aufgelöste Software-Verknüpfungen:</strong>
        <p>
          {createReport.unresolvedSoftwareLinks
            .map((item) => `${item.objectDataId}: ${item.from} -> ${item.to}`)
            .join(' | ')}
        </p>
      </div>
    {/if}
  </div>
{/if}
