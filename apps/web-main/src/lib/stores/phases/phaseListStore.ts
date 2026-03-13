import { createListStore } from '$lib/stores/list/listStore.js';
import { phaseListRepository } from '$lib/infrastructure/api/phase.adapter.js';

export const phaseListStore = createListStore(phaseListRepository, { pageSize: 20 });
