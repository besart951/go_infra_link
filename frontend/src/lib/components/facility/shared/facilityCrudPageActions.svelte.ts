import { ManageBuildingUseCase } from '$lib/application/useCases/facility/manageBuildingUseCase.js';
import { ManageObjectDataUseCase } from '$lib/application/useCases/facility/manageObjectDataUseCase.js';
import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';
import { addToast } from '$lib/components/toast.svelte';
import type {
  AlarmDefinition,
  Apparat,
  Building,
  NotificationClass,
  ObjectData,
  StateText,
  SystemPart,
  SystemType
} from '$lib/domain/facility/index.js';
import { alarmDefinitionRepository } from '$lib/infrastructure/api/alarmDefinitionRepository.js';
import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
import { notificationClassRepository } from '$lib/infrastructure/api/notificationClassRepository.js';
import { objectDataRepository } from '$lib/infrastructure/api/objectDataRepository.js';
import { stateTextRepository } from '$lib/infrastructure/api/stateTextRepository.js';
import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
import { systemTypeRepository } from '$lib/infrastructure/api/systemTypeRepository.js';
import { t as translate } from '$lib/i18n/index.js';
import {
  alarmDefinitionsStore,
  apparatsStore,
  buildingsStore,
  notificationClassesStore,
  objectDataStore,
  stateTextsStore,
  systemPartsStore,
  systemTypesStore
} from '$lib/stores/list/entityStores.js';
import { confirm } from '$lib/stores/confirm-dialog.js';
import { CrudPageActions } from './crudPageActions.svelte.js';

interface StandardFacilityCrudActionsOptions<TItem> {
  reload: () => void | Promise<void>;
  deleteItem: (item: TItem) => Promise<void>;
  getDeleteMessage: (item: TItem) => string;
  getDeleteSuccessMessage: () => string;
  getDeleteFailureMessage: () => string;
  getDeleteTitle?: () => string;
}

function createStandardFacilityCrudActions<TItem>(
  options: StandardFacilityCrudActionsOptions<TItem>
): CrudPageActions<TItem> {
  return new CrudPageActions<TItem>({
    reload: options.reload,
    deleteItem: options.deleteItem,
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: options.getDeleteTitle ?? (() => translate('common.delete')),
    getDeleteMessage: options.getDeleteMessage,
    getDeleteConfirmText: () => translate('common.delete'),
    getDeleteCancelText: () => translate('common.cancel'),
    getDeleteSuccessMessage: options.getDeleteSuccessMessage,
    getDeleteFailureMessage: options.getDeleteFailureMessage
  });
}

export function createBuildingActions(): CrudPageActions<Building> {
  const manageBuilding = new ManageBuildingUseCase(buildingRepository);
  return createStandardFacilityCrudActions<Building>({
    reload: () => buildingsStore.reload(),
    deleteItem: (building) => manageBuilding.delete(building.id),
    getDeleteMessage: (building) =>
      translate('facility.delete_building_confirm').replace('{name}', building.iws_code),
    getDeleteSuccessMessage: () => translate('facility.building_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_building_failed')
  });
}

export function createSystemTypeActions(): CrudPageActions<SystemType> {
  const manageSystemType = new ManageEntityUseCase(systemTypeRepository);
  return createStandardFacilityCrudActions<SystemType>({
    reload: () => systemTypesStore.reload(),
    deleteItem: (item) => manageSystemType.delete(item.id),
    getDeleteMessage: (item) =>
      translate('facility.delete_system_type_confirm').replace('{name}', item.name),
    getDeleteSuccessMessage: () => translate('facility.system_type_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_system_type_failed')
  });
}

export function createSystemPartActions(): CrudPageActions<SystemPart> {
  const manageSystemPart = new ManageEntityUseCase(systemPartRepository);
  return createStandardFacilityCrudActions<SystemPart>({
    reload: () => systemPartsStore.reload(),
    deleteItem: (item) => manageSystemPart.delete(item.id),
    getDeleteMessage: (item) =>
      translate('facility.delete_system_part_confirm').replace(
        '{name}',
        item.short_name ?? item.name
      ),
    getDeleteSuccessMessage: () => translate('facility.system_part_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_system_part_failed')
  });
}

