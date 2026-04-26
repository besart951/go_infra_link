import { describe, expect, it } from 'vitest';

import { groupSystemTypesByController } from './groupSystemTypesByController.js';

describe('groupSystemTypesByController', () => {
  it('groups items by controller id and ignores missing controller ids', () => {
    const grouped = groupSystemTypesByController([
      {
        id: 'type-1',
        sps_controller_id: 'controller-1',
        system_type_id: 'system-1',
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z'
      },
      {
        id: 'type-2',
        sps_controller_id: 'controller-1',
        system_type_id: 'system-2',
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z'
      },
      {
        id: 'type-3',
        sps_controller_id: 'controller-2',
        system_type_id: 'system-3',
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z'
      },
      {
        id: 'type-4',
        sps_controller_id: '',
        system_type_id: 'system-4',
        created_at: '2026-01-01T00:00:00Z',
        updated_at: '2026-01-01T00:00:00Z'
      }
    ]);

    expect(grouped).toEqual({
      'controller-1': [
        {
          id: 'type-1',
          sps_controller_id: 'controller-1',
          system_type_id: 'system-1',
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z'
        },
        {
          id: 'type-2',
          sps_controller_id: 'controller-1',
          system_type_id: 'system-2',
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z'
        }
      ],
      'controller-2': [
        {
          id: 'type-3',
          sps_controller_id: 'controller-2',
          system_type_id: 'system-3',
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z'
        }
      ]
    });
  });
});
