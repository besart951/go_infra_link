package facilitycache

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

// WrapReferenceData caches rarely changing facility reference data behind the
// existing repository interfaces. Any successful write invalidates both caches,
// because Apparat and SystemPart share the system_part_apparats relation.
func WrapReferenceData(
	apparats domainFacility.ApparatRepository,
	systemParts domainFacility.SystemPartRepository,
) (domainFacility.ApparatRepository, domainFacility.SystemPartRepository) {
	cache := newReferenceDataCache()
	return &cachedApparatRepository{next: apparats, cache: cache},
		&cachedSystemPartRepository{next: systemParts, cache: cache}
}

type referenceDataCache struct {
	mu sync.RWMutex

	apparatByIDs map[string][]*domainFacility.Apparat
	apparatPages map[string]*domain.PaginatedList[domainFacility.Apparat]

	systemPartByIDs map[string][]*domainFacility.SystemPart
	systemPartPages map[string]*domain.PaginatedList[domainFacility.SystemPart]
}

func newReferenceDataCache() *referenceDataCache {
	return &referenceDataCache{
		apparatByIDs:    make(map[string][]*domainFacility.Apparat),
		apparatPages:    make(map[string]*domain.PaginatedList[domainFacility.Apparat]),
		systemPartByIDs: make(map[string][]*domainFacility.SystemPart),
		systemPartPages: make(map[string]*domain.PaginatedList[domainFacility.SystemPart]),
	}
}

func (c *referenceDataCache) invalidateAll() {
	c.mu.Lock()
	c.apparatByIDs = make(map[string][]*domainFacility.Apparat)
	c.apparatPages = make(map[string]*domain.PaginatedList[domainFacility.Apparat])
	c.systemPartByIDs = make(map[string][]*domainFacility.SystemPart)
	c.systemPartPages = make(map[string]*domain.PaginatedList[domainFacility.SystemPart])
	c.mu.Unlock()
}

