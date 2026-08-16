import { expect, login, test, uniqueName } from './fixtures';

function waitForMutation(page: Parameters<typeof login>[0], method: string, pathname: RegExp) {
  return page.waitForResponse(
    (response) =>
      response.request().method() === method && pathname.test(new URL(response.url()).pathname)
  );
}

test.describe('Rollen und Berechtigungen', () => {
  test('Superadmin verwaltet Berechtigungsdefinitionen und weist sie einer Rolle zu', async ({
    page
  }) => {
    await login(page);
    await page.goto('/users/roles');
    await expect(page.getByRole('heading', { name: 'Rollen & Berechtigungen' })).toBeVisible();

    await page.getByRole('tab', { name: 'Berechtigungen' }).click();
    const resource = `e2erole${Date.now()}`;
    const permissionName = `${resource}.read`;
    const description = uniqueName('E2E Rollenberechtigung');
    await page.getByLabel('Berechtigung erstellen').first().click();

    const createDialog = page.getByRole('dialog');
    await createDialog.locator('#resource').fill(resource);
    await createDialog.locator('#action').fill('read');
    await createDialog.locator('#description').fill(description);
    const createPermissionResponse = waitForMutation(page, 'POST', /^\/api\/v1\/permissions$/u);
    await createDialog.getByRole('button', { name: 'Erstellen' }).click();
    expect((await createPermissionResponse).status()).toBe(201);

    const permissionRow = page.locator('tr:visible').filter({ hasText: permissionName });
    await expect(permissionRow).toBeVisible();
    await permissionRow.getByRole('button', { name: 'Menü öffnen' }).click();
    await page.getByRole('menuitem', { name: 'Bearbeiten' }).click();

    const editDialog = page.getByRole('dialog');
    const updatedDescription = `${description} aktualisiert`;
    await editDialog.locator('#description').fill(updatedDescription);
    const updatePermissionResponse = waitForMutation(
      page,
      'PUT',
      /^\/api\/v1\/permissions\/[0-9a-f-]+$/u
    );
    await editDialog.getByRole('button', { name: 'Aktualisieren' }).click();
    expect((await updatePermissionResponse).ok()).toBe(true);
    await expect(permissionRow.getByText(updatedDescription, { exact: true })).toBeVisible();

    await page.getByRole('tab', { name: 'Rollen' }).click();
    const plannerCard = page
      .locator('[class~="group"]')
      .filter({ hasText: 'Planner with project access and limited management' });
    await expect(plannerCard).toBeVisible();
    await plannerCard.getByRole('button').first().click();

    const editor = page.getByRole('dialog').filter({ hasText: 'Berechtigungen' });
    await expect(editor).toBeVisible();
    await editor.getByPlaceholder('Berechtigungen suchen...').fill(updatedDescription);
    const permissionItem = editor.locator('label').filter({ hasText: updatedDescription });
    await expect(permissionItem).toBeVisible();
    await permissionItem.click();
    const updateRoleResponse = waitForMutation(
      page,
      'PUT',
      /^\/api\/v1\/roles\/planer\/permissions$/u
    );
    await editor.getByRole('button', { name: 'Änderungen speichern' }).click();
    expect((await updateRoleResponse).ok()).toBe(true);

    await page.getByRole('tab', { name: 'Berechtigungen' }).click();
    const permissionSearch = page.getByPlaceholder('Berechtigungen suchen...').first();
    await permissionSearch.fill(permissionName);
    await expect(permissionRow).toBeVisible();
    await permissionRow.getByRole('button', { name: 'Menü öffnen' }).click();
    await page.getByRole('menuitem', { name: 'Löschen', exact: true }).click();
    const deletePermissionResponse = waitForMutation(
      page,
      'DELETE',
      /^\/api\/v1\/permissions\/[0-9a-f-]+$/u
    );
    await page.getByRole('dialog').getByRole('button', { name: 'Löschen', exact: true }).click();
    expect((await deletePermissionResponse).status()).toBe(204);
    await expect(permissionRow).toHaveCount(0);
  });

  test('Superadmin erstellt, aktualisiert und entfernt Phasenregeln über die Rollenansicht', async ({
    page
  }) => {
    await login(page);
    const phaseName = uniqueName('E2E Phasenregel');
    await page.goto('/projects/phases');
    await page.getByLabel('Neue Phase', { exact: true }).click();
    await page.locator('#phase_name').fill(phaseName);
    const createPhaseResponse = waitForMutation(page, 'POST', /^\/api\/v1\/phases$/u);
    await page.getByRole('dialog').getByRole('button', { name: 'Erstellen' }).click();
    expect((await createPhaseResponse).status()).toBe(201);

    await page.goto('/users/roles');
    await page.getByRole('tab', { name: 'Phasenregeln' }).click();

    const phaseRow = page.locator('tr:visible').filter({ hasText: phaseName }).first();
    await expect(phaseRow).toBeVisible();
    await expect(phaseRow.getByText('Standard', { exact: true })).toBeVisible();

    const createRuleResponse = waitForMutation(page, 'POST', /^\/api\/v1\/phase-permissions$/u);
    await phaseRow.getByRole('button', { name: 'Lesen' }).click();
    expect((await createRuleResponse).status()).toBe(201);
    await expect(phaseRow.getByText('Regel aktiv', { exact: true })).toBeVisible();

    const updateRuleResponse = waitForMutation(
      page,
      'PATCH',
      /^\/api\/v1\/phase-permissions\/[0-9a-f-]+$/u
    );
    await phaseRow.getByRole('button', { name: 'Bearbeiten' }).click();
    expect((await updateRuleResponse).ok()).toBe(true);

    const deleteRuleResponse = waitForMutation(
      page,
      'DELETE',
      /^\/api\/v1\/phase-permissions\/[0-9a-f-]+$/u
    );
    await phaseRow.getByRole('button', { name: 'Standard nutzen' }).click();
    expect((await deleteRuleResponse).status()).toBe(204);
    await expect(phaseRow.getByText('Standard', { exact: true })).toBeVisible();

    await page.goto('/projects/phases');
    const phaseRowForCleanup = page.locator('tr:visible').filter({ hasText: phaseName });
    await phaseRowForCleanup.getByRole('button', { name: 'Löschen' }).click();
    const deletePhaseResponse = waitForMutation(page, 'DELETE', /^\/api\/v1\/phases\/[0-9a-f-]+$/u);
    await page.getByRole('dialog').getByRole('button', { name: 'Löschen', exact: true }).click();
    expect((await deletePhaseResponse).status()).toBe(204);
  });
});
