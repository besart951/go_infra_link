import { access, readFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { expect, test } from './fixtures';
import { facilityProjectMutationCoverage } from './crud-coverage';

const mutationMethods = new Set(['post', 'put', 'patch', 'delete']);

test('every documented facility and project mutation has an acceptance scenario', async () => {
  const contractPath = new URL('../src/lib/api/generated/openapi.json', import.meta.url);
  const contract = JSON.parse(await readFile(contractPath, 'utf8')) as {
    paths: Record<string, Record<string, unknown>>;
  };
  const mutations = Object.entries(contract.paths).flatMap(([path, operations]) =>
    Object.keys(operations)
      .filter((method) => mutationMethods.has(method))
      .filter(
        () =>
          path.startsWith('/api/v1/facility/') ||
          path.startsWith('/api/v1/projects') ||
          path.startsWith('/api/v1/phases')
      )
      .map((method) => `${method.toUpperCase()} ${path}`)
  );

  expect(mutations.sort()).toEqual(Object.keys(facilityProjectMutationCoverage).sort());
});

test('every mapped acceptance scenario or handler contract exists', async () => {
  const targets = [...new Set(Object.values(facilityProjectMutationCoverage))];
  await Promise.all(
    targets.map((target) => {
      const file = target.startsWith('backend/')
        ? resolve(process.cwd(), '..', target)
        : resolve(process.cwd(), 'e2e', target);
      return expect(access(file)).resolves.toBeUndefined();
    })
  );
});
