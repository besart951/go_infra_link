import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import InlineAutosaveField from './InlineAutosaveField.svelte';

describe('InlineAutosaveField', () => {
  it('normalizes a constrained value before autosaving it', async () => {
    const onSave = vi.fn(async () => undefined);
    render(InlineAutosaveField, {
      label: 'GA-Gerät',
      value: 'AAA',
      maxLength: 3,
      transform: (value: string) => value.toUpperCase(),
      onSave
    });

    const input = screen.getByRole('textbox', { name: 'GA-Gerät' }) as HTMLInputElement;
    expect(input.maxLength).toBe(3);

    await fireEvent.focus(input);
    await fireEvent.input(input, { target: { value: 'abc' } });
    await fireEvent.blur(input);

    await waitFor(() => expect(onSave).toHaveBeenCalledWith('ABC'));
  });

  it('renders calculated values as read-only text', () => {
    const onSave = vi.fn(async () => undefined);
    render(InlineAutosaveField, {
      label: 'Name',
      value: 'E001_AK01_ABC',
      disabled: true,
      onSave
    });

    expect(screen.queryByRole('textbox', { name: 'Name' })).not.toBeInTheDocument();
    expect(screen.getByText('E001_AK01_ABC')).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });
});
