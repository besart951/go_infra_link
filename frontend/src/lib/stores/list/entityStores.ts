import { createApiAdapter } from '$lib/infrastructure/api/apiListAdapter.js';
import { createListStore } from './listStore.js';
import type { Building } from '$lib/domain/facility/index.js';
import type { ControlCabinet } from '$lib/domain/facility/index.js';
import type { SPSController } from '$lib/domain/facility/index.js';
import type { Apparat } from '$lib/domain/facility/index.js';
import type { SystemPart } from '$lib/domain/entities/systemPart.js';
import type { SPSControllerSystemType } from '$lib/domain/entities/spsControllerSystemType.js';
import type { ObjectData } from '$lib/domain/facility/index.js';
import type { Project } from '$lib/domain/entities/project.js';
import type { Team } from '$lib/domain/entities/team.js';
import type { User } from '$lib/domain/entities/user.js';

/**
 * Buildings store
 */
export const buildingsStore = createListStore<Building>(
	createApiAdapter<Building>('/facility/buildings'),
	{ pageSize: 10 }
);

/**
 * Control Cabinets store
 */
export const controlCabinetsStore = createListStore<ControlCabinet>(
	createApiAdapter<ControlCabinet>('/facility/control-cabinets'),
	{ pageSize: 10 }
);

/**
 * SPS Controllers store
 */
export const spsControllersStore = createListStore<SPSController>(
	createApiAdapter<SPSController>('/facility/sps-controllers'),
	{ pageSize: 10 }
);

/**
 * Apparats store
 */
export const apparatsStore = createListStore<Apparat>(
	createApiAdapter<Apparat>('/facility/apparats'),
	{ pageSize: 10 }
);

/**
 * System Parts store
 */
export const systemPartsStore = createListStore<SystemPart>(
	createApiAdapter<SystemPart>('/facility/system-parts'),
	{ pageSize: 10 }
);

/**
 * SPS Controller System Types store
 */
export const spsControllerSystemTypesStore = createListStore<SPSControllerSystemType>(
	createApiAdapter<SPSControllerSystemType>('/facility/sps-controller-system-types'),
	{ pageSize: 10 }
);

/**
 * Object Data store
 */
export const objectDataStore = createListStore<ObjectData>(
	createApiAdapter<ObjectData>('/facility/object-data'),
	{ pageSize: 10 }
);

/**
 * Projects store
 */
export const projectsStore = createListStore<Project>(
	createApiAdapter<Project>('/projects'),
	{ pageSize: 10 }
);

/**
 * Teams store
 */
export const teamsStore = createListStore<Team>(createApiAdapter<Team>('/teams'), { pageSize: 10 });

/**
 * Users store
 */
export const usersStore = createListStore<User>(createApiAdapter<User>('/users'), { pageSize: 10 });
