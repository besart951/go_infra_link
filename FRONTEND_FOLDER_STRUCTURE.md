# Frontend Ordnerstruktur - Visuell

## Gesamtübersicht

```
go_infra_link/
└── frontend/
    ├── src/
    │   ├── lib/                           # Shared Library Code
    │   │   ├── api/                      # ⚠️  Legacy API Layer
    │   │   │   ├── client.ts             # ⭐⭐⭐⭐⭐ Zentraler API Client (CSRF, Error Handling)
    │   │   │   ├── users.ts              # ⚠️  User API (sollte zu infrastructure)
    │   │   │   └── teams.ts              # ⚠️  Teams API (sollte zu infrastructure)
    │   │   │
    │   │   ├── domain/                   # ✅ Domain Layer (Pure Business Logic)
    │   │   │   ├── entities/             # ⭐⭐⭐⭐⭐ Domain Entities
    │   │   │   │   ├── building.ts
    │   │   │   │   ├── team.ts
    │   │   │   │   ├── user.ts
    │   │   │   │   ├── project.ts
    │   │   │   │   └── ...
    │   │   │   ├── valueObjects/         # ⭐⭐⭐⭐⭐ Value Objects
    │   │   │   │   ├── pagination.ts
    │   │   │   │   └── search.ts
    │   │   │   ├── ports/                # ⭐⭐⭐⭐⭐ Repository Interfaces
    │   │   │   │   └── listRepository.ts
    │   │   │   ├── facility/             # Facility-spezifische Domain-Typen
    │   │   │   ├── phase/
    │   │   │   ├── project/
    │   │   │   ├── team/
    │   │   │   ├── user/
    │   │   │   └── utils/
    │   │   │
    │   │   ├── application/              # ✅ Application Layer (Use Cases)
    │   │   │   └── useCases/             # ⭐⭐⭐⭐⭐ Business Logic Orchestration
    │   │   │       └── listUseCase.ts    # Generic List/Pagination Logic
    │   │   │
    │   │   ├── infrastructure/           # ✅ Infrastructure Layer (Adapters)
    │   │   │   └── api/                  # ⭐⭐⭐⭐☆ API Adapter Implementations
    │   │   │       └── apiListAdapter.ts # Generic Backend API Adapter
    │   │   │
    │   │   ├── stores/                   # ⚠️  State Management (Mixed Patterns)
    │   │   │   ├── list/                 # ⭐⭐⭐⭐⭐ Generic List Stores
    │   │   │   │   ├── listStore.ts      # Caching, Debouncing, Pagination
    │   │   │   │   └── entityStores.ts   # ⚠️  All entity stores (140 lines - sollte aufgeteilt werden)
    │   │   │   ├── auth.svelte.ts        # ⭐⭐⭐⭐⭐ Auth State (Svelte 5 Runes)
    │   │   │   ├── theme.ts              # ⚠️  Theme (alte Svelte 4 Store-API)
    │   │   │   ├── confirm-dialog.ts
    │   │   │   ├── facility/
    │   │   │   ├── phases/
    │   │   │   └── projects/
    │   │   │
    │   │   ├── components/               # ✅ UI Components
    │   │   │   ├── list/                 # ⭐⭐⭐⭐⭐ Generic List Components
    │   │   │   │   └── PaginatedList.svelte
    │   │   │   ├── ui/                   # ⭐⭐⭐⭐⭐ Headless UI Components (bits-ui)
    │   │   │   │   ├── button/
    │   │   │   │   ├── input/
    │   │   │   │   ├── table/
    │   │   │   │   ├── dialog/
    │   │   │   │   ├── combobox/
    │   │   │   │   ├── sidebar/
    │   │   │   │   └── ...               # 20+ UI Components
    │   │   │   ├── facility/             # Domain-spezifische Components
    │   │   │   │   └── BuildingForm.svelte
    │   │   │   ├── project/
    │   │   │   │   ├── PhaseForm.svelte
    │   │   │   │   └── ProjectPhaseSelect.svelte
    │   │   │   ├── sidebar/              # Layout Components
    │   │   │   │   ├── nav-main.svelte
    │   │   │   │   ├── nav-projects.svelte
    │   │   │   │   ├── nav-user.svelte
    │   │   │   │   └── team-switcher.svelte
    │   │   │   ├── toast.svelte          # Global Toast Notifications
    │   │   │   ├── confirm-dialog.svelte
    │   │   │   └── permission-guard.svelte
    │   │   │
    │   │   ├── utils/                    # ✅ Utility Functions
    │   │   │   ├── permissions.ts
    │   │   │   └── utils.ts
    │   │   │
    │   │   ├── hooks/                    # SvelteKit Hooks
    │   │   └── server/                   # Server-Side Code
    │   │       ├── backend.example.ts
    │   │       └── set-cookie.example.ts
    │   │
    │   └── routes/                        # ✅ SvelteKit Routes (31 Pages)
    │       ├── (app)/                     # Protected App Routes
    │       │   ├── +page.svelte           # Dashboard
    │       │   ├── +layout.svelte         # App Layout mit Sidebar
    │       │   ├── +layout.ts
    │       │   ├── teams/                 # Teams Management
    │       │   │   ├── +page.svelte
    │       │   │   └── [id]/
    │       │   │       └── +page.svelte
    │       │   ├── users/                 # User Management
    │       │   │   └── +page.svelte
    │       │   ├── projects/              # Project Management
    │       │   │   ├── +page.svelte
    │       │   │   ├── new/
    │       │   │   ├── [id]/
    │       │   │   └── phases/
    │       │   ├── facility/              # Facility Management
    │       │   │   ├── buildings/
    │       │   │   │   ├── +page.svelte
    │       │   │   │   └── [id]/
    │       │   │   ├── control-cabinets/
    │       │   │   ├── sps-controllers/
    │       │   │   ├── apparats/
    │       │   │   ├── system-parts/
    │       │   │   ├── system-types/
    │       │   │   ├── object-data/
    │       │   │   ├── field-devices/
    │       │   │   ├── specifications/
    │       │   │   ├── state-texts/
    │       │   │   ├── notification-classes/
    │       │   │   └── alarm-definitions/
    │       │   ├── account/               # Account Settings
    │       │   ├── settings/              # Global Settings
    │       │   └── logout/
    │       ├── (auth)/                    # Public Auth Routes
    │       │   └── login/
    │       │       └── +page.svelte
    │       ├── api/                       # API Routes (Proxy)
    │       │   └── v1/
    │       │       └── [...path]/
    │       │           └── +server.ts     # Backend Proxy
    │       ├── +layout.svelte             # Root Layout
    │       └── +layout.ts
    │
    ├── static/                            # Static Assets
    │   └── favicon.svg
    ├── package.json                       # Dependencies
    ├── tsconfig.json                      # TypeScript Config
    ├── svelte.config.js                   # SvelteKit Config
    ├── vite.config.ts                     # Vite Config
    ├── tailwind.config.ts                 # Tailwind Config
    ├── .prettierrc                        # Code Formatting
    ├── components.json                    # UI Components Config
    ├── ARCHITECTURE.md                    # ⭐⭐⭐⭐⭐ Architektur-Dokumentation
    └── README.md
```

