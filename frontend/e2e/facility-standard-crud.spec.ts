import { expect, login, navigateToFacility, test, uniqueName } from './fixtures';
import type { Page } from '@playwright/test';

test.describe('facility standard CRUD', () => {
  test('covers the standalone reference catalogs through visible controls', async ({ page }) => {
    await login(page);

    // The E2E seed deliberately keeps this short range free. A fixed,
    // documented reservation is more reliable than guessing around the
    // domain's occupied system-type ranges.
    const rangeStart = 7800;
    await createUpdateDelete({
      page,
      destination: 'Systemtypen',
      pathname: '/facility/system-types',
      createLabel: 'Neuer Systemtyp',
      endpoint: '/api/v1/facility/system-types',
      rowName: uniqueName('E2E Systemtyp'),
      updatedRowName: undefined,
      create: async (name) => {
        await page.locator('#system_type_name').fill(name);
        await page.locator('#system_type_min').fill(String(rangeStart));
        await page.locator('#system_type_max').fill(String(rangeStart + 1));
      },
      update: async (name) => {
        await page.locator('#system_type_name').fill(name);
      }
    });

    const stateReference = 90_000 + Math.floor(Math.random() * 9_000);
    await createUpdateDelete({
      page,
      destination: 'Statustextgruppen',
      pathname: '/facility/state-texts',
      createLabel: 'Neuer Statustext',
      endpoint: '/api/v1/facility/state-texts',
      rowName: uniqueName('E2E Statustext'),
      create: async (name) => {
        await page.locator('#state_ref').fill(String(stateReference));
        await page.locator('#state_text1').fill(name);
      },
      update: async (name) => {
        await page.locator('#state_text1').fill(name);
      }
    });

    await createUpdateDelete({
      page,
      destination: 'Benachrichtigungsklassen',
      pathname: '/facility/notification-classes',
      createLabel: 'Neue Benachrichtigungsklasse',
      endpoint: '/api/v1/facility/notification-classes',
      rowName: uniqueName('E2E Kategorie'),
      create: async (name) => {
        await page.locator('#nc_event').fill(name);
        await page.locator('#nc_value').fill(String(Math.floor(Math.random() * 10_000) + 50_000));
        await page.locator('#nc_object_desc').fill('E2E Objektbeschreibung');
        await page.locator('#nc_internal_desc').fill('E2E Interne Beschreibung');
        await page.locator('#nc_meaning').fill('E2E Bedeutung');
      },
      update: async (name) => {
        await page.locator('#nc_event').fill(name);
      }
    });

    await createUpdateDelete({
      page,
      destination: 'Alarmdefinitionen',
      pathname: '/facility/alarm-definitions',
      createLabel: 'Neue Alarmdefinition',
      endpoint: '/api/v1/facility/alarm-definitions',
      rowName: uniqueName('E2E Alarmdefinition'),
      create: async (name) => {
        await page.locator('#alarm_name').fill(name);
        await page.locator('#alarm_note').fill('E2E Alarmnotiz');
        await page.getByRole('combobox').click();
        await page.getByRole('option').first().click();
      },
      update: async (name) => {
        await page.locator('#alarm_name').fill(name);
      }
    });

    await createUpdateDelete({
      page,
      destination: 'Objektdaten',
      pathname: '/facility/object-data',
      createLabel: 'Neue Objektdaten',
      endpoint: '/api/v1/facility/object-data',
      rowName: uniqueName('E2E Objektdaten'),
      create: async (name) => {
        await page.locator('#object_data_description').fill(name);
        await page.locator('#object_data_version').fill('1.0-e2e');
      },
      update: async (name) => {
        await page.locator('#object_data_description').fill(name);
      }
    });
  });
});

interface CrudScenario {
  page: Page;
  destination: string;
  pathname: string;
  createLabel: string;
  endpoint: string;
  rowName: string;
  updatedRowName?: string;
  create: (name: string) => Promise<void>;
  update: (name: string) => Promise<void>;
}

async function createUpdateDelete(scenario: CrudScenario): Promise<void> {
  const updatedRowName = scenario.updatedRowName ?? `${scenario.rowName} aktualisiert`;
  const { page, endpoint } = scenario;

  await navigateToFacility(page, scenario.destination, scenario.pathname);
  await page.getByLabel(scenario.createLabel, { exact: true }).click();
  await scenario.create(scenario.rowName);
  await expectSuccessfulMutation(page, 'POST', endpoint, () =>
    page.locator('form').getByRole('button', { name: 'Erstellen' }).click()
  );
  await expect(page.locator('tr', { hasText: scenario.rowName })).toBeVisible();

  const initialRow = page.locator('tr', { hasText: scenario.rowName });
  await initialRow.getByRole('button').click();
  await page.getByRole('menuitem', { name: 'Bearbeiten' }).click();
  await scenario.update(updatedRowName);
  await expectSuccessfulMutation(page, 'PUT', `${endpoint}/`, () =>
    page.locator('form').getByRole('button', { name: 'Aktualisieren' }).click()
  );
  await expect(page.locator('tr', { hasText: updatedRowName })).toBeVisible();

  const updatedRow = page.locator('tr', { hasText: updatedRowName });
  await updatedRow.getByRole('button').click();
  await page.getByRole('menuitem', { name: 'Löschen' }).click();
  await expectSuccessfulMutation(page, 'DELETE', `${endpoint}/`, () =>
    page.getByRole('dialog').getByRole('button', { name: 'Löschen' }).click()
  );
  await expect(page.locator('tr', { hasText: updatedRowName })).toHaveCount(0);
}

async function expectSuccessfulMutation(
  page: Page,
  method: string,
  endpointPrefix: string,
  action: () => Promise<void>
): Promise<void> {
  const response = page.waitForResponse(
    (candidate) =>
      candidate.request().method() === method &&
      new URL(candidate.url()).pathname.startsWith(endpointPrefix)
  );
  await action();
  expect((await response).ok()).toBe(true);
}
