import type { Phase } from '$lib/domain/phase/index.js';
import type {
	ListParams,
	ListRepository,
	PaginatedResponse
} from '$lib/domain/ports/listRepository.js';
import { calculateTotalPages } from '$lib/domain/valueObjects/pagination.js';
import { listProjects, createProject } from '$lib/infrastructure/api/project.adapter.js';

const MAX_PHASE_SAMPLES = 1000;

function buildPhaseList(phaseIds: string[], searchText: string): Phase[] {
	const normalized = searchText.trim().toLowerCase();
	const filtered = normalized
		? phaseIds.filter((id) => id.toLowerCase().includes(normalized))
		: phaseIds;

	return filtered.map((id) => ({ id, name: id }));
}

export class PhaseListRepository implements ListRepository<Phase> {
	async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Phase>> {
		const response = await listProjects({ page: 1, limit: MAX_PHASE_SAMPLES }, { signal });
		const unique = new Set<string>();
		for (const project of response.items ?? []) {
			const phaseId = project.phase_id?.trim();
			if (phaseId) unique.add(phaseId);
		}

		const phaseIds = Array.from(unique.values()).sort();
		const phases = buildPhaseList(phaseIds, params.search?.text ?? '');

		const pageSize = params.pagination.pageSize;
		const page = params.pagination.page;
		const total = phases.length;
		const totalPages = calculateTotalPages(total, pageSize);
		const start = (page - 1) * pageSize;
		const items = phases.slice(start, start + pageSize);

		return {
			items,
			metadata: {
				total,
				page,
				pageSize,
				totalPages
			}
		};
	}
}

export async function createPhase(
	id: string,
	name?: string,
	options?: RequestInit
): Promise<Phase> {
	const trimmed = id.trim();
	const displayName = name?.trim() || trimmed;

	const project = await createProject(
		{
			name: displayName || `Phase: ${trimmed}`,
			status: 'planned',
			phase_id: trimmed
		},
		options
	);

	const resolvedName = displayName || project.phase_id || trimmed;

	return {
		id: project.phase_id ?? trimmed,
		name: resolvedName
	};
}

export const phaseListRepository = new PhaseListRepository();
