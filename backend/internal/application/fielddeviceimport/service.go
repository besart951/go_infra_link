package fielddeviceimport

import (
	"context"
	"errors"
	"fmt"
)

var ErrInvalidWorkbook = errors.New("invalid field device workbook")

type Service struct {
	reader WorkbookReader
	store  StagingStore
	writer AggregateWriter
}

func NewService(reader WorkbookReader, store StagingStore, writer AggregateWriter) *Service {
	return &Service{reader: reader, store: store, writer: writer}
}

func (s *Service) Import(ctx context.Context, command Command) (Result, error) {
	result, session, err := s.stage(ctx, command)
	if err != nil {
		return result, err
	}
	if issues, validateErr := session.Validate(ctx); validateErr != nil || len(issues) > 0 {
		result.Issues = issues
		_ = session.Discard(ctx)
		return result, errors.Join(ErrInvalidWorkbook, validateErr)
	}
	return s.write(ctx, session, result)
}

func (s *Service) stage(ctx context.Context, command Command) (Result, Session, error) {
	id, session, err := s.store.Start(ctx, command.OwnerID)
	result := Result{ImportID: id}
	if err != nil {
		return result, nil, err
	}
	manifest, err := s.reader.Read(ctx, command.Source, session)
	if err != nil {
		_ = session.Discard(ctx)
		return result, nil, fmt.Errorf("%w: read import workbook: %v", ErrInvalidWorkbook, err)
	}
	result.Total = manifest.DeviceCount
	if manifest.SchemaVersion != SchemaVersion {
		_ = session.Discard(ctx)
		return result, nil, fmt.Errorf("%w: schema version %d", ErrInvalidWorkbook, manifest.SchemaVersion)
	}
	if err := session.Seal(ctx, manifest); err != nil {
		_ = session.Discard(ctx)
		return result, nil, err
	}
	return result, session, nil
}

func (s *Service) write(ctx context.Context, session Session, result Result) (Result, error) {
	cursor := ""
	for {
		page, err := session.Aggregates(ctx, cursor)
		if err != nil {
			return result, err
		}
		for index := range page.Items {
			if err := s.writer.Import(ctx, page.Items[index]); err != nil {
				result.Failed++
				result.Issues = append(result.Issues, Issue{
					Code: "aggregate_import_failed", Entity: "field_device",
					SourceID: page.Items[index].FieldDevice.ID, Message: err.Error(),
				})
				continue
			}
			result.Imported++
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	if err := session.Complete(ctx); err != nil {
		return result, err
	}
	return result, nil
}
