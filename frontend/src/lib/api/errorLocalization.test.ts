/// <reference types="vitest" />

import { localizeErrorText } from './errorLocalization.js';

describe('error localization', () => {
  it('localizes facility system type overlap errors', () => {
    expect(
      localizeErrorText(
        'number_min and number_max range must not overlap existing ranges',
        'systemtype.number_min'
      )
    ).toBe('Nummer von und Nummer bis dürfen sich nicht mit bestehenden Bereichen überschneiden.');
  });

  it('localizes building group scopes in uniqueness errors', () => {
    expect(
      localizeErrorText('iws_code must be unique within the building group', 'building.iws_code')
    ).toBe('IWS-Code muss innerhalb von Gebäudegruppe eindeutig sein.');
  });

  it('localizes simple unique field errors', () => {
    expect(localizeErrorText('short_name must be unique', 'apparat.short_name')).toBe(
      'Kurzname muss eindeutig sein.'
    );
    expect(localizeErrorText('name must be unique', 'apparat.name')).toBe(
      'Name muss eindeutig sein.'
    );
  });
});
