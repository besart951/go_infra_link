import type { MultiCreateSelection } from '$lib/domain/facility/fieldDeviceMultiCreate.js';
import type { FieldDevicePreselection } from '$lib/domain/facility/preselectionFilter.js';

export function createEmptyMultiCreateSelection(): MultiCreateSelection {
  return {
    spsControllerSystemTypeId: '',
    objectDataId: '',
    apparatId: '',
    systemPartId: ''
  };
}

export function createEmptyFieldDevicePreselection(): FieldDevicePreselection {
  return {
    objectDataId: '',
    apparatId: '',
    systemPartId: ''
  };
}

export function resetSelectionForSpsSystemType(value: string): MultiCreateSelection {
  return {
    ...createEmptyMultiCreateSelection(),
    spsControllerSystemTypeId: value
  };
}

export function applyPreselectionToSelection(
  selection: MultiCreateSelection,
  preselection: FieldDevicePreselection
): MultiCreateSelection {
  return {
    ...selection,
    objectDataId: preselection.objectDataId,
    apparatId: preselection.apparatId,
    systemPartId: preselection.systemPartId
  };
}

export function preselectionFromSelection(
  selection: MultiCreateSelection
): FieldDevicePreselection {
  return {
    objectDataId: selection.objectDataId,
    apparatId: selection.apparatId,
    systemPartId: selection.systemPartId
  };
}
