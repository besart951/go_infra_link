import { describe, expect, it } from 'vitest';

import { permission } from '../../helpers/permissions.js';
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

  it('redirects users without SMTP manage permission back to the dashboard', async () => {
    await expect(
      load({
        parent: async () => ({ user: buildUser() })
      } as never)
    ).rejects.toMatchObject({ status: 302, location: '/' });
  });

  it('allows users with SMTP manage permission through', async () => {
    const user = buildAdminUser({ permissions: [permission('notification.smtp', 'manage')] });

    await expect(
      load({
        parent: async () => ({ user })
      } as never)
    ).resolves.toEqual({ user });
  });
});
