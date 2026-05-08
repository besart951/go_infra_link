import { describe, expect, it } from 'vitest';
import { parseProjectCollaborationInboundMessage } from './projectCollaboration.svelte.js';

const projectId = '11111111-1111-4111-8111-111111111111';
const userId = '22222222-2222-4222-8222-222222222222';
const deviceId = '33333333-3333-4333-8333-333333333333';
const bacnetObjectId = '44444444-4444-4444-8444-444444444444';

describe('project collaboration realtime message validation', () => {
  it('accepts a valid snapshot message', () => {
    const parsed = parseProjectCollaborationInboundMessage({
      type: 'snapshot',
      project_id: projectId,
      presence: [
        {
          user_id: userId,
          connected_at: '2026-01-01T10:00:00Z',
          last_seen_at: '2026-01-01T10:00:01Z'
        }
      ],
      edit_states: [
        {
          user_id: userId,
          devices: [
            {
              device_id: deviceId,
              changed_fields: ['text_fix', `bacnet_objects.${bacnetObjectId}.software_type`],
              field_values: {
                text_fix: 'TF-1',
                [`bacnet_objects.${bacnetObjectId}.software_type`]: 'ai'
              }
            }
          ],
          updated_at: '2026-01-01T10:00:02Z'
        }
      ],
      at: '2026-01-01T10:00:03Z'
    });

    expect(parsed.type).toBe('snapshot');
  });

  it('rejects unknown message types and invalid UUIDs', () => {
    expect(() =>
      parseProjectCollaborationInboundMessage({
        type: 'snapshot',
        project_id: projectId,
        presence: [{ user_id: 'not-a-uuid', connected_at: 'now', last_seen_at: 'now' }],
        edit_states: []
      })
    ).toThrow();

    expect(() =>
      parseProjectCollaborationInboundMessage({
        type: 'delete_everything'
      })
    ).toThrow();
  });

  it('rejects field values that are not declared as changed fields', () => {
    expect(() =>
      parseProjectCollaborationInboundMessage({
        type: 'edit_states',
        project_id: projectId,
        edit_states: [
          {
            user_id: userId,
            devices: [
              {
                device_id: deviceId,
                changed_fields: ['text_fix'],
                field_values: {
                  description: 'not declared'
                }
              }
            ],
            updated_at: '2026-01-01T10:00:00Z'
          }
        ]
      })
    ).toThrow();
  });

  it('rejects unknown entity delta fields before they reach UI state', () => {
    expect(() =>
      parseProjectCollaborationInboundMessage({
        type: 'entity_delta',
        project_id: projectId,
        scope: 'field_device',
        field_devices: [
          {
            id: deviceId,
            apparat_nr: 1,
            sps_controller_system_type_id: projectId,
            apparat_id: projectId,
            created_at: '2026-01-01T10:00:00Z',
            updated_at: '2026-01-01T10:00:00Z',
            admin_note: 'not allowed'
          }
        ],
        at: '2026-01-01T10:00:00Z'
      })
    ).toThrow();
  });
});
