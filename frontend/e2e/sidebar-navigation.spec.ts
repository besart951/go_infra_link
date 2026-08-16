import { expect, login, test } from './fixtures';

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
});
