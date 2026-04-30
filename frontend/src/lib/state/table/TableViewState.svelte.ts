export type TableDensity = 'small' | 'medium' | 'large';

export interface TableDensityOption<TDensity extends string = TableDensity> {
  value: TDensity;
  labelKey: string;
  tableClass: string;
}

export interface TableGroupingDefinition<TItem, TKey extends string> {
  key: TKey;
  labelKey: string;
  resolveId: (item: TItem) => string | undefined;
  resolveLabel: (item: TItem, groupId: string) => string;
}

export interface TableGroupNode<TItem, TKey extends string> {
  id: string;
  key: TKey;
  label: string;
  count: number;
  level: number;
  items: TItem[];
  children: TableGroupNode<TItem, TKey>[];
}

export interface TableViewPreferences<TKey extends string, TDensity extends string> {
  density?: TDensity;
  activeGroupKeys?: TKey[];
}

interface TableViewStateOptions<TItem, TKey extends string, TDensity extends string> {
  densityOptions?: TableDensityOption<TDensity>[];
  defaultDensity: TDensity;
  groupingDefinitions: TableGroupingDefinition<TItem, TKey>[];
  storageKey?: string;
}

interface StateMutationOptions {
  notify?: boolean;
}

export const DEFAULT_TABLE_DENSITY_OPTIONS: TableDensityOption[] = [
  {
    value: 'small',
    labelKey: 'field_device.view.density_small',
    tableClass:
      'text-xs [&_td]:px-1 [&_td]:py-0.5 [&_th]:h-8 [&_th]:px-1 [&_th]:py-1 [&_td_button:not([data-slot=checkbox])]:h-6 [&_td_button:not([data-slot=checkbox])]:min-h-6 [&_td_button:not([data-slot=checkbox])]:px-1.5 [&_td_button:not([data-slot=checkbox])]:py-0.5 [&_td_button:not([data-slot=checkbox])]:text-xs [&_td_input]:h-6 [&_td_input]:px-1.5 [&_td_input]:py-0.5 [&_td_input]:text-xs'
  },
  {
    value: 'medium',
    labelKey: 'field_device.view.density_medium',
    tableClass:
      'text-sm [&_td]:px-1.5 [&_td]:py-1 [&_th]:h-9 [&_th]:px-1.5 [&_th]:py-1.5 [&_td_button:not([data-slot=checkbox])]:h-7 [&_td_button:not([data-slot=checkbox])]:min-h-7 [&_td_button:not([data-slot=checkbox])]:px-2 [&_td_button:not([data-slot=checkbox])]:py-1 [&_td_input]:h-7 [&_td_input]:px-2 [&_td_input]:py-1'
  },
  {
    value: 'large',
    labelKey: 'field_device.view.density_large',
    tableClass: 'text-sm [&_td]:p-2 [&_th]:h-10 [&_th]:px-2 [&_th]:py-2'
  }
];

export class TableDensityState<TDensity extends string = TableDensity> {
  readonly options: TableDensityOption<TDensity>[];
  private readonly defaultValue: TDensity;
  value = $state<TDensity>(undefined as unknown as TDensity);

  constructor(
    options: TableDensityOption<TDensity>[],
    defaultValue: TDensity,
    private readonly onChange?: () => void
  ) {
    this.options = options;
    this.defaultValue = defaultValue;
    this.value = defaultValue;
  }

  get current(): TableDensityOption<TDensity> {
    return this.options.find((option) => option.value === this.value) ?? this.options[0];
  }

  get tableClass(): string {
    return this.current?.tableClass ?? '';
  }

  get isDefault(): boolean {
    return this.value === this.defaultValue;
  }

  set(value: TDensity, options: StateMutationOptions = {}): void {
    if (!this.options.some((option) => option.value === value)) return;
    const changed = this.value !== value;
    this.value = value;

    if (changed && options.notify !== false) {
      this.onChange?.();
    }
  }
}

export class TableGroupingState<TItem, TKey extends string> {
  activeKeys = $state<TKey[]>([]);
  collapsedGroupIds = $state<Set<string>>(new Set());

  constructor(
    readonly definitions: TableGroupingDefinition<TItem, TKey>[],
    private readonly onChange?: () => void
  ) {}

  get activeDefinitions(): TableGroupingDefinition<TItem, TKey>[] {
    const active = new Set(this.activeKeys);
    return this.definitions.filter((definition) => active.has(definition.key));
  }

  get isGrouped(): boolean {
    return this.activeKeys.length > 0;
  }

  isActive(key: TKey): boolean {
    return this.activeKeys.includes(key);
  }

  setActiveKeys(keys: TKey[], options: StateMutationOptions = {}): void {
    const nextKeys = this.definitions
      .map((definition) => definition.key)
      .filter((definitionKey) => keys.includes(definitionKey));
    const changed = nextKeys.join('|') !== this.activeKeys.join('|');

    this.activeKeys = nextKeys;
    this.collapsedGroupIds = new Set();

    if (changed && options.notify !== false) {
      this.onChange?.();
    }
  }