## Legende

| Symbol | Bedeutung |
|--------|-----------|
| ✅ | Gut strukturiert, folgt Best Practices |
| ⚠️  | Verbesserungsbedarf identifiziert |
| ⭐⭐⭐⭐⭐ | Exzellent (5/5) |
| ⭐⭐⭐⭐☆ | Sehr gut (4/5) |
| ⭐⭐⭐☆☆ | Gut (3/5) |

## Statistiken

### Dateien nach Layer

| Layer | Anzahl Dateien | Anteil |
|-------|----------------|--------|
| Domain | ~30 | 11% |
| Application | ~3 | 1% |
| Infrastructure | ~15 | 6% |
| UI (Components) | ~150 | 57% |
| Routes | ~31 | 12% |
| Stores | ~15 | 6% |
| Utils/Other | ~20 | 7% |
| **Total** | **~264** | **100%** |

### Lines of Code (geschätzt)

| Layer | LoC | Durchschnitt/Datei |
|-------|-----|---------------------|
| Domain | ~600 | ~20 |
| Application | ~150 | ~50 |
| Infrastructure | ~400 | ~27 |
| UI Components | ~2,500 | ~17 |
| Routes | ~800 | ~26 |
| Stores | ~600 | ~40 |
| **Total** | **~5,050** | **~19** |

