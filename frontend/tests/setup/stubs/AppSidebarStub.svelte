<script lang="ts">
  let { user } = $props<{ user?: { email?: string } }>();
</script>

<div data-testid="app-sidebar-stub">{user?.email ?? ''}</div>
