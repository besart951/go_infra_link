import { describe, expect, expectTypeOf, it } from 'vitest';
import { apiClient } from './generated/client.js';
import type { components } from './generated/schema.js';

type CreateProjectRequest =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_project.CreateProjectRequest'];

describe('generated OpenAPI contract', () => {
  it('preserves required fields and valid route parameters at compile time', () => {
    const request: CreateProjectRequest = {
      name: 'Terminal A',
      phase_id: 'phase-1'
    };

    expectTypeOf(request.name).toEqualTypeOf<string>();

    if (false) {
      apiClient.GET('/api/v1/projects/{id}', {
        params: { path: { id: 'project-1' } }
      });

      // @ts-expect-error Project creation requires a name.
      apiClient.POST('/api/v1/projects', { body: { phase_id: 'phase-1' } });

      // @ts-expect-error The generated route requires its id path parameter.
      apiClient.GET('/api/v1/projects/{id}', { params: { path: {} } });

      const invalidProject: CreateProjectRequest = {
        name: 'Terminal A',
        phase_id: 'phase-1',
        // @ts-expect-error Unknown request properties are rejected by the generated schema.
        unsupported: true
      };
      expect(invalidProject).toBeDefined();
    }

    expect(request.phase_id).toBe('phase-1');
  });
});
