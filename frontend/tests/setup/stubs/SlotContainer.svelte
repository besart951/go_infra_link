<script lang="ts">
  let {
    as = 'div',
    children,
    ...restProps
  } = $props<{ as?: string; children?: () => unknown } & Record<string, unknown>>();
</script>

<svelte:element this={as} {...restProps}>
  {@render children?.()}
</svelte:element>