func (c *referenceDataCache) getApparatsByIDs(key string) ([]*domainFacility.Apparat, bool) {
	c.mu.RLock()
	items, ok := c.apparatByIDs[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return cloneApparatPtrs(items), true
}

func (c *referenceDataCache) setApparatsByIDs(key string, items []*domainFacility.Apparat) {
	c.mu.Lock()
	c.apparatByIDs[key] = cloneApparatPtrs(items)
	c.mu.Unlock()
}

func (c *referenceDataCache) getApparatPage(key string) (*domain.PaginatedList[domainFacility.Apparat], bool) {
	c.mu.RLock()
	page, ok := c.apparatPages[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return cloneApparatPage(page), true
}

func (c *referenceDataCache) setApparatPage(key string, page *domain.PaginatedList[domainFacility.Apparat]) {
	if page == nil {
		return
	}
	c.mu.Lock()
	c.apparatPages[key] = cloneApparatPage(page)
	c.mu.Unlock()
}

func (c *referenceDataCache) getSystemPartsByIDs(key string) ([]*domainFacility.SystemPart, bool) {
	c.mu.RLock()
	items, ok := c.systemPartByIDs[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return cloneSystemPartPtrs(items), true
}

func (c *referenceDataCache) setSystemPartsByIDs(key string, items []*domainFacility.SystemPart) {
	c.mu.Lock()
	c.systemPartByIDs[key] = cloneSystemPartPtrs(items)
	c.mu.Unlock()
}

func (c *referenceDataCache) getSystemPartPage(key string) (*domain.PaginatedList[domainFacility.SystemPart], bool) {
	c.mu.RLock()
	page, ok := c.systemPartPages[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return cloneSystemPartPage(page), true
}

func (c *referenceDataCache) setSystemPartPage(key string, page *domain.PaginatedList[domainFacility.SystemPart]) {
	if page == nil {
		return
	}
	c.mu.Lock()
	c.systemPartPages[key] = cloneSystemPartPage(page)
	c.mu.Unlock()
}

type cachedApparatRepository struct {
	next  domainFacility.ApparatRepository
	cache *referenceDataCache
}

func (r *cachedApparatRepository) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.Apparat, error) {
	if len(ids) == 0 {
		return []*domainFacility.Apparat{}, nil
	}
	key := idsKey(ids)
	if items, ok := r.cache.getApparatsByIDs(key); ok {
		return items, nil
	}
	items, err := r.next.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	r.cache.setApparatsByIDs(key, items)
	return cloneApparatPtrs(items), nil
}

func (r *cachedApparatRepository) Create(ctx context.Context, entity *domainFacility.Apparat) error {
	if err := r.next.Create(ctx, entity); err != nil {
		return err
	}
	r.cache.invalidateAll()
	return nil
}

func (r *cachedApparatRepository) Update(ctx context.Context, entity *domainFacility.Apparat) error {
	if err := r.next.Update(ctx, entity); err != nil {
		return err
	}
	r.cache.invalidateAll()
	return nil
}

func (r *cachedApparatRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if err := r.next.DeleteByIds(ctx, ids); err != nil {
		return err
	}
	r.cache.invalidateAll()
	return nil
}

func (r *cachedApparatRepository) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	key := pageKey("apparat:list", params)
	if page, ok := r.cache.getApparatPage(key); ok {
		return page, nil
	}
	page, err := r.next.GetPaginatedList(ctx, params)
	if err != nil {
		return nil, err
	}
	r.cache.setApparatPage(key, page)
	return cloneApparatPage(page), nil
}

func (r *cachedApparatRepository) ExistsShortName(ctx context.Context, shortName string, excludeID *uuid.UUID) (bool, error) {
	return r.next.ExistsShortName(ctx, shortName, excludeID)
}

func (r *cachedApparatRepository) ExistsName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	return r.next.ExistsName(ctx, name, excludeID)
}

func (r *cachedApparatRepository) GetPaginatedListWithFilters(ctx context.Context, params domain.PaginationParams, filters domainFacility.ApparatFilterParams) (*domain.PaginatedList[domainFacility.Apparat], error) {
	key := pageKey("apparat:filters:"+filterID(filters.ObjectDataID)+":"+filterID(filters.SystemPartID), params)
	if page, ok := r.cache.getApparatPage(key); ok {
		return page, nil
	}
	page, err := r.next.GetPaginatedListWithFilters(ctx, params, filters)
	if err != nil {
		return nil, err
	}
	r.cache.setApparatPage(key, page)
	return cloneApparatPage(page), nil
}

type cachedSystemPartRepository struct {
	next  domainFacility.SystemPartRepository
	cache *referenceDataCache
}

func (r *cachedSystemPartRepository) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.SystemPart, error) {
	if len(ids) == 0 {
		return []*domainFacility.SystemPart{}, nil
	}
	key := idsKey(ids)
	if items, ok := r.cache.getSystemPartsByIDs(key); ok {
		return items, nil
	}
	items, err := r.next.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	r.cache.setSystemPartsByIDs(key, items)
	return cloneSystemPartPtrs(items), nil
}

func (r *cachedSystemPartRepository) Create(ctx context.Context, entity *domainFacility.SystemPart) error {
	if err := r.next.Create(ctx, entity); err != nil {
		return err
	}
	r.cache.invalidateAll()
	return nil
}

func (r *cachedSystemPartRepository) Update(ctx context.Context, entity *domainFacility.SystemPart) error {
	if err := r.next.Update(ctx, entity); err != nil {
		return err
	}
	r.cache.invalidateAll()
	return nil
}

func (r *cachedSystemPartRepository) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if err := r.next.DeleteByIds(ctx, ids); err != nil {
		return err
	}
	r.cache.invalidateAll()
	return nil
}

func (r *cachedSystemPartRepository) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SystemPart], error) {
	key := pageKey("system_part:list", params)
	if page, ok := r.cache.getSystemPartPage(key); ok {
		return page, nil
	}
	page, err := r.next.GetPaginatedList(ctx, params)
	if err != nil {
		return nil, err
	}
	r.cache.setSystemPartPage(key, page)
	return cloneSystemPartPage(page), nil
}

