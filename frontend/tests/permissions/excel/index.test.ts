import { render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { permission } from '../../helpers/permissions.js';

const state = vi.hoisted(() => {
  const grantedPermissions = new Set<string>();

  return {
    addToastMock: vi.fn(),
    setPermissions(permissions: string[]) {
      grantedPermissions.clear();
      for (const granted of permissions) {
        grantedPermissions.add(granted);
      }
    },
    resetPermissions() {
      grantedPermissions.clear();
    },
    canPerform(action: string, resource: string) {
      return grantedPermissions.has(`${resource}.${action}`);
    },
    executeMock: vi.fn(),
    cancelMock: vi.fn()
  };
});

vi.mock('$lib/utils/permissions.js', () => ({
  canPerform: (action: string, resource: string) => state.canPerform(action, resource)
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: state.addToastMock
}));

vi.mock('$lib/application/useCases/excel/startExcelReadSessionUseCase.js', () => ({
  StartExcelReadSessionUseCase: class {
    execute = state.executeMock;
    cancel = state.cancelMock;
  }
}));

vi.mock('$lib/infrastructure/excel/excelWorkerReaderAdapter.js', () => ({
  ExcelWorkerReaderAdapter: class {}
}));

vi.mock('$lib/components/excel/ExcelUploadDropzone.svelte', async () => {
  const { default: ExcelUploadDropzoneStub } =
    await import('../../setup/stubs/ExcelUploadDropzoneStub.svelte');
  return { default: ExcelUploadDropzoneStub };
});

vi.mock('$lib/components/excel/ExcelReadProgressCard.svelte', async () => {
  const { default: ExcelReadProgressCardStub } =
    await import('../../setup/stubs/ExcelReadProgressCardStub.svelte');
  return { default: ExcelReadProgressCardStub };
});

vi.mock('$lib/components/excel/ExcelSessionSummary.svelte', async () => {
  const { default: ExcelSessionSummaryStub } =
    await import('../../setup/stubs/ExcelSessionSummaryStub.svelte');
  return { default: ExcelSessionSummaryStub };
});

import ExcelPage from '../../../src/routes/(app)/excel/+page.svelte';

describe('/excel permission surface', () => {
  beforeEach(() => {
    state.resetPermissions();
    state.addToastMock.mockReset();
    state.executeMock.mockReset();
    state.cancelMock.mockReset();
  });

  it('shows the unauthorized message when objectdata.create is missing', () => {
    render(ExcelPage);

    expect(
      screen.getByText('Sie haben keine Berechtigung, Excel-Daten zu importieren.')
    ).toBeInTheDocument();
    expect(screen.queryByTestId('excel-upload-dropzone')).not.toBeInTheDocument();
  });

  it('still denies a user who only has the sidebar-level read permission', () => {
    state.setPermissions([permission('objectdata')]);

    render(ExcelPage);

    expect(
      screen.getByText('Sie haben keine Berechtigung, Excel-Daten zu importieren.')
    ).toBeInTheDocument();
    expect(screen.queryByTestId('excel-upload-dropzone')).not.toBeInTheDocument();
  });

  it('renders the upload workflow when objectdata.create is granted', () => {
    state.setPermissions([permission('objectdata', 'create')]);

    render(ExcelPage);

    expect(screen.getByTestId('excel-upload-dropzone')).toBeInTheDocument();
    expect(screen.getByTestId('excel-read-progress')).toBeInTheDocument();
    expect(
      screen.queryByText('Sie haben keine Berechtigung, Excel-Daten zu importieren.')
    ).not.toBeInTheDocument();
  });
});
