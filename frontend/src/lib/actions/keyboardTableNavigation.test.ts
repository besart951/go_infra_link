/// <reference types="vitest" />

import { keyboardTableNavigation } from './keyboardTableNavigation.js';

function makeCell(
  row: string,
  column: string,
  content: string,
  activate: 'focus' | 'edit' = 'focus'
): string {
  return `
    <div
      data-keyboard-table-cell
      data-keyboard-table-row="${row}"
      data-keyboard-table-column="${column}"
      data-keyboard-table-activate="${activate}"
    >
      ${content}
    </div>
  `;
}

function dispatchKey(element: Element, key: string, options: KeyboardEventInit = {}): void {
  element.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true, ...options }));
}

async function afterNavigation(): Promise<void> {
  await Promise.resolve();
}

describe('keyboardTableNavigation', () => {
  let root: HTMLElement;
  let destroy: () => void;

  beforeEach(() => {
    root = document.createElement('div');
    document.body.append(root);
  });

  afterEach(() => {
    destroy?.();
    root.remove();
    document.body.innerHTML = '';
  });

  it('moves horizontally with arrow keys', async () => {
    root.innerHTML = [
      makeCell('row-1', 'bmk', '<button type="button" id="bmk">BMK</button>'),
      makeCell(
        'row-1',
        'description',
        '<button type="button" id="description">Description</button>'
      )
    ].join('');
    destroy = keyboardTableNavigation(root).destroy;

    document.querySelector<HTMLElement>('#bmk')?.focus();
    dispatchKey(document.activeElement!, 'ArrowRight');
    await afterNavigation();

    expect(document.activeElement).toBe(document.querySelector('#description'));
  });

  it('moves vertically with Tab', async () => {
    root.innerHTML = [
      makeCell('row-1', 'bmk', '<button type="button" id="row-1-bmk">Row 1 BMK</button>'),
      makeCell(
        'row-1',
        'description',
        '<button type="button" id="row-1-description">Row 1 Description</button>'
      ),
      makeCell('row-2', 'bmk', '<button type="button" id="row-2-bmk">Row 2 BMK</button>'),
      makeCell(
        'row-2',
        'description',
        '<button type="button" id="row-2-description">Row 2 Description</button>'
      )
    ].join('');
    destroy = keyboardTableNavigation(root).destroy;

    document.querySelector<HTMLElement>('#row-1-description')?.focus();
    dispatchKey(document.activeElement!, 'Tab');
    await afterNavigation();

    expect(document.activeElement).toBe(document.querySelector('#row-2-description'));
  });

  it('activates the next edit cell after Enter from an input', async () => {
    let clicked = false;
    root.innerHTML = [
      makeCell('row-1', 'bmk', '<input id="bmk" value="FD-001" />', 'edit'),
      makeCell(
        'row-1',
        'description',
        '<button type="button" id="description">Description</button>',
        'edit'
      )
    ].join('');
    destroy = keyboardTableNavigation(root).destroy;
    document.querySelector('#description')?.addEventListener('click', () => {
      clicked = true;
    });

    document.querySelector<HTMLInputElement>('#bmk')?.focus();
    dispatchKey(document.activeElement!, 'Enter');
    await afterNavigation();

    expect(document.activeElement).toBe(document.querySelector('#description'));
    expect(clicked).toBe(true);
  });

  it('lets Enter open a focused button instead of navigating away', async () => {
    root.innerHTML = [
      makeCell('row-1', 'apparat', '<button type="button" id="apparat">Apparat</button>'),
      makeCell(
        'row-1',
        'system-part',
        '<button type="button" id="system-part">System Part</button>'
      )
    ].join('');
    destroy = keyboardTableNavigation(root).destroy;

    document.querySelector<HTMLElement>('#apparat')?.focus();
    dispatchKey(document.activeElement!, 'Enter');
    await afterNavigation();

    expect(document.activeElement).toBe(document.querySelector('#apparat'));
  });

  it('skips disabled cells', async () => {
    root.innerHTML = [
      makeCell('row-1', 'bmk', '<button type="button" id="bmk">BMK</button>'),
      makeCell(
        'row-1',
        'description',
        '<button type="button" id="description" disabled>Description</button>'
      ),
      makeCell('row-1', 'text-fix', '<button type="button" id="text-fix">Text Fix</button>')
    ].join('');
    destroy = keyboardTableNavigation(root).destroy;

    document.querySelector<HTMLElement>('#bmk')?.focus();
    dispatchKey(document.activeElement!, 'ArrowRight');
    await afterNavigation();

    expect(document.activeElement).toBe(document.querySelector('#text-fix'));
  });

  it('moves between nested subrows by column', async () => {
    root.innerHTML = [
      makeCell('field-device-1', 'bmk', '<button type="button" id="field-bmk">BMK</button>'),
      makeCell(
        'bacnet:object-1',
        'bacnet.description',
        '<button type="button" id="object-1-description">Object 1 Description</button>'
      ),
      makeCell(
        'bacnet:object-1',
        'bacnet.optional',
        '<button type="button" id="object-1-optional">Object 1 Optional</button>'
      ),
      makeCell(
        'bacnet:object-2',
        'bacnet.description',
        '<button type="button" id="object-2-description">Object 2 Description</button>'
      ),
      makeCell(
        'bacnet:object-2',
        'bacnet.optional',
        '<button type="button" id="object-2-optional">Object 2 Optional</button>'
      )
    ].join('');
    destroy = keyboardTableNavigation(root).destroy;

    document.querySelector<HTMLElement>('#object-1-description')?.focus();
    dispatchKey(document.activeElement!, 'ArrowDown');
    await afterNavigation();

    expect(document.activeElement).toBe(document.querySelector('#object-2-description'));
  });

  it('does not intercept keys from ignored nested controls', async () => {
    root.innerHTML = [
      makeCell(
        'row-1',
        'type',
        '<select id="type" data-keyboard-table-ignore><option>ai</option></select>'
      ),
      makeCell('row-2', 'type', '<button type="button" id="next-type">Next Type</button>')
    ].join('');
    destroy = keyboardTableNavigation(root).destroy;

    document.querySelector<HTMLElement>('#type')?.focus();
    dispatchKey(document.activeElement!, 'ArrowDown');
    await afterNavigation();

    expect(document.activeElement).toBe(document.querySelector('#type'));
  });
});
