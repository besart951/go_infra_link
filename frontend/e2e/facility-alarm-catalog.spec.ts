import { expect, login, test, uniqueName } from './fixtures';

test.describe('facility alarm catalog CRUD', () => {
  test('creates, updates, and deletes an alarm unit through visible controls', async ({ page }) => {
    await login(page);
    await page.goto('/facility/alarm-catalog');
    await expect(page.getByRole('heading', { name: 'Alarm-Katalog' })).toBeVisible();

    const code = `E${Math.floor(Math.random() * 900 + 100)}`;
    const name = uniqueName('E2E Alarm Unit');
    await page.locator('#unit-code').fill(code);
    await page.locator('#unit-symbol').fill('u');
    await page.locator('#unit-name').fill(name);
    const createResponse = page.waitForResponse(
      (response) =>
        response.request().method() === 'POST' &&
        new URL(response.url()).pathname === '/api/v1/facility/alarm-units'
    );
    await page.getByRole('button', { name: 'Einheit erstellen' }).click();
    expect((await createResponse).ok()).toBe(true);

    const row = page.locator('tr', { hasText: name });
    await expect(row).toBeVisible();
    await row.getByRole('button', { name: 'Bearbeiten' }).click();
    const updatedName = `${name} aktualisiert`;
    await page.locator('#unit-name').fill(updatedName);
    const updateResponse = page.waitForResponse(
      (response) =>
        response.request().method() === 'PUT' &&
        new URL(response.url()).pathname.startsWith('/api/v1/facility/alarm-units/')
    );
    await page.getByRole('button', { name: 'Speichern' }).click();
    expect((await updateResponse).ok()).toBe(true);

    const updatedRow = page.locator('tr', { hasText: updatedName });
    await expect(updatedRow).toBeVisible();
    const deleteResponse = page.waitForResponse(
      (response) =>
        response.request().method() === 'DELETE' &&
        new URL(response.url()).pathname.startsWith('/api/v1/facility/alarm-units/')
    );
    await updatedRow.getByRole('button', { name: 'Einheit löschen' }).click();
    expect((await deleteResponse).ok()).toBe(true);
    await expect(page.locator('tr', { hasText: updatedName })).toHaveCount(0);
  });

  test('creates, updates and deletes an alarm field, type and mapping through visible controls', async ({
    page
  }) => {
    await login(page);
    await page.goto('/facility/alarm-catalog');
    await expect(page.getByRole('heading', { name: 'Alarm-Katalog' })).toBeVisible();

    const suffix = `${Date.now()}${Math.floor(Math.random() * 10_000)}`;
    const fieldKey = `e2e_field_${suffix}`;
    const fieldLabel = uniqueName('E2E Alarmfeld');
    const updatedFieldLabel = `${fieldLabel} aktualisiert`;
    const typeCode = `E${suffix.slice(-6)}`;
    const typeName = uniqueName('E2E Alarmtyp');
    const updatedTypeName = `${typeName} aktualisiert`;

    await page.locator('#field-key').fill(fieldKey);
    await page.locator('#field-label').fill(fieldLabel);
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/alarm-fields', () =>
      page.getByRole('button', { name: 'Alarmfeld erstellen' }).click()
    );
    const fieldRow = page.locator('tr', { hasText: fieldLabel });
    await expect(fieldRow).toBeVisible();
    await fieldRow.getByRole('button', { name: 'Bearbeiten' }).click();
    await page.locator('#field-label').fill(updatedFieldLabel);
    await expectSuccessfulMutation(page, 'PUT', '/api/v1/facility/alarm-fields/', () =>
      page.getByRole('button', { name: 'Speichern' }).click()
    );

    await page.locator('#type-code').fill(typeCode);
    await page.locator('#type-name').fill(typeName);
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/alarm-types', () =>
      page.getByRole('button', { name: 'Alarmtyp erstellen' }).click()
    );
    const typeRow = page.locator('tr', { hasText: typeName });
    await expect(typeRow).toBeVisible();
    await typeRow.getByRole('button', { name: 'Bearbeiten' }).click();
    await page.locator('#type-name').fill(updatedTypeName);
    await expectSuccessfulMutation(page, 'PUT', '/api/v1/facility/alarm-types/', () =>
      page.getByRole('button', { name: 'Speichern' }).click()
    );

    const typeValue = await page
      .locator('#mapping-type option', { hasText: updatedTypeName })
      .getAttribute('value');
    const fieldValue = await page
      .locator('#mapping-field option', { hasText: updatedFieldLabel })
      .getAttribute('value');
    if (!typeValue || !fieldValue)
      throw new Error('The created alarm type or field is missing from mapping selectors.');
    await page.locator('#mapping-type').selectOption(typeValue);
    await page.locator('#mapping-field').selectOption(fieldValue);
    await page.locator('#mapping-order').fill('7');
    await expectSuccessfulMutation(page, 'POST', '/api/v1/facility/alarm-types/', () =>
      page.getByRole('button', { name: 'Zuordnung erstellen' }).click()
    );
    const mappingRow = page
      .locator('tr', { hasText: updatedFieldLabel })
      .filter({ has: page.getByRole('button', { name: 'Zuordnung löschen' }) });
    await expect(mappingRow).toBeVisible();
    await mappingRow.getByRole('button', { name: 'Bearbeiten' }).click();
    await page.locator('#mapping-group').fill('e2e');
    await expectSuccessfulMutation(page, 'PUT', '/api/v1/facility/alarm-type-fields/', () =>
      page.getByRole('button', { name: 'Speichern' }).click()
    );
    const updatedMappingRow = page
      .locator('tr', { hasText: updatedFieldLabel })
      .filter({ has: page.getByRole('button', { name: 'Zuordnung löschen' }) });
    await expectSuccessfulMutation(page, 'DELETE', '/api/v1/facility/alarm-type-fields/', () =>
      updatedMappingRow.getByRole('button', { name: 'Zuordnung löschen' }).click()
    );

    const updatedTypeRow = page.locator('tr', { hasText: updatedTypeName });
    await expectSuccessfulMutation(page, 'DELETE', '/api/v1/facility/alarm-types/', () =>
      updatedTypeRow.getByRole('button', { name: 'Alarmtyp löschen' }).click()
    );

    const updatedFieldRow = page
      .locator('tr', { hasText: updatedFieldLabel })
      .filter({ has: page.getByRole('button', { name: 'Alarmfeld löschen' }) });
    await expectSuccessfulMutation(page, 'DELETE', '/api/v1/facility/alarm-fields/', () =>
      updatedFieldRow.getByRole('button', { name: 'Alarmfeld löschen' }).click()
    );
  });
});

async function expectSuccessfulMutation(
  page: import('@playwright/test').Page,
  method: string,
  endpointPrefix: string,
  action: () => Promise<void>
): Promise<void> {
  const response = page.waitForResponse((candidate) => {
    if (candidate.request().method() !== method) return false;
    const pathname = new URL(candidate.url()).pathname;
    if (method === 'POST' && endpointPrefix.endsWith('/')) {
      return pathname.startsWith(endpointPrefix) && !pathname.endsWith('/validate');
    }
    return method === 'POST' ? pathname === endpointPrefix : pathname.startsWith(endpointPrefix);
  });
  await action();
  expect((await response).ok()).toBe(true);
}
