import {
  createProject,
  expectActiveSocketCount,
  expectNoConcurrentSockets,
  login,
  navigateToProjectOverview,
  logout,
  observeWebSockets,
  test
} from './fixtures';

test.describe('websocket lifecycle', () => {
  test('keeps one global stream per concern and disposes project and global streams', async ({
    page
  }) => {
    const sockets = observeWebSockets(page);

    await login(page);
    await expectActiveSocketCount(sockets, '/api/v1/facility/reference-data/stream', 1);
    await expectActiveSocketCount(sockets, '/api/v1/account/notifications/stream', 1);
    expectNoConcurrentSockets(sockets, '/api/v1/facility/reference-data/stream');
    expectNoConcurrentSockets(sockets, '/api/v1/account/notifications/stream');

    const projectID = await createProject(page);
    await expectActiveSocketCount(sockets, `/api/v1/projects/${projectID}/collaboration`, 1);
    expectNoConcurrentSockets(sockets, `/api/v1/projects/${projectID}/collaboration`);

    await navigateToProjectOverview(page);
    await expectActiveSocketCount(sockets, `/api/v1/projects/${projectID}/collaboration`, 0);
    await expectActiveSocketCount(sockets, '/api/v1/facility/reference-data/stream', 1);
    await expectActiveSocketCount(sockets, '/api/v1/account/notifications/stream', 1);
    expectNoConcurrentSockets(sockets, `/api/v1/projects/${projectID}/collaboration`);
    expectNoConcurrentSockets(sockets, '/api/v1/facility/reference-data/stream');
    expectNoConcurrentSockets(sockets, '/api/v1/account/notifications/stream');

    await logout(page);
    await expectActiveSocketCount(sockets, '/api/v1/facility/reference-data/stream', 0);
    await expectActiveSocketCount(sockets, '/api/v1/account/notifications/stream', 0);
  });
});