func (r *cachedSystemPartRepository) ExistsShortName(ctx context.Context, shortName string, excludeID *uuid.UUID) (bool, error) {
	return r.next.ExistsShortName(ctx, shortName, excludeID)
}

func (r *cachedSystemPartRepository) ExistsName(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	return r.next.ExistsName(ctx, name, excludeID)
}

func idsKey(ids []uuid.UUID) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = id.String()
	}
	return strings.Join(parts, "|")
}

func pageKey(scope string, params domain.PaginationParams) string {
	return fmt.Sprintf("%s|page=%d|limit=%d|search=%s", scope, params.Page, params.Limit, params.Search)
}

func filterID(id *uuid.UUID) string {
	if id == nil {
		return "-"
	}
	return id.String()
}

func cloneApparatPage(page *domain.PaginatedList[domainFacility.Apparat]) *domain.PaginatedList[domainFacility.Apparat] {
	if page == nil {
		return nil
	}
	items := make([]domainFacility.Apparat, len(page.Items))
	for i := range page.Items {
		items[i] = *cloneApparat(&page.Items[i])
	}
	return &domain.PaginatedList[domainFacility.Apparat]{
		Items:      items,
		Total:      page.Total,
		Page:       page.Page,
		TotalPages: page.TotalPages,
	}
}

func cloneSystemPartPage(page *domain.PaginatedList[domainFacility.SystemPart]) *domain.PaginatedList[domainFacility.SystemPart] {
	if page == nil {
		return nil
	}
	items := make([]domainFacility.SystemPart, len(page.Items))
	for i := range page.Items {
		items[i] = *cloneSystemPart(&page.Items[i])
	}
	return &domain.PaginatedList[domainFacility.SystemPart]{
		Items:      items,
		Total:      page.Total,
		Page:       page.Page,
		TotalPages: page.TotalPages,
	}
}

func cloneApparatPtrs(items []*domainFacility.Apparat) []*domainFacility.Apparat {
	if items == nil {
		return nil
	}
	out := make([]*domainFacility.Apparat, len(items))
	for i, item := range items {
		out[i] = cloneApparat(item)
	}
	return out
}

func cloneSystemPartPtrs(items []*domainFacility.SystemPart) []*domainFacility.SystemPart {
	if items == nil {
		return nil
	}
	out := make([]*domainFacility.SystemPart, len(items))
	for i, item := range items {
		out[i] = cloneSystemPart(item)
	}
	return out
}

func cloneApparat(item *domainFacility.Apparat) *domainFacility.Apparat {
	if item == nil {
		return nil
	}
	clone := *item
	clone.Description = cloneStringPtr(item.Description)
	clone.SystemParts = cloneSystemPartPtrsWithoutApparats(item.SystemParts)
	clone.FieldDevices = append([]domainFacility.FieldDevice(nil), item.FieldDevices...)
	return &clone
}

func cloneSystemPart(item *domainFacility.SystemPart) *domainFacility.SystemPart {
	if item == nil {
		return nil
	}
	clone := *item
	clone.Description = cloneStringPtr(item.Description)
	clone.Apparats = cloneApparatPtrsWithoutSystemParts(item.Apparats)
	return &clone
}

func cloneSystemPartPtrsWithoutApparats(items []*domainFacility.SystemPart) []*domainFacility.SystemPart {
	if items == nil {
		return nil
	}
	out := make([]*domainFacility.SystemPart, len(items))
	for i, item := range items {
		if item == nil {
			continue
		}
		clone := *item
		clone.Description = cloneStringPtr(item.Description)
		clone.Apparats = nil
		out[i] = &clone
	}
	return out
}

func cloneApparatPtrsWithoutSystemParts(items []*domainFacility.Apparat) []*domainFacility.Apparat {
	if items == nil {
		return nil
	}
	out := make([]*domainFacility.Apparat, len(items))
	for i, item := range items {
		if item == nil {
			continue
		}
		clone := *item
		clone.Description = cloneStringPtr(item.Description)
		clone.SystemParts = nil
		clone.FieldDevices = append([]domainFacility.FieldDevice(nil), item.FieldDevices...)
		out[i] = &clone
	}
	return out
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}
