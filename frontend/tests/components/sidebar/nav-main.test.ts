import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';

import NavMainCollapsedHarness from '../../setup/NavMainCollapsedHarness.svelte';

describe('NavMain', () => {
  it('renders collapsed submenus before their dynamic refs are assigned', () => {
    render(NavMainCollapsedHarness);

    expect(screen.getByRole('button', { name: 'Facility' })).toBeInTheDocument();
  });

  it('opens the hovered menu and closes the previously active menu', async () => {
    render(NavMainCollapsedHarness);

    const facilityButton = screen.getByRole('button', { name: 'Facility' });
    const projectsButton = screen.getByRole('button', { name: 'Projects' });

    await fireEvent.pointerMove(facilityButton, { pointerType: 'mouse' });
    await waitFor(() => expect(facilityButton).toHaveAttribute('aria-expanded', 'true'));
    expect(projectsButton).toHaveAttribute('aria-expanded', 'false');

    await fireEvent.pointerMove(projectsButton, { pointerType: 'mouse' });
    await waitFor(() => expect(projectsButton).toHaveAttribute('aria-expanded', 'true'));
    expect(facilityButton).toHaveAttribute('aria-expanded', 'false');
  });
});
