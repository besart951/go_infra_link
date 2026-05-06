import { render } from '@testing-library/svelte';
import Button from './button.svelte';

describe('Button', () => {
  it('uses the default press effect for regular buttons', () => {
    const { container } = render(Button, { type: 'button' });
    const button = container.querySelector('[data-slot="button"]');

    expect(button).toHaveAttribute('data-press-effect', 'default');
    expect(button?.className).toContain('active:scale-[0.98]');
    expect(button?.className).toContain('active:shadow-inner');
  });

  it('uses a subtler press effect for combobox triggers', () => {
    const { container } = render(Button, {
      type: 'button',
      role: 'combobox',
      'aria-label': 'Control cabinet'
    });
    const button = container.querySelector('[data-slot="button"]');

    expect(button).toHaveAttribute('data-press-effect', 'subtle');
    expect(button?.className).toContain('active:scale-[0.995]');
    expect(button?.className).toContain('active:brightness-95');
    expect(button?.className).not.toContain('active:shadow-inner');
  });
});
