export interface ExcelReadProgress {
	percent: number;
	currentSheet: number;
	totalSheets: number;
	message: string;
}

export interface BacnetObjectExcel {
	id: string;
	text_fix: string;
	description: string;
	gms_visible: boolean;
	is_optional: boolean;
	text_individual: string;
	software_type: string;
	software_number: string;
	hardware_label: string;
	software_reference_label: string;
	state_text_label: string;
	notification_class_label: string;
	alarm_definition_label: string;
	apparat_label: string;
}

export interface ObjectDataExcel {
	id: string;
	description: string;
	is_optional_anchor: boolean;
	bacnet_objects: BacnetObjectExcel[];
}

export interface ExcelReadSession {
	fileName: string;
	fileSize: number;
	objectDataExcel: ObjectDataExcel[];
	createdAt: string;
}
