package facility

import (
	"context"
	"sync"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type fieldDeviceOptionsRevisioner interface {
	GetFieldDeviceOptionsRevision(ctx context.Context, projectID *uuid.UUID) (string, error)
}

type fieldDeviceOptionsCacheEntry struct {
	revision string
	options  *domainFacility.FieldDeviceOptions
}

type fieldDeviceOptionsCache struct {
	mu      sync.RWMutex
	entries map[string]fieldDeviceOptionsCacheEntry
}

func newFieldDeviceOptionsCache() *fieldDeviceOptionsCache {
	return &fieldDeviceOptionsCache{
		entries: make(map[string]fieldDeviceOptionsCacheEntry),
	}
}

func (c *fieldDeviceOptionsCache) get(key, revision string) (*domainFacility.FieldDeviceOptions, bool) {
	if c == nil || revision == "" {
		return nil, false
	}

	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok || entry.revision != revision {
		return nil, false
	}

	return cloneFieldDeviceOptions(entry.options), true
}

func (c *fieldDeviceOptionsCache) set(key, revision string, options *domainFacility.FieldDeviceOptions) {
	if c == nil || revision == "" || options == nil {
		return
	}

	c.mu.Lock()
	c.entries[key] = fieldDeviceOptionsCacheEntry{
		revision: revision,
		options:  cloneFieldDeviceOptions(options),
	}
	c.mu.Unlock()
}

func cloneFieldDeviceOptions(options *domainFacility.FieldDeviceOptions) *domainFacility.FieldDeviceOptions {
	if options == nil {
		return nil
	}

	clone := &domainFacility.FieldDeviceOptions{
		Apparats:            append([]domainFacility.Apparat(nil), options.Apparats...),
		SystemParts:         append([]domainFacility.SystemPart(nil), options.SystemParts...),
		ObjectDatas:         append([]domainFacility.ObjectData(nil), options.ObjectDatas...),
		ApparatToSystemPart: make(map[uuid.UUID][]uuid.UUID, len(options.ApparatToSystemPart)),
		ObjectDataToApparat: make(map[uuid.UUID][]uuid.UUID, len(options.ObjectDataToApparat)),
	}

	for apparatID, systemPartIDs := range options.ApparatToSystemPart {
		clone.ApparatToSystemPart[apparatID] = append([]uuid.UUID(nil), systemPartIDs...)
	}
	for objectDataID, apparatIDs := range options.ObjectDataToApparat {
		clone.ObjectDataToApparat[objectDataID] = append([]uuid.UUID(nil), apparatIDs...)
	}

	return clone
}
