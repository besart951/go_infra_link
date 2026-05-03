package historysql

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/besart951/go_infra_link/backend/internal/service/auditctx"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const defaultTimelineLimit = 50
const maxTimelineLimit = 200

var tableWhitelist = map[string]struct{}{
	"buildings":                   {},
	"projects":                    {},
	"control_cabinets":            {},
	"sps_controllers":             {},
	"sps_controller_system_types": {},
	"field_devices":               {},
	"specifications":              {},
	"bacnet_objects":              {},
	"bacnet_object_alarm_values":  {},
	"object_data":                 {},
	"state_texts":                 {},
	"notification_classes":        {},
	"alarm_definitions":           {},
	"alarm_fields":                {},
	"alarm_types":                 {},
	"alarm_type_fields":           {},
	"units":                       {},
	"system_types":                {},
	"system_parts":                {},
	"apparats":                    {},
	"project_control_cabinets":    {},
	"project_sps_controllers":     {},
	"project_field_devices":       {},
}

type Store struct {
	db *gorm.DB
}

type Mutation struct {
	Action      domainHistory.Action
	EntityTable string
	EntityID    uuid.UUID
	BeforeJSON  domainHistory.JSONB
	AfterJSON   domainHistory.JSONB
	BatchID     *uuid.UUID
	Summary     string
	Metadata    map[string]any
	ActorID     *uuid.UUID
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&domainHistory.ChangeEvent{}, &domainHistory.ChangeEventScope{}, &domainHistory.EntityVersion{}); err != nil {
		return err
	}

	statements := []string{
		`CREATE INDEX IF NOT EXISTS idx_change_events_entity_time ON change_events (entity_table, entity_id, occurred_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_change_events_batch ON change_events (batch_id, occurred_at DESC) WHERE batch_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_change_event_scopes_scope_time ON change_event_scopes (scope_type, scope_id, occurred_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_entity_versions_latest ON entity_versions (entity_table, entity_id, version_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_entity_versions_change_event ON entity_versions (change_event_id)`,
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) WithDB(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) LoadRow(ctx context.Context, table string, id uuid.UUID) (domainHistory.JSONB, bool, error) {
	if !allowedTable(table) {
		return nil, false, fmt.Errorf("history table not allowed: %s", table)
	}
	if id == uuid.Nil {
		return nil, false, nil
	}

	var row struct {
		Data domainHistory.JSONB `gorm:"column:data"`
	}
	err := s.db.WithContext(ctx).
		Raw(fmt.Sprintf(`SELECT to_jsonb(t) AS data FROM (SELECT * FROM %s WHERE id = ?) AS t`, quoteIdent(table)), id).
		Scan(&row).Error
	if err != nil {
		return nil, false, err
	}
	if len(row.Data) == 0 {
		return nil, false, nil
	}
	return row.Data, true, nil
}

func (s *Store) LoadRows(ctx context.Context, table string, ids []uuid.UUID) (map[uuid.UUID]domainHistory.JSONB, error) {
	if !allowedTable(table) {
		return nil, fmt.Errorf("history table not allowed: %s", table)
	}
	out := make(map[uuid.UUID]domainHistory.JSONB, len(ids))
	for _, chunk := range uuidChunks(ids, 500) {
		var rows []rawRow
		if err := s.db.WithContext(ctx).
			Raw(fmt.Sprintf(`SELECT id, to_jsonb(t) AS data FROM (SELECT * FROM %s WHERE id IN ?) AS t`, quoteIdent(table)), chunk).
			Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			out[row.ID] = row.Data
		}
	}
	return out, nil
}

func (s *Store) LoadRowsWhere(ctx context.Context, table string, where string, args ...any) (map[uuid.UUID]domainHistory.JSONB, error) {
	if !allowedTable(table) {
		return nil, fmt.Errorf("history table not allowed: %s", table)
	}
	out := map[uuid.UUID]domainHistory.JSONB{}
	query := fmt.Sprintf(`SELECT id, to_jsonb(t) AS data FROM (SELECT * FROM %s WHERE %s) AS t`, quoteIdent(table), where)
	var rows []rawRow
	if err := s.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.ID] = row.Data
	}
	return out, nil
}

