import { fireEvent, render, screen } from '@testing-library/svelte';
import ProjectPhaseSelect from './ProjectPhaseSelect.svelte';

vi.mock('$lib/infrastructure/api/phase.adapter.js', () => ({
  getPhase: vi.fn(),
  listPhases: vi.fn().mockResolvedValue({ items: [] })
}));

const PhaseSelect = ProjectPhaseSelect as any;

describe('ProjectPhaseSelect', () => {
  beforeAll(() => {
    HTMLElement.prototype.scrollIntoView = vi.fn();
  });

  it('uses localized German defaults for its trigger and search', async () => {
    render(PhaseSelect, {});

    const trigger = screen.getByRole('combobox');
    expect(trigger).toHaveTextContent('Phase auswählen...');

    await fireEvent.click(trigger);

    expect(await screen.findByPlaceholderText('Phasen suchen...')).toBeInTheDocument();
  });
});
