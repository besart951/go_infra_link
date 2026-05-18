<script lang="ts">
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import { getPhase, listPhases } from '$lib/infrastructure/api/phase.adapter.js';
  import type { Phase } from '$lib/domain/phase/index.js';

  interface ProjectPhaseSelectProps {
    value?: string;
    width?: string;
    id?: string;
    disabled?: boolean;
    clearable?: boolean;
    clearText?: string;
    placeholder?: string;
    searchPlaceholder?: string;
    emptyText?: string;
    onValueChange?: (value: string) => void;
  }

  let {
    value = $bindable(''),
    width = 'w-[260px]',
    id,
    disabled = false,
    clearable = false,
    clearText = 'Auswahl aufheben',
    placeholder = 'Select phase...',
    searchPlaceholder = 'Search phases...',
    emptyText = 'No phases found.',
    onValueChange
  }: ProjectPhaseSelectProps = $props();

  const MAX_PHASE_SAMPLES = 100;

  async function fetcher(search: string): Promise<Phase[]> {
    const res = await listPhases({ page: 1, limit: MAX_PHASE_SAMPLES, search });
    return (res.items ?? []).map((phase) => ({
      ...phase,
      name: phase.name || phase.id
    }));
  }

  async function fetchById(id: string): Promise<Phase | null> {
    const phase = await getPhase(id);
    return { ...phase, name: phase.name || phase.id };
  }
</script>

<AsyncCombobox
  bind:value
  {fetcher}
  {fetchById}
  labelKey="name"
  {placeholder}
  {searchPlaceholder}
  {emptyText}
  {clearable}
  {clearText}
  {onValueChange}
  {width}
  {id}
  {disabled}
/>