  toggle(key: TKey, options: StateMutationOptions = {}): void {
    const next = new Set(this.activeKeys);
    if (next.has(key)) {
      next.delete(key);
    } else {
      next.add(key);
    }

    this.setActiveKeys(
      this.definitions
        .map((definition) => definition.key)
        .filter((definitionKey) => next.has(definitionKey)),
      options
    );
  }

  groupItems(items: TItem[]): TableGroupNode<TItem, TKey>[] {
    if (!this.isGrouped) return [];
    return this.buildGroups(items, 0, []);
  }

  isGroupExpanded(groupId: string): boolean {
    return !this.collapsedGroupIds.has(groupId);
  }

  toggleGroupExpansion(groupId: string): void {
    const next = new Set(this.collapsedGroupIds);
    if (next.has(groupId)) {
      next.delete(groupId);
    } else {
      next.add(groupId);
    }
    this.collapsedGroupIds = next;
  }

  private buildGroups(
    items: TItem[],
    level: number,
    parentPath: string[]
  ): TableGroupNode<TItem, TKey>[] {
    const definition = this.activeDefinitions[level];
    if (!definition) return [];

    const buckets = new Map<string, { label: string; items: TItem[] }>();
    for (const item of items) {
      const groupId = definition.resolveId(item) ?? 'unassigned';
      const label = definition.resolveLabel(item, groupId);
      const bucket = buckets.get(groupId);

      if (bucket) {
        bucket.items.push(item);
      } else {
        buckets.set(groupId, { label, items: [item] });
      }
    }

    return [...buckets.entries()]
      .map(([groupId, bucket]) => {
        const id = [...parentPath, this.createPathPart(definition.key, groupId)].join('|');
        const children = this.buildGroups(bucket.items, level + 1, [
          ...parentPath,
          this.createPathPart(definition.key, groupId)
        ]);

        return {
          id,
          key: definition.key,
          label: bucket.label,
          count: bucket.items.length,
          level,
          items: children.length > 0 ? [] : bucket.items,
          children
        };
      })
      .sort((a, b) => a.label.localeCompare(b.label, undefined, { numeric: true }));
  }

  private createPathPart(key: TKey, groupId: string): string {
    return `${key}:${encodeURIComponent(groupId)}`;
  }
}

class TableViewPreferenceStore<TKey extends string, TDensity extends string> {
  constructor(private readonly storageKey?: string) {}

  load(): TableViewPreferences<TKey, TDensity> | undefined {
    if (!this.storageKey || typeof localStorage === 'undefined') return undefined;

    try {
      const raw = localStorage.getItem(this.storageKey);
      if (!raw) return undefined;

      const parsed = JSON.parse(raw) as Partial<TableViewPreferences<TKey, TDensity>>;
      return {
        density: typeof parsed.density === 'string' ? parsed.density : undefined,
        activeGroupKeys: Array.isArray(parsed.activeGroupKeys)
          ? parsed.activeGroupKeys.filter((key): key is TKey => typeof key === 'string')
          : undefined
      };
    } catch {
      return undefined;
    }
  }

  save(preferences: TableViewPreferences<TKey, TDensity>): void {
    if (!this.storageKey || typeof localStorage === 'undefined') return;

    try {
      localStorage.setItem(this.storageKey, JSON.stringify(preferences));
    } catch {
      // Ignore storage failures so table interactions keep working.
    }
  }
}

export class TableViewState<TItem, TKey extends string, TDensity extends string = TableDensity> {
  readonly density: TableDensityState<TDensity>;
  readonly grouping: TableGroupingState<TItem, TKey>;
  private readonly preferences: TableViewPreferenceStore<TKey, TDensity>;

  constructor(options: TableViewStateOptions<TItem, TKey, TDensity>) {
    this.preferences = new TableViewPreferenceStore(options.storageKey);
    const persist = () => this.savePreferences();

    this.density = new TableDensityState<TDensity>(
      options.densityOptions ?? (DEFAULT_TABLE_DENSITY_OPTIONS as TableDensityOption<TDensity>[]),
      options.defaultDensity,
      persist
    );
    this.grouping = new TableGroupingState(options.groupingDefinitions, persist);
    this.restorePreferences();
  }

  get tableClass(): string {
    return this.density.tableClass;
  }

  get isCustomized(): boolean {
    return !this.density.isDefault || this.grouping.isGrouped;
  }

  groupItems(items: TItem[]): TableGroupNode<TItem, TKey>[] {
    return this.grouping.groupItems(items);
  }

  private restorePreferences(): void {
    const saved = this.preferences.load();
    if (!saved) return;

    if (saved.density) {
      this.density.set(saved.density, { notify: false });
    }

    if (saved.activeGroupKeys) {
      this.grouping.setActiveKeys(saved.activeGroupKeys, { notify: false });
    }
  }

  private savePreferences(): void {
    this.preferences.save({
      density: this.density.value,
      activeGroupKeys: this.grouping.activeKeys
    });
  }
}
