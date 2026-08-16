import { expect, createProject, login, test, uniqueName } from './fixtures';
import type { Locator, Page, Response } from '@playwright/test';

test.describe('project facility links', () => {
  test('links, copies and removes a control cabinet through the project UI', async ({ page }) => {
    await login(page);
    const projectID = await createProject(page, uniqueName('E2E Projektanlagen'));
    const cabinetNumber = `EP${Date.now().toString().slice(-9)}`;

    const cabinetLabel = await createProjectControlCabinet(page, projectID, cabinetNumber);
    await copyProjectControlCabinet(page, projectID, cabinetLabel);
    await removeProjectLink(page, projectID, cabinetLabel, 'control-cabinets');
  });

  test('links, copies and removes an SPS controller through the project UI', async ({ page }) => {
    await login(page);
    const projectID = await createProject(page, uniqueName('E2E Projektsysteme'));
    const cabinetNumber = `EP${Date.now().toString().slice(-9)}`;

    const cabinetLabel = await createProjectControlCabinet(page, projectID, cabinetNumber);

    const controllerLabel = await createProjectSPSController(page, projectID);
    await copyProjectSPSControllerSystemType(page, projectID, controllerLabel);
    await copyProjectSPSController(page, projectID, controllerLabel);
    await removeProjectLink(page, projectID, controllerLabel, 'sps-controllers');
    await removeProjectLink(page, projectID, cabinetLabel, 'control-cabinets');
  });

  test('creates and removes a field device through the project UI', async ({ page }) => {
    await login(page);
    const projectID = await createProject(page, uniqueName('E2E Projektfeldgeräte'));
    const cabinetNumber = `EP${Date.now().toString().slice(-9)}`;
    const cabinetLabel = await createProjectControlCabinet(page, projectID, cabinetNumber);
    const controllerLabel = await createProjectSPSController(page, projectID);

    await createAndRemoveProjectFieldDevice(page, projectID);
    await removeProjectLink(page, projectID, controllerLabel, 'sps-controllers');
    await removeProjectLink(page, projectID, cabinetLabel, 'control-cabinets');
  });
});

async function createProjectControlCabinet(
  page: Page,
  projectID: string,
  cabinetNumber: string
): Promise<string> {
  const createButton = page.getByRole('button', { name: 'Neuer Schaltschrank' });
  const form = page.locator('form:has(#control_cabinet_nr)');
  await createButton.click();
  if ((await form.count()) === 0) {
    // The lazy project tab can finish its initial load immediately after the
    // first click. Retry only when that render transition replaced the form.
    await expect(createButton).toBeVisible();
    await createButton.click();
  }
  await expect(form).toBeVisible();
  await selectFirstOption(form.getByRole('combobox'));
  await form.locator('#control_cabinet_nr').fill(cabinetNumber);

  const create = waitForMutation(page, 'POST', '/api/v1/facility/control-cabinets');
  const link = waitForMutation(page, 'POST', `/api/v1/projects/${projectID}/control-cabinets`);
  await form.getByRole('button', { name: 'Erstellen' }).click();
  expect((await create).ok()).toBe(true);
  expect((await link).ok()).toBe(true);

  const row = visibleRow(page, cabinetNumber);
  await expect(row).toBeVisible();
  return cabinetNumber;
}

async function copyProjectControlCabinet(
  page: Page,
  projectID: string,
  label: string
): Promise<void> {
  const row = visibleRow(page, label);
  await openRowMenu(row);
  const response = waitForMutation(
    page,
    'POST',
    new RegExp(`^/api/v1/projects/${projectID}/control-cabinets/[0-9a-f-]{36}/copy$`, 'iu')
  );
  await page.getByRole('menuitem', { name: 'Duplizieren' }).click();
  expect((await response).ok()).toBe(true);
  await page.keyboard.press('Escape');
}

async function createProjectSPSController(page: Page, projectID: string): Promise<string> {
  await page.getByRole('tab', { name: 'SPS-Regler' }).click();
  await page.getByRole('button', { name: 'Neuer SPS-Regler' }).click();
  const form = page.locator('form');
  const selects = form.getByRole('combobox');
  await selectFirstOption(selects.first());
  await expect(form.locator('#device_name')).not.toHaveValue('');
  await selectFirstOption(selects.last());
  await form.getByRole('button', { name: 'Hinzufügen', exact: true }).click();
  const controllerName = await form.locator('#device_name').inputValue();

  const create = waitForMutation(page, 'POST', '/api/v1/facility/sps-controllers');
  const link = waitForMutation(page, 'POST', `/api/v1/projects/${projectID}/sps-controllers`);
  await form.getByRole('button', { name: 'Erstellen' }).click();
  expect((await create).ok()).toBe(true);
  expect((await link).ok()).toBe(true);

  const row = visibleRow(page, controllerName);
  await expect(row).toBeVisible();
  return controllerName;
}

