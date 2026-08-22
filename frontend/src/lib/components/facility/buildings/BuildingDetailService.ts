import { ManageBuildingUseCase } from '$lib/application/useCases/facility/manageBuildingUseCase.js';
import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';

const manageBuilding = new ManageBuildingUseCase(buildingRepository);

export const buildingDetailService = {
  validate(data: Parameters<ManageBuildingUseCase['validate']>[0]) {
    return manageBuilding.validate(data);
  },

  update(id: string, data: Parameters<ManageBuildingUseCase['update']>[1]) {
    return manageBuilding.update(id, data);
  },

  delete(command: Parameters<ManageBuildingUseCase['delete']>[0]) {
    return manageBuilding.delete(command);
  }
};
