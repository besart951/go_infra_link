package historysql

import (
	"context"
	"strconv"
	"strings"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

func (s *Store) enrichActorNames(ctx context.Context, events []domainHistory.ChangeEvent) error {
	if len(events) == 0 {
		return nil
	}

	seen := map[uuid.UUID]struct{}{}
	ids := make([]uuid.UUID, 0, len(events))
	for i := range events {
		if events[i].ActorID == nil || *events[i].ActorID == uuid.Nil {
			continue
		}
		if _, ok := seen[*events[i].ActorID]; ok {
			continue
		}
		seen[*events[i].ActorID] = struct{}{}
		ids = append(ids, *events[i].ActorID)
	}
	if len(ids) == 0 {
		return nil
	}

	var actors []actorNameRow
	if err := s.db.WithContext(ctx).
		Table("users").
		Select("id, first_name, last_name, email").
		Where("id IN ?", ids).
		Scan(&actors).Error; err != nil {
		return err
	}

	names := make(map[uuid.UUID]string, len(actors))
	for _, actor := range actors {
		if name := actor.displayName(); name != "" {
			names[actor.ID] = name
		}
	}
	for i := range events {
		if events[i].ActorID == nil {
			continue
		}
		name, ok := names[*events[i].ActorID]
		if !ok {
			continue
		}
		events[i].ActorName = stringPtr(name)
	}
	return nil
}

func (s *Store) enrichScopeSummaries(ctx context.Context, events []domainHistory.ChangeEvent) error {
	if len(events) == 0 {
		return nil
	}

	eventIndexes := make(map[uuid.UUID]int, len(events))
	eventIDs := make([]uuid.UUID, 0, len(events))
	for i := range events {
		eventIndexes[events[i].ID] = i
		eventIDs = append(eventIDs, events[i].ID)
	}

	var rows []scopeSummaryRow
	if err := s.db.WithContext(ctx).
		Table("change_event_scopes").
		Select("change_event_id, scope_type, scope_id").
		Where("change_event_id IN ?", eventIDs).
		Order("scope_type ASC, scope_id ASC").
		Scan(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	scopeIDs := map[string]map[uuid.UUID]struct{}{}
	for _, row := range rows {
		if _, ok := scopeIDs[row.ScopeType]; !ok {
			scopeIDs[row.ScopeType] = map[uuid.UUID]struct{}{}
		}
		scopeIDs[row.ScopeType][row.ScopeID] = struct{}{}
	}
	labels, err := s.loadScopeLabels(ctx, scopeIDs)
	if err != nil {
		return err
	}

	for _, row := range rows {
		index, ok := eventIndexes[row.ChangeEventID]
		if !ok {
			continue
		}
		summary := domainHistory.Scope{
			ScopeType: row.ScopeType,
			ScopeID:   row.ScopeID,
		}
		if label, ok := labels[scopeLabelKey(row.ScopeType, row.ScopeID)]; ok {
			summary.Label = stringPtr(label)
		}
		events[index].Scopes = append(events[index].Scopes, summary)
	}
	return nil
}

func (s *Store) loadScopeLabels(ctx context.Context, scopeIDs map[string]map[uuid.UUID]struct{}) (map[string]string, error) {
	labels := map[string]string{}
	if err := s.addBuildingLabels(ctx, labels, scopeIDs[scopeBuilding]); err != nil {
		return nil, err
	}
	if err := s.addControlCabinetLabels(ctx, labels, scopeIDs[scopeControlCabinet]); err != nil {
		return nil, err
	}
	if err := s.addSPSControllerLabels(ctx, labels, scopeIDs[scopeSPSController]); err != nil {
		return nil, err
	}
	if err := s.addSPSControllerSystemTypeLabels(ctx, labels, scopeIDs[scopeSPSControllerSystemType]); err != nil {
		return nil, err
	}
	if err := s.addFieldDeviceLabels(ctx, labels, scopeIDs[scopeFieldDevice]); err != nil {
		return nil, err
	}
	if err := s.addBacnetObjectLabels(ctx, labels, scopeIDs[scopeBacnetObject]); err != nil {
		return nil, err
	}
	return labels, nil
}

func (s *Store) addBuildingLabels(ctx context.Context, labels map[string]string, ids map[uuid.UUID]struct{}) error {
	if len(ids) == 0 {
		return nil
	}
	var rows []struct {
		ID            uuid.UUID `gorm:"column:id"`
		IWSCode       string    `gorm:"column:iws_code"`
		BuildingGroup int       `gorm:"column:building_group"`
	}
	if err := s.db.WithContext(ctx).
		Table("buildings").
		Select("id, iws_code, building_group").
		Where("id IN ?", sortedUUIDs(ids)).
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		if label := firstNonBlank(row.IWSCode, strconv.Itoa(row.BuildingGroup)); label != "" {
			labels[scopeLabelKey(scopeBuilding, row.ID)] = label
		}
	}
	return nil
}

func (s *Store) addControlCabinetLabels(ctx context.Context, labels map[string]string, ids map[uuid.UUID]struct{}) error {
	if len(ids) == 0 {
		return nil
	}
	var rows []struct {
		ID uuid.UUID `gorm:"column:id"`
		Nr *string   `gorm:"column:control_cabinet_nr"`
	}
	if err := s.db.WithContext(ctx).
		Table("control_cabinets").
		Select("id, control_cabinet_nr").
		Where("id IN ?", sortedUUIDs(ids)).
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		if label := firstNonBlank(ptrString(row.Nr)); label != "" {
			labels[scopeLabelKey(scopeControlCabinet, row.ID)] = label
		}
	}
	return nil
}