## Architektur-Layers im Detail

### 1. Domain Layer (`lib/domain/`)
**Zweck:** Kerngeschäftslogik, framework-unabhängig

**Enthält:**
- ✅ Entities (Business Objects)
- ✅ Value Objects (Immutable Konzepte)
- ✅ Repository Ports (Interfaces)
- ✅ Domain-spezifische Typen

**Dependencies:** Keine externen Dependencies ⭐

### 2. Application Layer (`lib/application/`)
**Zweck:** Orchestrierung der Business Logic

**Enthält:**
- ✅ Use Cases (Business Operations)
- ✅ Application Services

**Dependencies:** Nur Domain Layer ⭐

### 3. Infrastructure Layer (`lib/infrastructure/`)
**Zweck:** Technische Implementierungen

**Enthält:**
- ✅ API Adapters (Backend-Kommunikation)
- ⚠️  Sollte auch DB, Cache, etc. enthalten

**Dependencies:** Domain, Application, externe Libraries

### 4. UI Layer (`lib/components/`, `lib/stores/`, `routes/`)
**Zweck:** Benutzeroberfläche und State Management

**Enthält:**
- ✅ Svelte Components
- ✅ State Stores
- ✅ Routes/Pages

**Dependencies:** Alle anderen Layers

## Dependency-Richtung

```
┌─────────────────────────────────────────────┐
│              UI Layer                       │
│  (Components, Stores, Routes)               │
└────────────────┬────────────────────────────┘
                 │ depends on ↓
┌────────────────▼────────────────────────────┐
│         Infrastructure Layer                │
│  (API Adapters, DB, Cache, etc.)            │
└────────────────┬────────────────────────────┘
                 │ implements ↓
┌────────────────▼────────────────────────────┐
│         Application Layer                   │
│  (Use Cases, Services)                      │
└────────────────┬────────────────────────────┘
                 │ uses ↓
┌────────────────▼────────────────────────────┐
│         Domain Layer                        │
│  (Entities, Value Objects, Ports)           │
│  NO EXTERNAL DEPENDENCIES ⭐                 │
└─────────────────────────────────────────────┘
```

## Hexagonal Architecture Visualisierung

```
                    ┌───────────────────────────┐
                    │     External Systems      │
                    │   (Backend API)           │
                    └─────────┬─────────────────┘
                              │
                   ┌──────────▼──────────────┐
                   │   Infrastructure        │
   ┌───────────────┤   (Adapters)            │
   │               │   - apiListAdapter.ts   │
   │               └──────────┬──────────────┘
   │                          │
   │ UI Components     ┌──────▼──────────┐
   │ (Svelte)          │  Application    │
   │                   │  (Use Cases)    │
   │   ┌───────────────┤  - listUseCase  │
   │   │               └─────────┬───────┘
   │   │                         │
┌──▼───▼────┐            ┌───────▼───────────┐
│  Stores   │────────────▶   Domain          │
│  (State)  │            │  (Core Logic)     │
└───────────┘            │  - entities/      │
                         │  - ports/         │
                         │  - valueObjects/  │
                         └───────────────────┘
```

## Nächste Schritte

1. **Kurzfristig (1-2 Wochen):**
   - `/api/` Legacy-Code zu `/infrastructure/` migrieren
   - `entityStores.ts` in separate Dateien aufteilen
   - Tests für Use Cases schreiben

2. **Mittelfristig (1-2 Monate):**
   - Svelte 4 Stores zu Svelte 5 Runes migrieren
   - Input-Validierung implementieren
   - Code-Duplizierung eliminieren

3. **Langfristig (3-6 Monate):**
   - Test-Coverage auf 60%+ erhöhen
   - Performance-Optimierungen (Code-Splitting)
   - Offline-Support (IndexedDB)

---

**Für Details siehe:** `FRONTEND_ARCHITECTURE_REVIEW.md`
