<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import ProjectPhaseSelect from '$lib/components/project/ProjectPhaseSelect.svelte';
  import { ProjectSettingsPageState } from '$lib/components/project/settings/ProjectSettingsPageState.svelte.js';
  import { ArrowLeft, Pencil } from '@lucide/svelte';
  import ObjectDataForm from '$lib/components/facility/forms/ObjectDataForm.svelte';

  const t = createTranslator();

  const projectId = $derived($page.params.id ?? '');
  const state = new ProjectSettingsPageState(() => projectId);

  $effect(() => {
    state.ensureActiveTabLoaded();
  });

  onMount(() => {
    void state.load();
  });
</script>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div class="flex items-start gap-3">
    <Button variant="outline" onclick={() => goto(`/projects/${projectId}`)}>
      <ArrowLeft class="mr-2 h-4 w-4" />
      {$t('common.back')}
    </Button>
    <div>
      <h1 class="text-3xl font-bold tracking-tight">{$t('projects.settings.title')}</h1>
      <p class="mt-1 text-muted-foreground">{$t('projects.settings.description')}</p>
    </div>
  </div>

  {#if state.error}
    <div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
      <p class="font-medium">{$t('projects.errors.load_title')}</p>
      <p class="text-sm">{state.error}</p>
    </div>
  {/if}

  <div class="rounded-lg border bg-background">
    <div class="flex flex-wrap gap-2 border-b px-6 py-3">
      <Button
        variant={state.activeTab === 'settings' ? 'default' : 'ghost'}
        onclick={() => (state.activeTab = 'settings')}
      >
        {$t('projects.settings.tabs.settings')}
      </Button>
      <Button
        variant={state.activeTab === 'users' ? 'default' : 'ghost'}
        onclick={() => (state.activeTab = 'users')}
      >
        {$t('projects.settings.tabs.users')}
      </Button>
      <Button
        variant={state.activeTab === 'object-data' ? 'default' : 'ghost'}
        onclick={() => (state.activeTab = 'object-data')}
      >
        {$t('projects.settings.tabs.object_data')}
      </Button>
    </div>

    {#if state.loading}
      <div class="p-6">
        <div class="grid gap-4 md:grid-cols-2">
          {#each Array(6) as _}
            <Skeleton class="h-6 w-full" />
          {/each}
        </div>
      </div>
    {:else if !state.project}
      <div class="p-6 text-sm text-muted-foreground">{$t('projects.errors.not_found')}</div>
    {:else if state.activeTab === 'settings'}
      <div class="p-6">
        <div class="grid gap-4 md:grid-cols-2">
          <div class="flex flex-col gap-2">
            <label class="text-sm font-medium" for="project_name">{$t('common.name')}</label>
            <Input
              id="project_name"
              bind:value={state.form.name}
              disabled={state.saving || !state.canUpdateProject}
            />
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-sm font-medium" for="project_status">{$t('common.status')}</label>
            <select
              id="project_status"
              class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
              bind:value={state.form.status}
              disabled={state.saving || !state.canUpdateProject}
            >
              {#each state.statusOptions as opt}
                <option value={opt.value}>{$t(opt.label)}</option>
              {/each}
            </select>
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-sm font-medium" for="project_start">
              {$t('projects.settings.start_date')}
            </label>
            <Input
              id="project_start"
              type="date"
              bind:value={state.form.start_date}
              disabled={state.saving || !state.canUpdateProject}
            />
          </div>

          <div class="flex flex-col gap-2">
            <label class="text-sm font-medium" for="project_phase_edit"
              >{$t('projects.settings.phase')}</label
            >
            <ProjectPhaseSelect
              id="project_phase_edit"
              bind:value={state.form.phase_id}
              width="w-full"
              disabled={state.saving || !state.canUpdateProject}
            />
          </div>

          <div class="flex flex-col gap-2 md:col-span-2">
            <label class="text-sm font-medium" for="project_desc">{$t('common.description')}</label>
            <Textarea
              id="project_desc"
              rows={4}
              bind:value={state.form.description}
              disabled={state.saving || !state.canUpdateProject}
            />
          </div>
        </div>

        <div class="mt-6 flex items-center justify-end gap-2">
          <Button
            variant="outline"
            onclick={() => state.resetForm()}
            disabled={state.saving || !state.canUpdateProject}
          >
            {$t('common.reset')}
          </Button>
          <Button
            onclick={() => state.saveSettings()}
            disabled={state.saving || !state.canUpdateProject}
          >
            {$t('projects.settings.save_changes')}
          </Button>
        </div>

        <div class="mt-8 grid gap-6 md:grid-cols-2">
          <div class="space-y-2">
            <div class="text-xs text-muted-foreground uppercase">{$t('common.created')}</div>
            <div class="text-sm font-medium">{state.formatDate(state.project.created_at)}</div>
          </div>
          <div class="space-y-2">
            <div class="text-xs text-muted-foreground uppercase">{$t('common.modified')}</div>
            <div class="text-sm font-medium">{state.formatDate(state.project.updated_at)}</div>
          </div>
        </div>
      </div>
    {:else if state.activeTab === 'users'}
      <div class="p-6">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div class="flex w-full max-w-sm flex-col gap-2">
            <label class="text-sm font-medium" for="project_user_search">
              {$t('projects.users.search_label')}
            </label>
            <Input
              id="project_user_search"
              placeholder={$t('projects.users.search_placeholder')}
              bind:value={state.userSearch}
              oninput={() => state.handleUserSearchInput()}
            />
          </div>
          <Button variant="outline" onclick={() => state.loadUsers()} disabled={state.usersLoading}>
            {$t('common.refresh')}
          </Button>
        </div>

        <div class="mt-6 rounded-lg border bg-background">
          <Table.Root>
            <Table.Header>
              <Table.Row>
                <Table.Head>{$t('common.name')}</Table.Head>
                <Table.Head>{$t('auth.email')}</Table.Head>
                <Table.Head>{$t('common.status')}</Table.Head>
                <Table.Head class="w-44"></Table.Head>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {#if state.usersLoading}
                {#each Array(5) as _}
                  <Table.Row>
                    <Table.Cell><Skeleton class="h-4 w-40" /></Table.Cell>
                    <Table.Cell><Skeleton class="h-4 w-60" /></Table.Cell>
                    <Table.Cell><Skeleton class="h-8 w-20" /></Table.Cell>
                  </Table.Row>
                {/each}
              {:else if state.availableUsers.length === 0}
                <Table.Row>
                  <Table.Cell colspan={4} class="h-20 text-center text-sm text-muted-foreground">
                    {$t('projects.users.empty')}
                  </Table.Cell>
                </Table.Row>
              {:else}
                {#each state.availableUsers as user (user.id)}
                  <Table.Row>
                    <Table.Cell class="font-medium">
                      {user.first_name}
                      {user.last_name}
                    </Table.Cell>
                    <Table.Cell class="text-muted-foreground">{user.email}</Table.Cell>
                    <Table.Cell>
                      {state.isUserInProject(user.id) ? $t('common.active') : $t('common.inactive')}
                    </Table.Cell>
                    <Table.Cell class="text-right">
                      {#if state.canUpdateProject}
                        <Button
                          variant={state.isUserInProject(user.id) ? 'outline' : 'default'}
                          onclick={() => state.toggleUser(user)}
                        >
                          {state.isUserInProject(user.id) ? $t('common.remove') : $t('common.add')}
                        </Button>
                      {/if}
                    </Table.Cell>
                  </Table.Row>
                {/each}
              {/if}
            </Table.Body>
          </Table.Root>
        </div>
      </div>
    {:else}
      <div class="p-6">
        {#if state.showObjectDataForm && state.canUpdateObjectData}
          <ObjectDataForm
            initialData={state.editingObjectData}
            onSuccess={() => state.handleObjectDataSuccess()}
            onCancel={() => state.handleObjectDataCancel()}
          />
        {/if}

        <div class="flex flex-wrap items-end justify-between gap-3">
          <div class="flex w-full max-w-sm flex-col gap-2">
            <label class="text-sm font-medium" for="project_object_data_search">
              {$t('projects.object_data.search_label')}
            </label>
            <Input
              id="project_object_data_search"
              placeholder={$t('projects.object_data.search_placeholder')}
              bind:value={state.objectDataSearch}
              oninput={() => state.handleObjectDataSearchInput()}
            />
          </div>
          <div class="flex items-end gap-3">
            <div class="flex flex-col gap-2">
              <label class="text-sm font-medium" for="project_object_data_status">
                {$t('common.status')}
              </label>
              <select
                id="project_object_data_status"
                class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
                bind:value={state.objectDataStatusFilter}
              >
                <option value="all">{$t('projects.object_data.status_all')}</option>
                <option value="active">{$t('common.active')}</option>
                <option value="inactive">{$t('common.inactive')}</option>
              </select>
            </div>
            <Button
              variant="outline"
              onclick={() => state.loadObjectData()}
              disabled={state.objectDataLoading}>{$t('common.refresh')}</Button
            >
          </div>
        </div>

        <div class="mt-6 rounded-lg border bg-background">
          <Table.Root>
            <Table.Header>
              <Table.Row>
                <Table.Head>{$t('common.description')}</Table.Head>
                <Table.Head>{$t('projects.object_data.version')}</Table.Head>
                <Table.Head>{$t('projects.object_data.project_status')}</Table.Head>
                <Table.Head class="w-44"></Table.Head>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {#if state.objectDataLoading}
                {#each Array(5) as _}
                  <Table.Row>
                    <Table.Cell><Skeleton class="h-4 w-60" /></Table.Cell>
                    <Table.Cell><Skeleton class="h-4 w-24" /></Table.Cell>
                    <Table.Cell><Skeleton class="h-4 w-16" /></Table.Cell>
                    <Table.Cell><Skeleton class="h-8 w-20" /></Table.Cell>
                  </Table.Row>
                {/each}
              {:else if state.getFilteredObjectData().length === 0}
                <Table.Row>
                  <Table.Cell colspan={4} class="h-20 text-center text-sm text-muted-foreground">
                    {$t('projects.object_data.empty')}
                  </Table.Cell>
                </Table.Row>
              {:else}
                {#each state.getFilteredObjectData() as obj (obj.id)}
                  <Table.Row>
                    <Table.Cell class="font-medium">{obj.description}</Table.Cell>
                    <Table.Cell class="text-muted-foreground">{obj.version}</Table.Cell>
                    <Table.Cell>
                      {state.isObjectDataActive(obj) ? $t('common.active') : $t('common.inactive')}
                    </Table.Cell>
                    <Table.Cell class="text-right">
                      <div class="flex items-center justify-end gap-2">
                        {#if state.canUpdateObjectData}
                          <Button variant="outline" onclick={() => state.editObjectData(obj)}>
                            <Pencil class="mr-2 h-4 w-4" />
                            {$t('common.edit')}
                          </Button>
                        {/if}
                        {#if state.canUpdateProject}
                          <Button
                            variant={state.isObjectDataActive(obj) ? 'outline' : 'default'}
                            onclick={() => state.toggleObjectData(obj)}
                          >
                            {state.isObjectDataActive(obj)
                              ? $t('projects.object_data.deactivate')
                              : $t('projects.object_data.activate')}
                          </Button>
                        {/if}
                      </div>
                    </Table.Cell>
                  </Table.Row>
                {/each}
              {/if}
            </Table.Body>
          </Table.Root>
        </div>
      </div>
    {/if}
  </div>
</div>