func (s *Store) addSPSControllerLabels(ctx context.Context, labels map[string]string, ids map[uuid.UUID]struct{}) error {
	if len(ids) == 0 {
		return nil
	}
	var rows []struct {
		ID         uuid.UUID `gorm:"column:id"`
		DeviceName string    `gorm:"column:device_name"`
		GADevice   *string   `gorm:"column:ga_device"`
		IPAddress  *string   `gorm:"column:ip_address"`
	}
	if err := s.db.WithContext(ctx).
		Table("sps_controllers").
		Select("id, device_name, ga_device, ip_address").
		Where("id IN ?", sortedUUIDs(ids)).
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		label := firstNonBlank(row.DeviceName, ptrString(row.GADevice), ptrString(row.IPAddress))
		if label != "" {
			labels[scopeLabelKey(scopeSPSController, row.ID)] = label
		}
	}
	return nil
}

func (s *Store) addSPSControllerSystemTypeLabels(ctx context.Context, labels map[string]string, ids map[uuid.UUID]struct{}) error {
	if len(ids) == 0 {
		return nil
	}
	var rows []struct {
		ID           uuid.UUID `gorm:"column:id"`
		Number       *int      `gorm:"column:number"`
		DocumentName *string   `gorm:"column:document_name"`
		SystemName   *string   `gorm:"column:system_name"`
	}
	if err := s.db.WithContext(ctx).
		Table("sps_controller_system_types AS st").
		Select("st.id, st.number, st.document_name, sy.name AS system_name").
		Joins("LEFT JOIN system_types AS sy ON sy.id = st.system_type_id").
		Where("st.id IN ?", sortedUUIDs(ids)).
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		label := firstNonBlank(ptrString(row.DocumentName), ptrString(row.SystemName), intLabel(row.Number))
		if label != "" {
			labels[scopeLabelKey(scopeSPSControllerSystemType, row.ID)] = label
		}
	}
	return nil
}

func (s *Store) addFieldDeviceLabels(ctx context.Context, labels map[string]string, ids map[uuid.UUID]struct{}) error {
	if len(ids) == 0 {
		return nil
	}
	var rows []struct {
		ID          uuid.UUID `gorm:"column:id"`
		BMK         *string   `gorm:"column:bmk"`
		Description *string   `gorm:"column:description"`
	}
	if err := s.db.WithContext(ctx).
		Table("field_devices").
		Select("id, bmk, description").
		Where("id IN ?", sortedUUIDs(ids)).
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		label := firstNonBlank(ptrString(row.BMK), ptrString(row.Description))
		if label != "" {
			labels[scopeLabelKey(scopeFieldDevice, row.ID)] = label
		}
	}
	return nil
}

func (s *Store) addBacnetObjectLabels(ctx context.Context, labels map[string]string, ids map[uuid.UUID]struct{}) error {
	if len(ids) == 0 {
		return nil
	}
	var rows []struct {
		ID          uuid.UUID `gorm:"column:id"`
		TextFix     string    `gorm:"column:text_fix"`
		Description *string   `gorm:"column:description"`
	}
	if err := s.db.WithContext(ctx).
		Table("bacnet_objects").
		Select("id, text_fix, description").
		Where("id IN ?", sortedUUIDs(ids)).
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		label := firstNonBlank(row.TextFix, ptrString(row.Description))
		if label != "" {
			labels[scopeLabelKey(scopeBacnetObject, row.ID)] = label
		}
	}
	return nil
}

type actorNameRow struct {
	ID        uuid.UUID `gorm:"column:id"`
	FirstName string    `gorm:"column:first_name"`
	LastName  string    `gorm:"column:last_name"`
	Email     string    `gorm:"column:email"`
}

type scopeSummaryRow struct {
	ChangeEventID uuid.UUID `gorm:"column:change_event_id"`
	ScopeType     string    `gorm:"column:scope_type"`
	ScopeID       uuid.UUID `gorm:"column:scope_id"`
}

func (r actorNameRow) displayName() string {
	name := strings.TrimSpace(strings.Join([]string{
		strings.TrimSpace(r.FirstName),
		strings.TrimSpace(r.LastName),
	}, " "))
	if name != "" {
		return name
	}
	return strings.TrimSpace(r.Email)
}

func stringPtr(value string) *string {
	return &value
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func ptrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intLabel(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func scopeLabelKey(scopeType string, id uuid.UUID) string {
	return scopeType + ":" + id.String()
}
