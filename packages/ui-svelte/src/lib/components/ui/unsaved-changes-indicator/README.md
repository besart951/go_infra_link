# Unsaved Changes System - Architecture Documentation

## 📐 Hexagonal Architecture Overview

Die Implementierung folgt strikt der hexagonalen Architektur (Ports & Adapters) mit klarer Separation of Concerns:

```
┌─────────────────────────────────────────────────────┐
│                   Presentation Layer                 │
│  (UI Components - Svelte Components)                │
│  • FieldDeviceListView.svelte                       │
│  • FieldDeviceFloatingSaveBar.svelte                │
│  • UnsavedChangesIndicator.svelte                   │
└────────────────────┬────────────────────────────────┘
                     │
┌────────────────────┴────────────────────────────────┐
│                  Application Layer                   │
│  (Hooks/Composables - Business Logic)               │
│  • useFieldDeviceEditing.svelte.ts                  │
│  • useUnsavedChangesWarning.svelte.ts               │
└────────────────────┬────────────────────────────────┘
                     │
┌────────────────────┴────────────────────────────────┐
│                 Infrastructure Layer                 │
│  (Services - External Dependencies)                  │
│  • sessionStorageService.ts                         │
└─────────────────────────────────────────────────────┘
```

## 🎯 Implementierte Lösungen

### 1. SessionStorage Persistierung ✅

**Zweck**: Änderungen überleben Seitenwechsel innerhalb der Browser-Session

**Files**:

- `lib/services/sessionStorageService.ts` - Infrastructure Layer
- `lib/hooks/useFieldDeviceEditing.svelte.ts` - Erweitert mit Persistierung

**Features**:

- ✅ Automatisches Speichern bei jeder Änderung ($effect)
- ✅ Automatisches Laden beim Initialisieren
- ✅ Stale-Data Protection (24h TTL)
- ✅ Projekt-spezifische Storage Keys
- ✅ SSR-kompatibel (graceful degradation)
- ✅ Fehlerresistenz (try-catch wrapping)

**Usage**:

```typescript
// Automatisch aktiviert in useFieldDeviceEditing
const editing = useFieldDeviceEditing(projectId);
// Edits werden automatisch in sessionStorage persistiert
```

### 2. Browser-Warnung für ungespeicherte Änderungen ✅

**Zweck**: Verhindert versehentliches Verlassen bei ungespeicherten Änderungen

**Files**:

- `lib/hooks/useUnsavedChangesWarning.svelte.ts` - Wiederverwendbares Hook

**Features**:

- ✅ Native Browser beforeunload Warning
- ✅ Funktioniert bei: Tab schliessen, Browser schliessen, Seite neu laden
- ✅ Opt-in/opt-out Support
- ✅ Custom Nachrichten (legacy Browser)

**Usage**:

```typescript
const editing = useFieldDeviceEditing();

// Aktiviere Browser-Warnung
useUnsavedChangesWarning(() => editing.hasUnsavedChanges);
```

### 3. Visueller Hinweis im UI ✅

**Zweck**: Klare Kommunikation des Persistierungs-Status

**Files**:

- `lib/components/facility/FieldDeviceFloatingSaveBar.svelte` - Erweitert
- `lib/components/ui/unsaved-changes-indicator/UnsavedChangesIndicator.svelte` - Neue wiederverwendbare Komponente

**Features**:

- ✅ Floating Save Bar mit Status-Message
- ✅ Generische wiederverwendbare Indicator-Komponente
- ✅ 3 Varianten: badge, inline, card
- ✅ shadcn-Style Design

**Usage**:

```svelte
<!-- In FieldDeviceListView -->
<FieldDeviceFloatingSaveBar {editing} onSave={...} onDiscard={...} />

<!-- Standalone Indicator (wiederverwendbar) -->
<UnsavedChangesIndicator
  count={editing.pendingCount}
  variant="card"
/>
```

## 🏗️ Clean Code Prinzipien

### 1. Single Responsibility Principle (SRP)

- **SessionStorageService**: Nur Storage-Operationen
- **useUnsavedChangesWarning**: Nur Browser-Warnungen
- **useFieldDeviceEditing**: Nur Edit-State Management
- **UnsavedChangesIndicator**: Nur UI-Darstellung

### 2. Dependency Inversion Principle (DIP)

```typescript
// Interface (Port)
export interface SessionStorageAdapter {
  save<T>(key: string, value: T): void;
  load<T>(key: string): T | null;
  // ...
}

// Implementation (Adapter)
export class BrowserSessionStorage implements SessionStorageAdapter {
  // ...
}
```

### 3. Open/Closed Principle (OCP)

- Komponenten sind offen für Erweiterungen (Props, Varianten)
- Aber geschlossen für Modifikationen (stabile Interfaces)

### 4. Interface Segregation Principle (ISP)

