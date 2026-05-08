<script lang="ts">
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import type { NotificationClass } from '$lib/domain/facility/index.js';
  import { createCachedFetchById } from '$lib/infrastructure/api/createCachedFetchById.js';
  import { notificationClassRepository } from '$lib/infrastructure/api/notificationClassRepository.js';
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
    width = 'w-[300px]',
    popupWidth = 'w-[360px]',
    disabled = false,
    clearable = true,
    onValueChange
  }: Props = $props();

  const t = createTranslator();
  const fetchNotificationClassByIdCached = createCachedFetchById(
    'notification-class-select',
    (id) => notificationClassRepository.get(id)
  );

  async function fetcher(search: string): Promise<NotificationClass[]> {
    const res = await notificationClassRepository.list({
      pagination: { page: 1, pageSize: 20 },
      search: { text: search }
    });
    return res.items;
  }

  async function fetchById(id: string): Promise<NotificationClass | null | undefined> {
    return fetchNotificationClassByIdCached(id);
  }

  function formatLabel(item: NotificationClass): string {
    return `NC ${item.nc} - ${item.object_description}`;
  }

  function formatTitle(item: NotificationClass): string {
    return [`NC ${item.nc}`, item.object_description, item.internal_description, item.meaning]
      .filter((value) => value?.trim())
      .join('\n');
  }
</script>

<AsyncCombobox
  bind:value
  {id}
  {fetcher}
  {fetchById}
  labelKey="nc"
  labelFormatter={formatLabel}
  itemTitleFormatter={formatTitle}
  selectedTitleFormatter={formatTitle}
  placeholder={$t('facility.selects.notification_class')}
  searchPlaceholder={$t('common.search')}
  emptyText={$t('facility.selects.notification_class_empty')}
  clearText={$t('facility.selects.clear')}
  {width}
  {popupWidth}
  {disabled}
  {clearable}
  {onValueChange}
/>
