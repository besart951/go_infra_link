<script lang="ts">
  import ModuleCardGrid, {
    type ModuleCardItem
  } from '$lib/components/navigation/ModuleCardGrid.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import AlarmClockIcon from '@lucide/svelte/icons/alarm-clock';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import BellRingIcon from '@lucide/svelte/icons/bell-ring';
  import BoxesIcon from '@lucide/svelte/icons/boxes';
  import Building2Icon from '@lucide/svelte/icons/building-2';
  import ClipboardListIcon from '@lucide/svelte/icons/clipboard-list';
  import CpuIcon from '@lucide/svelte/icons/cpu';
  import DatabaseIcon from '@lucide/svelte/icons/database';
  import FileSlidersIcon from '@lucide/svelte/icons/file-sliders';
  import HardDriveIcon from '@lucide/svelte/icons/hard-drive';
  import LayersIcon from '@lucide/svelte/icons/layers';
  import ListTreeIcon from '@lucide/svelte/icons/list-tree';
  import PanelTopIcon from '@lucide/svelte/icons/panel-top';

  const t = createTranslator();

  const facilityCards = $derived.by<ModuleCardItem[]>(() =>
    [
      {
        title: $t('facility.buildings'),
        description: $t('facility.buildings_desc'),
        href: '/facility/buildings',
        icon: Building2Icon,
        tone: 'facility',
        hasAccess: canPerform('read', 'building')
      },
      {
        title: $t('facility.control_cabinets'),
        description: $t('facility.control_cabinets_desc'),
        href: '/facility/control-cabinets',
        icon: PanelTopIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'controlcabinet')
      },
      {
        title: $t('facility.sps_controllers'),
        description: $t('facility.sps_controllers_desc'),
        href: '/facility/sps-controllers',
        icon: CpuIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'spscontroller')
      },
      {
        title: $t('facility.field_devices'),
        description: $t('facility.field_devices_desc'),
        href: '/facility/field-devices',
        icon: HardDriveIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'fielddevice')
      },
      {
        title: $t('facility.system_types'),
        description: $t('facility.system_types_desc'),
        href: '/facility/system-types',
        icon: LayersIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'systemtype')
      },
      {
        title: $t('facility.system_parts'),
        description: $t('facility.system_parts_desc'),
        href: '/facility/system-parts',
        icon: ListTreeIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'systempart')
      },
      {
        title: $t('facility.apparats'),
        description: $t('facility.apparats_desc'),
        href: '/facility/apparats',
        icon: BoxesIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'apparat')
      },
      {
        title: $t('facility.object_data'),
        description: $t('facility.object_data_desc'),
        href: '/facility/object-data',
        icon: DatabaseIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'objectdata')
      },
      {
        title: $t('facility.state_texts'),
        description: $t('facility.state_texts_desc'),
        href: '/facility/state-texts',
        icon: ClipboardListIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'statetext')
      },
      {
        title: $t('facility.alarm_definitions'),
        description: $t('facility.alarm_definitions_desc'),
        href: '/facility/alarm-definitions',
        icon: AlarmClockIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'alarmdefinition')
      },
      {
        title: $t('navigation.alarm_catalog'),
        description: $t('hub.facility.alarm_catalog_desc'),
        href: '/facility/alarm-catalog',
        icon: BellRingIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'alarmtype')
      },
      {
        title: $t('facility.notification_classes'),
        description: $t('facility.notification_classes_desc'),
        href: '/facility/notification-classes',
        icon: BellRingIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'notificationclass')
      },
      {
        title: $t('facility.specifications'),
        description: $t('facility.specifications_desc'),
        href: '/facility/specifications',
        icon: FileSlidersIcon,
        tone: 'facility',
        hasAccess: canPerform('read', 'specification') || canPerform('read', 'fielddevice')
      }
    ].filter((item) => item.hasAccess)
  );
</script>

<svelte:head>
  <title>{$t('facility.facility_overview')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <header class="flex flex-col gap-4 border-b pb-5 sm:flex-row sm:items-end sm:justify-between">
    <div class="min-w-0 space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight sm:text-3xl">
        {$t('facility.facility_overview')}
      </h1>
      <p class="max-w-3xl text-sm leading-6 text-muted-foreground">
        {$t('hub.facility.description')}
      </p>
    </div>
    <Button variant="outline" href="/" class="w-full sm:w-auto">
      <ArrowLeftIcon class="size-4" />
      {$t('hub.back_to_dashboard')}
    </Button>
  </header>

  <ModuleCardGrid
    items={facilityCards}
    columns="lg:grid-cols-3 2xl:grid-cols-4"
    emptyMessage={$t('hub.no_access')}
  />
</div>
