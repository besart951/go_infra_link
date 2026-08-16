import { expect, login, navigateToFacility, test, uniqueName } from './fixtures';
import type { Page } from '@playwright/test';

test.describe('facility hierarchy CRUD', () => {
  test('creates, edits and deletes a building and control cabinet through the UI', async ({
    page
  }) => {
    await login(page);
    const buildingCode = uniqueBuildingCode();
    const cabinetNumber = `E${Date.now().toString().slice(-9)}${Math.floor(Math.random() * 10)}`;
    const updatedCabinetNumber = `U${cabinetNumber.slice(1)}`;

    await navigateToFacility(page, 'Gebäude', '/facility/buildings');
    await page.getByLabel('Neues Gebäude', { exact: true }).click();
    await page.locator('#iws_code').fill(buildingCode);
    await page.locator('#building_group').fill('1');
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/buildings', () =>
      page.locator('form').getByRole('button', { name: 'Erstellen' }).click()
    );
    await expect(page.locator('tr', { hasText: buildingCode })).toBeVisible();

    await navigateToFacility(page, 'Schaltschränke', '/facility/control-cabinets');
    await page.getByRole('button', { name: 'Neuer Schaltschrank' }).click();
    await selectAsyncOption(page, buildingCode);
    await page.locator('#control_cabinet_nr').fill(cabinetNumber);
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/control-cabinets', () =>
      page.locator('form').getByRole('button', { name: 'Erstellen' }).click()
    );
    await expect(page.locator('tr', { hasText: cabinetNumber })).toBeVisible();

    const cabinetRow = page.locator('tr', { hasText: cabinetNumber });
    await cabinetRow.getByRole('button').last().click();
    await page.getByRole('menuitem', { name: 'Bearbeiten' }).click();
    await page.locator('#control_cabinet_nr').fill(updatedCabinetNumber);
    await expectSuccessfulMutation(page, 'PUT', '/api/v1/facility/control-cabinets/', () =>
      page.locator('form').getByRole('button', { name: 'Aktualisieren' }).click()
    );
    await expect(page.locator('tr', { hasText: updatedCabinetNumber })).toBeVisible();

    const updatedCabinetRow = page.locator('tr', { hasText: updatedCabinetNumber });
    await updatedCabinetRow.getByRole('button').last().click();
    await expectSuccessfulMutation(page, 'DELETE', '/api/v1/facility/control-cabinets/', () =>
      page.getByRole('menuitem', { name: 'Löschen' }).click()
    );
    await expect(page.locator('tr', { hasText: updatedCabinetNumber })).toHaveCount(0);

    await navigateToFacility(page, 'Gebäude', '/facility/buildings');
    const buildingRow = page.locator('tr', { hasText: buildingCode });
    await buildingRow.getByRole('button').last().click();
    await page.getByRole('menuitem', { name: 'Bearbeiten' }).click();
    await page.locator('#building_group').fill('2');
    await expectSuccessfulMutation(page, 'PUT', '/api/v1/facility/buildings/', () =>
      page.locator('form').getByRole('button', { name: 'Aktualisieren' }).click()
    );

    const updatedBuildingRow = page.locator('tr', { hasText: buildingCode });
    await updatedBuildingRow.getByRole('button').last().click();
    await page.getByRole('menuitem', { name: 'Löschen' }).click();
    await expectSuccessfulMutation(page, 'DELETE', '/api/v1/facility/buildings/', () =>
      page.getByRole('dialog').getByRole('button', { name: 'Löschen' }).click()
    );
  });

  test('starts a control-cabinet copy from the visible row action', async ({ page }) => {
    await login(page);
    const buildingCode = uniqueBuildingCode();
    const cabinetNumber = `E${Date.now().toString().slice(-9)}${Math.floor(Math.random() * 10)}`;

    await navigateToFacility(page, 'Gebäude', '/facility/buildings');
    await page.getByLabel('Neues Gebäude', { exact: true }).click();
    await page.locator('#iws_code').fill(buildingCode);
    await page.locator('#building_group').fill('1');
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/buildings', () =>
      page.locator('form').getByRole('button', { name: 'Erstellen' }).click()
    );

    await navigateToFacility(page, 'Schaltschränke', '/facility/control-cabinets');
    await page.getByRole('button', { name: 'Neuer Schaltschrank' }).click();
    await selectAsyncOption(page, buildingCode);
    await page.locator('#control_cabinet_nr').fill(cabinetNumber);
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/control-cabinets', () =>
      page.locator('form').getByRole('button', { name: 'Erstellen' }).click()
    );

    const cabinetRow = page.locator('tr', { hasText: cabinetNumber });
    await cabinetRow.getByRole('button').last().click();
    const copyResponse = page.waitForResponse(
      (candidate) =>
        candidate.request().method() === 'POST' &&
        new URL(candidate.url()).pathname.startsWith('/api/v1/facility/control-cabinets/') &&
        new URL(candidate.url()).pathname.endsWith('/copy')
    );
    await page.getByRole('menuitem', { name: 'Duplizieren' }).click();
    expect((await copyResponse).status()).toBe(202);
  });

  test('creates, updates, copies and deletes an SPS controller and its system type through the UI', async ({
    page
  }) => {
    await login(page);
    const buildingCode = uniqueBuildingCode();
    const cabinetNumber = `E${Date.now().toString().slice(-9)}${Math.floor(Math.random() * 10)}`;
    const ipAddress = '10.23.45.67';
    const documentName = uniqueName('E2E SPS Systemtyp');

    await navigateToFacility(page, 'Gebäude', '/facility/buildings');
    await page.getByLabel('Neues Gebäude', { exact: true }).click();
    await page.locator('#iws_code').fill(buildingCode);
    await page.locator('#building_group').fill('1');
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/buildings', () =>
      page.locator('form').getByRole('button', { name: 'Erstellen' }).click()
    );

    await navigateToFacility(page, 'Schaltschränke', '/facility/control-cabinets');
    await page.getByRole('button', { name: 'Neuer Schaltschrank' }).click();
    await selectAsyncOption(page, buildingCode);
    await page.locator('#control_cabinet_nr').fill(cabinetNumber);
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/control-cabinets', () =>
      page.locator('form').getByRole('button', { name: 'Erstellen' }).click()
    );

    await navigateToFacility(page, 'SPS-Regler', '/facility/sps-controllers');
    await page.getByRole('button', { name: 'Neuer SPS-Regler' }).click();
    await selectAsyncOption(page, cabinetNumber);
    await expect(page.locator('#device_name')).not.toHaveValue('');
    const deviceName = await page.locator('#device_name').inputValue();
    await selectFirstOption(page.getByRole('combobox').last());
    await page.locator('form').getByRole('button', { name: 'Hinzufügen', exact: true }).click();
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/sps-controllers', () =>
      page.locator('form').getByRole('button', { name: 'Erstellen' }).click()
    );

    const controllerRow = page.locator('tr', { hasText: deviceName });
    await expect(controllerRow).toBeVisible();
    await controllerRow.getByRole('button').last().click();
    await page.getByRole('menuitem', { name: 'Bearbeiten' }).click();
    await page.locator('#ip_address').fill(ipAddress);
    await expectSuccessfulMutation(page, 'PUT', '/api/v1/facility/sps-controllers/', () =>
      page.locator('form').getByRole('button', { name: 'Aktualisieren' }).click()
    );
    await expect(controllerRow).toContainText(ipAddress);

    await controllerRow.getByRole('button').last().click();
    await page.getByRole('menuitem', { name: 'Bearbeiten' }).click();
    const systemType = page.locator('[data-testid^="sps-system-type-"]').first();
    await expect(systemType).toBeVisible();
    await systemType.locator('input[maxlength="250"]').fill(documentName);
    await expectSuccessfulMutation(
      page,
      'PUT',
      '/api/v1/facility/sps-controller-system-types/',
      () => page.locator('form').getByRole('button', { name: 'Aktualisieren' }).click()
    );

    await controllerRow.getByRole('button').last().click();
    await page.getByRole('menuitem', { name: 'Bearbeiten' }).click();
    const persistedSystemType = page.locator('[data-testid^="sps-system-type-"]').first();
    await expect(persistedSystemType).toBeVisible();
    await expectSuccessfulMutation(
      page,
      'POST',
      '/api/v1/facility/sps-controller-system-types/',
      () => persistedSystemType.getByRole('button', { name: 'Kopieren' }).click(),
      202
    );
    await expectSuccessfulMutation(
      page,
      'DELETE',
      '/api/v1/facility/sps-controller-system-types/',
      () => persistedSystemType.getByRole('button', { name: 'Entfernen' }).click()
    );
    await expect(persistedSystemType).toHaveCount(0);
    await page.locator('form').getByRole('button', { name: 'Abbrechen' }).click();

    await controllerRow.getByRole('button').last().click();
    const deletion = page.waitForResponse(
      (candidate) =>
        candidate.request().method() === 'DELETE' &&
        new URL(candidate.url()).pathname.startsWith('/api/v1/facility/sps-controllers/')
    );
    await page.getByRole('menuitem', { name: 'Löschen' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Löschen' }).click();
    expect((await deletion).ok()).toBe(true);
    await expect(controllerRow).toHaveCount(0);

    await page.getByRole('button', { name: 'Neuer SPS-Regler' }).click();
    await selectAsyncOption(page, cabinetNumber);
    await expect(page.locator('#device_name')).not.toHaveValue('');
    const copySourceName = await page.locator('#device_name').inputValue();
    await selectFirstOption(page.getByRole('combobox').last());
    await page.locator('form').getByRole('button', { name: 'Hinzufügen', exact: true }).click();
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/sps-controllers', () =>
      page.locator('form').getByRole('button', { name: 'Erstellen' }).click()
    );

    const copySourceRow = page.locator('tr', { hasText: copySourceName });
    await expect(copySourceRow).toBeVisible();
    await copySourceRow.getByRole('button').last().click();
    const controllerCopy = page.waitForResponse(
      (candidate) =>
        candidate.request().method() === 'POST' &&
        new URL(candidate.url()).pathname.startsWith('/api/v1/facility/sps-controllers/') &&
        new URL(candidate.url()).pathname.endsWith('/copy')
    );
    await page.getByRole('menuitem', { name: 'Duplizieren' }).click();
    expect((await controllerCopy).status()).toBe(202);
  });
});

