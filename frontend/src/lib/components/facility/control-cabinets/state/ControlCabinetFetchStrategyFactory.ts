import { ContextualControlCabinetFetchStrategy } from './strategies/ContextualControlCabinetFetchStrategy.js';

export class ControlCabinetFetchStrategyFactory {
  private readonly resolveProjectId: () => string | undefined;

  constructor(resolveProjectId: () => string | undefined) {
    this.resolveProjectId = resolveProjectId;
  }

  create(): ContextualControlCabinetFetchStrategy {
    return new ContextualControlCabinetFetchStrategy(this.resolveProjectId);
  }
}
