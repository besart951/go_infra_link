export interface EntityRefreshRequest {
  key: string | number;
  entityIds?: string[];
}

export interface EntityDeltaRequest<T> {
  key: string | number;
  items: T[];
}

export interface EntityChangeEvent<T> {
  entityIds?: string[];
  items?: T[];
}