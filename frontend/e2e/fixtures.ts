import {
  expect,
  test as base,
  type BrowserContext,
  type Page,
  type TestInfo
} from '@playwright/test';
import { mkdir, rm } from 'node:fs/promises';

const DESKTOP_CHROME_CONTEXT = {
  userAgent:
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.7922.34 Safari/537.36',
  viewport: { width: 1280, height: 720 },
  screen: { width: 1920, height: 1080 },
  deviceScaleFactor: 1,
  isMobile: false,
  hasTouch: false,
  locale: 'de-CH',
  timezoneId: 'Europe/Zurich'
} as const;

export const e2eUsers = {
  superadmin: {
    email: process.env.E2E_SUPERADMIN_EMAIL ?? 'besart_morina@hotmail.com',
    password: process.env.E2E_SUPERADMIN_PASSWORD ?? 'password'
  },
  planner: {
    email: process.env.E2E_PLANNER_EMAIL ?? 'planner@example.test',
    password: process.env.E2E_PLANNER_PASSWORD ?? 'planner-password'
  }
} as const;

type Credentials = (typeof e2eUsers)[keyof typeof e2eUsers];

interface ManagedContext {
  context: BrowserContext;
  page: Page;
}

type E2EFixtures = {
  createContext: () => Promise<ManagedContext>;
};

export const test = base.extend<E2EFixtures>({
  createContext: async ({ browser }, use, testInfo) => {
    const managedContexts: ManagedContext[] = [];
    const artifactDirectory = testInfo.outputPath('multi-context');

    await use(async () => {
      await mkdir(artifactDirectory, { recursive: true });
      const context = await browser.newContext({
        ...DESKTOP_CHROME_CONTEXT,
        recordVideo: { dir: artifactDirectory }
      });
      const page = await context.newPage();

      const managed = { context, page };
      managedContexts.push(managed);
      return managed;
    });

    const retainArtifacts = testInfo.status !== testInfo.expectedStatus;
    try {
      for (const [index, managed] of managedContexts.entries()) {
        await disposeManagedContext(managed, testInfo, index + 1, retainArtifacts);
      }
    } finally {
      await Promise.allSettled(managedContexts.map(({ context }) => context.close()));
      if (!retainArtifacts) {
        await rm(artifactDirectory, { recursive: true, force: true });
      }
    }
  }
});

export { expect };

export interface SocketObservation {
  readonly url: string;
  closed: boolean;
}

export interface SocketTracker {
  activeCount(path: string): number;
  maximumActiveCount(path: string): number;
}

export function observeWebSockets(page: Page): SocketTracker {
  const sockets: SocketObservation[] = [];
  const maximumActiveCounts = new Map<string, number>();

  page.on('websocket', (socket) => {
    const observed: SocketObservation = { url: socket.url(), closed: false };
    sockets.push(observed);
    const path = new URL(observed.url).pathname;
    maximumActiveCounts.set(path, Math.max(maximumActiveCounts.get(path) ?? 0, activeCount(path)));
    socket.on('close', () => {
      observed.closed = true;
    });
  });

  function activeCount(path: string): number {
    return sockets.filter((socket) => socket.url.includes(path) && !socket.closed).length;
  }

  return {
    activeCount,
    maximumActiveCount: (path) =>
      Math.max(
        0,
        ...[...maximumActiveCounts.entries()]
          .filter(([socketPath]) => socketPath.includes(path))
          .map(([, count]) => count)
      )
  };
}

export async function expectActiveSocketCount(
  sockets: SocketTracker,
  path: string,
  count: number
): Promise<void> {
  await expect.poll(() => sockets.activeCount(path)).toBe(count);
}

export function expectNoConcurrentSockets(sockets: SocketTracker, path: string): void {
  expect(sockets.maximumActiveCount(path)).toBeLessThanOrEqual(1);
}

export async function login(
  page: Page,
  credentials: Credentials = e2eUsers.superadmin
): Promise<void> {
  await page.goto('/login');
  await page.locator('#email').fill(credentials.email);
  await page.locator('#password').fill(credentials.password);
  await page.getByRole('button', { name: 'Anmelden' }).click();
  await page.waitForURL((url) => url.pathname === '/');
  await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
}

