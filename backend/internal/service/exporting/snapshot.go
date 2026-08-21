package exporting

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	domainExport "github.com/besart951/go_infra_link/backend/internal/domain/exporting"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

const exportSnapshotSchemaVersion = 1

type exportSnapshotManifest struct {
	SchemaVersion int                       `json:"schema_version"`
	SnapshotAt    time.Time                 `json:"snapshot_at"`
	Controllers   []domainExport.Controller `json:"controllers"`
	DeviceCount   int64                     `json:"device_count"`
}

type snapshotReadState struct {
	file    *os.File
	gzip    *gzip.Reader
	decoder *json.Decoder
	lastID  uuid.UUID
	done    bool
}

type snapshotDataProvider struct {
	directory string
	manifest  exportSnapshotManifest
	mu        sync.Mutex
	states    map[uuid.UUID]*snapshotReadState
}

func createOrOpenSnapshot(
	ctx context.Context,
	directory string,
	req domainExport.Request,
	live domainExport.SnapshotDataProvider,
	pageSize int,
	report func(processed int64),
) (*snapshotDataProvider, error) {
	if existing, err := openSnapshot(directory); err == nil {
		return existing, nil
	}
	if err := os.RemoveAll(directory); err != nil {
		return nil, fmt.Errorf("clear incomplete export snapshot: %w", err)
	}
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return nil, fmt.Errorf("create export snapshot directory: %w", err)
	}

	manifest := exportSnapshotManifest{SchemaVersion: exportSnapshotSchemaVersion, SnapshotAt: time.Now().UTC()}
	err := live.WithinSnapshot(ctx, func(source domainExport.DataProvider) error {
		controllers, err := source.ResolveControllers(ctx, req)
		if err != nil {
			return err
		}
		for _, controller := range controllers {
			count, err := writeControllerSnapshot(ctx, directory, controller, source, req, pageSize, func(delta int64) {
				manifest.DeviceCount += delta
				if report != nil {
					report(manifest.DeviceCount)
				}
			})
			if err != nil {
				return err
			}
			if count > 0 {
				manifest.Controllers = append(manifest.Controllers, controller)
			}
		}
		return nil
	})
	if err != nil {
		_ = os.RemoveAll(directory)
		return nil, fmt.Errorf("create consistent export snapshot: %w", err)
	}
	if len(manifest.Controllers) == 0 {
		_ = os.RemoveAll(directory)
		return nil, errors.New("export selection contains no field devices")
	}
	if err := writeSnapshotManifest(directory, manifest); err != nil {
		_ = os.RemoveAll(directory)
		return nil, err
	}
	return openSnapshot(directory)
}

func writeControllerSnapshot(
	ctx context.Context,
	directory string,
	controller domainExport.Controller,
	source domainExport.DataProvider,
	req domainExport.Request,
	pageSize int,
	report func(int64),
) (int64, error) {
	finalPath := snapshotControllerPath(directory, controller.ID)
	partialPath := finalPath + ".partial"
	file, err := os.OpenFile(partialPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o640)
	if err != nil {
		return 0, err
	}
	compressed := gzip.NewWriter(file)
	encoder := json.NewEncoder(compressed)
	var count int64
	afterID := uuid.Nil
	writeErr := error(nil)
	for {
		if err := ctx.Err(); err != nil {
			writeErr = err
			break
		}
		items, err := source.ListFieldDevicesByControllerAfter(ctx, controller.ID, req, afterID, pageSize)
		if err != nil {
			writeErr = err
			break
		}
		if len(items) == 0 {
			break
		}
		for i := range items {
			if err := encoder.Encode(&items[i]); err != nil {
				writeErr = err
				break
			}
		}
		if writeErr != nil {
			break
		}
		count += int64(len(items))
		report(int64(len(items)))
		afterID = items[len(items)-1].ID
		if len(items) < pageSize {
			break
		}
	}
	writeErr = errors.Join(writeErr, compressed.Close(), file.Close())
	if writeErr != nil {
		_ = os.Remove(partialPath)
		return 0, writeErr
	}
	if count == 0 {
		_ = os.Remove(partialPath)
		return 0, nil
	}
	if err := os.Rename(partialPath, finalPath); err != nil {
		_ = os.Remove(partialPath)
		return 0, err
	}
	return count, nil
}

func writeSnapshotManifest(directory string, manifest exportSnapshotManifest) error {
	partial := filepath.Join(directory, "manifest.json.partial")
	final := filepath.Join(directory, "manifest.json")
	raw, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	if err := os.WriteFile(partial, raw, 0o640); err != nil {
		return err
	}
	return os.Rename(partial, final)
}

func openSnapshot(directory string) (*snapshotDataProvider, error) {
	raw, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var manifest exportSnapshotManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, err
	}
	if manifest.SchemaVersion != exportSnapshotSchemaVersion || len(manifest.Controllers) == 0 {
		return nil, errors.New("unsupported or empty export snapshot")
	}
	for _, controller := range manifest.Controllers {
		if _, err := os.Stat(snapshotControllerPath(directory, controller.ID)); err != nil {
			return nil, err
		}
	}
	return &snapshotDataProvider{directory: directory, manifest: manifest, states: make(map[uuid.UUID]*snapshotReadState)}, nil
}

func (s *snapshotDataProvider) ResolveControllers(context.Context, domainExport.Request) ([]domainExport.Controller, error) {
	return append([]domainExport.Controller(nil), s.manifest.Controllers...), nil
}

func (s *snapshotDataProvider) WithinSnapshot(ctx context.Context, consume func(domainExport.DataProvider) error) error {
	return consume(s)
}

func (s *snapshotDataProvider) ListFieldDevicesByControllerAfter(_ context.Context, controllerID uuid.UUID, _ domainExport.Request, afterID uuid.UUID, limit int) ([]domainFacility.FieldDevice, error) {
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	state, err := s.state(controllerID)
	if err != nil {
		return nil, err
	}
	if state.done {
		return []domainFacility.FieldDevice{}, nil
	}
	if state.lastID != afterID {
		return nil, errors.New("export snapshot cursor does not match sequential reader")
	}
	items := make([]domainFacility.FieldDevice, 0, limit)
	for range limit {
		var item domainFacility.FieldDevice
		if err := state.decoder.Decode(&item); err != nil {
			if errors.Is(err, io.EOF) {
				state.done = true
				break
			}
			return nil, err
		}
		items = append(items, item)
		state.lastID = item.ID
	}
	return items, nil
}

func (s *snapshotDataProvider) state(controllerID uuid.UUID) (*snapshotReadState, error) {
	if state := s.states[controllerID]; state != nil {
		return state, nil
	}
	file, err := os.Open(snapshotControllerPath(s.directory, controllerID))
	if err != nil {
		return nil, err
	}
	compressed, err := gzip.NewReader(file)
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	state := &snapshotReadState{file: file, gzip: compressed, decoder: json.NewDecoder(compressed)}
	s.states[controllerID] = state
	return state, nil
}

func (s *snapshotDataProvider) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	for _, state := range s.states {
		errs = append(errs, state.gzip.Close(), state.file.Close())
	}
	s.states = make(map[uuid.UUID]*snapshotReadState)
	return errors.Join(errs...)
}

func snapshotControllerPath(directory string, controllerID uuid.UUID) string {
	return filepath.Join(directory, controllerID.String()+".ndjson.gz")
}
