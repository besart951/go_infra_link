import { render, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const state = vi.hoisted(() => ({
  gotoMock: vi.fn(),
  loadAuthMock: vi.fn().mockResolvedValue(undefined),
  pageStore: {
    subscribe(run: (value: unknown) => void) {
      run({ url: new URL('http://localhost/'), params: {}, data: {} });
      return () => {};
    }
  }
}));

vi.mock('$app/stores', () => ({
  page: state.pageStore
}));

vi.mock('$app/navigation', () => ({
  goto: state.gotoMock
}));

vi.mock('$lib/stores/auth.svelte.js', () => ({
  loadAuth: state.loadAuthMock
}));

vi.mock('$lib/i18n/index.js', () => ({
  t: (key: string) => key
}));

vi.mock('$lib/i18n/translator.js', () => ({
  createTranslator: () => ({
    subscribe(run: (value: (key: string) => string) => void) {
      run((key: string) => key);
      return () => {};
    }
  })
}));

vi.mock('$lib/components/app-sidebar.svelte', async () => {
  const { default: AppSidebarStub } = await import('../../setup/stubs/AppSidebarStub.svelte');
  return { default: AppSidebarStub };
});

vi.mock('$lib/components/toast.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

vi.mock('$lib/components/ui/separator/index.js', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { Separator: SlotContainer };
});

vi.mock('$lib/components/ui/sidebar/index.js', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return {
    Provider: SlotContainer,
    Inset: SlotContainer,
    Trigger: SlotContainer
  };
});

vi.mock('$lib/components/ui/breadcrumb/index.js', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return {
    Root: SlotContainer,
    List: SlotContainer,
    Item: SlotContainer,
    Link: SlotContainer,
    Separator: SlotContainer,
    Page: SlotContainer
  };
});

import AppLayout from '../../../src/routes/(app)/+layout.svelte';

describe('authenticated app layout', () => {
  beforeEach(() => {
    state.gotoMock.mockReset();
    state.loadAuthMock.mockReset();
    state.loadAuthMock.mockResolvedValue(undefined);
  });

  it('redirects guests to /login when the backend is available', async () => {
    render(AppLayout, {
      data: {
        backendAvailable: true,
        user: null,
        teams: [],
        projects: []
      }
    });

    await waitFor(() => {
      expect(state.gotoMock).toHaveBeenCalledWith('/login');
    });
    expect(state.loadAuthMock).toHaveBeenCalledTimes(1);
  });

  it('does not redirect when the backend is already marked unavailable', async () => {
    render(AppLayout, {
      data: {
        backendAvailable: false,
        user: null,
        teams: [],
        projects: []
      }
    });

    await waitFor(() => {
      expect(state.loadAuthMock).toHaveBeenCalledTimes(1);
    });
    expect(state.gotoMock).not.toHaveBeenCalled();
  });
});