- Kleine, fokussierte Interfaces
- Keine "fat interfaces"

## 📦 Wiederverwendbare Komponenten

### UnsavedChangesIndicator

**Varianten**:

```svelte
<!-- Kompakter Badge -->
<UnsavedChangesIndicator count={3} variant="badge" />

<!-- Inline Text -->
<UnsavedChangesIndicator count={3} variant="inline" />

<!-- Card mit Details -->
<UnsavedChangesIndicator count={3} variant="card" message="Custom message" />
```

**Anwendungsfälle**:

- Header/Navbar Status-Anzeige
- Inline in Formularen
- Standalone Warnings

### SessionStorageService

**Wiederverwendbar für alle Storage-Bedürfnisse**:

```typescript
import { sessionStorage } from '$lib/services/sessionStorageService';

// Speichern
sessionStorage.save('my-key', { data: 'value' });

// Laden
const data = sessionStorage.load<MyType>('my-key');

// Prüfen
if (sessionStorage.has('my-key')) { ... }

// Löschen
sessionStorage.remove('my-key');
```

### useUnsavedChangesWarning

**Universal für jeden unsaved-state**:

```typescript
const formState = useFormState();
useUnsavedChangesWarning(() => formState.isDirty);

const editingState = useTableEditing();
useUnsavedChangesWarning(() => editingState.hasChanges);
```

## 🔄 Data Flow

```
User Edit
    ↓
queueEdit() / queueBacnetEdit()
    ↓
pendingEdits Map update (reactive $state)
    ↓
$effect triggers
    ↓
savePersistedState()
    ↓
sessionStorage.save()
    ↓
Browser SessionStorage
```

**Beim Page Load**:

```
Component Mount
    ↓
useFieldDeviceEditing(projectId) init
    ↓
loadPersistedState(storageKey)
    ↓
sessionStorage.load()
    ↓
Check staleness (24h TTL)
    ↓
Initialize pendingEdits Map
    ↓
User sieht wiederhergestellte Edits
```

## 🧪 Testing Considerations

### Unit Tests

```typescript
// sessionStorageService.test.ts
describe("BrowserSessionStorage", () => {
  it("should save and load data");
  it("should handle SSR gracefully");
  it("should clear stale data");
});

// useFieldDeviceEditing.test.ts
describe("useFieldDeviceEditing", () => {
  it("should persist edits to sessionStorage");
  it("should restore edits on init");
  it("should clear storage on discard");
});
```

### Integration Tests

```typescript
// FieldDeviceListView.test.ts
describe("FieldDeviceListView with persistence", () => {
  it("should restore edits after navigation");
  it("should show browser warning on leave");
  it("should clear storage after save");
});
```

## 🚀 Erweiterungsmöglichkeiten

### LocalStorage für permanente Persistierung

```typescript
export class BrowserLocalStorage implements SessionStorageAdapter {
  // Same interface, different implementation
}
```

### Backend Draft System

```typescript
export class BackendDraftStorage implements SessionStorageAdapter {
  async save<T>(key: string, value: T) {
    await api.saveDraft(key, value);
  }
}
```

### Multi-Tab Synchronisation

```typescript
// Mit BroadcastChannel API
export class SyncedSessionStorage implements SessionStorageAdapter {
  private channel = new BroadcastChannel("storage-sync");
  // ...
}
```

## 📝 Best Practices

1. **Immer mit projectId arbeiten** für projekt-spezifische Isolation:

   ```typescript
   const editing = useFieldDeviceEditing(projectId);
   ```

2. **TTL für Stale-Data setzen** (aktuell 24h):

   ```typescript
   const MAX_AGE_MS = 24 * 60 * 60 * 1000;
   ```

3. **Storage bei Success clearen**:

   ```typescript
   if (saveSuccess) {
     sessionStorage.remove(storageKey);
   }
   ```

4. **Graceful Degradation** bei fehlender Browser-Unterstützung:
   ```typescript
   if (typeof sessionStorage === "undefined") return null;
   ```

## 🎨 UI/UX Guidelines

- ✅ Klare visuelle Kennzeichnung ungespeicherter Änderungen
- ✅ Eindeutige Action-Buttons (Save All / Discard)
- ✅ Informative Tooltip-Messages
- ✅ Konsistente Farbgebung (amber für warnings)
- ✅ Accessibility: Screen-reader friendly

## 🔐 Security Considerations

- ⚠️ SessionStorage ist **nicht encrypted**
- ⚠️ Keine sensitiven Daten (Passwords, Tokens) speichern
- ✅ Nur UI-State und Form-Edits
- ✅ Automatischer Cleanup bei Tab-Close
- ✅ TTL für data staleness

## 📚 Weiterführende Ressourcen

- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
- [Web Storage API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Storage_API)
- [BeforeUnload Event](https://developer.mozilla.org/en-US/docs/Web/API/Window/beforeunload_event)
