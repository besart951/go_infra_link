import { fireEvent, render, screen } from '@testing-library/svelte';
import ProjectListToolbar from './ProjectListToolbar.svelte';

vi.mock('$lib/infrastructure/api/phase.adapter.js', () => ({
  getPhase: vi.fn(),
  listPhases: vi.fn().mockResolvedValue({ items: [] })
}));

const Toolbar = ProjectListToolbar as any;

describe('ProjectListToolbar', () => {
  beforeAll(() => {
    HTMLElement.prototype.scrollIntoView = vi.fn();
  });

  it('renders the status filter as a searchable combobox instead of a native select', async () => {
    const onStatusChange = vi.fn();
    const { container } = render(Toolbar, {
      statusLabel: 'Status',
      statusValue: 'all',
      options: [
        { value: 'all', label: 'Alle Status' },
        { value: 'planned', label: 'Geplant' }
      ],
      phaseLabel: 'Phase',
      phaseValue: '',
      allPhasesLabel: 'Alle Phasen',
      statusSearchPlaceholder: 'Status suchen...',
      statusEmptyText: 'Keine Status gefunden.',
      phaseSearchPlaceholder: 'Phasen suchen...',
      phaseEmptyText: 'Keine Phasen gefunden.',
      onStatusChange,
      onPhaseChange: vi.fn()
    });

    expect(container.querySelector('select')).toBeNull();

    const statusTrigger = container.querySelector('#project_status_filter');
    expect(statusTrigger).toHaveAttribute('role', 'combobox');

    await fireEvent.click(statusTrigger!);
    expect(await screen.findByPlaceholderText('Status suchen...')).toBeInTheDocument();

    await fireEvent.click(screen.getByText('Geplant'));
    expect(onStatusChange).toHaveBeenCalledWith('planned');
  });
});
