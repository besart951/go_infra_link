package historysql

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"time"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type V2BackfillRequest struct {
	AfterOccurredAt time.Time
	AfterID         uuid.UUID
	Limit           int
}

type V2BackfillResult struct {
	Processed      int
	NextOccurredAt time.Time
	NextID         uuid.UUID
	Done           bool
}

type V2Verification struct {
	SourceCount  int64
	TargetCount  int64
	SourceDigest string
	TargetDigest string
}

func (s *Store) VerifyAndEnableV2Reads(ctx context.Context) error {
	start, end, err := s.historyBounds(ctx)
	if err != nil {
		return err
	}
	for month := monthStart(start); !month.After(monthStart(end)); month = month.AddDate(0, 1, 0) {
		report, err := s.VerifyV2Month(ctx, month)
		if err != nil {
			return err
		}
		for kind, verification := range report {
			if !verification.Matches() {
				return fmt.Errorf("history V2 verification failed for %s/%s", month.Format("2006-01"), kind)
			}
		}
	}
	return s.EnableV2Reads(ctx)
}

func (s *Store) historyBounds(ctx context.Context) (time.Time, time.Time, error) {
	var first, last domainHistory.ChangeEvent
	firstResult := s.db.WithContext(ctx).Select("occurred_at").Order("occurred_at ASC, id ASC").Take(&first)
	if firstResult.Error == gorm.ErrRecordNotFound {
		now := time.Now().UTC()
		return now, now, nil
	}
	if firstResult.Error != nil {
		return time.Time{}, time.Time{}, firstResult.Error
	}
	if err := s.db.WithContext(ctx).Select("occurred_at").Order("occurred_at DESC, id DESC").Take(&last).Error; err != nil {
		return time.Time{}, time.Time{}, err
	}
	return first.OccurredAt.UTC(), last.OccurredAt.UTC(), nil
}

func (v V2Verification) Matches() bool {
	return v.SourceCount == v.TargetCount && v.SourceDigest == v.TargetDigest
}

func (s *Store) BackfillV2(ctx context.Context, request V2BackfillRequest) (V2BackfillResult, error) {
	request.Limit = normalizeBackfillLimit(request.Limit)
	events, err := s.loadBackfillEvents(ctx, request)
	if err != nil || len(events) == 0 {
		return V2BackfillResult{Done: len(events) == 0}, err
	}
	if err := s.writeBackfillBatch(ctx, events); err != nil {
		return V2BackfillResult{}, err
	}
	last := events[len(events)-1]
	return V2BackfillResult{
		Processed: len(events), NextOccurredAt: last.OccurredAt, NextID: last.ID,
		Done: len(events) < request.Limit,
	}, nil
}

func (s *Store) VerifyV2Month(ctx context.Context, month time.Time) (map[string]V2Verification, error) {
	start := monthStart(month.UTC())
	end := start.AddDate(0, 1, 0)
	pairs := map[string][2]string{
		"events":   {"change_events", "change_events_v2"},
		"scopes":   {"change_event_scopes", "change_event_scopes_v2"},
		"versions": {"entity_versions", "entity_versions_v2"},
	}
	report := make(map[string]V2Verification, len(pairs))
	for name, tables := range pairs {
		verification, err := verifyHistoryPair(ctx, s.db, tables, start, end)
		if err != nil {
			return nil, err
		}
		report[name] = verification
	}
	return report, nil
}

func normalizeBackfillLimit(limit int) int {
	if limit < 1 || limit > 500 {
		return 500
	}
	return limit
}

func (s *Store) loadBackfillEvents(ctx context.Context, request V2BackfillRequest) ([]domainHistory.ChangeEvent, error) {
	query := s.db.WithContext(ctx).Order("occurred_at ASC, id ASC").Limit(request.Limit)
	if !request.AfterOccurredAt.IsZero() {
		query = query.Where("occurred_at > ? OR (occurred_at = ? AND id > ?)", request.AfterOccurredAt, request.AfterOccurredAt, request.AfterID)
	}
	var events []domainHistory.ChangeEvent
	return events, query.Find(&events).Error
}

func (s *Store) writeBackfillBatch(ctx context.Context, events []domainHistory.ChangeEvent) error {
	ids := make([]uuid.UUID, len(events))
	eventRecords := make([]changeEventV2Record, len(events))
	for index := range events {
		ids[index] = events[index].ID
		eventRecords[index] = changeEventV2FromDomain(&events[index])
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := createBackfillRecords(tx, eventRecords); err != nil {
			return err
		}
		return backfillRelatedRecords(tx, ids)
	})
}

