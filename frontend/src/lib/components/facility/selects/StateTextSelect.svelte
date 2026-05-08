<script lang="ts">
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import type { StateText } from '$lib/domain/facility/index.js';
  import { createCachedFetchById } from '$lib/infrastructure/api/createCachedFetchById.js';
  import { stateTextRepository } from '$lib/infrastructure/api/stateTextRepository.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  type Props = {
    value?: string;
    id?: string;
    width?: string;
    popupWidth?: string;
    disabled?: boolean;
    clearable?: boolean;
    onValueChange?: (value: string) => void;
  };

  let {
    value = $bindable(''),
    id,
    width = 'w-[220px]',
    popupWidth = 'w-[320px]',
    disabled = false,
    clearable = true,
    onValueChange
  }: Props = $props();

  const t = createTranslator();
  const fetchStateTextByIdCached = createCachedFetchById('state-text-select', (id) =>
    stateTextRepository.get(id)
  );

  async function fetcher(search: string): Promise<StateText[]> {
    const res = await stateTextRepository.list({
      pagination: { page: 1, pageSize: 20 },
      search: { text: search }
    });
    return res.items;
  }

  async function fetchById(id: string): Promise<StateText | null | undefined> {
    return fetchStateTextByIdCached(id);
  }

  function formatLabel(item: StateText): string {
    const preview = stateTextValues(item).slice(0, 2).join(', ');
    return preview ? `${item.ref_number} - ${preview}` : String(item.ref_number);
  }

  function formatTitle(item: StateText): string {
    const lines = [`#${item.ref_number}`];
    for (const [index, value] of stateTextValues(item).entries()) {
      lines.push(`${index + 1}. ${value}`);
    }
    return lines.join('\n');
  }

  function stateTextValues(item: StateText): string[] {
    const values: string[] = [];
    for (let index = 1; index <= 16; index++) {
      const key = `state_text${index}` as keyof StateText;
      const value = item[key];
      if (typeof value === 'string' && value.trim()) {
        values.push(value.trim());
      }
    }
    return values;
  }
</script>

<AsyncCombobox
  bind:value
  {id}
  {fetcher}
  {fetchById}
  labelKey="ref_number"
  labelFormatter={formatLabel}
  itemTitleFormatter={formatTitle}
  selectedTitleFormatter={formatTitle}
  placeholder={$t('facility.selects.state_text')}
  searchPlaceholder={$t('common.search')}
  emptyText={$t('facility.selects.state_text_empty')}
  clearText={$t('facility.selects.clear')}
  {width}
  {popupWidth}
  {disabled}
  {clearable}
  {onValueChange}
/>
