export type BacnetReferenceResource =
  | 'apparat'
  | 'system_part'
  | 'system_type'
  | 'state_text'
  | 'notification_class'
  | 'alarm_type'
  | 'alarm_definition'
  | 'object_data';

export interface BacnetReferenceUsage {
  resource: BacnetReferenceResource;
  id: string;
  bacnet_object_count: number;
}

export interface BacnetReferenceUsageListResponse {
  items: BacnetReferenceUsage[];
}
