<script lang="ts">
  import { goto } from '$app/navigation';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import PhaseForm from '$lib/components/project/PhaseForm.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';

  const t = createTranslator();

  function handleSuccess() {
    goto('/projects/phases');
  }

  function handleCancel() {
    goto('/projects/phases');
  }
</script>

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={$t('phases.new.title')}
    description={$t('phases.new.description')}
    infoLabel={$t('common.info')}
    backHref="/projects/phases"
    backLabel={$t('common.back')}
  />

  <PhaseForm onSuccess={handleSuccess} onCancel={handleCancel} />
</div>