func (s *Store) RecordCreate(ctx context.Context, table string, id uuid.UUID) error {
	after, ok, err := s.LoadRow(ctx, table, id)
	if err != nil || !ok {
		return err
	}
	return s.RecordMutation(ctx, Mutation{
		Action:      domainHistory.ActionCreate,
		EntityTable: table,
		EntityID:    id,
		AfterJSON:   after,
	})
}

func (s *Store) RecordUpdate(ctx context.Context, table string, id uuid.UUID, before domainHistory.JSONB) error {
	after, ok, err := s.LoadRow(ctx, table, id)
	if err != nil || !ok {
		return err
	}
	if jsonEqual(before, after) {
		return nil
	}
	return s.RecordMutation(ctx, Mutation{
		Action:      domainHistory.ActionUpdate,
		EntityTable: table,
		EntityID:    id,
		BeforeJSON:  before,
		AfterJSON:   after,
	})
}

func (s *Store) RecordDelete(ctx context.Context, table string, id uuid.UUID, before domainHistory.JSONB) error {
	if len(before) == 0 {
		return nil
	}
	return s.RecordMutation(ctx, Mutation{
		Action:      domainHistory.ActionDelete,
		EntityTable: table,
		EntityID:    id,
		BeforeJSON:  before,
	})
}

func (s *Store) RecordMutation(ctx context.Context, mutation Mutation) error {
	if !allowedTable(mutation.EntityTable) {
		return fmt.Errorf("history table not allowed: %s", mutation.EntityTable)
	}
	if mutation.EntityID == uuid.Nil {
		return nil
	}

	now := time.Now().UTC()
	eventID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	actorID := mutation.ActorID
	if actorID == nil {
		if actor, ok := auditctx.ActorID(ctx); ok {
			actorID = actor
		}
	}
	metadataJSON, err := marshalJSON(mutation.Metadata)
	if err != nil {
		return err
	}
	diffJSON, err := diffJSON(mutation.BeforeJSON, mutation.AfterJSON)
	if err != nil {
		return err
	}

	var summary *string
	if strings.TrimSpace(mutation.Summary) != "" {
		trimmed := strings.TrimSpace(mutation.Summary)
		summary = &trimmed
	}

	event := &domainHistory.ChangeEvent{
		ID:           eventID,
		OccurredAt:   now,
		ActorID:      actorID,
		Action:       mutation.Action,
		EntityTable:  mutation.EntityTable,
		EntityID:     mutation.EntityID,
		BatchID:      mutation.BatchID,
		Summary:      summary,
		BeforeJSON:   mutation.BeforeJSON,
		AfterJSON:    mutation.AfterJSON,
		DiffJSON:     diffJSON,
		MetadataJSON: metadataJSON,
	}
	if err := s.db.WithContext(ctx).Create(event).Error; err != nil {
		return err
	}

	scopeSnapshot := mutation.AfterJSON
	if len(scopeSnapshot) == 0 {
		scopeSnapshot = mutation.BeforeJSON
	}
	scopes, err := s.resolveScopes(ctx, mutation.EntityTable, mutation.EntityID, scopeSnapshot)
	if err != nil {
		return err
	}
	if len(scopes) > 0 {
		rows := make([]domainHistory.ChangeEventScope, 0, len(scopes))
		for _, scope := range scopes {
			scopeID, err := uuid.NewV7()
			if err != nil {
				return err
			}
			rows = append(rows, domainHistory.ChangeEventScope{
				ID:            scopeID,
				ChangeEventID: eventID,
				ScopeType:     scope.Type,
				ScopeID:       scope.ID,
				OccurredAt:    now,
			})
		}
		if err := s.db.WithContext(ctx).CreateInBatches(rows, 500).Error; err != nil {
			return err
		}
	}

	versionID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	version := &domainHistory.EntityVersion{
		ID:            versionID,
		ChangeEventID: eventID,
		EntityTable:   mutation.EntityTable,
		EntityID:      mutation.EntityID,
		VersionAt:     now,
		Action:        mutation.Action,
	}
	if mutation.Action == domainHistory.ActionDelete {
		version.SnapshotJSON = nil
	} else {
		version.SnapshotJSON = mutation.AfterJSON
	}
	return s.db.WithContext(ctx).Create(version).Error
}

