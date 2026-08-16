<script lang="ts">
  import { page } from '$app/stores';
  import FacilityHierarchyDetailPage from '$lib/components/facility/detail/FacilityHierarchyDetailPage.svelte';
</script>

<FacilityHierarchyDetailPage kind="sps-controller-system-types" id={$page.params.id ?? ''} />
