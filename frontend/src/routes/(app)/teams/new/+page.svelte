<script lang="ts">
  import { goto } from '$app/navigation';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import { TeamListPageState } from '$lib/components/teams/TeamListPageState.svelte.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  const t = createTranslator();
  const state = new TeamListPageState();

  function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    void state.submitCreate();
  }
</script>

<svelte:head>
  <title>{$t('pages.create_team')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={$t('pages.create_team')}
    description={$t('pages.teams_desc')}
    backHref="/teams"
    backLabel={$t('common.back')}
  />

  <form class="max-w-3xl rounded-lg border bg-background p-6 shadow-xs" onsubmit={handleSubmit}>
    <div class="grid gap-5">
      <div class="grid gap-2">
        <label class="text-sm font-medium" for="team_name">{$t('common.name')}</label>
        <Input
          id="team_name"
          placeholder={$t('messages.team_name_placeholder')}
          bind:value={state.form.name}
          disabled={state.createBusy}
        />
      </div>

      <div class="grid gap-2">
        <label class="text-sm font-medium" for="team_desc">{$t('messages.team_description')}</label>
        <Textarea
          id="team_desc"
          placeholder={$t('pages.optional')}
          bind:value={state.form.description}
          disabled={state.createBusy}
          rows={4}
        />
      </div>
    </div>

    <div class="mt-6 flex justify-end gap-2">
      <Button variant="outline" onclick={() => goto('/teams')} disabled={state.createBusy}>
        {$t('common.cancel')}
      </Button>
      <Button type="submit" disabled={!state.canSubmitCreate()}>
        {$t('common.create')}
      </Button>
    </div>
  </form>
</div>
