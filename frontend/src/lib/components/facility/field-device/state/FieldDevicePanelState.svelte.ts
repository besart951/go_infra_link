export class FieldDevicePanelState {
  showMultiCreateForm = $state(false);
  bulkEditPanelOpen = $state(false);
  showExportPanel = $state(false);
  showFilterPanel = $state(false);
  showSpecifications = $state(false);
  expandedBacnetRows = $state<Set<string>>(new Set());
  loadingBacnetRows = $state<Set<string>>(new Set());

  openMultiCreateForm(): void {
    this.showMultiCreateForm = true;
  }

  closeMultiCreateForm(): void {
    this.showMultiCreateForm = false;
  }

  closeBulkEditPanel(): void {
    this.bulkEditPanelOpen = false;
  }

  toggleBulkEditPanel(): void {
    this.bulkEditPanelOpen = !this.bulkEditPanelOpen;
  }

  toggleExportPanel(): void {
    this.showExportPanel = !this.showExportPanel;
  }

  toggleFilterPanel(): void {
    this.showFilterPanel = !this.showFilterPanel;
  }

  toggleSpecifications(): boolean {
    this.showSpecifications = !this.showSpecifications;
    return this.showSpecifications;
  }

  toggleBacnetExpansion(deviceId: string): boolean {
    const nextExpanded = new Set(this.expandedBacnetRows);
    if (nextExpanded.has(deviceId)) {
      nextExpanded.delete(deviceId);
    } else {
      nextExpanded.add(deviceId);
    }

    this.expandedBacnetRows = nextExpanded;
    return nextExpanded.has(deviceId);
  }

  isBacnetExpanded(deviceId: string): boolean {
    return this.expandedBacnetRows.has(deviceId);
  }

  isBacnetLoading(deviceId: string): boolean {
    return this.loadingBacnetRows.has(deviceId);
  }

  markBacnetLoading(deviceId: string): void {
    const nextLoading = new Set(this.loadingBacnetRows);
    nextLoading.add(deviceId);
    this.loadingBacnetRows = nextLoading;
  }

  clearBacnetLoading(deviceId: string): void {
    const nextLoading = new Set(this.loadingBacnetRows);
    nextLoading.delete(deviceId);
    this.loadingBacnetRows = nextLoading;
  }
}
