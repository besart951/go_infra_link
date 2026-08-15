import type { Project, ProjectCapabilities } from '$lib/domain/project/index.js';

export interface ProjectContext {
  project: Project;
  capabilities: ProjectCapabilities;
}

interface ProjectContextCacheEntry {
  expiresAt: number;
  context: ProjectContext;
}

interface ProjectContextRequest {
  id: number;
  force: boolean;
  promise: Promise<ProjectContext>;
}

export interface ProjectContextCacheOptions {
  ttlMs?: number;
  now?: () => number;
}

export const PROJECT_CONTEXT_CACHE_TTL_MS = 30 * 60 * 1000;

/**
 * Keeps the authorization-sensitive project context in memory for the active
 * browser session. Entries are bounded by a TTL and are explicitly refreshed
 * after project collaboration events, so navigation does not repeat the two
 * context requests unnecessarily.
 */
export class ProjectContextCache {
  private readonly entries = new Map<string, ProjectContextCacheEntry>();
  private readonly requests = new Map<string, ProjectContextRequest>();
  private readonly ttlMs: number;
  private readonly now: () => number;
  private cacheGeneration = 0;
  private requestID = 0;

  constructor(
    private readonly fetchContext: (projectId: string) => Promise<ProjectContext>,
    options: ProjectContextCacheOptions = {}
  ) {
    this.ttlMs = options.ttlMs ?? PROJECT_CONTEXT_CACHE_TTL_MS;
    this.now = options.now ?? Date.now;
  }

  load(projectId: string): Promise<ProjectContext> {
    return this.fetch(projectId, false);
  }

  refresh(projectId: string): Promise<ProjectContext> {
    return this.fetch(projectId, true);
  }

  invalidate(projectId: string): void {
    this.entries.delete(projectId);
  }

  clear(): void {
    this.cacheGeneration += 1;
    this.entries.clear();
    this.requests.clear();
  }

  private fetch(projectId: string, force: boolean): Promise<ProjectContext> {
    const cached = this.entries.get(projectId);
    if (!force && cached && cached.expiresAt > this.now()) {
      return Promise.resolve(cloneProjectContext(cached.context));
    }

    const activeRequest = this.requests.get(projectId);
    if (activeRequest && (!force || activeRequest.force)) {
      return activeRequest.promise.then(cloneProjectContext);
    }

    const request = this.createRequest(projectId, force);
    this.requests.set(projectId, request);
    return request.promise.then(cloneProjectContext);
  }

  private createRequest(projectId: string, force: boolean): ProjectContextRequest {
    const requestID = ++this.requestID;
    const generation = this.cacheGeneration;
    const promise = this.fetchContext(projectId)
      .then((context) => {
        if (generation !== this.cacheGeneration || this.requests.get(projectId)?.id !== requestID) {
          return context;
        }
        this.entries.set(projectId, {
          context: cloneProjectContext(context),
          expiresAt: this.now() + this.ttlMs
        });
        return context;
      })
      .finally(() => {
        if (this.requests.get(projectId)?.id === requestID) {
          this.requests.delete(projectId);
        }
      });

    return { id: requestID, force, promise };
  }
}

function cloneProjectContext(context: ProjectContext): ProjectContext {
  return {
    project: {
      ...context.project,
      phase: context.project.phase ? { ...context.project.phase } : context.project.phase
    },
    capabilities: { permissions: [...context.capabilities.permissions] }
  };
}
