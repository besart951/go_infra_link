import { getContext, setContext } from 'svelte';
import { ControlCabinetDetailState } from './ControlCabinetDetailState.svelte.js';

import type { ControlCabinetDetailStateOptions } from './ControlCabinetDetailState.svelte.js';

const CONTROL_CABINET_DETAIL_STATE_KEY = Symbol.for('go-infra-link.control-cabinet-detail.state');

export function createControlCabinetDetailState(
  options: ControlCabinetDetailStateOptions
): ControlCabinetDetailState {
  return new ControlCabinetDetailState(options);
}

export function provideControlCabinetDetailState(
  options: ControlCabinetDetailStateOptions
): ControlCabinetDetailState {
  const state = createControlCabinetDetailState(options);
  setContext(CONTROL_CABINET_DETAIL_STATE_KEY, state);
  return state;
}

export function useControlCabinetDetailState(): ControlCabinetDetailState {
  return getContext(CONTROL_CABINET_DETAIL_STATE_KEY);
}
