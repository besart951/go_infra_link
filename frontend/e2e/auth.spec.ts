import { e2eUsers, expect, login, logout, test } from './fixtures';

test.describe('authentication', () => {
  test('redirects protected URLs, rejects invalid credentials, then logs out through the UI', async ({
    page
  }) => {
    await page.goto('/timeline');
    await page.waitForURL((url) => url.pathname === '/login');

    await page.locator('#email').fill(e2eUsers.superadmin.email);
    await page.locator('#password').fill('incorrect-password');
    await page.getByRole('button', { name: 'Anmelden' }).click();
    await expect(page.getByRole('alert')).toBeVisible();
    await expect(page).toHaveURL(/\/login$/);

    await page.goto('/');
    await page.waitForURL((url) => url.pathname === '/login');

    await login(page);
    await logout(page);

    await page.goto('/projects/list');
    await page.waitForURL((url) => url.pathname === '/login');
  });
});
