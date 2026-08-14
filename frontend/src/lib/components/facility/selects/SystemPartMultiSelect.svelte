<script lang="ts">
  import { onMount } from 'svelte';
  import AsyncMultiSelect from '$lib/components/ui/combobox/AsyncMultiSelect.svelte';
  import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
  import type { SystemPart } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { facilityReferenceDataCache } from '$lib/services/facilityReferenceDataCache.js';
  import { matchesSystemPartSearch } from './facilityReferenceDataSearch.js';

  type Props = {
    value?: string[];
    width?: string;
    disabled?: boolean;
    id?: string;
  };

  let { value = $bindable([]), width = 'w-full', disabled = false, id }: Props = $props();

  const t = createTranslator();
  const maxOptions = 50;
  let referenceDataRefreshKey = $state(0);

  onMount(() =>
    facilityReferenceDataCache.subscribe(() => {
      referenceDataRefreshKey += 1;
    })
  );

  async function fetcher(search: string): Promise<SystemPart[]> {
    try {
      const { systemParts } = await facilityReferenceDataCache.load();
      return systemParts
        .filter((systemPart) => matchesSystemPartSearch(systemPart, search))
        .slice(0, maxOptions);
    } catch {
      const res = await systemPartRepository.list({
        pagination: { page: 1, pageSize: maxOptions },
        search: { text: search }
      });
      return res.items;
    }
  }

  async function fetchByIds(ids: string[]): Promise<SystemPart[]> {
    if (ids.length === 0) return [];

    try {
      const { systemParts } = await facilityReferenceDataCache.load();
      const systemPartsByID = new Map(systemParts.map((systemPart) => [systemPart.id, systemPart]));
      const selectedSystemParts = ids
        .map((id) => systemPartsByID.get(id))
        .filter((systemPart): systemPart is SystemPart => Boolean(systemPart));

      if (selectedSystemParts.length === ids.length) {
        return selectedSystemParts;
      }
    } catch {
      // Preserve the existing repository fallback when the reference cache cannot be loaded.
    }

    const results = await Promise.allSettled(ids.map((id) => systemPartRepository.get(id)));
    return results
      .filter(
        (result): result is PromiseFulfilledResult<SystemPart> => result.status === 'fulfilled'
      )
      .map((result) => result.value);
  }
</script>

<AsyncMultiSelect
  bind:value
  {fetcher}
  {fetchByIds}
  labelKey="name"
  placeholder={$t('facility.multi_selects.system_parts_placeholder')}
  searchPlaceholder={$t('facility.multi_selects.system_parts_search')}
  emptyText={$t('facility.multi_selects.system_parts_empty')}
  refreshKey={referenceDataRefreshKey}
  {width}
  {disabled}
  {id}
/>
