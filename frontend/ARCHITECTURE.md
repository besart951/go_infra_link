# Frontend Hexagonal Architecture Implementation

## Overview

This document describes the implementation of a hexagonal (clean) architecture for the frontend, following SOLID principles and best practices for maintainable code.

## Architecture Layers

### 1. Domain Layer (`src/lib/domain/`)

The domain layer contains the core business logic and entities, independent of any framework or infrastructure.

#### **Entities** (`domain/entities/`)

Pure TypeScript interfaces representing domain models:

- `building.ts`, `controlCabinet.ts`, `spsController.ts`
- `apparat.ts`, `systemPart.ts`, `spsControllerSystemType.ts`
- `objectData.ts`, `project.ts`, `team.ts`, `user.ts`

#### **Value Objects** (`domain/valueObjects/`)

Immutable objects representing domain concepts:

- `pagination.ts` - Pagination parameters and metadata
- `search.ts` - Search query value object

#### **Ports** (`domain/ports/`)

Interfaces defining contracts for infrastructure adapters:

- `listRepository.ts` - Repository interface for paginated list operations

### 2. Application Layer (`src/lib/application/`)

Orchestrates domain logic and coordinates between layers, completely framework-agnostic.

#### **Use Cases** (`application/useCases/`)

- `listUseCase.ts` - Handles list, search, and pagination logic
  - `execute()` - Fetch paginated data
  - `getById()` - Fetch single item
  - `createInitialState()` - Initialize empty state

### 3. Infrastructure Layer (`src/lib/infrastructure/`)

Implements domain ports with concrete technology choices.

#### **API Adapters** (`infrastructure/api/`)

- `apiListAdapter.ts` - Implements `ListRepository` port using the backend API
  - Generic adapter for any entity type
  - Transforms backend responses to domain models
  - Handles HTTP requests via the API client

### 4. UI Layer

#### **Stores** (`src/lib/stores/list/`)

Svelte stores that wrap use cases for reactive state management:

- `listStore.ts` - Generic factory for creating list stores
  - Reactive state management
  - Request caching (30s TTL)
  - Deduplication of identical requests
  - Debounced search (300ms)
  - Pagination controls
- `entityStores.ts` - Concrete store instances for each entity

**Features:**

- `load(searchText)` - Load first page
- `reload()` - Force refresh (bypass cache)
- `goToPage(page)` - Navigate to specific page
- `nextPage()` / `previousPage()` - Navigation
- `search(text)` - Debounced search
- `clearCache()` / `reset()` - State management

#### **Components** (`src/lib/components/list/`)

- `PaginatedList.svelte` - Generic reusable list component
  - Supports any entity type via generics
  - Search bar with debounced input
  - Pagination controls
  - Loading skeletons
  - Error handling
  - Empty state messaging

#### **Pages**

Refactored pages using the new architecture:

- `routes/(app)/teams/+page.svelte`
- `routes/(app)/users/+page.svelte`
- `routes/(app)/projects/+page.svelte`
- `routes/(app)/facility/buildings/+page.svelte`
- `routes/(app)/facility/control-cabinets/+page.svelte`
- `routes/(app)/facility/sps-controllers/+page.svelte`

## Design Principles

### SOLID Compliance

1. **Single Responsibility Principle (SRP)**
   - Each module has one reason to change
   - Use cases handle business logic
   - Adapters handle infrastructure concerns
   - Components handle presentation

2. **Open/Closed Principle (OCP)**
   - System is open for extension (new entities)
   - Closed for modification (core logic unchanged)
   - Adding new entities requires minimal code

3. **Liskov Substitution Principle (LSP)**
   - All repository implementations are interchangeable
   - Can swap backend API for mock adapter in tests

4. **Interface Segregation Principle (ISP)**
   - Repository ports define minimal interfaces
   - Clients depend only on methods they use

5. **Dependency Inversion Principle (DIP)**
   - High-level modules (use cases) don't depend on low-level modules (API adapters)
   - Both depend on abstractions (repository ports)