export function createStateTextActions(): CrudPageActions<StateText> {
  const manageStateText = new ManageEntityUseCase(stateTextRepository);
  return createStandardFacilityCrudActions<StateText>({
    reload: () => stateTextsStore.reload(),
    deleteItem: (item) => manageStateText.delete(item.id),
    getDeleteTitle: () => translate('facility.delete_state_text_confirm').replace('{ref}', ''),
    getDeleteMessage: (item) =>
      translate('facility.delete_state_text_confirm').replace(
        '{ref}',
        String(item.ref_number || '')
      ),
    getDeleteSuccessMessage: () => translate('facility.state_text_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_state_text_failed')
  });
}

export function createApparatActions(): CrudPageActions<Apparat> {
  const manageApparat = new ManageEntityUseCase(apparatRepository);
  return createStandardFacilityCrudActions<Apparat>({
    reload: () => apparatsStore.reload(),
    deleteItem: (item) => manageApparat.delete(item.id),
    getDeleteMessage: (item) =>
      translate('facility.delete_apparat_confirm').replace('{name}', item.short_name ?? item.name),
    getDeleteSuccessMessage: () => translate('facility.apparat_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_apparat_failed')
  });
}

export function createNotificationClassActions(): CrudPageActions<NotificationClass> {
  const manageNotificationClass = new ManageEntityUseCase(notificationClassRepository);
  return createStandardFacilityCrudActions<NotificationClass>({
    reload: () => notificationClassesStore.reload(),
    deleteItem: (item) => manageNotificationClass.delete(item.id),
    getDeleteTitle: () =>
      translate('facility.delete_notification_class_confirm').replace('{name}', ''),
    getDeleteMessage: (item) =>
      translate('facility.delete_notification_class_confirm').replace(
        '{name}',
        item.event_category || ''
      ),
    getDeleteSuccessMessage: () => translate('facility.notification_class_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_notification_class_failed')
  });
}

export function createAlarmDefinitionActions(): CrudPageActions<AlarmDefinition> {
  const manageAlarmDefinition = new ManageEntityUseCase(alarmDefinitionRepository);
  return createStandardFacilityCrudActions<AlarmDefinition>({
    reload: () => alarmDefinitionsStore.reload(),
    deleteItem: (item) => manageAlarmDefinition.delete(item.id),
    getDeleteTitle: () =>
      translate('facility.delete_alarm_definition_confirm').replace('{name}', ''),
    getDeleteMessage: (item) =>
      translate('facility.delete_alarm_definition_confirm').replace('{name}', item.name || ''),
    getDeleteSuccessMessage: () => translate('facility.alarm_definition_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_alarm_definition_failed')
  });
}

export type ObjectDataActions = CrudPageActions<ObjectData> & {
  editFresh: (item: ObjectData) => Promise<void>;
};

export function createObjectDataActions(): ObjectDataActions {
  const manageObjectData = new ManageObjectDataUseCase(objectDataRepository);
  const actions = createStandardFacilityCrudActions<ObjectData>({
    reload: () => objectDataStore.reload(),
    deleteItem: (item) => manageObjectData.delete(item.id),
    getDeleteTitle: () => translate('facility.delete_object_data_confirm').replace('{desc}', ''),
    getDeleteMessage: (item) =>
      translate('facility.delete_object_data_confirm').replace('{desc}', item.description || ''),
    getDeleteSuccessMessage: () => translate('facility.object_data_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_object_data_failed')
  }) as ObjectDataActions;

  actions.editFresh = async (item: ObjectData) => {
    try {
      actions.edit(await manageObjectData.get(item.id));
    } catch (error) {
      console.error(error);
      actions.edit(item);
    }
  };

  return actions;
}
