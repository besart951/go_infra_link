export type PressEffect = 'default' | 'subtle' | 'none';

const pressEffectClasses: Record<PressEffect, string> = {
  default: 'active:scale-[0.98] active:brightness-90 active:shadow-inner',
  subtle: 'active:scale-[0.995] active:brightness-95',
  none: ''
};

export function getPressEffectClasses(effect: PressEffect): string {
  return pressEffectClasses[effect];
}
