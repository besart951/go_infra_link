import { e2eUsers, expect, login, test, uniqueName } from './fixtures';

function waitForMutation(page: Parameters<typeof login>[0], method: string, pathname: RegExp) {
  return page.waitForResponse(
    (response) =>
      response.request().method() === method && pathname.test(new URL(response.url()).pathname)
  );
}

test.describe('Benutzerverwaltung', () => {
  test('Superadmin verwaltet Einladungen, Rollen, Status und Wiederherstellung über die UI', async ({
    page
  }) => {
    await login(page);
    await page.goto('/users/directory');
    await expect(page.getByRole('heading', { name: 'Benutzerverwaltung' })).toBeVisible();

    const email = `e2e-user-${Date.now()}@example.test`;
    await page.getByLabel('Benutzer erstellen').click();

    const invitationDialog = page.getByRole('dialog');
    await invitationDialog.locator('#email').fill(email);
    await invitationDialog.getByRole('combobox').click();
    await page.locator('[role="option"][data-value="planer"]:visible').click();

    const invitationResponse = waitForMutation(page, 'POST', /^\/api\/v1\/users\/invitations$/u);
    await invitationDialog.getByRole('button', { name: 'Einladung senden' }).click();
    expect((await invitationResponse).status()).toBe(201);

    const userRow = page.locator('tr').filter({ hasText: email });
    await expect(userRow).toBeVisible();

    await userRow.getByRole('button', { name: 'Aktionen' }).click();
    await expect(page.getByText('Einladung erneut senden', { exact: true })).toBeVisible();
    const roleResponse = waitForMutation(
      page,
      'POST',
      /^\/api\/v1\/admin\/users\/[0-9a-f-]+\/role$/u
    );
    await page.getByRole('menuitem', { name: 'Entrepreneur', exact: true }).click();
    expect((await roleResponse).ok()).toBe(true);
    await expect(userRow.getByText('Entrepreneur', { exact: true })).toBeVisible();

    await userRow.getByRole('button', { name: 'Aktionen' }).click();
    const disableResponse = waitForMutation(
      page,
      'POST',
      /^\/api\/v1\/admin\/users\/[0-9a-f-]+\/disable$/u
    );
    await page.getByRole('menuitem', { name: 'Benutzer deaktivieren' }).click();
    expect((await disableResponse).ok()).toBe(true);

    const search = page.getByLabel('Benutzer nach Name oder E-Mail suchen...');
    await search.fill(e2eUsers.planner.email);
    await search.press('Enter');
    const plannerRow = page.locator('tr').filter({ hasText: e2eUsers.planner.email });
    await expect(plannerRow).toBeVisible();
    await plannerRow.getByRole('button', { name: 'Aktionen' }).click();
    const disablePlannerResponse = waitForMutation(
      page,
      'POST',
      /^\/api\/v1\/admin\/users\/[0-9a-f-]+\/disable$/u
    );
    await page.getByRole('menuitem', { name: 'Benutzer deaktivieren' }).click();
    expect((await disablePlannerResponse).ok()).toBe(true);

    await plannerRow.getByRole('button', { name: 'Aktionen' }).click();
    const enableResponse = waitForMutation(
      page,
      'POST',
      /^\/api\/v1\/admin\/users\/[0-9a-f-]+\/enable$/u
    );
    await page.getByRole('menuitem', { name: 'Benutzer aktivieren' }).click();
    expect((await enableResponse).ok()).toBe(true);

    await search.fill(email);
    await search.press('Enter');
    await expect(userRow).toBeVisible();

    await userRow.getByRole('button', { name: 'Aktionen' }).click();
    await page.getByRole('menuitem', { name: 'Benutzer löschen', exact: true }).click();
    const deleteResponse = waitForMutation(page, 'DELETE', /^\/api\/v1\/users\/[0-9a-f-]+$/u);
    await page.getByRole('dialog').getByRole('button', { name: 'Löschen', exact: true }).click();
    expect((await deleteResponse).status()).toBe(204);
    await expect(userRow).toHaveCount(0);

    await page.getByText('Gelöschte Benutzer anzeigen', { exact: true }).click();
    await expect(userRow).toBeVisible();

    await userRow.getByRole('button', { name: 'Aktionen' }).click();
    await page.getByRole('menuitem', { name: 'Benutzer wiederherstellen' }).click();
    const restoreResponse = waitForMutation(
      page,
      'POST',
      /^\/api\/v1\/admin\/users\/[0-9a-f-]+\/restore$/u
    );
    await page.getByRole('dialog').getByRole('button', { name: 'Nur Eintrag' }).click();
    expect((await restoreResponse).ok()).toBe(true);
    await expect(userRow).toBeVisible();
  });

  test('Superadmin verwaltet Teams vollständig über die UI', async ({ page }) => {
    await login(page);
    await page.goto('/teams');
    await expect(page.getByRole('heading', { name: 'Teams' })).toBeVisible();

    const teamName = uniqueName('E2E Team');
    const updatedTeamName = `${teamName} aktualisiert`;
    await page.getByLabel('Team erstellen').click();

    const createDialog = page.getByRole('dialog');
    await createDialog.locator('#team_name').fill(teamName);
    await createDialog.locator('#team_desc').fill('E2E Team-Beschreibung');
    const createResponse = waitForMutation(page, 'POST', /^\/api\/v1\/teams$/u);
    await createDialog.getByRole('button', { name: 'Erstellen' }).click();
    expect((await createResponse).status()).toBe(201);
    await page.waitForURL(/\/teams\/[0-9a-f-]{36}$/u);
    await expect(page.getByRole('heading', { name: teamName })).toBeVisible();

    const teamID = page.url().split('/').at(-1);
    if (!teamID) throw new Error('Team creation did not produce a team ID.');

    await page.getByRole('button', { name: 'Bearbeiten', exact: true }).click();
    const editDialog = page.getByRole('dialog');
    await editDialog.locator('#team_edit_name').fill(updatedTeamName);
    await editDialog.locator('#team_edit_description').fill('Aktualisierte E2E Team-Beschreibung');
    const updateResponse = waitForMutation(
      page,
      'PUT',
      new RegExp(`^/api/v1/teams/${teamID}$`, 'u')
    );
    await editDialog.getByRole('button', { name: 'Änderungen speichern' }).click();
    expect((await updateResponse).ok()).toBe(true);
    await expect(page.getByRole('heading', { name: updatedTeamName })).toBeVisible();

    await page.getByRole('button', { name: 'Mitglied hinzufügen' }).click();
    await page.getByPlaceholder('Benutzer suchen...').fill(e2eUsers.planner.email);
    await expect(page.getByText(e2eUsers.planner.email, { exact: true })).toBeVisible();
    const addMemberResponse = waitForMutation(
      page,
      'POST',
      new RegExp(`^/api/v1/teams/${teamID}/members$`, 'u')
    );
    await page.getByText(e2eUsers.planner.email, { exact: true }).click();
    expect((await addMemberResponse).ok()).toBe(true);

    const memberRow = page.locator('tr').filter({ hasText: e2eUsers.planner.email });
    await expect(memberRow).toBeVisible();
    await memberRow.getByRole('combobox').click();
    const changeMemberRoleResponse = waitForMutation(
      page,
      'POST',
      new RegExp(`^/api/v1/teams/${teamID}/members$`, 'u')
    );
    await page.locator('[role="option"][data-value="manager"]:visible').click();
    expect((await changeMemberRoleResponse).ok()).toBe(true);

    const removeMemberResponse = waitForMutation(
      page,
      'DELETE',
      new RegExp(`^/api/v1/teams/${teamID}/members/[0-9a-f-]+$`, 'u')
    );
    await memberRow.getByRole('button', { name: 'Entfernen' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Entfernen' }).click();
    expect((await removeMemberResponse).status()).toBe(204);
    await expect(memberRow).toHaveCount(0);

    await page.goto('/teams');
    const teamRow = page.locator('tr').filter({ hasText: updatedTeamName });
    await expect(teamRow).toBeVisible();
    const deleteTeamResponse = waitForMutation(
      page,
      'DELETE',
      new RegExp(`^/api/v1/teams/${teamID}$`, 'u')
    );
    await teamRow.getByRole('button', { name: 'Team löschen' }).click();
    await page.getByRole('dialog').getByRole('button', { name: 'Löschen', exact: true }).click();
    expect((await deleteTeamResponse).status()).toBe(204);
    await expect(teamRow).toHaveCount(0);
  });

  test('Benutzer können ihr eigenes Profil und Passwort über die Kontoansicht ändern', async ({
    page
  }) => {
    await login(page, e2eUsers.collaborator);
    await page.goto('/account');
    await expect(page.locator('#first_name')).toBeVisible();

    const originalFirstName = await page.locator('#first_name').inputValue();
    const updatedFirstName = uniqueName('E2E').slice(0, 100);
    const informationForm = page.locator('#first_name').locator('xpath=ancestor::form');
    await page.locator('#first_name').fill(updatedFirstName);
    const updateProfileResponse = waitForMutation(page, 'PUT', /^\/api\/v1\/users\/[0-9a-f-]+$/u);
    await informationForm.getByRole('button', { name: 'Änderungen speichern' }).click();
    expect((await updateProfileResponse).ok()).toBe(true);
    await expect(page.locator('#first_name')).toHaveValue(updatedFirstName);

    await page.getByRole('button', { name: 'Passwort', exact: true }).click();
    const newPassword = `e2e-${Date.now()}-password`;
    const passwordForm = page.locator('#current_password').locator('xpath=ancestor::form');
    await page.locator('#current_password').fill(e2eUsers.collaborator.password);
    await page.locator('#new_password').fill(newPassword);
    await page.locator('#confirm_password').fill(newPassword);
    const updatePasswordResponse = waitForMutation(
      page,
      'PUT',
      /^\/api\/v1\/users\/me\/password$/u
    );
    await passwordForm.getByRole('button', { name: 'Passwort speichern' }).click();
    expect((await updatePasswordResponse).ok()).toBe(true);

    await page.locator('#current_password').fill(newPassword);
    await page.locator('#new_password').fill(e2eUsers.collaborator.password);
    await page.locator('#confirm_password').fill(e2eUsers.collaborator.password);
    const restorePasswordResponse = waitForMutation(
      page,
      'PUT',
      /^\/api\/v1\/users\/me\/password$/u
    );
    await passwordForm.getByRole('button', { name: 'Passwort speichern' }).click();
    expect((await restorePasswordResponse).ok()).toBe(true);

    await page.getByRole('button', { name: 'Information', exact: true }).click();
    await page.locator('#first_name').fill(originalFirstName);
    const restoreProfileResponse = waitForMutation(page, 'PUT', /^\/api\/v1\/users\/[0-9a-f-]+$/u);
    await informationForm.getByRole('button', { name: 'Änderungen speichern' }).click();
    expect((await restoreProfileResponse).ok()).toBe(true);
  });

  test('eingeschränkte Benutzer sehen keine Verwaltung und erhalten bei Deep Links 403', async ({
    page
  }) => {
    await login(page, e2eUsers.planner);
    await page.goto('/users');
    await expect(page.getByRole('link', { name: 'Benutzerverwaltung' })).toHaveCount(0);
    await expect(page.getByRole('link', { name: 'Teams' })).toHaveCount(0);
    await expect(page.getByRole('link', { name: 'Rollen & Berechtigungen' })).toHaveCount(0);

    for (const path of ['/users/directory', '/teams', '/users/roles']) {
      await page.goto(path);
      await expect(page.getByText('403 Forbidden')).toBeVisible();
      expect(new URL(page.url()).pathname).toBe('/errors/403');
    }
  });
});
