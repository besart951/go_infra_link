import type { Apparat, FieldDeviceOptions, ObjectData, SystemPart } from './index.js';

export interface FieldDevicePreselection {
	objectDataId: string;
	apparatId: string;
	systemPartId: string;
}

export interface FilteredPreselectionOptions {
	objectDatas: ObjectData[];
	apparats: Apparat[];
	systemParts: SystemPart[];
}

function uniq<T>(items: T[]): T[] {
	return Array.from(new Set(items));
}

function invertMap(map: Record<string, string[]>): Record<string, string[]> {
	const inverted: Record<string, string[]> = {};
	for (const [key, values] of Object.entries(map)) {
		for (const value of values) {
			(inverted[value] ??= []).push(key);
		}
	}
	return inverted;
}

export function getFilteredFieldDevicePreselectionOptions(
	options: FieldDeviceOptions,
	selection: FieldDevicePreselection
): FilteredPreselectionOptions {
	const { objectDataId, apparatId, systemPartId } = selection;

	const objectDataToApparat = options.object_data_to_apparat ?? {};
	const apparatToSystemPart = options.apparat_to_system_part ?? {};

	const apparatToObjectData = invertMap(objectDataToApparat); // apparat_id -> [object_data_ids]
	const systemPartToApparat = invertMap(apparatToSystemPart); // system_part_id -> [apparat_ids]

	// --- Filter apparats ---
	let allowedApparatIds = options.apparats.map((a) => a.id);
	if (objectDataId) {
		allowedApparatIds = allowedApparatIds.filter((id) =>
			(objectDataToApparat[objectDataId] ?? []).includes(id)
		);
	}
	if (systemPartId) {
		allowedApparatIds = allowedApparatIds.filter((id) =>
			(systemPartToApparat[systemPartId] ?? []).includes(id)
		);
	}
	const filteredApparats = options.apparats.filter((a) => allowedApparatIds.includes(a.id));

	// --- Filter system parts ---
	let basisApparatIdsForSystemParts = allowedApparatIds;
	if (apparatId) {
		basisApparatIdsForSystemParts = basisApparatIdsForSystemParts.filter((id) => id === apparatId);
	}
	const allowedSystemPartIds = uniq(
		basisApparatIdsForSystemParts.flatMap((id) => apparatToSystemPart[id] ?? [])
	);
	const filteredSystemParts = options.system_parts.filter((sp) =>
		allowedSystemPartIds.includes(sp.id)
	);

	// --- Filter object datas ---
	let allowedObjectDataIds = options.object_datas.map((od) => od.id);
	if (apparatId) {
		allowedObjectDataIds = allowedObjectDataIds.filter((id) =>
			(apparatToObjectData[apparatId] ?? []).includes(id)
		);
	}
	if (systemPartId) {
		// objectData must have at least one apparat that contains the selected system part
		allowedObjectDataIds = allowedObjectDataIds.filter((odId) => {
			const apparatIds = objectDataToApparat[odId] ?? [];
			return apparatIds.some((aId) => (apparatToSystemPart[aId] ?? []).includes(systemPartId));
		});
	}
	const filteredObjectDatas = options.object_datas.filter((od) =>
		allowedObjectDataIds.includes(od.id)
	);

	return {
		objectDatas: filteredObjectDatas,
		apparats: filteredApparats,
		systemParts: filteredSystemParts
	};
}
