import { ManageSPSControllerUseCase } from '$lib/application/useCases/facility/manageSPSControllerUseCase.js';
import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
import { copyProjectSPSControllerSystemType } from '$lib/infrastructure/api/project.adapter.js';
import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
import { spsControllerSystemTypeRepository } from '$lib/infrastructure/api/spsControllerSystemTypeRepository.js';
import { systemTypeRepository } from '$lib/infrastructure/api/systemTypeRepository.js';
import type {
  CreateSPSControllerRequest,
  FacilityJob,
  SPSController,
  SPSControllerSystemType,
  UpdateSPSControllerRequest
} from '$lib/domain/facility/index.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';

const manageSPSController = new ManageSPSControllerUseCase(spsControllerRepository);

export const spsControllerFormService = {
  validate(data: Parameters<ManageSPSControllerUseCase['validate']>[0]) {
    return manageSPSController.validate(data);
  },

  create(data: CreateSPSControllerRequest): Promise<SPSController> {
    return manageSPSController.create(data);
  },

  update(id: string, data: UpdateSPSControllerRequest): Promise<SPSController> {
    return manageSPSController.update(id, data);
  },

  getNextGADevice(controlCabinetId: string, excludeId?: string) {
    return manageSPSController.getNextGADevice(controlCabinetId, excludeId);
  },

  listSystemTypes(params: ListParams): Promise<PaginatedResponse<SPSControllerSystemType>> {
    return spsControllerSystemTypeRepository.list(params);
  },

  updateSystemType(
    entry: Pick<SPSControllerSystemType, 'id' | 'aggregate_version' | 'number' | 'document_name'>
  ): Promise<SPSControllerSystemType> {
    return spsControllerSystemTypeRepository.update(entry.id, {
      base_version: entry.aggregate_version,
      number: entry.number,
      document_name: entry.document_name
    });
  },

  deleteSystemType(
    entry: Pick<SPSControllerSystemType, 'id' | 'aggregate_version'>
  ): Promise<void> {
    return spsControllerSystemTypeRepository.delete({
      id: entry.id,
      base_version: entry.aggregate_version
    });
  },

  copySystemType(id: string, operationId: string): Promise<FacilityJob> {
    return spsControllerSystemTypeRepository.copy(id, operationId);
  },

  copyProjectSystemType(projectId: string, id: string, operationId: string): Promise<FacilityJob> {
    return copyProjectSPSControllerSystemType(projectId, id, operationId);
  },

  getSystemType(id: string) {
    return systemTypeRepository.get(id);
  },

  getControlCabinet(id: string) {
    return controlCabinetRepository.get(id);
  },

  getBuilding(id: string) {
    return buildingRepository.get(id);
  }
};
