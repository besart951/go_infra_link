import { getContext, setContext } from 'svelte';
import { SPSControllerState } from './SPSControllerState.svelte.js';
import type { SPSControllerStateProps } from './types.js';

const SPS_CONTROLLER_STATE_KEY = Symbol.for('go-infra-link.sps-controller.state');

export function createSPSControllerState(props: SPSControllerStateProps = {}) {
  return new SPSControllerState(props);
}

export function provideSPSControllerState(props: SPSControllerStateProps = {}) {
  const state = createSPSControllerState(props);
  return setContext(SPS_CONTROLLER_STATE_KEY, state);
}

export function useSPSControllerState(): SPSControllerState {
  return getContext(SPS_CONTROLLER_STATE_KEY);
}
