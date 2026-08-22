package fielddeviceimport

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	facility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type readerStub struct {
	manifest Manifest
	err      error
}

func (r readerStub) Read(context.Context, io.Reader, Sink) (Manifest, error) {
	return r.manifest, r.err
}

type storeStub struct{ session *sessionStub }

func (s storeStub) Start(context.Context, uuid.UUID) (uuid.UUID, Session, error) {
	return uuid.New(), s.session, nil
}

type sessionStub struct {
	noopSink
	issues    []Issue
	pages     []AggregatePage
	page      int
	discarded bool
	completed bool
}

func (s *sessionStub) Validate(context.Context) ([]Issue, error) { return s.issues, nil }
func (s *sessionStub) Seal(context.Context, Manifest) error      { return nil }
func (s *sessionStub) Aggregates(context.Context, string) (AggregatePage, error) {
	page := s.pages[s.page]
	s.page++
	return page, nil
}
func (s *sessionStub) Complete(context.Context) error { s.completed = true; return nil }
func (s *sessionStub) Discard(context.Context) error  { s.discarded = true; return nil }

type writerStub struct{ calls int }

func (w *writerStub) Import(context.Context, Aggregate) error { w.calls++; return nil }

func TestServiceRejectsWorkbookBeforeMutation(t *testing.T) {
	session := &sessionStub{issues: []Issue{{Code: "orphan", Message: "orphan object"}}}
	writer := &writerStub{}
	service := NewService(readerStub{manifest: Manifest{SchemaVersion: SchemaVersion}}, storeStub{session}, writer)

	result, err := service.Import(context.Background(), Command{Source: strings.NewReader("data")})

	if !errors.Is(err, ErrInvalidWorkbook) || len(result.Issues) != 1 {
		t.Fatalf("expected invalid workbook result, got result=%+v err=%v", result, err)
	}
	if writer.calls != 0 || !session.discarded {
		t.Fatalf("mutation happened before validation: calls=%d discarded=%v", writer.calls, session.discarded)
	}
}

func TestServiceImportsValidatedAggregatesPageByPage(t *testing.T) {
	session := &sessionStub{pages: []AggregatePage{
		{Items: []Aggregate{{}}, NextCursor: "next"},
		{Items: []Aggregate{{}}},
	}}
	writer := &writerStub{}
	service := NewService(readerStub{manifest: Manifest{SchemaVersion: SchemaVersion, DeviceCount: 2}}, storeStub{session}, writer)

	result, err := service.Import(context.Background(), Command{Source: strings.NewReader("data")})

	if err != nil || result.Imported != 2 || writer.calls != 2 || !session.completed {
		t.Fatalf("unexpected import result=%+v calls=%d completed=%v err=%v", result, writer.calls, session.completed, err)
	}
}

type noopSink struct{}

func (noopSink) FieldDevices(context.Context, []facility.FieldDevice) error           { return nil }
func (noopSink) Specifications(context.Context, []facility.Specification) error       { return nil }
func (noopSink) BacnetObjects(context.Context, []facility.BacnetObject) error         { return nil }
func (noopSink) SoftwareReferences(context.Context, []SoftwareReference) error        { return nil }
func (noopSink) AlarmValues(context.Context, []facility.BacnetObjectAlarmValue) error { return nil }