func (s *Store) ListTimeline(ctx context.Context, filter domainHistory.TimelineFilter) (*domain.PaginatedList[domainHistory.ChangeEvent], error) {
	page, limit := normalizeTimelinePagination(filter.Page, filter.Limit)
	offset := (page - 1) * limit

	query := s.db.WithContext(ctx).Model(&domainHistory.ChangeEvent{})
	if filter.EntityTable != "" && filter.EntityID != uuid.Nil {
		query = query.Where("entity_table = ? AND entity_id = ?", filter.EntityTable, filter.EntityID)
	}
	if filter.ScopeType != "" && filter.ScopeID != uuid.Nil {
		sub := s.db.WithContext(ctx).Model(&domainHistory.ChangeEventScope{}).
			Select("change_event_id").
			Where("scope_type = ? AND scope_id = ?", filter.ScopeType, filter.ScopeID)
		query = query.Where("id IN (?)", sub)
	}
	if filter.SecondaryScopeType != "" && filter.SecondaryScopeID != uuid.Nil {
		sub := s.db.WithContext(ctx).Model(&domainHistory.ChangeEventScope{}).
			Select("change_event_id").
			Where("scope_type = ? AND scope_id = ?", filter.SecondaryScopeType, filter.SecondaryScopeID)
		query = query.Where("id IN (?)", sub)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainHistory.ChangeEvent
	if err := query.Order("occurred_at DESC, id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}
	if err := s.enrichActorNames(ctx, items); err != nil {
		return nil, err
	}
	if err := s.enrichScopeSummaries(ctx, items); err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainHistory.ChangeEvent]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (s *Store) GetEvent(ctx context.Context, id uuid.UUID) (*domainHistory.ChangeEvent, error) {
	var event domainHistory.ChangeEvent
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&event).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	events := []domainHistory.ChangeEvent{event}
	if err := s.enrichActorNames(ctx, events); err != nil {
		return nil, err
	}
	if err := s.enrichScopeSummaries(ctx, events); err != nil {
		return nil, err
	}
	event = events[0]
	return &event, nil
}

func normalizeTimelinePagination(page, limit int) (int, int) {
	page, limit = domain.NormalizePagination(page, limit, defaultTimelineLimit)
	if limit > maxTimelineLimit {
		limit = maxTimelineLimit
	}
	return page, limit
}

type rawRow struct {
	ID   uuid.UUID           `gorm:"column:id"`
	Data domainHistory.JSONB `gorm:"column:data"`
}

func allowedTable(table string) bool {
	_, ok := tableWhitelist[table]
	return ok
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func uuidChunks(ids []uuid.UUID, size int) [][]uuid.UUID {
	if len(ids) == 0 {
		return nil
	}
	if size <= 0 {
		size = 500
	}
	chunks := make([][]uuid.UUID, 0, (len(ids)+size-1)/size)
	for start := 0; start < len(ids); start += size {
		end := start + size
		if end > len(ids) {
			end = len(ids)
		}
		chunks = append(chunks, ids[start:end])
	}
	return chunks
}

func marshalJSON(value any) (domainHistory.JSONB, error) {
	if value == nil {
		return nil, nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if string(b) == "null" || string(b) == "{}" {
		return nil, nil
	}
	return domainHistory.JSONB(b), nil
}

func jsonEqual(a, b domainHistory.JSONB) bool {
	var am, bm any
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	if json.Unmarshal(a, &am) != nil || json.Unmarshal(b, &bm) != nil {
		return string(a) == string(b)
	}
	return reflect.DeepEqual(am, bm)
}

func diffJSON(before, after domainHistory.JSONB) (domainHistory.JSONB, error) {
	if len(before) == 0 || len(after) == 0 {
		return nil, nil
	}
	var oldMap map[string]any
	var newMap map[string]any
	if err := json.Unmarshal(before, &oldMap); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(after, &newMap); err != nil {
		return nil, err
	}

	keys := make(map[string]struct{}, len(oldMap)+len(newMap))
	for key := range oldMap {
		keys[key] = struct{}{}
	}
	for key := range newMap {
		keys[key] = struct{}{}
	}

	out := map[string]map[string]any{}
	for key := range keys {
		oldValue, oldOK := oldMap[key]
		newValue, newOK := newMap[key]
		if !oldOK || !newOK || !reflect.DeepEqual(oldValue, newValue) {
			out[key] = map[string]any{"before": oldValue, "after": newValue}
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	return marshalJSON(out)
}

func sortedUUIDs(set map[uuid.UUID]struct{}) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out
}
