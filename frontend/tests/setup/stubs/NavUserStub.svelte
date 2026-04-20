<script lang="ts">
  let { user } = $props<{ user?: { email?: string } }>();
</script>

<div data-testid="nav-user">{user?.email ?? ''}</div>
