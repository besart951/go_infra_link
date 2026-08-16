import {
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

test.describe('facility realtime changes', () => {
  test('applies apparatus create, update and delete to another user’s open list without navigation', async ({
    createContext
  }) => {
    const { page: readerPage } = await createContext();
    const readerSockets = await observeWebSockets(readerPage);
    await login(readerPage, e2eUsers.collaborator);
    await navigateToFacility(readerPage, 'Apparate', '/facility/apparats');
    await expectActiveSocketCount(readerSockets, '/api/v1/facility/reference-data/stream', 1);
    await expectNoConcurrentSockets(readerSockets, '/api/v1/facility/reference-data/stream');

    const { page: writerPage } = await createContext();
    await login(writerPage, e2eUsers.superadmin);
    await writerPage.goto('/facility/apparats');
    await writerPage.getByLabel('Neuer Apparat', { exact: true }).click();
    const name = uniqueName('E2E Live Apparat');
    await writerPage.locator('#apparat_short').fill('ELV');
    await writerPage.locator('#apparat_name').fill(name);
    const createResponse = writerPage.waitForResponse(
      (response) =>
        response.request().method() === 'POST' &&
        new URL(response.url()).pathname === '/api/v1/facility/apparats'
    );
    await writerPage.locator('form').getByRole('button', { name: 'Erstellen' }).click();
    expect((await createResponse).ok()).toBe(true);

    await expect(readerPage.getByText(name, { exact: true })).toBeVisible();
    const updatedName = `${name} aktualisiert`;
    const writerRow = writerPage.locator('tr', { hasText: name });
    await writerRow.getByRole('button').click();
    await writerPage.getByRole('menuitem', { name: 'Bearbeiten' }).click();
    await writerPage.locator('#apparat_name').fill(updatedName);
    const updateResponse = writerPage.waitForResponse(
      (response) =>
        response.request().method() === 'PUT' &&
        new URL(response.url()).pathname.startsWith('/api/v1/facility/apparats/')
    );
    await writerPage.locator('form').getByRole('button', { name: 'Aktualisieren' }).click();
    expect((await updateResponse).ok()).toBe(true);
    await expect(readerPage.getByText(updatedName, { exact: true })).toBeVisible();

    const updatedWriterRow = writerPage.locator('tr', { hasText: updatedName });
    await updatedWriterRow.getByRole('button').click();
    const deleteResponse = writerPage.waitForResponse(
      (response) =>
        response.request().method() === 'DELETE' &&
        new URL(response.url()).pathname.startsWith('/api/v1/facility/apparats/')
    );
    await writerPage.getByRole('menuitem', { name: 'Löschen' }).click();
    await writerPage.getByRole('dialog').getByRole('button', { name: 'Löschen' }).click();
    expect((await deleteResponse).ok()).toBe(true);
    await expect(readerPage.getByText(updatedName, { exact: true })).toHaveCount(0);
    await expectActiveSocketCount(readerSockets, '/api/v1/facility/reference-data/stream', 1);
  });
});
