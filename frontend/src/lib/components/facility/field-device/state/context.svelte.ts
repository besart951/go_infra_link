import { getContext, setContext } from 'svelte';
import { FieldDeviceState } from './FieldDeviceState.svelte.js';
import type { FieldDeviceStateProps } from './types.js';

const FIELD_DEVICE_STATE_KEY = Symbol.for('go-infra-link.field-device.state');

export function createFieldDeviceState(props: FieldDeviceStateProps = {}) {
  return new FieldDeviceState(props);
}

export function provideFieldDeviceState(props: FieldDeviceStateProps = {}) {
  const state = createFieldDeviceState(props);
  return setContext(FIELD_DEVICE_STATE_KEY, state);
}

export function useFieldDeviceState(): FieldDeviceState {
  return getContext(FIELD_DEVICE_STATE_KEY);
}
