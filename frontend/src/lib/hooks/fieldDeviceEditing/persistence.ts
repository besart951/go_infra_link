import type { BacnetObjectInput, BulkUpdateFieldDeviceItem } from '$lib/domain/facility/index.js';
import type { SessionStorageAdapter } from '$lib/services/sessionStorageService.js';

export interface PersistedFieldDeviceEditingState {
  edits: Array<[string, Partial<BulkUpdateFieldDeviceItem>]>;
  bacnetEdits: Array<[string, Array<[string, Partial<BacnetObjectInput>]>]>;
  timestamp: number;
}

const STORAGE_KEY_PREFIX = 'fielddevice-editing';
const MAX_AGE_MS = 24 * 60 * 60 * 1000;

export function getFieldDeviceEditingStorageKey(projectId?: string): string {
  return projectId ? `${STORAGE_KEY_PREFIX}-${projectId}` : STORAGE_KEY_PREFIX;
}

export function loadPersistedFieldDeviceEditingState(
  storage: SessionStorageAdapter,
  key: string,
  now = Date.now()
): PersistedFieldDeviceEditingState | null {
  const loaded = storage.load<PersistedFieldDeviceEditingState>(key);
  if (!loaded) return null;

  if (now - loaded.timestamp > MAX_AGE_MS) {
    storage.remove(key);
    return null;
  }

  return loaded;
}

export function savePersistedFieldDeviceEditingState(
  storage: SessionStorageAdapter,
  key: string,
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  pendingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>,
  now = Date.now()
): void {
  if (pendingEdits.size === 0 && pendingBacnetEdits.size === 0) {
    storage.remove(key);
    return;
  }

  storage.save<PersistedFieldDeviceEditingState>(key, {
    edits: Array.from(pendingEdits.entries()),
    bacnetEdits: Array.from(pendingBacnetEdits.entries()).map(([deviceId, objMap]) => [
      deviceId,
      Array.from(objMap.entries())
    ]),
    timestamp: now
  });
}

export function removePersistedFieldDeviceEditingState(
  storage: SessionStorageAdapter,
  key: string
): void {
  storage.remove(key);
}
