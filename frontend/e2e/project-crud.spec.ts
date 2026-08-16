import { expect, createProject, e2eUsers, login, test, uniqueName } from './fixtures';

test.describe('project and phase CRUD', () => {
  test('creates, updates and deletes a phase through the visible phase management UI', async ({
    page
  }) => {
    await login(page);
    const phaseName = uniqueName('E2E Phase');
    const updatedPhaseName = `${phaseName} aktualisiert`;

    await page.goto('/projects/phases');
    await expect(page.getByRole('heading', { name: 'Phasen' })).toBeVisible();
    await page.getByLabel('Neue Phase', { exact: true }).click();
    await page.locator('#phase_name').fill(phaseName);
    await expectSuccessfulMutation(page, 'POST', '/api/v1/phases', () =>
      page.getByRole('dialog').getByRole('button', { name: 'Erstellen' }).click()
    );

    const phaseRow = page.locator('tr', { hasText: phaseName });
    await expect(phaseRow).toBeVisible();
    await phaseRow.getByRole('button', { name: 'Bearbeiten' }).click();
    await page.locator('#phase_name').fill(updatedPhaseName);
    await expectSuccessfulMutation(page, 'PATCH', '/api/v1/phases/', () =>
      page.getByRole('dialog').getByRole('button', { name: 'Aktualisieren' }).click()
    );

    const updatedPhaseRow = page.locator('tr', { hasText: updatedPhaseName });
    await expect(updatedPhaseRow).toBeVisible();
    await updatedPhaseRow.getByRole('button', { name: 'Löschen' }).click();
    await expectSuccessfulMutation(page, 'DELETE', '/api/v1/phases/', () =>
      page.getByRole('dialog').getByRole('button', { name: 'Löschen' }).click()
    );
    await expect(updatedPhaseRow).toHaveCount(0);
  });

  test('creates, updates, manages membership and deletes a project through the UI', async ({
    page
  }) => {
    await login(page);
    const name = uniqueName('E2E Projekt');
    const updatedName = `${name} aktualisiert`;
    const projectID = await createProject(page, name);

    await page.getByRole('button', { name: 'Einstellungen' }).click();
    await page.waitForURL(`/projects/${projectID}/settings`);
    await page.locator('#project_name').fill(updatedName);
    await expectSuccessfulMutation(page, 'PATCH', `/api/v1/projects/${projectID}`, () =>
      page.getByRole('button', { name: 'Änderungen speichern' }).click()
    );
    await expect(page.locator('#project_name')).toHaveValue(updatedName);

    await page.getByRole('main').getByRole('button', { name: 'Benutzer' }).click();
    await page.locator('#project_user_search').fill(e2eUsers.collaborator.email);
    await expect(page.locator('tr', { hasText: e2eUsers.collaborator.email })).toBeVisible();
    const collaboratorRow = page.locator('tr', { hasText: e2eUsers.collaborator.email });
    if ((await collaboratorRow.getByRole('button', { name: 'Hinzufügen' }).count()) === 0) {
      await removeProjectMember(page, projectID, collaboratorRow);
    }
    await expectSuccessfulMutation(page, 'POST', `/api/v1/projects/${projectID}/users`, () =>
      collaboratorRow.getByRole('button', { name: 'Hinzufügen' }).click()
    );
    await expect(collaboratorRow.getByRole('button', { name: 'Entfernen' })).toBeVisible();
    await removeProjectMember(page, projectID, collaboratorRow);

    await page.getByRole('main').getByRole('button', { name: 'Einstellungen' }).click();
    await expect(page.getByRole('button', { name: 'Projekt löschen' })).toBeVisible();
    await page.getByRole('button', { name: 'Projekt löschen' }).click();
    await expectSuccessfulMutation(page, 'DELETE', `/api/v1/projects/${projectID}`, () =>
      page.getByRole('dialog').getByRole('button', { name: 'Projekt löschen' }).click()
    );
    await page.waitForURL('/projects/list');
    await expect(page.locator('tr', { hasText: updatedName })).toHaveCount(0);
  });

  test('assigns and removes object data through the project settings UI', async ({ page }) => {
    await login(page);
    const projectID = await createProject(page, uniqueName('E2E Projektobjektdaten'));

    await page.getByRole('button', { name: 'Einstellungen' }).click();
    await page.waitForURL(`/projects/${projectID}/settings`);
    await page.getByRole('main').getByRole('button', { name: 'Objektdaten' }).click();
    const deactivateObjectData = page.getByRole('button', { name: 'Deaktivieren' }).first();
    await expect(deactivateObjectData).toBeVisible();
    const objectDataDescription = await deactivateObjectData
      .locator('xpath=ancestor::tr')
      .locator('td')
      .first()
      .innerText();
    const objectDataRow = page.locator('tr', { hasText: objectDataDescription });

    await deactivateObjectData.click();
    const confirmationDialog = page.getByRole('dialog');
    const removeObjectData = page.waitForResponse(
      (response) =>
        response.request().method() === 'DELETE' &&
        new URL(response.url()).pathname.startsWith(`/api/v1/projects/${projectID}/object-data/`)
    );
    await confirmationDialog.getByRole('button', { name: 'Deaktivieren' }).click();
    expect((await removeObjectData).ok()).toBe(true);
    await expect(confirmationDialog).toHaveCount(0);
    const activateObjectData = objectDataRow.getByRole('button', { name: 'Aktivieren' });
    await expect(activateObjectData).toBeVisible();

    const addObjectData = page.waitForResponse(
      (response) =>
        response.request().method() === 'POST' &&
        new URL(response.url()).pathname === `/api/v1/projects/${projectID}/object-data`
    );
    await activateObjectData.click();
    expect((await addObjectData).ok()).toBe(true);
    await expect(objectDataRow.getByRole('button', { name: 'Deaktivieren' })).toBeVisible();

    await page.getByRole('main').getByRole('button', { name: 'Einstellungen' }).click();
    await page.getByRole('button', { name: 'Projekt löschen' }).click();
    await expectSuccessfulMutation(page, 'DELETE', `/api/v1/projects/${projectID}`, () =>
      page.getByRole('dialog').getByRole('button', { name: 'Projekt löschen' }).click()
    );
    await page.waitForURL('/projects/list');
  });
});

async function expectSuccessfulMutation(
  page: import('@playwright/test').Page,
  method: string,
  pathname: string,
  action: () => Promise<void>
): Promise<void> {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.request().method() === method &&
      new URL(candidate.url()).pathname.startsWith(pathname)
  );
  await action();
  expect((await response).ok()).toBe(true);
}

async function removeProjectMember(
  page: import('@playwright/test').Page,
  projectID: string,
  memberRow: import('@playwright/test').Locator
): Promise<void> {
  await memberRow.getByRole('button', { name: 'Entfernen' }).click();
  await expectSuccessfulMutation(page, 'DELETE', `/api/v1/projects/${projectID}/users/`, () =>
    page.getByRole('dialog').getByRole('button', { name: 'Entfernen' }).click()
  );
  await expect(memberRow.getByRole('button', { name: 'Hinzufügen' })).toBeVisible();
}
