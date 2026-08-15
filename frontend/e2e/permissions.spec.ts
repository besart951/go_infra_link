import { createProject, e2eUsers, expect, login, test } from './fixtures';

test.describe('permission gates', () => {
  test('hides protected navigation and controls, and redirects global and project deep links', async ({
    createContext
  }) => {
    const { page: administratorPage } = await createContext();
    await login(administratorPage);
    const projectID = await createProject(administratorPage);

    const { page: plannerPage } = await createContext();
    await login(plannerPage, e2eUsers.planner);

    await expect(plannerPage.getByRole('link', { name: 'Timeline' })).toHaveCount(0);
    await expect(plannerPage.getByRole('button', { name: 'Erstellen' })).toHaveCount(0);

    await plannerPage.goto('/timeline');
    await plannerPage.waitForURL(/\/errors\/403/);
    await expect(plannerPage.getByText('403 Forbidden')).toBeVisible();

    await plannerPage.goto(`/projects/${projectID}`);
    await plannerPage.waitForURL(/\/errors\/403/);
    await expect(plannerPage.getByText('403 Forbidden')).toBeVisible();

    await administratorPage.goto('/projects/list');
    await expect(administratorPage.getByLabel('Erstellen', { exact: true })).toBeVisible();
  });
});
