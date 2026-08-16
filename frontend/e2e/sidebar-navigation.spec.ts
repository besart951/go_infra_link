import { expect, login, test } from './fixtures';

const collapsibleNavigationGroups = ['Benutzer', 'Anlage', 'Projekte', 'Benachrichtigungen'];

async function expandNavigationGroups(page: Parameters<typeof login>[0]): Promise<void> {
  const sidebarContent = page.locator('[data-sidebar="content"]');

  for (const name of collapsibleNavigationGroups) {
    const trigger = sidebarContent.getByRole('button', { name, exact: true });
    if ((await trigger.getAttribute('aria-expanded')) !== 'true') {
      await trigger.click();
    }
  }
}

test.describe('collapsed sidebar navigation', () => {
  test('shows the hovered icon menu above the main content and switches panels', async ({
    page
  }) => {
    await login(page);

    await page.locator('[data-slot="sidebar-trigger"]').click();
    await expect(page.locator('[data-slot="sidebar"]')).toHaveAttribute('data-state', 'collapsed');

    const userTrigger = page.getByRole('button', { name: 'Benutzer', exact: true });
    const facilityTrigger = page.getByRole('button', { name: 'Anlage', exact: true });
    const userPanel = page
      .locator('[data-slot="navigation-menu-content"]')
      .filter({ has: page.getByText('Alle Benutzer', { exact: true }) });
    const facilityPanel = page
      .locator('[data-slot="navigation-menu-content"]')
      .filter({ has: page.getByText('Gebäude', { exact: true }) });

    await expect(userTrigger).toHaveCSS('background-color', 'rgba(0, 0, 0, 0)');
    await userTrigger.hover();
    await expect(userPanel).toBeVisible();
    await expect(userTrigger).toHaveAttribute('aria-expanded', 'true');
    await expect(userPanel.locator(':scope > div').first().locator('svg')).toHaveCount(0);

    const panelReceivesPointerEvents = await userPanel.evaluate((panel) => {
      const firstLink = panel.querySelector('a');
      if (!firstLink) return false;

      const bounds = firstLink.getBoundingClientRect();
      const elementAtCenter = document.elementFromPoint(
        bounds.left + bounds.width / 2,
        bounds.top + bounds.height / 2
      );

      return elementAtCenter?.closest('[data-slot="navigation-menu-content"]') === panel;
    });
    expect(panelReceivesPointerEvents).toBe(true);

    await facilityTrigger.hover();
    await expect(facilityPanel).toBeVisible();
    await expect(facilityTrigger).toHaveAttribute('aria-expanded', 'true');
    await expect(userTrigger).toHaveAttribute('aria-expanded', 'false');
  });

  test('uses a narrow transparent scrollbar that follows the sidebar theme', async ({ page }) => {
    await login(page);

    const sidebarContent = page.locator('[data-sidebar="content"]');
    const scrollbarStyles = await sidebarContent.evaluate((element) => {
      const scrollbar = getComputedStyle(element, '::-webkit-scrollbar');
      const track = getComputedStyle(element, '::-webkit-scrollbar-track');
      const thumb = getComputedStyle(element, '::-webkit-scrollbar-thumb');

      return {
        width: scrollbar.width,
        trackBackgroundColor: track.backgroundColor,
        thumbBorderRadius: thumb.borderRadius
      };
    });

    expect(scrollbarStyles).toEqual({
      width: '8px',
      trackBackgroundColor: 'rgba(0, 0, 0, 0)',
      thumbBorderRadius: '9999px'
    });
  });

  test('uses a pointer cursor for every enabled sidebar control', async ({ page }) => {
    await login(page);
    await expandNavigationGroups(page);

    const cursors = await page
      .locator('[data-sidebar="content"] :is(a[href], button, [role="button"], [role="link"])')
      .evaluateAll((elements) =>
        elements
          .filter(
            (element) =>
              element.checkVisibility() && !element.matches(':disabled, [aria-disabled="true"]')
          )
          .map((element) => getComputedStyle(element).cursor)
      );

    expect(cursors.length).toBeGreaterThan(25);
    expect(cursors.every((cursor) => cursor === 'pointer')).toBe(true);
  });

  test('opens every visible sidebar navigation target', async ({ page }) => {
    test.setTimeout(90_000);

    await login(page);
    await expandNavigationGroups(page);

    const sidebarContent = page.locator('[data-sidebar="content"]');
    const targets = await sidebarContent.locator('a[href^="/"]').evaluateAll((links) => {
      return [
        ...new Set(
          links
            .map((link) => link.getAttribute('href'))
            .filter((href): href is string => href !== null)
        )
      ];
    });

    expect(targets.length).toBeGreaterThanOrEqual(25);

    for (const target of targets) {
      await expandNavigationGroups(page);

      const link = sidebarContent.locator(`a[href="${target}"]`);
      await expect(link).toBeVisible();
      await Promise.all([page.waitForURL((url) => url.pathname === target), link.click()]);
      await expect(page).not.toHaveURL(/\/errors\/403$/);
    }
  });
});
