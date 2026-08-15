import { ProjectContextCache, PROJECT_CONTEXT_CACHE_TTL_MS } from './projectContextCache.js';
import type { ProjectContext } from './projectContextCache.js';

const projectId = 'project-1';

function context(name: string, permissions = ['project.fielddevice.read']): ProjectContext {
  return {
    project: {
      id: projectId,
      name,
      description: '',
      status: 'planned',
      phase_id: 'phase-1',
      creator_id: 'creator-1',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z'
    },
    capabilities: { permissions: permissions as ProjectContext['capabilities']['permissions'] }
  };
}

describe('ProjectContextCache', () => {
  it('serves one project context from the cache for 30 minutes', async () => {
    let now = 0;
    const fetchContext = vi.fn().mockResolvedValue(context('Initial'));
    const cache = new ProjectContextCache(fetchContext, { now: () => now });

    await cache.load(projectId);
    now = PROJECT_CONTEXT_CACHE_TTL_MS - 1;
    const cached = await cache.load(projectId);

    expect(fetchContext).toHaveBeenCalledOnce();
    expect(cached.project.name).toBe('Initial');

    now += 1;
    await cache.load(projectId);

    expect(fetchContext).toHaveBeenCalledTimes(2);
  });

  it('deduplicates concurrent navigation requests', async () => {
    let resolveContext: ((value: ProjectContext) => void) | undefined;
    const fetchContext = vi.fn(
      () =>
        new Promise<ProjectContext>((resolve) => {
          resolveContext = resolve;
        })
    );
    const cache = new ProjectContextCache(fetchContext);

    const first = cache.load(projectId);
    const second = cache.load(projectId);
    resolveContext?.(context('Initial'));

    await expect(Promise.all([first, second])).resolves.toHaveLength(2);
    expect(fetchContext).toHaveBeenCalledOnce();
  });

  it('keeps the newer realtime refresh when an older navigation request finishes later', async () => {
    let resolveInitial: ((value: ProjectContext) => void) | undefined;
    let resolveRefresh: ((value: ProjectContext) => void) | undefined;
    const fetchContext = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise<ProjectContext>((resolve) => {
            resolveInitial = resolve;
          })
      )
      .mockImplementationOnce(
        () =>
          new Promise<ProjectContext>((resolve) => {
            resolveRefresh = resolve;
          })
      );
    const cache = new ProjectContextCache(fetchContext);

    const initial = cache.load(projectId);
    const refresh = cache.refresh(projectId);
    resolveRefresh?.(context('Realtime', ['project.update']));
    resolveInitial?.(context('Stale'));

    await Promise.all([initial, refresh]);
    const cached = await cache.load(projectId);

    expect(cached.project.name).toBe('Realtime');
    expect(cached.capabilities.permissions).toEqual(['project.update']);
  });

  it('clears cached context on session teardown', async () => {
    const fetchContext = vi
      .fn()
      .mockResolvedValueOnce(context('First session'))
      .mockResolvedValueOnce(context('Second session'));
    const cache = new ProjectContextCache(fetchContext);

    await cache.load(projectId);
    cache.clear();
    const next = await cache.load(projectId);

    expect(fetchContext).toHaveBeenCalledTimes(2);
    expect(next.project.name).toBe('Second session');
  });
});
