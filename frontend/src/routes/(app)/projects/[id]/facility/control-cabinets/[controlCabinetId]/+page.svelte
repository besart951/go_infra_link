<script lang="ts">
  import { page } from '$app/stores';
  import FacilityHierarchyDetailPage from '$lib/components/facility/detail/FacilityHierarchyDetailPage.svelte';
</script>

<FacilityHierarchyDetailPage
  kind="control-cabinets"
  id={$page.params.controlCabinetId ?? ''}
  projectId={$page.params.id ?? ''}
/>
