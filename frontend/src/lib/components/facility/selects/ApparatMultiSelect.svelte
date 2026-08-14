<script lang="ts">
  import { onMount } from 'svelte';
  import AsyncMultiSelect from '$lib/components/ui/combobox/AsyncMultiSelect.svelte';
  import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
  import type { Apparat } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { facilityReferenceDataCache } from '$lib/services/facilityReferenceDataCache.js';
  import { matchesApparatSearch } from './facilityReferenceDataSearch.js';

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

  type ApparatOption = Apparat & { label: string };

  function toOption(apparat: Apparat): ApparatOption {
    return {
      ...apparat,
      label: `${apparat.short_name} - ${apparat.name}`
    };
  }

  async function fetcher(search: string): Promise<ApparatOption[]> {
    try {
      const { apparats } = await facilityReferenceDataCache.load();
      return apparats
        .filter((apparat) => matchesApparatSearch(apparat, search))
        .slice(0, maxOptions)
        .map(toOption);
    } catch {
      const res = await apparatRepository.list({
        pagination: { page: 1, pageSize: maxOptions },
        search: { text: search }
      });
      return res.items.map(toOption);
    }
  }

  async function fetchByIds(ids: string[]): Promise<ApparatOption[]> {
    if (ids.length === 0) return [];

    try {
      const { apparats } = await facilityReferenceDataCache.load();
      const apparatsByID = new Map(apparats.map((apparat) => [apparat.id, apparat]));
      const selectedApparats = ids
        .map((id) => apparatsByID.get(id))
        .filter((apparat): apparat is Apparat => Boolean(apparat));

      if (selectedApparats.length === ids.length) {
        return selectedApparats.map(toOption);
      }
    } catch {
      // Preserve the existing repository fallback when the reference cache cannot be loaded.
    }

    const items = await apparatRepository.getBulk(ids);
    return items.map(toOption);
  }
</script>

<AsyncMultiSelect
  bind:value
  {fetcher}
  {fetchByIds}
  labelKey="label"
  placeholder={$t('facility.multi_selects.apparats_placeholder')}
  searchPlaceholder={$t('facility.multi_selects.apparats_search')}
  emptyText={$t('facility.multi_selects.apparats_empty')}
  refreshKey={referenceDataRefreshKey}
  {width}
  {disabled}
  {id}
/>
