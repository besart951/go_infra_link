/// <reference types="vitest" />

import {
  fieldErrorPathMatches,
  findIndexedFieldPathSegment,
  getRelevantFieldPathSegments,
  resolveFieldError,
  splitFieldPath
} from './fieldPath.js';

describe('field path helpers', () => {
  it('normalizes wrapper, index, and domain alias segments for matching', () => {
    expect(
      fieldErrorPathMatches(
        'data.field_devices[0].bacnet_objects[0].alarm_type_id',
        'fielddevice.bacnetobjects.alarm_type_id'
      )
    ).toBe(true);
    expect(
      fieldErrorPathMatches('data.field_devices[0].bacnet_objects[0].alarm_type_id', 'fielddevice')
    ).toBe(false);
  });

  it('extracts segments after indexed update collections', () => {
    expect(getRelevantFieldPathSegments('updates[1].specifications.specification_brand')).toEqual([
      'specifications',
      'specification_brand'
    ]);
  });

  it('resolves field errors through aliases', () => {
    const errors = {
      'data.sps_controller.control_cabinet_id': 'cabinet required'
    };

    expect(resolveFieldError(errors, 'control_cabinet_id', ['spscontroller'])).toBe(
      'cabinet required'
    );
  });

  it('finds indexed collection segments', () => {
    const segments = splitFieldPath('field_devices[2].bmk');

    expect(findIndexedFieldPathSegment(segments, ['fielddevices'])).toEqual({
      segmentIndex: 0,
      index: 2
    });
  });
});
