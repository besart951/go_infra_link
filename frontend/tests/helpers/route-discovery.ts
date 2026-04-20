import { readdirSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

export type DiscoveredRoute = {
  path: string;
  files: string[];
};

const currentFile = fileURLToPath(import.meta.url);
const currentDir = path.dirname(currentFile);
const frontendRoot = path.resolve(currentDir, '..', '..');
const routesRoot = path.resolve(frontendRoot, 'src', 'routes');

function toPosix(filePath: string): string {
  return filePath.split(path.sep).join('/');
}

function collectPageFiles(dirPath: string): string[] {
  const entries = readdirSync(dirPath, { withFileTypes: true });

  return entries
    .flatMap((entry) => {
      const absolutePath = path.join(dirPath, entry.name);

      if (entry.isDirectory()) {
        return collectPageFiles(absolutePath);
      }

      return /\+page\.(svelte|ts)$/.test(entry.name) ? [absolutePath] : [];
    })
    .sort((left, right) => left.localeCompare(right));
}

function toRoutePath(relativeFilePath: string): string {
  const segments = relativeFilePath.split('/');
  const routeSegments = segments
    .slice(2, -1)
    .filter((segment) => !(segment.startsWith('(') && segment.endsWith(')')))
    .map((segment) =>
      segment.startsWith('[') && segment.endsWith(']') ? `:${segment.slice(1, -1)}` : segment
    );

  return routeSegments.length === 0 ? '/' : `/${routeSegments.join('/')}`;
}

export function discoverRoutePages(): DiscoveredRoute[] {
  const byRoute = new Map<string, string[]>();

  for (const absoluteFilePath of collectPageFiles(routesRoot)) {
    const relativePath = toPosix(path.relative(frontendRoot, absoluteFilePath));
    const routePath = toRoutePath(relativePath);
    const files = byRoute.get(routePath) ?? [];

    files.push(relativePath);
    files.sort((left, right) => left.localeCompare(right));
    byRoute.set(routePath, files);
  }

  return [...byRoute.entries()]
    .map(([routePath, files]) => ({ path: routePath, files }))
    .sort((left, right) => left.path.localeCompare(right.path));
}
