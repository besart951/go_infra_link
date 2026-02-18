import type {
	ControlCabinet,
	CreateControlCabinetRequest,
	UpdateControlCabinetRequest,
	ControlCabinetDeleteImpact
} from '$lib/domain/facility/index.js';
import type { ControlCabinetRepository } from '$lib/domain/ports/facility/controlCabinetRepository.js';
import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';

export class ManageControlCabinetUseCase extends ManageEntityUseCase<
	ControlCabinet,
	CreateControlCabinetRequest,
	UpdateControlCabinetRequest
> {
	constructor(private repo: ControlCabinetRepository) {
		super(repo);
	}

	async copy(id: string, signal?: AbortSignal): Promise<ControlCabinet> {
		return this.repo.copy(id, signal);
	}

	async validate(
		data: { id?: string; building_id: string; control_cabinet_nr?: string },
		signal?: AbortSignal
	): Promise<void> {
		return this.repo.validate(data, signal);
	}

	async getDeleteImpact(id: string, signal?: AbortSignal): Promise<ControlCabinetDeleteImpact> {
		return this.repo.getDeleteImpact(id, signal);
	}
}
