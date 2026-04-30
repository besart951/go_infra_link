import { describe, expect, it } from 'vitest';
import { readdirSync, readFileSync, statSync } from 'node:fs';
import path from 'node:path';

const srcRoot = path.join(process.cwd(), 'src');
const routeRoot = path.join(srcRoot, 'routes');

function listSourceFiles(dir = srcRoot): string[] {
  return readdirSync(dir).flatMap((entry) => {
    const fullPath = path.join(dir, entry);
    const stats = statSync(fullPath);

    if (stats.isDirectory()) {
      return listSourceFiles(fullPath);
    }

    return /\.(svelte|ts)$/.test(entry) && !/\.test\.ts$/.test(entry) ? [fullPath] : [];
  });
}

function listRouteSourceFiles(): string[] {
  return listSourceFiles(routeRoot);
}

function toRoutePath(filePath: string): string {
  return path.relative(process.cwd(), filePath).replaceAll(path.sep, '/');
}

function findMatchingImports(pattern: RegExp, files = listRouteSourceFiles()): string[] {
  return files
    .map((filePath) => ({
      path: toRoutePath(filePath),
      text: readFileSync(filePath, 'utf8')
    }))
    .filter((file) => pattern.test(file.text))
    .map((file) => file.path)
    .sort();
}

describe('route import boundaries', () => {
  it('keeps infrastructure adapters out of route modules', () => {
    expect(findMatchingImports(/\$lib\/infrastructure\/api/)).toEqual([]);
  });

  it('keeps legacy user and team endpoint modules out of route modules', () => {
    expect(findMatchingImports(/\$lib\/api\/(?:users|teams)/)).toEqual([]);
  });

  it('keeps legacy user and team endpoint modules behind repository wrappers', () => {
    expect(findMatchingImports(/\$lib\/api\/(?:users|teams)/, listSourceFiles())).toEqual([
      'src/lib/infrastructure/api/teamRepository.ts',
      'src/lib/infrastructure/api/userRepository.ts'
    ]);
  });

  it('limits direct api client imports to auth and load-module composition points', () => {
    expect(findMatchingImports(/\$lib\/api\/client/)).toEqual([
      'src/routes/(app)/+layout.ts',
      'src/routes/(app)/facility/buildings/[id]/+page.svelte',
      'src/routes/(app)/facility/buildings/[id]/+page.ts',
      'src/routes/(app)/facility/control-cabinets/[id]/+page.ts',
      'src/routes/(app)/facility/sps-controller-system-type/[id]/+page.ts',
      'src/routes/(app)/facility/sps-controllers/[id]/+page.ts',
      'src/routes/(app)/logout/+page.svelte',
      'src/routes/(auth)/login/+page.svelte'
    ]);
  });
});
