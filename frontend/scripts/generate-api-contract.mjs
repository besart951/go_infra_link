import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import openapiTS, { astToString } from 'openapi-typescript';
import swagger2openapi from 'swagger2openapi';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const frontendDir = resolve(scriptDir, '..');
const repositoryDir = resolve(frontendDir, '..');
const swaggerPath = resolve(repositoryDir, 'backend/docs/swagger.json');
const generatedDir = resolve(frontendDir, 'src/lib/api/generated');
const openapiPath = resolve(generatedDir, 'openapi.json');
const schemaPath = resolve(generatedDir, 'schema.ts');
const clientPath = resolve(generatedDir, 'client.ts');

const swagger = JSON.parse(await readFile(swaggerPath, 'utf8'));
const { openapi } = await swagger2openapi.convertObj(swagger, {
  patch: true,
  targetVersion: '3.0.3'
});

await mkdir(generatedDir, { recursive: true });
await writeFile(openapiPath, `${JSON.stringify(openapi, null, 2)}\n`);

const schemaAst = await openapiTS(openapi, {
  alphabetize: true,
  exportType: true
});
const schema = astToString(schemaAst).trimEnd();

await writeFile(
  schemaPath,
  [
    '/* eslint-disable */',
    '// This file is generated from backend/docs/swagger.json. Do not edit manually.',
    '',
    schema,
    ''
  ].join('\n')
);

await writeFile(
  clientPath,
  [
    '/* eslint-disable */',
    '// This file is generated from backend/docs/swagger.json. Do not edit manually.',
    '',
    "import createClient from 'openapi-fetch';",
    "import { apiFetch, assertApiSuccess } from '../client.js';",
    "import type { paths } from './schema.js';",
    '',
    "const baseUrl = typeof window === 'undefined' ? 'http://localhost' : window.location.origin;",
    '',
    'export const apiClient = createClient<paths>({',
    '  baseUrl,',
    '  fetch: apiFetch',
    '});',
    '',
    'apiClient.use({',
    '  onResponse: async ({ response }) => assertApiSuccess(response)',
    '});',
    ''
  ].join('\n')
);
