import {
  createProject,
  expect,
  expectActiveSocketCount,
  expectNoConcurrentSockets,
  login,
  observeWebSockets,
  test,
  uniqueName
} from './fixtures';

test.describe('project activity realtime', () => {
  test('shows a live activity notice, loads the durable history, and opens a field diff', async ({
    createContext
  }) => {
    const { page: editorPage } = await createContext();
    await login(editorPage);
    const projectName = uniqueName('E2E Verlauf');
    const projectID = await createProject(editorPage, projectName);

    const { page: viewerPage } = await createContext();
    const viewerSockets = observeWebSockets(viewerPage);
    await login(viewerPage);
    await viewerPage.goto(`/projects/${projectID}`);
    await expectActiveSocketCount(viewerSockets, `/api/v1/projects/${projectID}/collaboration`, 1);
    expectNoConcurrentSockets(viewerSockets, `/api/v1/projects/${projectID}/collaboration`);
    await viewerPage.getByLabel('Verlauf', { exact: true }).click();
    await expect(viewerPage.getByRole('dialog')).toContainText('Projektverlauf');

    const updatedName = uniqueName('E2E Verlauf aktualisiert');
    await editorPage.goto(`/projects/${projectID}/settings`);
    await editorPage.locator('#project_name').fill(updatedName);
    const updateResponse = editorPage.waitForResponse(
      (response) =>
        response.request().method() === 'PATCH' &&
        new URL(response.url()).pathname === `/api/v1/projects/${projectID}`
    );
    await editorPage.getByRole('button', { name: 'Änderungen speichern' }).click();
    expect((await updateResponse).ok()).toBe(true);

    const liveActivityNotice = viewerPage.getByRole('button', {
      name: /neue Aktivität(?:en)? anzeigen/
    });
    await expect(liveActivityNotice).toBeVisible();
    await liveActivityNotice.click();

    const updatedActivity = viewerPage.locator('article').filter({ hasText: 'Aktualisiert' });
    await expect(updatedActivity).toContainText('Projekt');
    await updatedActivity.getByRole('button', { name: 'Projekt' }).click();

    await expect(viewerPage.getByRole('dialog', { name: 'Änderungsdetails' })).toBeVisible();
    await expect(viewerPage.getByRole('dialog', { name: 'Änderungsdetails' })).toContainText(
      'Name'
    );
    await expect(viewerPage.getByRole('dialog', { name: 'Änderungsdetails' })).toContainText(
      projectName
    );
    await expect(viewerPage.getByRole('dialog', { name: 'Änderungsdetails' })).toContainText(
      updatedName
    );
  });
});
