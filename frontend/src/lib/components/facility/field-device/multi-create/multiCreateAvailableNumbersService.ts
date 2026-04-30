import { ManageFieldDeviceUseCase } from '$lib/application/useCases/facility/manageFieldDeviceUseCase.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import type { AvailableApparatNumbersResponse } from '$lib/domain/facility/index.js';
import type {
  FieldDeviceRowData,
  MultiCreateSelection
} from '$lib/domain/facility/fieldDeviceMultiCreate.js';
import { getUsedApparatNumbers } from '$lib/domain/facility/fieldDeviceMultiCreate.js';

export class MultiCreateAvailableNumbersService {
  private readonly manageUseCase = new ManageFieldDeviceUseCase(fieldDeviceRepository);

  fetch(
    selection: MultiCreateSelection,
    signal?: AbortSignal
  ): Promise<AvailableApparatNumbersResponse> {
    return this.manageUseCase.getAvailableApparatNumbers(
      selection.spsControllerSystemTypeId,
      selection.apparatId,
      selection.systemPartId,
      signal
    );
  }
}

export function autoAssignApparatNumbers(
  rows: FieldDeviceRowData[],
  availableNumbers: number[]
): { rows: FieldDeviceRowData[]; changed: boolean } {
  const usedNumbers = getUsedApparatNumbers(rows);
  let changed = false;

  const nextRows = rows.map((row) => {
    if (row.apparatNr !== null) {
      return row;
    }

    const nextAvailable = availableNumbers.find((candidate) => !usedNumbers.has(candidate));
    if (nextAvailable === undefined) {
      return row;
    }

    usedNumbers.add(nextAvailable);
    changed = true;
    return { ...row, apparatNr: nextAvailable };
  });

  return { rows: nextRows, changed };
}
