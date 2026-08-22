package historysql

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/cursor"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

const timelineCursorKind = "history_timeline"

type timelineFingerprint struct {
	ScopeType          string   `json:"scope_type"`
	ScopeID            string   `json:"scope_id"`
	SecondaryScopeType string   `json:"secondary_scope_type"`
	SecondaryScopeID   string   `json:"secondary_scope_id"`
	EntityTable        string   `json:"entity_table"`
	EntityID           string   `json:"entity_id"`
	ActorID            string   `json:"actor_id"`
	OccurredFrom       string   `json:"occurred_from"`
	OccurredTo         string   `json:"occurred_to"`
	Actions            []string `json:"actions,omitempty"`
	Fields             []string `json:"fields,omitempty"`
}

type timelineCursorScan struct {
	filter domainHistory.TimelineFilter
	anchor *domainHistory.TimelineCursorAnchor
	limit  int
}

type timelinePageBuild struct {
	items       []domainHistory.ChangeEvent
	fingerprint string
	anchor      *domainHistory.TimelineCursorAnchor
	hasMore     bool
}

func (s *Store) ListTimelineCursor(ctx context.Context, filter domainHistory.TimelineFilter) (*domainHistory.TimelineCursorPage, error) {
	_, limit := normalizeTimelinePagination(1, filter.Limit)
	fingerprint, err := timelineQueryFingerprint(filter)
	if err != nil {
		return nil, err
	}
	anchor, err := decodeTimelineCursor(filter.Cursor, fingerprint)
	if err != nil {
		return nil, err
	}
	items, hasMore, err := s.scanTimelineCursor(ctx, timelineCursorScan{filter: filter, anchor: anchor, limit: limit})
	if err != nil {
		return nil, err
	}
	if err := s.enrichActorNames(ctx, items); err != nil {
		return nil, err
	}
	if err := s.enrichScopeSummaries(ctx, items); err != nil {
		return nil, err
	}
	return buildTimelineCursorPage(timelinePageBuild{items: items, fingerprint: fingerprint, anchor: anchor, hasMore: hasMore})
}

func (s *Store) scanTimelineCursor(ctx context.Context, scan timelineCursorScan) ([]domainHistory.ChangeEvent, bool, error) {
	query, err := s.timelineQuery(ctx, scan.filter)
	if err != nil {
		return nil, false, err
	}
	direction := timelineDirection(scan.anchor)
	if scan.anchor != nil {
		operator := "<"
		if direction == "previous" {
			operator = ">"
		}
		query = query.Where("occurred_at "+operator+" ? OR (occurred_at = ? AND id "+operator+" ?)", scan.anchor.OccurredAt, scan.anchor.OccurredAt, scan.anchor.ID)
	}
	order := "occurred_at DESC, id DESC"
	if direction == "previous" {
		order = "occurred_at ASC, id ASC"
	}
	var items []domainHistory.ChangeEvent
	if err := query.Order(order).Limit(scan.limit + 1).Find(&items).Error; err != nil {
		return nil, false, err
	}
	hasMore := len(items) > scan.limit
	if hasMore {
		items = items[:scan.limit]
	}
	if direction == "previous" {
		slices.Reverse(items)
	}
	return items, hasMore, nil
}

func buildTimelineCursorPage(build timelinePageBuild) (*domainHistory.TimelineCursorPage, error) {
	page := &domainHistory.TimelineCursorPage{Items: build.items}
	if len(build.items) == 0 {
		return page, nil
	}
	direction := timelineDirection(build.anchor)
	hasPrevious := (build.anchor != nil && direction == "next") || (direction == "previous" && build.hasMore)
	hasNext := (direction == "next" && build.hasMore) || (build.anchor != nil && direction == "previous")
	var err error
	if hasPrevious {
		page.PreviousCursor, err = encodeTimelineCursor(build.items[0], build.fingerprint, "previous")
	}
	if err == nil && hasNext {
		page.NextCursor, err = encodeTimelineCursor(build.items[len(build.items)-1], build.fingerprint, "next")
	}
	return page, err
}

func encodeTimelineCursor(event domainHistory.ChangeEvent, fingerprint, direction string) (string, error) {
	return cursor.Encode(timelineCursorKind, domainHistory.TimelineCursorAnchor{
		Direction: direction, Fingerprint: fingerprint, OccurredAt: event.OccurredAt, ID: event.ID,
	})
}

func decodeTimelineCursor(value, fingerprint string) (*domainHistory.TimelineCursorAnchor, error) {
	if value == "" {
		return nil, nil
	}
	var anchor domainHistory.TimelineCursorAnchor
	if err := cursor.Decode(value, timelineCursorKind, &anchor); err != nil {
		return nil, err
	}
	if anchor.Fingerprint != fingerprint || anchor.ID == uuid.Nil || anchor.OccurredAt.IsZero() {
		return nil, cursor.ErrInvalid
	}
	if anchor.Direction != "next" && anchor.Direction != "previous" {
		return nil, cursor.ErrInvalid
	}
	return &anchor, nil
}

func timelineDirection(anchor *domainHistory.TimelineCursorAnchor) string {
	if anchor == nil {
		return "next"
	}
	return anchor.Direction
}

func timelineQueryFingerprint(filter domainHistory.TimelineFilter) (string, error) {
	actions := make([]string, len(filter.Actions))
	for index, action := range filter.Actions {
		actions[index] = string(action)
	}
	fields := append([]string(nil), filter.Fields...)
	slices.Sort(actions)
	slices.Sort(fields)
	return cursor.Fingerprint(timelineFingerprint{
		ScopeType: strings.TrimSpace(filter.ScopeType), ScopeID: cursorUUID(filter.ScopeID),
		SecondaryScopeType: strings.TrimSpace(filter.SecondaryScopeType), SecondaryScopeID: cursorUUID(filter.SecondaryScopeID),
		EntityTable: strings.TrimSpace(filter.EntityTable), EntityID: cursorUUID(filter.EntityID), ActorID: cursorUUID(filter.ActorID),
		OccurredFrom: cursorTime(filter.OccurredFrom), OccurredTo: cursorTime(filter.OccurredTo), Actions: actions, Fields: fields,
	})
}

func cursorUUID(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

func cursorTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