func backfillRelatedRecords(tx *gorm.DB, eventIDs []uuid.UUID) error {
	var scopes []domainHistory.ChangeEventScope
	if err := tx.Where("change_event_id IN ?", eventIDs).Find(&scopes).Error; err != nil {
		return err
	}
	var versions []domainHistory.EntityVersion
	if err := tx.Where("change_event_id IN ?", eventIDs).Find(&versions).Error; err != nil {
		return err
	}
	scopeRecords := make([]changeEventScopeV2Record, len(scopes))
	for index := range scopes {
		scopeRecords[index] = changeEventScopeV2FromDomain(scopes[index])
	}
	versionRecords := make([]entityVersionV2Record, len(versions))
	for index := range versions {
		versionRecords[index] = entityVersionV2FromDomain(&versions[index])
	}
	return createBackfillRecords(tx, scopeRecords, versionRecords)
}

func createBackfillRecords(tx *gorm.DB, groups ...any) error {
	for _, records := range groups {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(records, 500).Error; err != nil {
			return err
		}
	}
	return nil
}

func verifyHistoryPair(ctx context.Context, db *gorm.DB, tables [2]string, start, end time.Time) (V2Verification, error) {
	sourceCount, sourceDigest, err := historyDigest(ctx, db, tables[0], start, end)
	if err != nil {
		return V2Verification{}, err
	}
	targetCount, targetDigest, err := historyDigest(ctx, db, tables[1], start, end)
	return V2Verification{
		SourceCount: sourceCount, TargetCount: targetCount,
		SourceDigest: sourceDigest, TargetDigest: targetDigest,
	}, err
}

func historyDigest(ctx context.Context, db *gorm.DB, table string, start, end time.Time) (int64, string, error) {
	spec, ok := historyDigestSpecs[table]
	if !ok {
		return 0, "", fmt.Errorf("unsupported history verification table: %s", table)
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s >= ? AND %s < ? ORDER BY %s, id", spec.columns, table, spec.timeColumn, spec.timeColumn, spec.timeColumn)
	rows, err := db.WithContext(ctx).Raw(query, start, end).Rows()
	if err != nil {
		return 0, "", err
	}
	defer rows.Close()
	return digestHistoryRows(rows, sha256.New(), spec.columnCount)
}

func digestHistoryRows(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}, digest hash.Hash, columnCount int) (int64, string, error) {
	var count int64
	for rows.Next() {
		values, destinations := historyDigestDestinations(columnCount)
		if err := rows.Scan(destinations...); err != nil {
			return 0, "", err
		}
		_, _ = fmt.Fprintln(digest, values...)
		count++
	}
	return count, hex.EncodeToString(digest.Sum(nil)), rows.Err()
}

func historyDigestDestinations(count int) ([]any, []any) {
	values := make([]any, count)
	destinations := make([]any, count)
	for index := range values {
		destinations[index] = &values[index]
	}
	return values, destinations
}

type historyDigestSpec struct {
	timeColumn  string
	columns     string
	columnCount int
}

var historyDigestSpecs = map[string]historyDigestSpec{
	"change_events":          eventDigestSpec(),
	"change_events_v2":       eventDigestSpec(),
	"change_event_scopes":    scopeDigestSpec(),
	"change_event_scopes_v2": scopeDigestSpec(),
	"entity_versions":        versionDigestSpec(),
	"entity_versions_v2":     versionDigestSpec(),
}

func eventDigestSpec() historyDigestSpec {
	return historyDigestSpec{
		timeColumn: "occurred_at", columnCount: 12,
		columns: "id, occurred_at, actor_id, action, entity_table, entity_id, batch_id, summary, before_json, after_json, diff_json, metadata_json",
	}
}

func scopeDigestSpec() historyDigestSpec {
	return historyDigestSpec{
		timeColumn: "occurred_at", columnCount: 5,
		columns: "id, change_event_id, scope_type, scope_id, occurred_at",
	}
}

func versionDigestSpec() historyDigestSpec {
	return historyDigestSpec{
		timeColumn: "version_at", columnCount: 7,
		columns: "id, change_event_id, entity_table, entity_id, version_at, action, snapshot_json",
	}
}