async function copyProjectSPSController(
  page: Page,
  projectID: string,
  label: string
): Promise<void> {
  const row = visibleRow(page, label);
  await openRowMenu(row);
  const response = waitForMutation(
    page,
    'POST',
    new RegExp(`^/api/v1/projects/${projectID}/sps-controllers/[0-9a-f-]{36}/copy$`, 'iu')
  );
  await page.getByRole('menuitem', { name: 'Duplizieren' }).click();
  expect((await response).ok()).toBe(true);
  await page.keyboard.press('Escape');
}

async function copyProjectSPSControllerSystemType(
  page: Page,
  projectID: string,
  controllerLabel: string
): Promise<void> {
  const row = visibleRow(page, controllerLabel);
  await openRowMenu(row);
  await page.getByRole('menuitem', { name: 'Bearbeiten' }).click();
  const systemType = page.locator('[data-testid^="sps-system-type-"]').first();
  await expect(systemType).toBeVisible();
  const response = waitForMutation(
    page,
    'POST',
    new RegExp(
      `^/api/v1/projects/${projectID}/sps-controller-system-types/[0-9a-f-]{36}/copy$`,
      'iu'
    )
  );
  await systemType.getByRole('button', { name: 'Kopieren' }).click();
  expect((await response).status()).toBe(202);
}

async function createAndRemoveProjectFieldDevice(page: Page, projectID: string): Promise<void> {
  await page.getByRole('tab', { name: 'Feldgeräte' }).click();
  await page.getByRole('button', { name: 'Mehrfach anlegen' }).click();
  await selectFirstOption(page.locator('#sps-system-type'));
  await selectFirstOption(page.locator('#fd-apparat'));
  await selectFirstOption(page.locator('#fd-system-part'));
  await selectFirstOption(page.locator('#fd-object-data'));
  await page.getByRole('button', { name: 'Feldgerät hinzufügen' }).click();
  const bmk = `P${Date.now().toString(36).toUpperCase().slice(-8)}`;
  await page.locator('#bmk-0').fill(bmk);
  const create = waitForMutation(
    page,
    'POST',
    `/api/v1/projects/${projectID}/field-devices/multi-create`
  );
  await page.getByRole('button', { name: /Erstelle .* Feldgerät/ }).click();
  const result = await create;
  expect(result.ok()).toBe(true);
  await expect(result.json()).resolves.toMatchObject({ success_count: 1, failure_count: 0 });

  const row = visibleRow(page, bmk);
  await expect(row).toBeVisible();
  await openRowMenu(row);
  page.once('dialog', (dialog) => dialog.accept());
  const remove = waitForMutation(
    page,
    'DELETE',
    new RegExp(`^/api/v1/projects/${projectID}/field-devices/[0-9a-f-]{36}$`, 'iu')
  );
  await page.getByRole('menuitem', { name: 'Löschen' }).click();
  expect((await remove).ok()).toBe(true);
  await expect(row).toHaveCount(0);
}

async function removeProjectLink(
  page: Page,
  projectID: string,
  label: string,
  resource: 'control-cabinets' | 'sps-controllers'
): Promise<void> {
  if (resource === 'sps-controllers') {
    await page.getByRole('tab', { name: 'SPS-Regler' }).click();
  } else {
    await page.getByRole('tab', { name: 'Schaltschränke' }).click();
  }
  const row = visibleRow(page, label);
  await expect(row).toBeVisible();
  await openRowMenu(row);
  const response = waitForMutation(
    page,
    'DELETE',
    new RegExp(`^/api/v1/projects/${projectID}/${resource}/[0-9a-f-]{36}$`, 'iu')
  );
  await page.getByRole('menuitem', { name: 'Löschen' }).click();
  await page
    .getByRole('dialog')
    .getByRole('button', { name: resource === 'sps-controllers' ? 'Löschen' : 'Entfernen' })
    .click();
  expect((await response).ok()).toBe(true);
  await expect(row).toHaveCount(0);
}

async function openRowMenu(row: Locator): Promise<void> {
  await row.getByRole('button').last().click();
}

function visibleRow(page: Page, label: string): Locator {
  return page.locator('tr:visible', { hasText: label });
}

async function selectFirstOption(combobox: Locator): Promise<void> {
  for (let attempt = 0; attempt < 2; attempt += 1) {
    try {
      await expect(combobox).toBeVisible();
      await combobox.click({ timeout: 3_000 });
      const option = combobox.page().getByRole('option').first();
      await expect(option).toBeVisible();
      await option.click();
      return;
    } catch (error) {
      if (
        attempt === 1 ||
        !(error instanceof Error) ||
        !/(detached|not stable)/iu.test(error.message)
      ) {
        throw error;
      }
      await combobox.page().keyboard.press('Escape');
    }
  }
}

function waitForMutation(page: Page, method: string, pathname: string | RegExp): Promise<Response> {
  return page.waitForResponse(
    (candidate) =>
      candidate.request().method() === method &&
      (typeof pathname === 'string'
        ? new URL(candidate.url()).pathname === pathname
        : pathname.test(new URL(candidate.url()).pathname))
  );
}
