package facilityjobsql

import (
	"context"
	"fmt"
	"time"

	facilityjobs "github.com/besart951/go_infra_link/backend/internal/application/facilityjobs"
	"github.com/besart951/go_infra_link/backend/internal/postgresjson"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FieldDeviceUpdatePlanStore struct {
	db *gorm.DB
}

func NewFieldDeviceUpdatePlanStore(db *gorm.DB) *FieldDeviceUpdatePlanStore {
	return &FieldDeviceUpdatePlanStore{db: db}
}

type fieldDeviceUpdatePlanRecord struct {
	OwnerID           uuid.UUID `gorm:"primaryKey"`
	JobID             uuid.UUID `gorm:"primaryKey"`
	Ordinal           int64     `gorm:"primaryKey"`
	GroupOrdinal      int64
	DependencyGroupID uuid.UUID
	FieldDeviceID     uuid.UUID
	Command           postgresjson.Document `gorm:"type:jsonb"`
	CreatedAt         time.Time
}

func (fieldDeviceUpdatePlanRecord) TableName() string {
	return "field_device_bulk_update_plan_items"
}

func (s *FieldDeviceUpdatePlanStore) Save(ctx context.Context, items []facilityjobs.FieldDeviceUpdatePlanItem) error {
	for start := 0; start < len(items); start += 500 {
		end := min(start+500, len(items))
		records := planRecords(items[start:end])
		if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&records).Error; err != nil {
			return err
		}
	}
	return nil
}

func planRecords(items []facilityjobs.FieldDeviceUpdatePlanItem) []fieldDeviceUpdatePlanRecord {
	records := make([]fieldDeviceUpdatePlanRecord, len(items))
	for index, item := range items {
		records[index] = planRecord(item)
	}
	return records
}

func (s *FieldDeviceUpdatePlanStore) Plan(ctx context.Context, ownerID, jobID uuid.UUID) error {
	if s.db.Dialector == nil || s.db.Dialector.Name() != "postgres" {
		return fmt.Errorf("relational FieldDevice update planning requires PostgreSQL")
	}
	parameters := map[string]any{"owner_id": ownerID, "job_id": jobID}
	if err := s.db.WithContext(ctx).Exec(fieldDeviceUpdateComponentSQL, parameters).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Exec(fieldDeviceUpdateGroupIDSQL, parameters).Error
}

func (s *FieldDeviceUpdatePlanStore) List(ctx context.Context, ownerID, jobID uuid.UUID) ([]facilityjobs.FieldDeviceUpdatePlanItem, error) {
	var records []fieldDeviceUpdatePlanRecord
	err := s.db.WithContext(ctx).Where("owner_id=? AND job_id=?", ownerID, jobID).
		Order("group_ordinal ASC, ordinal ASC").Find(&records).Error
	items := make([]facilityjobs.FieldDeviceUpdatePlanItem, len(records))
	for index, record := range records {
		items[index] = record.toDomain()
	}
	return items, err
}

const fieldDeviceUpdateComponentSQL = `
WITH RECURSIVE command_keys AS MATERIALIZED (
  SELECT plan.ordinal,
    ROW(device.sps_controller_system_type_id,device.system_part_id,device.apparat_id,device.apparat_nr)::text AS current_key,
    ROW(
      device.sps_controller_system_type_id,
      COALESCE(NULLIF(plan.command->>'SystemPartID','')::uuid,device.system_part_id),
      COALESCE(NULLIF(plan.command->>'ApparatID','')::uuid,device.apparat_id),
      COALESCE(NULLIF(plan.command->>'ApparatNr','')::integer,device.apparat_nr)
    )::text AS target_key,
    plan.field_device_id
  FROM field_device_bulk_update_plan_items plan
  JOIN field_devices device ON device.id=plan.field_device_id
  WHERE plan.owner_id=@owner_id AND plan.job_id=@job_id
), edges AS MATERIALIZED (
  SELECT source.ordinal AS left_ordinal,target.ordinal AS right_ordinal
  FROM command_keys source JOIN command_keys target ON source.target_key=target.current_key AND source.ordinal<>target.ordinal
  UNION
  SELECT source.ordinal,target.ordinal
  FROM command_keys source JOIN command_keys target ON source.target_key=target.target_key AND source.ordinal<target.ordinal
  UNION
  SELECT source.ordinal,target.ordinal
  FROM command_keys source JOIN command_keys target ON source.field_device_id=target.field_device_id AND source.ordinal<target.ordinal
), undirected_edges AS (
  SELECT left_ordinal,right_ordinal FROM edges
  UNION SELECT right_ordinal,left_ordinal FROM edges
), reach(root_ordinal,node_ordinal) AS (
  SELECT ordinal,ordinal FROM command_keys
  UNION
  SELECT reach.root_ordinal,edge.right_ordinal
  FROM reach JOIN undirected_edges edge ON edge.left_ordinal=reach.node_ordinal
), components AS (
  SELECT node_ordinal,min(root_ordinal) AS group_ordinal FROM reach GROUP BY node_ordinal
)
UPDATE field_device_bulk_update_plan_items plan SET group_ordinal=components.group_ordinal
FROM components
WHERE plan.owner_id=@owner_id AND plan.job_id=@job_id AND plan.ordinal=components.node_ordinal`

const fieldDeviceUpdateGroupIDSQL = `
WITH groups AS (
  SELECT group_ordinal,md5(string_agg(field_device_id::text,',' ORDER BY field_device_id))::uuid AS dependency_group_id
  FROM field_device_bulk_update_plan_items
  WHERE owner_id=@owner_id AND job_id=@job_id
  GROUP BY group_ordinal
)
UPDATE field_device_bulk_update_plan_items plan SET dependency_group_id=groups.dependency_group_id
FROM groups
WHERE plan.owner_id=@owner_id AND plan.job_id=@job_id AND plan.group_ordinal=groups.group_ordinal`

func planRecord(item facilityjobs.FieldDeviceUpdatePlanItem) fieldDeviceUpdatePlanRecord {
	return fieldDeviceUpdatePlanRecord{
		OwnerID: item.OwnerID, JobID: item.JobID, Ordinal: item.Ordinal,
		GroupOrdinal: item.GroupOrdinal, DependencyGroupID: item.DependencyGroupID,
		FieldDeviceID: item.FieldDeviceID, Command: postgresjson.Document(item.Command), CreatedAt: time.Now().UTC(),
	}
}

func (record fieldDeviceUpdatePlanRecord) toDomain() facilityjobs.FieldDeviceUpdatePlanItem {
	return facilityjobs.FieldDeviceUpdatePlanItem{
		OwnerID: record.OwnerID, JobID: record.JobID, Ordinal: record.Ordinal,
		GroupOrdinal: record.GroupOrdinal, DependencyGroupID: record.DependencyGroupID,
		FieldDeviceID: record.FieldDeviceID, Command: record.Command.Bytes(),
	}
}
