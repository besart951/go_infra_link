<script lang="ts">
  type NavItem = {
    title: string;
    url: string;
    items?: Array<{ title: string; url: string }>;
  };

  let { items = [] } = $props<{ items?: NavItem[] }>();
</script>

<nav data-testid="nav-main">
  {#each items as item}
    <section data-testid={`nav-group:${item.url}`}>
      <div>{item.title}</div>
      {#if item.items && item.items.length > 0}
        {#each item.items as child}
          <a href={child.url} data-testid={`nav-link:${child.url}`}>{child.title}</a>
        {/each}
      {:else}
        <a href={item.url} data-testid={`nav-link:${item.url}`}>{item.title}</a>
      {/if}
    </section>
  {/each}
</nav>
