import {
  filterApparatsForRelationSource,
  filterApparatsForSystemPart,
  filterSystemPartsForRelationSource,
  filterSystemPartsForApparat,
  formatRelationSelectLabel,
  isSystemPartAllowedForApparat,
  mergeSelectedRelationOption
} from './relationSelectOptions.js';

describe('relation select options', () => {
  it('formats relation labels as short_name - name', () => {
    expect(
      formatRelationSelectLabel({
        id: 'apparat-1',
        short_name: 'Ven',
        name: 'Ventilator'
      })
    ).toBe('Ven - Ventilator');
  });

  it('keeps the selected embedded relation available when lookup items are empty', () => {
    const selected = {
      id: 'apparat-1',
      short_name: 'Ven',
      name: 'Ventilator'
    };

    expect(mergeSelectedRelationOption([], selected)).toEqual([selected]);
  });

  it('does not duplicate the selected relation when it is already in the lookup items', () => {
    const selected = {
      id: 'apparat-1',
      short_name: 'Ven',
      name: 'Ventilator'
    };

    expect(mergeSelectedRelationOption([selected], selected)).toEqual([selected]);
  });

  it('filters system parts by selected apparat relation', () => {
    const apparats = [
      {
        id: 'apparat-1',
        short_name: 'Ven',
        name: 'Ventilator',
        system_parts: [{ id: 'system-part-1', short_name: 'Ele', name: 'Elektro' }]
      },
      {
        id: 'apparat-2',
        short_name: 'Pmp',
        name: 'Pumpe',
        system_parts: [{ id: 'system-part-2', short_name: 'Hyd', name: 'Hydraulik' }]
      }
    ];
    const systemParts = [
      { id: 'system-part-1', short_name: 'Ele', name: 'Elektro' },
      { id: 'system-part-2', short_name: 'Hyd', name: 'Hydraulik' }
    ];

    expect(filterSystemPartsForApparat(systemParts, apparats, 'apparat-1')).toEqual([
      systemParts[0]
    ]);
  });

  it('filters apparats by selected system part relation', () => {
    const apparats = [
      {
        id: 'apparat-1',
        short_name: 'Ven',
        name: 'Ventilator',
        system_parts: [{ id: 'system-part-1', short_name: 'Ele', name: 'Elektro' }]
      },
      {
        id: 'apparat-2',
        short_name: 'Pmp',
        name: 'Pumpe',
        system_parts: [{ id: 'system-part-2', short_name: 'Hyd', name: 'Hydraulik' }]
      }
    ];

    expect(filterApparatsForSystemPart(apparats, 'system-part-2')).toEqual([apparats[1]]);
  });

  it('treats empty selections as unfiltered', () => {
    const apparats = [
      {
        id: 'apparat-1',
        short_name: 'Ven',
        name: 'Ventilator',
        system_parts: [{ id: 'system-part-1', short_name: 'Ele', name: 'Elektro' }]
      }
    ];
    const systemParts = [{ id: 'system-part-1', short_name: 'Ele', name: 'Elektro' }];

    expect(filterApparatsForSystemPart(apparats, '')).toEqual(apparats);
    expect(filterSystemPartsForApparat(systemParts, apparats, '')).toEqual(systemParts);
  });

  it('filters only from the actively changed relation side', () => {
    const apparats = [
      {
        id: 'apparat-1',
        short_name: 'Ven',
        name: 'Ventilator',
        system_parts: [{ id: 'system-part-1', short_name: 'Ele', name: 'Elektro' }]
      },
      {
        id: 'apparat-2',
        short_name: 'Pmp',
        name: 'Pumpe',
        system_parts: [{ id: 'system-part-2', short_name: 'Hyd', name: 'Hydraulik' }]
      }
    ];
    const systemParts = [
      { id: 'system-part-1', short_name: 'Ele', name: 'Elektro' },
      { id: 'system-part-2', short_name: 'Hyd', name: 'Hydraulik' }
    ];

    expect(filterApparatsForRelationSource(apparats, 'system-part-2', null)).toEqual(apparats);
    expect(filterApparatsForRelationSource(apparats, 'system-part-2', 'system_part_id')).toEqual([
      apparats[1]
    ]);
    expect(filterSystemPartsForRelationSource(systemParts, apparats, 'apparat-1', null)).toEqual(
      systemParts
    );
    expect(
      filterSystemPartsForRelationSource(systemParts, apparats, 'apparat-1', 'apparat_id')
    ).toEqual([systemParts[0]]);
  });

  it('checks whether a system part is allowed for an apparat', () => {
    const apparats = [
      {
        id: 'apparat-1',
        short_name: 'Ven',
        name: 'Ventilator',
        system_parts: [{ id: 'system-part-1', short_name: 'Ele', name: 'Elektro' }]
      }
    ];

    expect(isSystemPartAllowedForApparat(apparats, 'apparat-1', 'system-part-1')).toBe(true);
    expect(isSystemPartAllowedForApparat(apparats, 'apparat-1', 'system-part-2')).toBe(false);
  });
});
