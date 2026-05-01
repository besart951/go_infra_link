import type { DataTableFetchStrategy, DataTableQuery, TableFilterRecord } from './contracts.js';

type ProjectStrategyFactory<TProjectStrategy> = (projectId: string) => TProjectStrategy;

export class ContextualDataTableFetchStrategy<
  TItem,
  TFilters extends TableFilterRecord,
  TProjectStrategy extends DataTableFetchStrategy<TItem, TFilters>
> implements DataTableFetchStrategy<TItem, TFilters> {
  private projectStrategy: TProjectStrategy | null = null;
  private activeProjectId: string | undefined;

  constructor(
    private readonly resolveProjectId: () => string | undefined,
    private readonly facilityStrategy: DataTableFetchStrategy<TItem, TFilters>,
    private readonly createProjectStrategy: ProjectStrategyFactory<TProjectStrategy>
  ) {}

  fetch(query: DataTableQuery<TFilters>, signal?: AbortSignal) {
    return this.getActiveStrategy().fetch(query, signal);
  }

  protected getActiveProjectStrategy(): TProjectStrategy | null {
    return this.projectStrategy;
  }

  private getActiveStrategy(): DataTableFetchStrategy<TItem, TFilters> {
    const projectId = this.resolveProjectId();
    if (!projectId) {
      this.activeProjectId = undefined;
      return this.facilityStrategy;
    }

    if (!this.projectStrategy || this.activeProjectId !== projectId) {
      this.projectStrategy = this.createProjectStrategy(projectId);
      this.activeProjectId = projectId;
    }

    return this.projectStrategy;
  }
}