export async function logout(
  page: Page,
  credentials: Credentials = e2eUsers.superadmin
): Promise<void> {
  await page.getByRole('button', { name: credentials.email }).click();
  await page.getByRole('menuitem', { name: 'Abmelden' }).click();
  await page.waitForURL((url) => url.pathname === '/login');
}

export async function navigateToFacility(
  page: Page,
  destination: string,
  pathname: string
): Promise<void> {
  const facilityMenu = page.getByRole('button', { name: 'Anlage', exact: true });
  if ((await facilityMenu.getAttribute('aria-expanded')) !== 'true') {
    await facilityMenu.click();
  }

  await Promise.all([
    page.waitForURL((url) => url.pathname === pathname),
    page.getByRole('link', { name: destination, exact: true }).click()
  ]);
}

export async function navigateToProjectList(page: Page): Promise<void> {
  const projectMenu = page.getByRole('button', { name: 'Projekte', exact: true });
  if ((await projectMenu.getAttribute('aria-expanded')) !== 'true') {
    await projectMenu.click();
  }

  await Promise.all([
    page.waitForURL((url) => url.pathname === '/projects/list'),
    page.getByRole('link', { name: 'Projekte', exact: true }).click()
  ]);
}

export async function navigateToProjectOverview(page: Page): Promise<void> {
  const projectMenu = page.getByRole('button', { name: 'Projekte', exact: true });
  if ((await projectMenu.getAttribute('aria-expanded')) !== 'true') {
    await projectMenu.click();
  }

  await Promise.all([
    page.waitForURL((url) => url.pathname === '/projects'),
    page.getByRole('link', { name: 'Übersicht', exact: true }).click()
  ]);
}

export async function createProject(page: Page, name = uniqueName('E2E Projekt')): Promise<string> {
  await navigateToProjectList(page);
  await expect(page.getByRole('heading', { name: 'Projekte' })).toBeVisible();
  await page.getByLabel('Erstellen', { exact: true }).click();

  const dialog = page.getByRole('dialog');
  await dialog.locator('#project_name_create').fill(name);
  await dialog.locator('#project_phase_create').click();
  await page.getByRole('option', { name: 'SIA 21', exact: true }).click();
  const createResponse = page.waitForResponse(
    (response) =>
      response.request().method() === 'POST' &&
      new URL(response.url()).pathname === '/api/v1/projects'
  );
  await dialog.getByRole('button', { name: 'Erstellen' }).click();
  expect((await createResponse).ok()).toBe(true);
  await page.waitForURL(/\/projects\/[0-9a-f-]{36}$/iu);

  const projectID = page.url().split('/').at(-1);
  if (!projectID) {
    throw new Error('The project creation flow did not yield a project ID.');
  }
  return projectID;
}

export function uniqueName(prefix: string): string {
  return `${prefix} ${Date.now()}-${Math.floor(Math.random() * 10_000)}`;
}

export function countRequests(page: Page, pathname: string): () => number {
  let count = 0;
  page.on('request', (request) => {
    if (request.method() !== 'GET') return;
    if (new URL(request.url()).pathname === pathname) count += 1;
  });
  return () => count;
}

async function disposeManagedContext(
  managed: ManagedContext,
  testInfo: TestInfo,
  index: number,
  retainArtifacts: boolean
): Promise<void> {
  const prefix = `browser-context-${index}`;
  const video = managed.page.video();

  try {
    if (retainArtifacts) {
      const screenshotPath = testInfo.outputPath(`${prefix}.png`);
      const savedScreenshot = await managed.page
        .screenshot({ path: screenshotPath, fullPage: true })
        .then(() => true)
        .catch(() => false);
      if (savedScreenshot) {
        await testInfo.attach(`${prefix}-screenshot`, {
          path: screenshotPath,
          contentType: 'image/png'
        });
      }
    }
  } finally {
    await managed.context.close();
  }

  if (retainArtifacts && video) {
    await testInfo.attach(`${prefix}-video`, {
      path: await video.path(),
      contentType: 'video/webm'
    });
  }
}