### Key Benefits

1. **Testability**
   - Domain logic can be tested without UI or API
   - Infrastructure can be mocked easily
   - Use cases are pure functions

2. **Maintainability**
   - Clear separation of concerns
   - Easy to locate and modify code
   - Consistent patterns across entities

3. **Extensibility**
   - New entities require minimal boilerplate
   - Can add new data sources without changing domain
   - UI components are reusable

4. **Performance**
   - Built-in caching prevents duplicate requests
   - Debounced search reduces API calls
   - AbortController cancels stale requests

## Adding a New Entity

To add support for a new entity, follow these steps:

### 1. Create Domain Entity

```typescript
// src/lib/domain/entities/newEntity.ts
export interface NewEntity {
	id: string;
	name: string;
	created_at: string;
	updated_at: string;
}
```

### 2. Add Store Instance

```typescript
// src/lib/stores/list/entityStores.ts
import type { NewEntity } from '$lib/domain/entities/newEntity.js';

export const newEntitiesStore = createListStore<NewEntity>(
	createApiAdapter<NewEntity>('/api/new-entities'),
	{ pageSize: 10 }
);
```

### 3. Create Page Component

```svelte
<script lang="ts">
	import { onMount } from 'svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { newEntitiesStore } from '$lib/stores/list/entityStores.js';
	import type { NewEntity } from '$lib/domain/entities/newEntity.js';
	import * as Table from '$lib/components/ui/table/index.js';

	onMount(() => {
		newEntitiesStore.load();
	});
</script>

<PaginatedList
	state={$newEntitiesStore}
	columns={[
		{ key: 'name', label: 'Name' },
		{ key: 'created', label: 'Created' }
	]}
	searchPlaceholder="Search entities..."
	onSearch={(text) => newEntitiesStore.search(text)}
	onPageChange={(page) => newEntitiesStore.goToPage(page)}
	onReload={() => newEntitiesStore.reload()}
>
	{#snippet rowSnippet(entity: NewEntity)}
		<Table.Cell>{entity.name}</Table.Cell>
		<Table.Cell>
			{new Date(entity.created_at).toLocaleDateString()}
		</Table.Cell>
	{/snippet}
</PaginatedList>
```

That's it! The new entity now has full pagination, search, and caching support.

## State Sharing

The store-based approach enables state sharing between different UI components:

```typescript
// Page component uses the store
const items = $newEntitiesStore.items;

// Combobox component can reuse the same store
import { newEntitiesStore } from '$lib/stores/list/entityStores.js';
// No duplicate requests - data is cached and shared
```

## Technical Details

### Caching Strategy

- **TTL**: 30 seconds by default
- **Key**: JSON.stringify({ page, searchText, pageSize })
- **Invalidation**: Manual via `reload()` or automatic on TTL expiry

### Request Handling

- **Debouncing**: Search requests debounced by 300ms
- **Cancellation**: Previous requests cancelled via AbortController
- **Error Handling**: Errors stored in state, displayed in UI

### Type Safety

- Full TypeScript support throughout
- Generic types for reusable components
- Type-safe repository ports and adapters

## Future Enhancements

Potential improvements for the architecture:

1. **Searchable Combobox Component**
   - Complete implementation for dropdown selectors
   - Reuses list store data
   - Supports lazy loading

2. **Advanced Filtering**
   - Add filter value objects
   - Support multiple filter criteria
   - Filter UI components

3. **Sorting**
   - Add sort value objects
   - Multi-column sorting support
   - Sort UI controls

4. **Offline Support**
   - IndexedDB adapter
   - Sync strategy
   - Conflict resolution

5. **Real-time Updates**
   - WebSocket adapter
   - Auto-refresh on external changes
   - Optimistic UI updates

## Conclusion

This hexagonal architecture provides a solid foundation for a maintainable, testable, and extensible frontend application. The clean separation of concerns and adherence to SOLID principles ensures the codebase remains manageable as it grows.
