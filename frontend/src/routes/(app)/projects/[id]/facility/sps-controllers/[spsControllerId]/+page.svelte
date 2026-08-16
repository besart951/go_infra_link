<script lang="ts">
  import { page } from '$app/stores';
  import FacilityHierarchyDetailPage from '$lib/components/facility/detail/FacilityHierarchyDetailPage.svelte';
</script>

<FacilityHierarchyDetailPage
  kind="sps-controllers"
  id={$page.params.spsControllerId ?? ''}
  projectId={$page.params.id ?? ''}
/>
