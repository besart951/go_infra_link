import { getContext, setContext } from 'svelte';
import { RolePermissionEditorState } from './RolePermissionEditorState.svelte.js';

import type { RolePermissionEditorStateOptions } from './RolePermissionEditorState.svelte.js';

const ROLE_PERMISSION_EDITOR_STATE_KEY = Symbol.for('go-infra-link.role-permission-editor.state');

export function createRolePermissionEditorState(
  options: RolePermissionEditorStateOptions
): RolePermissionEditorState {
  return new RolePermissionEditorState(options);
}

export function provideRolePermissionEditorState(
  options: RolePermissionEditorStateOptions
): RolePermissionEditorState {
  const state = createRolePermissionEditorState(options);
  setContext(ROLE_PERMISSION_EDITOR_STATE_KEY, state);
  return state;
}

export function useRolePermissionEditorState(): RolePermissionEditorState {
  return getContext(ROLE_PERMISSION_EDITOR_STATE_KEY);
}
