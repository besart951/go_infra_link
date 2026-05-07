package db

import "gorm.io/gorm"

func migrateRealtimeEvents(db *gorm.DB) error {
	if err := db.Exec(`
		create table if not exists realtime_events (
			id uuid primary key,
			topic text not null,
			source text not null,
			payload jsonb not null,
			published_at timestamptz not null,
			expires_at timestamptz not null
		)
	`).Error; err != nil {
		return err
	}
	return db.Exec(`
		create index if not exists realtime_events_expires_at_idx
		on realtime_events (expires_at)
	`).Error
}
