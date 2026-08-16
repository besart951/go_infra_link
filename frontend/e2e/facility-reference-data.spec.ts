import {
  countRequests,
  e2eUsers,
  expect,
  expectActiveSocketCount,
  expectNoConcurrentSockets,
  login,
  navigateToFacility,
  observeWebSockets,
  test,
  uniqueName
} from './fixtures';

test.describe('facility reference-data cache', () => {
  test('serves comboboxes from the cache and refreshes it through the facility websocket', async ({
    createContext
  }) => {
    const { page: readerPage } = await createContext();
    const readerSockets = await observeWebSockets(readerPage);
    const apparatRequests = countRequests(readerPage, '/api/v1/facility/apparats');
    const systemPartRequests = countRequests(readerPage, '/api/v1/facility/system-parts');

    await login(readerPage, e2eUsers.collaborator);
    await expectActiveSocketCount(readerSockets, '/api/v1/facility/reference-data/stream', 1);
    await expectNoConcurrentSockets(readerSockets, '/api/v1/facility/reference-data/stream');
    await expect.poll(apparatRequests).toBeGreaterThan(0);
    await expect.poll(systemPartRequests).toBeGreaterThan(0);

    const initialApparatRequests = apparatRequests();
    await navigateToFacility(readerPage, 'Objektdaten', '/facility/object-data');
    await readerPage.getByLabel('Neue Objektdaten', { exact: true }).click();
    await readerPage.locator('#object_data_apparats').click();
    await expect(readerPage.getByPlaceholder('Apparate suchen...')).toBeVisible();
    await expect(readerPage.locator('[role="option"]').first()).toBeVisible();
    await expect.poll(apparatRequests).toBe(initialApparatRequests);

    const { page: writerPage } = await createContext();
    await login(writerPage, e2eUsers.superadmin);
    const createdApparatName = uniqueName('E2E Cache Apparat');

    await writerPage.goto('/facility/apparats');
    await writerPage.getByLabel('Neuer Apparat', { exact: true }).click();
    await writerPage.locator('#apparat_short').fill('E2A');
    await writerPage.locator('#apparat_name').fill(createdApparatName);
    const apparatRefresh = readerPage.waitForResponse(
      (response) =>
        response.request().method() === 'GET' &&
        new URL(response.url()).pathname === '/api/v1/facility/apparats'
    );
    await writerPage.locator('form').getByRole('button', { name: 'Erstellen' }).click();
    expect((await apparatRefresh).ok()).toBe(true);

    await readerPage.getByPlaceholder('Apparate suchen...').fill(createdApparatName);
    await expect(
      readerPage.getByRole('option', { name: new RegExp(createdApparatName) })
    ).toBeVisible();
    await expect.poll(apparatRequests).toBe(initialApparatRequests + 1);

    const initialSystemPartRequests = systemPartRequests();
    await readerPage.keyboard.press('Escape');
    await navigateToFacility(readerPage, 'Apparate', '/facility/apparats');
    await readerPage.getByLabel('Neuer Apparat', { exact: true }).click();
    await readerPage.locator('#apparat_system_parts').click();
    await expect(readerPage.getByPlaceholder('Systemteile suchen...')).toBeVisible();
    await expect(readerPage.locator('[role="option"]').first()).toBeVisible();
    await expect.poll(systemPartRequests).toBe(initialSystemPartRequests);

    const createdSystemPartName = uniqueName('E2E Cache Systemteil');
    await writerPage.goto('/facility/system-parts');
    await writerPage.getByLabel('Neues Systemteil', { exact: true }).click();
    await writerPage.locator('#system_part_short').fill('E2S');
    await writerPage.locator('#system_part_name').fill(createdSystemPartName);
    const systemPartRefresh = readerPage.waitForResponse(
      (response) =>
        response.request().method() === 'GET' &&
        new URL(response.url()).pathname === '/api/v1/facility/system-parts'
    );
    await writerPage.locator('form').getByRole('button', { name: 'Erstellen' }).click();
    expect((await systemPartRefresh).ok()).toBe(true);

    await readerPage.getByPlaceholder('Systemteile suchen...').fill(createdSystemPartName);
    await expect(
      readerPage.getByRole('option', { name: new RegExp(createdSystemPartName) })
    ).toBeVisible();
    await expect.poll(systemPartRequests).toBe(initialSystemPartRequests + 1);
  });
});
