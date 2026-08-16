<script lang="ts">
  import NavMain from '$lib/components/sidebar/nav-main.svelte';
  import * as Sidebar from '$lib/components/ui/sidebar/index.js';
  import FolderIcon from '@lucide/svelte/icons/folder';

  const items = [
    {
      title: 'Facility',
      url: '/facility',
      icon: FolderIcon,
      items: [
        {
          title: 'Buildings',
          url: '/facility/buildings',
          icon: FolderIcon
        }
      ]
    },
    {
      title: 'Projects',
      url: '/projects',
      icon: FolderIcon,
      items: [
        {
          title: 'Project list',
          url: '/projects/list',
          icon: FolderIcon
        }
      ]
    }
  ];
</script>

<Sidebar.Provider open={false}>
  <NavMain {items} />
</Sidebar.Provider>