async function selectAsyncOption(page: Page, text: string): Promise<void> {
  await page.getByRole('combobox').click();
  await page.getByPlaceholder('Suchen...', { exact: true }).fill(text);
  await page.getByRole('option', { name: new RegExp(text) }).click();
}

async function selectFirstOption(combobox: ReturnType<Page['getByRole']>): Promise<void> {
  await combobox.click();
  const option = combobox.page().getByRole('option').first();
  await expect(option).toBeVisible();
  await option.click();
}

async function expectSuccessfulMutation(
  page: Page,
  method: string,
  endpointPrefix: string,
  action: () => Promise<void>,
  expectedStatus?: number
): Promise<void> {
  const response = page.waitForResponse((candidate) => {
    if (candidate.request().method() !== method) return false;
    const pathname = new URL(candidate.url()).pathname;
    return endpointPrefix.endsWith('/')
      ? pathname.startsWith(endpointPrefix)
      : pathname === endpointPrefix;
  });
  await action();
  const completed = await response;
  if (expectedStatus !== undefined) {
    expect(completed.status()).toBe(expectedStatus);
    return;
  }
  expect(completed.ok()).toBe(true);
}

function uniqueBuildingCode(): string {
  return `E${Math.floor(Math.random() * 36 ** 3)
    .toString(36)
    .toUpperCase()
    .padStart(3, '0')
    .slice(-3)}`;
}
