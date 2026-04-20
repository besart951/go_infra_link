import { render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const state = vi.hoisted(() => ({
  gotoMock: vi.fn()
}));

vi.mock('$app/navigation', () => ({
  goto: state.gotoMock
}));

vi.mock('$lib/i18n/translator.js', () => ({
  createTranslator: () => ({
    subscribe(run: (value: (key: string) => string) => void) {
      run((key: string) => key);
      return () => {};
    }
  })
}));

vi.mock('$lib/components/project/PhaseForm.svelte', async () => {
  const { default: PhaseFormStub } = await import('../../setup/stubs/PhaseFormStub.svelte');
  return { default: PhaseFormStub };
});

import ProjectPhaseNewPage from '../../../src/routes/(app)/projects/phases/new/+page.svelte';

describe('/projects/phases/new', () => {
  beforeEach(() => {
    state.gotoMock.mockReset();
  });

  it('renders the creation form without consulting phase.create permissions', () => {
    render(ProjectPhaseNewPage);

    expect(screen.getByText('phases.new.title')).toBeInTheDocument();
    expect(screen.getByTestId('phase-form')).toBeInTheDocument();
  });
});
