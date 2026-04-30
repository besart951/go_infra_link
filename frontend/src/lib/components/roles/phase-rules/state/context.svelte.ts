import { getContext, setContext } from 'svelte';
import {
  PhasePermissionRulesState,
  type PhasePermissionRulesStateOptions
} from './PhasePermissionRulesState.svelte.js';

const PHASE_PERMISSION_RULES_STATE_KEY = Symbol.for('go-infra-link.phase-permission-rules.state');

export function createPhasePermissionRulesState(
  options: PhasePermissionRulesStateOptions
): PhasePermissionRulesState {
  return new PhasePermissionRulesState(options);
}

export function providePhasePermissionRulesState(
  options: PhasePermissionRulesStateOptions
): PhasePermissionRulesState {
  const state = createPhasePermissionRulesState(options);
  setContext(PHASE_PERMISSION_RULES_STATE_KEY, state);
  return state;
}

export function usePhasePermissionRulesState(): PhasePermissionRulesState {
  return getContext(PHASE_PERMISSION_RULES_STATE_KEY);
}
