<script lang="ts">
  import { page } from '$app/stores';
  import FacilityHierarchyDetailPage from '$lib/components/facility/detail/FacilityHierarchyDetailPage.svelte';
</script>

<FacilityHierarchyDetailPage
  kind="buildings"
  id={$page.params.buildingId ?? ''}
  projectId={$page.params.id ?? ''}
/>
