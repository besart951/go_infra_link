import { expect, login, navigateToFacility, test, uniqueName } from './fixtures';

test.describe('facility reference CRUD', () => {
  test('creates, updates, and deletes unreferenced system parts and apparats through the UI', async ({
    page
  }) => {
    await login(page);
    const systemPartName = uniqueName('E2E CRUD Systemteil');
    const updatedSystemPartName = `${systemPartName} aktualisiert`;
    const apparatName = uniqueName('E2E CRUD Apparat');
    const updatedApparatName = `${apparatName} aktualisiert`;

    await navigateToFacility(page, 'Systemteile', '/facility/system-parts');
    await page.getByLabel('Neues Systemteil', { exact: true }).click();
    await page.locator('#system_part_short').fill(uniqueShortCode());
    await page.locator('#system_part_name').fill(systemPartName);
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/system-parts', async () => {
      await page.locator('form').getByRole('button', { name: 'Erstellen' }).click();
    });
    await expect(page.locator('tr', { hasText: systemPartName })).toBeVisible();

    await editRow(page, systemPartName, '#system_part_name', updatedSystemPartName);
    await expectSuccessfulMutation(page, 'PUT', '/api/v1/facility/system-parts/', async () => {
      await page.locator('form').getByRole('button', { name: 'Aktualisieren' }).click();
    });
    await deleteRow(page, updatedSystemPartName, '/api/v1/facility/system-parts/');

    await navigateToFacility(page, 'Apparate', '/facility/apparats');
    await page.getByLabel('Neuer Apparat', { exact: true }).click();
    await page.locator('#apparat_short').fill(uniqueShortCode());
    await page.locator('#apparat_name').fill(apparatName);
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/apparats', async () => {
      await page.locator('form').getByRole('button', { name: 'Erstellen' }).click();
    });
    await expect(page.locator('tr', { hasText: apparatName })).toBeVisible();

    await editRow(page, apparatName, '#apparat_name', updatedApparatName);
    await expectSuccessfulMutation(page, 'PUT', '/api/v1/facility/apparats/', async () => {
      await page.locator('form').getByRole('button', { name: 'Aktualisieren' }).click();
    });
    await deleteRow(page, updatedApparatName, '/api/v1/facility/apparats/');
  });

  test('blocks deleting apparatus and system parts that are linked to each other', async ({
    page
  }) => {
    await login(page);
    const systemPartName = uniqueName('E2E Delete Systemteil');
    const apparatName = uniqueName('E2E Delete Apparat');
    const systemPartShortName = uniqueShortCode();
    const apparatShortName = uniqueShortCode();

    await navigateToFacility(page, 'Systemteile', '/facility/system-parts');
    await page.getByLabel('Neues Systemteil', { exact: true }).click();
    await page.locator('#system_part_short').fill(systemPartShortName);
    await page.locator('#system_part_name').fill(systemPartName);
    const createSystemPart = page.waitForResponse(
      (response) =>
        response.request().method() === 'POST' &&
        new URL(response.url()).pathname === '/api/v1/facility/system-parts'
    );
    await page.locator('form').getByRole('button', { name: 'Erstellen' }).click();
    expect((await createSystemPart).ok()).toBe(true);

    await navigateToFacility(page, 'Apparate', '/facility/apparats');
    await page.getByLabel('Neuer Apparat', { exact: true }).click();
    await page.locator('#apparat_short').fill(apparatShortName);
    await page.locator('#apparat_name').fill(apparatName);
    await page.locator('#apparat_system_parts').click();
    await page.getByPlaceholder('Systemteile suchen...').fill(systemPartName);
    await page.getByRole('option', { name: new RegExp(systemPartName) }).click();
    const createApparat = page.waitForResponse(
      (response) =>
        response.request().method() === 'POST' &&
        new URL(response.url()).pathname === '/api/v1/facility/apparats'
    );
    await page.locator('form').getByRole('button', { name: 'Erstellen' }).click();
    expect((await createApparat).ok()).toBe(true);

    const apparatRow = page.locator('tr', { hasText: apparatName });
    await apparatRow.getByRole('button').click();
    const blockedApparatDelete = page.getByRole('menuitem', { name: /Löschen/ });
    await expect(blockedApparatDelete).toHaveAttribute('data-disabled');

    await page.keyboard.press('Escape');
    await navigateToFacility(page, 'Systemteile', '/facility/system-parts');
    const systemPartRow = page.locator('tr', { hasText: systemPartName });
    await systemPartRow.getByRole('button').click();
    const blockedSystemPartDelete = page.getByRole('menuitem', { name: /Löschen/ });
    await expect(blockedSystemPartDelete).toHaveAttribute('data-disabled');
  });
});

async function editRow(page: import('@playwright/test').Page, rowText: string, field: string, value: string) {
  const row = page.locator('tr', { hasText: rowText });
  await row.getByRole('button').click();
  await page.getByRole('menuitem', { name: 'Bearbeiten' }).click();
  await page.locator(field).fill(value);
}

async function deleteRow(page: import('@playwright/test').Page, rowText: string, endpointPrefix: string) {
  const row = page.locator('tr', { hasText: rowText });
  await row.getByRole('button').click();
  const deleteMenuItem = page.getByRole('menuitem', { name: 'Löschen' });
  await expect(deleteMenuItem).not.toHaveAttribute('data-disabled');
  await deleteMenuItem.click();
  await expectSuccessfulMutation(page, 'DELETE', endpointPrefix, async () => {
    await page.getByRole('dialog').getByRole('button', { name: 'Löschen' }).click();
  });
  await expect(page.locator('tr', { hasText: rowText })).toHaveCount(0);
}

async function expectSuccessfulMutation(
  page: import('@playwright/test').Page,
  method: string,
  endpointPrefix: string,
  action: () => Promise<void>
) {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.request().method() === method &&
      new URL(candidate.url()).pathname.startsWith(endpointPrefix)
  );
  await action();
  expect((await response).ok()).toBe(true);
}

function uniqueShortCode(): string {
  return `E${Math.floor(Math.random() * 36 ** 2)
    .toString(36)
    .toUpperCase()
    .padStart(2, '0')}`;
}
