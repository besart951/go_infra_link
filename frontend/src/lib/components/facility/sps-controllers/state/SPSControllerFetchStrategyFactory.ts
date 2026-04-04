import { ContextualSPSControllerFetchStrategy } from './strategies/ContextualSPSControllerFetchStrategy.js';

export class SPSControllerFetchStrategyFactory {
  private readonly resolveProjectId: () => string | undefined;

  constructor(resolveProjectId: () => string | undefined) {
    this.resolveProjectId = resolveProjectId;
  }

  create(): ContextualSPSControllerFetchStrategy {
    return new ContextualSPSControllerFetchStrategy(this.resolveProjectId);
  }
}
