import { describe, expect, it } from 'vitest';

import { buildAdminUser, buildUser } from '../../helpers/permissions.js';
import { load } from '../../../src/routes/(app)/admin/notifications/smtp/+page';

describe('/admin/notifications/smtp load guard', () => {
  it('redirects guests to /login', async () => {
    await expect(
      load({
        parent: async () => ({ user: null })
      } as never)
    ).rejects.toMatchObject({ status: 302, location: '/login' });
  });

  it('redirects non-superadmin users back to the dashboard', async () => {
    await expect(
      load({
        parent: async () => ({ user: buildUser() })
      } as never)
    ).rejects.toMatchObject({ status: 302, location: '/' });
  });

  it('allows superadmin users through', async () => {
    const user = buildAdminUser();

    await expect(
      load({
        parent: async () => ({ user })
      } as never)
    ).resolves.toEqual({ user });
  });
});
