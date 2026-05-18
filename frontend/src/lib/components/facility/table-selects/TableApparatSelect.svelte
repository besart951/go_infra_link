<script lang="ts">
  import StaticCombobox from '$lib/components/ui/combobox/StaticCombobox.svelte';
  import type { Apparat } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { formatRelationSelectLabel } from './relationSelectOptions.js';

  interface Props {
    items: Apparat[];
    value?: string;
    width?: string;
    popupWidth?: string;
    onValueChange?: (value: string) => void;
    disabled?: boolean;
    error?: string;
    clearable?: boolean;
    clearText?: string;
  }

  let {
    items,
    value = $bindable(''),
    width = 'w-full',
    popupWidth = width,
    onValueChange,
    disabled = false,
    error,
    clearable = false,
    clearText
  }: Props = $props();

  const t = createTranslator();

  const formattedItems = $derived(
    items.map((item) => {
      const label = formatRelationSelectLabel(item);
      return {
        ...item,
        display_name: label,
        option_name: label,
        tooltip_name: label
      };
    })
  );
</script>

<StaticCombobox
  items={formattedItems}
  bind:value
  labelKey="display_name"
  optionLabelKey="option_name"
  tooltipLabelKey="tooltip_name"
  placeholder={$t('field_device.table_select.apparat')}
  {clearable}
  clearText={clearText ?? $t('field_device.table_select.clear_apparat')}
  {width}
  {popupWidth}
  {onValueChange}
  {disabled}
  {error}
/>
