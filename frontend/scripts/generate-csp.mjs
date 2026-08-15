import { createHash } from 'node:crypto';
import { readdir, readFile, writeFile } from 'node:fs/promises';
import { join } from 'node:path';

const [buildDirectory, templatePath, targetPath] = process.argv.slice(2);

if (!buildDirectory || !templatePath || !targetPath) {
  throw new Error('Usage: node generate-csp.mjs <build-directory> <template> <target>');
}

const htmlFiles = await findHtmlFiles(buildDirectory);
const hashes = new Set();

for (const filePath of htmlFiles) {
  const html = await readFile(filePath, 'utf8');
  for (const script of inlineScripts(html)) {
    hashes.add(`'sha256-${createHash('sha256').update(script).digest('base64')}'`);
  }
}

if (hashes.size === 0) {
  throw new Error(
    `No inline scripts found in ${buildDirectory}; refusing to emit an incomplete CSP.`
  );
}

const template = await readFile(templatePath, 'utf8');
const placeholder = '__CSP_SCRIPT_HASHES__';
if (!template.includes(placeholder)) {
  throw new Error(`CSP placeholder ${placeholder} is missing from ${templatePath}.`);
}

const generated = template.replace(placeholder, [...hashes].sort().join(' '));
await writeFile(targetPath, generated);

async function findHtmlFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const nested = await Promise.all(
    entries.map(async (entry) => {
      const entryPath = join(directory, entry.name);
      if (entry.isDirectory()) return findHtmlFiles(entryPath);
      return entry.isFile() && entry.name.endsWith('.html') ? [entryPath] : [];
    })
  );
  return nested.flat();
}

function inlineScripts(html) {
  return [...html.matchAll(/<script(?:\s[^>]*)?>([\s\S]*?)<\/script>/giu)]
    .filter(([tag]) => !/\ssrc\s*=/iu.test(tag))
    .map(([, content]) => content);
}
