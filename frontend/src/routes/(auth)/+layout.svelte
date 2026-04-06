<script lang="ts">
  import AnimatedAirplane from '$lib/components/ui/animated-airplane/AnimatedAirplane.svelte';
  const { children } = $props();
</script>

<div class="auth-wrapper">
  <!-- Animated airplane background layer -->
  <div class="airplane-bg" aria-hidden="true">
    <AnimatedAirplane />
  </div>

  <!-- Brand logo -->
  <div class="absolute top-4 left-4 z-20 p-1">
    <a href="/" aria-label="CodeLinks Home" class="inline-flex items-center">
      <img src="/favicon.svg" alt="CodeLinks" class="h-8 w-auto dark:hidden" />
      <img src="/codelinks_logo_dark.svg" alt="CodeLinks" class="hidden h-8 w-auto dark:block" />
    </a>
  </div>

  <!-- Login form layer -->
  <div class="relative z-10 flex min-h-svh items-center justify-center p-6">
    <div
      class="w-full max-w-sm rounded-2xl border border-border/40 bg-background/85 p-8 shadow-xl backdrop-blur-md"
    >
      {@render children?.()}
    </div>
  </div>
</div>

<style>
  .auth-wrapper {
    position: relative;
    min-height: 100svh;
    overflow: hidden;
    background: var(--background);
  }

  .airplane-bg {
    position: absolute;
    inset: 0;
    z-index: 0;
    pointer-events: none;
  }
</style>
