import { ManageSPSControllerUseCase } from '$lib/application/useCases/facility/manageSPSControllerUseCase.js';
import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
import { copyProjectSPSControllerSystemType } from '$lib/infrastructure/api/project.adapter.js';
import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
import { spsControllerSystemTypeRepository } from '$lib/infrastructure/api/spsControllerSystemTypeRepository.js';
import { systemTypeRepository } from '$lib/infrastructure/api/systemTypeRepository.js';
import type {
  CreateSPSControllerRequest,
  SPSController,
  SPSControllerSystemType,
  SPSControllerSystemTypeInput,
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

  deleteSystemType(id: string): Promise<void> {
    return spsControllerSystemTypeRepository.delete(id);
  },

  copySystemType(id: string): Promise<SPSControllerSystemType> {
    return spsControllerSystemTypeRepository.copy(id);
  },

  copyProjectSystemType(projectId: string, id: string): Promise<SPSControllerSystemType> {
    return copyProjectSPSControllerSystemType(projectId, id);
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

export type SPSControllerSystemTypeEntry = SPSControllerSystemTypeInput & { id?: string };
