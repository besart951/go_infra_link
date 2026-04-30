<script lang="ts">
  import type { AccountTab, AccountTabItem } from './types.js';

  interface Props {
    activeTab: AccountTab;
    tabs: AccountTabItem[];
  }

  let { activeTab = $bindable(), tabs }: Props = $props();
</script>

<aside>
  <div class="rounded-lg border bg-card p-2">
    <nav class="flex flex-col gap-1">
      {#each tabs as tab (tab.value)}
        <button
          type="button"
          class={activeTab === tab.value
            ? 'rounded-md bg-muted px-3 py-2 text-left text-sm font-medium'
            : 'rounded-md px-3 py-2 text-left text-sm text-muted-foreground hover:bg-muted/60'}
          onclick={() => (activeTab = tab.value)}
        >
          {tab.label}
        </button>
      {/each}
    </nav>
  </div>
</aside>
