<script lang="ts">
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import { getProject, listProjects } from '$lib/infrastructure/api/project.adapter.js';
  import type { Project } from '$lib/domain/project/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  type Props = {
    value?: string;
    width?: string;
    disabled?: boolean;
    refreshKey?: string | number;
    onValueChange?: (value: string) => void;
  };

  let {
    value = $bindable(''),
    width = 'w-[250px]',
    disabled = false,
    refreshKey,
    onValueChange
  }: Props = $props();

  const t = createTranslator();

  async function fetcher(search: string): Promise<Project[]> {
    const res = await listProjects({ search, limit: 20 });
    return res.items || [];
  }

  async function fetchById(id: string): Promise<Project> {
    return getProject(id);
  }
</script>

<AsyncCombobox
  bind:value
  {fetcher}
  {fetchById}
  labelKey="name"
  placeholder={$t('project.select_placeholder')}
  {disabled}
  {refreshKey}
  {width}
  {onValueChange}
/>
