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

export function createBuildingActions(): CrudPageActions<Building> {
  const manageBuilding = new ManageBuildingUseCase(buildingRepository);
  return new CrudPageActions<Building>({
    reload: () => buildingsStore.reload(),
    deleteItem: (building) => manageBuilding.delete(building.id),
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: () => translate('common.delete'),
    getDeleteMessage: (building) =>
      translate('facility.delete_building_confirm').replace('{name}', building.iws_code),
    getDeleteConfirmText: () => translate('common.delete'),
    getDeleteCancelText: () => translate('common.cancel'),
    getDeleteSuccessMessage: () => translate('facility.building_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_building_failed')
  });
}

export function createSystemTypeActions(): CrudPageActions<SystemType> {
  const manageSystemType = new ManageEntityUseCase(systemTypeRepository);
  return new CrudPageActions<SystemType>({
    reload: () => systemTypesStore.reload(),
    deleteItem: (item) => manageSystemType.delete(item.id),
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: () => translate('common.delete'),
    getDeleteMessage: (item) =>
      translate('facility.delete_system_type_confirm').replace('{name}', item.name),
    getDeleteConfirmText: () => translate('common.delete'),
    getDeleteCancelText: () => translate('common.cancel'),
    getDeleteSuccessMessage: () => translate('facility.system_type_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_system_type_failed')
  });
}

export function createSystemPartActions(): CrudPageActions<SystemPart> {
  const manageSystemPart = new ManageEntityUseCase(systemPartRepository);
  return new CrudPageActions<SystemPart>({
    reload: () => systemPartsStore.reload(),
    deleteItem: (item) => manageSystemPart.delete(item.id),
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: () => translate('common.delete'),
    getDeleteMessage: (item) =>
      translate('facility.delete_system_part_confirm').replace(
        '{name}',
        item.short_name ?? item.name
      ),
    getDeleteConfirmText: () => translate('common.delete'),
    getDeleteCancelText: () => translate('common.cancel'),
    getDeleteSuccessMessage: () => translate('facility.system_part_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_system_part_failed')
  });
}

export function createStateTextActions(): CrudPageActions<StateText> {
  const manageStateText = new ManageEntityUseCase(stateTextRepository);
  return new CrudPageActions<StateText>({
    reload: () => stateTextsStore.reload(),
    deleteItem: (item) => manageStateText.delete(item.id),
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: () => translate('facility.delete_state_text_confirm').replace('{ref}', ''),
    getDeleteMessage: (item) =>
      translate('facility.delete_state_text_confirm').replace(
        '{ref}',
        String(item.ref_number || '')
      ),
    getDeleteConfirmText: () => translate('common.delete'),
    getDeleteCancelText: () => translate('common.cancel'),
    getDeleteSuccessMessage: () => translate('facility.state_text_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_state_text_failed')
  });
}

export function createApparatActions(): CrudPageActions<Apparat> {
  const manageApparat = new ManageEntityUseCase(apparatRepository);
  return new CrudPageActions<Apparat>({
    reload: () => apparatsStore.reload(),
    deleteItem: (item) => manageApparat.delete(item.id),
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: () => translate('common.delete'),
    getDeleteMessage: (item) =>
      translate('facility.delete_apparat_confirm').replace('{name}', item.short_name ?? item.name),
    getDeleteConfirmText: () => translate('common.delete'),
    getDeleteCancelText: () => translate('common.cancel'),
    getDeleteSuccessMessage: () => translate('facility.apparat_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_apparat_failed')
  });
}

export function createNotificationClassActions(): CrudPageActions<NotificationClass> {
  const manageNotificationClass = new ManageEntityUseCase(notificationClassRepository);
  return new CrudPageActions<NotificationClass>({
    reload: () => notificationClassesStore.reload(),
    deleteItem: (item) => manageNotificationClass.delete(item.id),
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: () =>
      translate('facility.delete_notification_class_confirm').replace('{name}', ''),
    getDeleteMessage: (item) =>
      translate('facility.delete_notification_class_confirm').replace(
        '{name}',
        item.event_category || ''
      ),
    getDeleteConfirmText: () => translate('common.delete'),
    getDeleteCancelText: () => translate('common.cancel'),
    getDeleteSuccessMessage: () => translate('facility.notification_class_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_notification_class_failed')
  });
}

export function createAlarmDefinitionActions(): CrudPageActions<AlarmDefinition> {
  const manageAlarmDefinition = new ManageEntityUseCase(alarmDefinitionRepository);
  return new CrudPageActions<AlarmDefinition>({
    reload: () => alarmDefinitionsStore.reload(),
    deleteItem: (item) => manageAlarmDefinition.delete(item.id),
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: () =>
      translate('facility.delete_alarm_definition_confirm').replace('{name}', ''),
    getDeleteMessage: (item) =>
      translate('facility.delete_alarm_definition_confirm').replace('{name}', item.name || ''),
    getDeleteConfirmText: () => translate('common.delete'),
    getDeleteCancelText: () => translate('common.cancel'),
    getDeleteSuccessMessage: () => translate('facility.alarm_definition_deleted'),
    getDeleteFailureMessage: () => translate('facility.delete_alarm_definition_failed')
  });
}

export type ObjectDataActions = CrudPageActions<ObjectData> & {
  editFresh: (item: ObjectData) => Promise<void>;
};

export function createObjectDataActions(): ObjectDataActions {
  const manageObjectData = new ManageObjectDataUseCase(objectDataRepository);
  const actions = new CrudPageActions<ObjectData>({
    reload: () => objectDataStore.reload(),
    deleteItem: (item) => manageObjectData.delete(item.id),
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: () => translate('facility.delete_object_data_confirm').replace('{desc}', ''),
    getDeleteMessage: (item) =>
      translate('facility.delete_object_data_confirm').replace('{desc}', item.description || ''),
    getDeleteConfirmText: () => translate('common.delete'),
    getDeleteCancelText: () => translate('common.cancel'),
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
