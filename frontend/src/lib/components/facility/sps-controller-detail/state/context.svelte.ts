import { getContext, setContext } from 'svelte';
import { SPSControllerDetailState } from './SPSControllerDetailState.svelte.js';

import type { SPSControllerDetailStateOptions } from './SPSControllerDetailState.svelte.js';

const SPS_CONTROLLER_DETAIL_STATE_KEY = Symbol.for('go-infra-link.sps-controller-detail.state');

export function createSPSControllerDetailState(
  options: SPSControllerDetailStateOptions
): SPSControllerDetailState {
  return new SPSControllerDetailState(options);
}

export function provideSPSControllerDetailState(
  options: SPSControllerDetailStateOptions
): SPSControllerDetailState {
  const state = createSPSControllerDetailState(options);
  setContext(SPS_CONTROLLER_DETAIL_STATE_KEY, state);
  return state;
}

export function useSPSControllerDetailState(): SPSControllerDetailState {
  return getContext(SPS_CONTROLLER_DETAIL_STATE_KEY);
}
