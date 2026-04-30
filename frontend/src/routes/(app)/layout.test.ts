/// <reference types="vitest" />

const mockApi = vi.hoisted(() => vi.fn());

vi.mock('$lib/api/client', () => ({
  ApiException: class ApiException extends Error {
    constructor(
      public status: number,
      public error: string,
      message: string
    ) {
      super(message);
    }
  },
  api: mockApi
}));

import { load } from './+layout.js';

function createLoadEvent() {
  return { fetch: vi.fn() as unknown as typeof fetch } as Parameters<typeof load>[0];
}

describe('(app) layout load', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('allows a signed-in user with no permissions without preloading protected lists', async () => {
    const user = {
      id: 'user-1',
      first_name: 'No',
      last_name: 'Permissions',
      email: 'noperms@example.com',
      is_active: true,
      role: 'planer',
      permissions: [],
      can_access_user_directory: false,
      created_at: '2026-04-30T00:00:00Z',
      updated_at: '2026-04-30T00:00:00Z',
      failed_login_attempts: 0
    };

    mockApi.mockResolvedValueOnce(user);

    const result = await load(createLoadEvent());

    expect(result).toMatchObject({
      backendAvailable: true,
      user,
      teams: [],
      projects: []
    });
    expect(mockApi).toHaveBeenCalledTimes(1);
    expect(mockApi).toHaveBeenCalledWith('/auth/me', { customFetch: expect.any(Function) });
  });

  it('preloads teams and projects only when the user has the matching permissions', async () => {
    const user = {
      id: 'user-1',
      first_name: 'Project',
      last_name: 'Reader',
      email: 'reader@example.com',
      is_active: true,
      role: 'admin',
      permissions: ['team.read', 'project.listAll'],
      can_access_user_directory: false,
      created_at: '2026-04-30T00:00:00Z',
      updated_at: '2026-04-30T00:00:00Z',
      failed_login_attempts: 0
    };
    const teams = [{ id: 'team-1', name: 'Team' }];
    const projects = [{ id: 'project-1', name: 'Project' }];

    mockApi.mockResolvedValueOnce(user).mockResolvedValueOnce(teams).mockResolvedValueOnce(projects);

    const result = await load(createLoadEvent());

    expect(result).toMatchObject({ user, teams, projects });
    expect(mockApi).toHaveBeenNthCalledWith(2, '/teams', {
      customFetch: expect.any(Function),
      skipHttpErrorNavigation: true
    });
    expect(mockApi).toHaveBeenNthCalledWith(3, '/projects', {
      customFetch: expect.any(Function),
      skipHttpErrorNavigation: true
    });
  });
});
