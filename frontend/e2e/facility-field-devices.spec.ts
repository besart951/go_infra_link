import { expect, login, navigateToFacility, test, uniqueName } from './fixtures';
import type { Page, Response } from '@playwright/test';

test.describe('facility field-device bulk CRUD', () => {
  test('creates field devices, bulk-updates them and deletes the selection through the UI', async ({
    page
  }) => {
    await login(page);
    const hierarchy = await createFieldDeviceHierarchy(page);
    await updateSPSControllerSystemTypeFromParentForm(page, hierarchy.deviceName);
    const bmk = uniqueBMK();
    const description = uniqueName('E2E Feldgerät');
    const updatedDescription = `${description} aktualisiert`;

    await navigateToFacility(page, 'Feldgeräte', '/facility/field-devices');
    await page.getByRole('button', { name: 'Mehrfach anlegen' }).click();
    await selectFirstOption(page, '#sps-system-type');
    await expect(page.locator('#fd-apparat')).toBeVisible();
    await selectFirstOption(page, '#fd-apparat');
    await selectFirstOption(page, '#fd-system-part');
    await selectFirstOption(page, '#fd-object-data');

    await expect(page.getByRole('button', { name: 'Feldgerät hinzufügen' })).toBeEnabled();
    await page.getByRole('button', { name: 'Feldgerät hinzufügen' }).click();
    await page.locator('#bmk-0').fill(bmk);
    await page.locator('#description-0').fill(description);
    await expectSuccessfulMutation(
      page,
      'POST',
      '/api/v1/facility/field-devices/multi-create',
      () => page.getByRole('button', { name: /Erstelle .* Feldgerät/ }).click()
    );

    const search = page.getByPlaceholder('Feldgeräte suchen...');
    await search.fill(bmk);
    const deviceRow = page.locator('tr', { hasText: bmk });
    await expect(deviceRow).toBeVisible();
    await deviceRow.getByLabel(new RegExp(`${bmk} auswählen`, 'i')).check();
    await page.getByLabel('Sammelbearbeitung').click();
    await page.getByPlaceholder('Beschreibung').fill(updatedDescription);
    await page.getByRole('button', { name: 'Auf Auswahl anwenden' }).click();
    await expectSuccessfulMutation(
      page,
      'PATCH',
      '/api/v1/facility/field-devices/bulk-update',
      () => page.getByLabel('Alles speichern').click()
    );
    await expect(deviceRow).toContainText(updatedDescription);

    page.once('dialog', (dialog) => dialog.accept());
    const deleteResponse = await expectSuccessfulMutation(
      page,
      'DELETE',
      '/api/v1/facility/field-devices/bulk-delete',
      () => page.getByLabel('Löschen', { exact: true }).click()
    );
    await expect(deleteResponse.json()).resolves.toMatchObject({ success_count: 1, failure_count: 0 });
    await expect(deviceRow).toHaveCount(0);

    await expect(page.getByRole('heading', { name: 'Feldgeräte' })).toBeVisible();
    expect(hierarchy).toBeDefined();
  });
});

async function createFieldDeviceHierarchy(page: Page): Promise<{
  buildingCode: string;
  cabinetNumber: string;
  deviceName: string;
}> {
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

  await navigateToFacility(page, 'SPS-Regler', '/facility/sps-controllers');
  await page.getByRole('button', { name: 'Neuer SPS-Regler' }).click();
  await selectAsyncOption(page, cabinetNumber);
  await expect(page.locator('#ga_device')).toBeVisible();
  await expect(page.locator('#device_name')).not.toHaveValue('');
  const deviceName = await page.locator('#device_name').inputValue();
  await selectFirstOption(page, page.getByRole('combobox').last());
  await page.locator('form').getByRole('button', { name: 'Hinzufügen', exact: true }).click();
  await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/sps-controllers', () =>
    page.locator('form').getByRole('button', { name: 'Erstellen' }).click()
  );

  return { buildingCode, cabinetNumber, deviceName };
}

async function updateSPSControllerSystemTypeFromParentForm(page: Page, deviceName: string): Promise<void> {
  await navigateToFacility(page, 'SPS-Regler', '/facility/sps-controllers');
  await page.getByPlaceholder('SPS-Regler suchen...').fill(deviceName);
  const controllerRow = page.locator('tr', { hasText: deviceName });
  await expect(controllerRow).toBeVisible();
  await controllerRow.getByRole('button').last().click();
  await page.getByRole('menuitem', { name: 'Bearbeiten' }).click();

  const systemType = page.locator('[data-testid^="sps-system-type-"]').first();
  await expect(systemType).toBeVisible();
  await systemType.locator('input[maxlength="250"]').fill(uniqueName('E2E Systemtyp-Dokument'));
  await expectSuccessfulMutation(
    page,
    'PUT',
    '/api/v1/facility/sps-controller-system-types/',
    () => page.locator('form').getByRole('button', { name: 'Aktualisieren' }).click()
  );
}

async function selectAsyncOption(page: Page, text: string): Promise<void> {
  await page.getByRole('combobox').click();
  await page.getByPlaceholder('Suchen...', { exact: true }).fill(text);
  await page.getByRole('option', { name: new RegExp(text) }).click();
}

async function selectFirstOption(
  page: Page,
  target: string | ReturnType<Page['getByRole']>
): Promise<void> {
  const combobox = typeof target === 'string' ? page.locator(target) : target;
  await combobox.click();
  const option = page.getByRole('option').first();
  await expect(option).toBeVisible();
  await option.click();
}

async function expectSuccessfulMutation(
  page: Page,
  method: string,
  pathname: string,
  action: () => Promise<void>
): Promise<Response> {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.request().method() === method &&
      (new URL(candidate.url()).pathname === pathname ||
        (pathname.endsWith('/') && new URL(candidate.url()).pathname.startsWith(pathname)))
  );
  await action();
  const completed = await response;
  expect(completed.ok()).toBe(true);
  return completed;
}

function uniqueBuildingCode(): string {
  return `E${Math.floor(Math.random() * 36 ** 3)
    .toString(36)
    .toUpperCase()
    .padStart(3, '0')
    .slice(-3)}`;
}

function uniqueBMK(): string {
  const timestamp = Date.now().toString(36).slice(-6).toUpperCase();
  const suffix = Math.floor(Math.random() * 36 ** 2)
    .toString(36)
    .toUpperCase()
    .padStart(2, '0');
  return `FD${timestamp}${suffix}`;
}
