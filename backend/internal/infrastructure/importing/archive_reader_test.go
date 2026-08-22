package importing

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestArchiveReaderReadsWorkbookPackage(t *testing.T) {
	firstID, secondID := uuid.New(), uuid.New()
	first := newImportFixtureForDevice(t, firstID, 2)
	second := newImportFixtureForDevice(t, secondID, 2)
	archive := workbookPackage(t, map[string][]byte{"cabinet-b.xlsx": second, "cabinet-a.xlsx": first})
	sink := &captureSink{}

	manifest, err := NewArchiveReader(NewExcelizeReader()).Read(context.Background(), bytes.NewReader(archive), sink)

	if err != nil {
		t.Fatal(err)
	}
	if manifest.DeviceCount != 2 || len(sink.devices) != 2 {
		t.Fatalf("manifest=%+v devices=%d", manifest, len(sink.devices))
	}
	if sink.devices[0].ID != firstID || sink.devices[1].ID != secondID {
		t.Fatalf("workbooks were not read deterministically: %v %v", sink.devices[0].ID, sink.devices[1].ID)
	}
}

func TestArchiveReaderRejectsInconsistentManifests(t *testing.T) {
	first := newImportFixtureForDevice(t, uuid.New(), 1)
	second := newImportFixtureForDevice(t, uuid.New(), 2)
	archive := workbookPackage(t, map[string][]byte{"a.xlsx": first, "b.xlsx": second})

	_, err := NewArchiveReader(NewExcelizeReader()).Read(context.Background(), bytes.NewReader(archive), &captureSink{})

	if err == nil {
		t.Fatal("expected inconsistent manifest error")
	}
}

func TestArchiveReaderStillReadsSingleWorkbook(t *testing.T) {
	fixture := newImportFixture(t)
	sink := &captureSink{}

	manifest, err := NewArchiveReader(NewExcelizeReader()).Read(context.Background(), bytes.NewReader(fixture), sink)

	if err != nil || manifest.DeviceCount != 1 || len(sink.devices) != 1 {
		t.Fatalf("manifest=%+v devices=%d err=%v", manifest, len(sink.devices), err)
	}
}

func workbookPackage(t *testing.T, workbooks map[string][]byte) []byte {
	t.Helper()
	var output bytes.Buffer
	archive := zip.NewWriter(&output)
	for name, content := range workbooks {
		entry, err := archive.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
