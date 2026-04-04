import { getContext, setContext } from 'svelte';
import { ControlCabinetState } from './ControlCabinetState.svelte.js';
import type { ControlCabinetStateProps } from './types.js';

const CONTROL_CABINET_STATE_KEY = Symbol.for('go-infra-link.control-cabinet.state');

export function createControlCabinetState(props: ControlCabinetStateProps = {}) {
  return new ControlCabinetState(props);
}

export function provideControlCabinetState(props: ControlCabinetStateProps = {}) {
  const state = createControlCabinetState(props);
  return setContext(CONTROL_CABINET_STATE_KEY, state);
}

export function useControlCabinetState(): ControlCabinetState {
  return getContext(CONTROL_CABINET_STATE_KEY);
}
