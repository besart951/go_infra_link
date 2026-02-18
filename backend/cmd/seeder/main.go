package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type permissionSeed struct {
	Name        string
	Description string
	Resource    string
	Action      string
}

func seedPermissions(database *gorm.DB) error {
	log.Println("Seeding permissions...")

	actions := []string{"create", "read", "update", "delete"}
	generalResources := []string{"user", "team", "project", "phase", "role", "permission"}
	facilityResources := []string{
		"building",
		"controlcabinet",
		"spscontroller",
		"spscontrollersystemtype",
		"fielddevice",
		"bacnetobject",
		"systempart",
		"systemtype",
		"specification",
		"apparat",
		"notificationclass",
		"statetext",
		"objectdata",
		"alarmdefinition",
	}
	projectSubResources := []string{
		"controlcabinet",
		"spscontroller",
		"spscontrollersystemtype",
		"fielddevice",
		"bacnetobject",
		"systemtype",
	}

	seeds := make([]permissionSeed, 0, 200)
	addPermission := func(name, resource, action string) {
		description := fmt.Sprintf("%s %s", strings.Title(action), strings.ReplaceAll(resource, ".", " "))
		seeds = append(seeds, permissionSeed{
			Name:        name,
			Description: description,
			Resource:    resource,
			Action:      action,
		})
	}

	for _, resource := range generalResources {
		for _, action := range actions {
			addPermission(fmt.Sprintf("%s.%s", resource, action), resource, action)
		}
	}

	for _, resource := range facilityResources {
		for _, action := range actions {
			addPermission(fmt.Sprintf("%s.%s", resource, action), resource, action)
		}
	}

	for _, resource := range projectSubResources {
		for _, action := range actions {
			name := fmt.Sprintf("project.%s.%s", resource, action)
			addPermission(name, fmt.Sprintf("project.%s", resource), action)
		}
	}

	if len(seeds) == 0 {
		return nil
	}

	names := make([]string, len(seeds))
	for i, seed := range seeds {
		names[i] = seed.Name
	}

	var existing []domainUser.Permission
	if err := database.Where("name IN ?", names).Find(&existing).Error; err != nil {
		return err
	}

	existingNames := make(map[string]struct{}, len(existing))
	for _, perm := range existing {
		existingNames[perm.Name] = struct{}{}
	}

	now := time.Now().UTC()
	missing := make([]domainUser.Permission, 0, len(seeds))
	for _, seed := range seeds {
		if _, ok := existingNames[seed.Name]; ok {
			continue
		}
		perm := domainUser.Permission{
			Name:        seed.Name,
			Description: seed.Description,
			Resource:    seed.Resource,
			Action:      seed.Action,
		}
		if err := perm.Base.InitForCreate(now); err != nil {
			return err
		}
		missing = append(missing, perm)
	}

	if len(missing) > 0 {
		if err := database.Create(&missing).Error; err != nil {
			return err
		}
		log.Printf("Seeded %d permissions", len(missing))
	} else {
		log.Println("No new permissions to seed")
	}

	// Ensure superadmin has all permissions.
	var rolePerms []domainUser.RolePermission
	if err := database.Where("role = ? AND permission IN ?", domainUser.RoleSuperAdmin, names).Find(&rolePerms).Error; err != nil {
		return err
	}
	existingRolePerms := make(map[string]struct{}, len(rolePerms))
	for _, rp := range rolePerms {
		existingRolePerms[rp.Permission] = struct{}{}
	}

	missingRolePerms := make([]domainUser.RolePermission, 0, len(names))
	for _, name := range names {
		if _, ok := existingRolePerms[name]; ok {
			continue
		}
		rp := domainUser.RolePermission{
			Role:       domainUser.RoleSuperAdmin,
			Permission: name,
		}
		if err := rp.Base.InitForCreate(now); err != nil {
			return err
		}
		missingRolePerms = append(missingRolePerms, rp)
	}

	if len(missingRolePerms) > 0 {
		if err := database.Create(&missingRolePerms).Error; err != nil {
			return err
		}
		log.Printf("Assigned %d permissions to superadmin", len(missingRolePerms))
	}

	return nil
}

type apparatSeedFile struct {
	Apparats []apparatSeed `json:"apparats"`
}

type apparatSeed struct {
	ID          int     `json:"id"`
	ShortName   string  `json:"short_name"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type systemPartSeedFile struct {
	SystemParts []systemPartSeed `json:"system_parts"`
}

type systemPartSeed struct {
	ID          int     `json:"id"`
	ShortName   string  `json:"short_name"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type apparatSystemPartSeedFile struct {
	ApparatSystemPart []apparatSystemPartSeed `json:"apparat_systempart"`
}

type apparatSystemPartSeed struct {
	ApparatID    int `json:"apparat_id"`
	SystemPartID int `json:"system_part_id"`
}

type buildingSeedFile struct {
	Buildings []buildingSeed `json:"buildings"`
}

type buildingSeed struct {
	ID            int    `json:"id"`
	IWSCode       string `json:"iws_code"`
	BuildingGroup int    `json:"building_group"`
}

type systemTypeSeedFile struct {
	SystemTypes []systemTypeSeed `json:"system_types"`
}

type systemTypeSeed struct {
	ID        int    `json:"id"`
	NumberMin int    `json:"number_min"`
	NumberMax int    `json:"number_max"`
	Name      string `json:"name"`
}

type notificationClassSeedFile struct {
	NotificationClasses []notificationClassSeed `json:"notification_classes"`
}

type notificationClassSeed struct {
	ID                   int    `json:"id"`
	EventCategory        string `json:"event_category"`
	Nc                   int    `json:"nc"`
	ObjectDescription    string `json:"object_description"`
	InternalDescription  string `json:"internal_description"`
	Meaning              string `json:"meaning"`
	AckRequiredNotNormal int    `json:"ack_required_not_normal"`
	AckRequiredError     int    `json:"ack_required_error"`
	AckRequiredNormal    int    `json:"ack_required_normal"`
	NormNotNormal        int    `json:"norm_not_normal"`
	NormError            int    `json:"norm_error"`
	NormNormal           int    `json:"norm_normal"`
}

type stateTextSeedFile struct {
	StateTexts []stateTextSeed `json:"state_texts"`
}

type stateTextSeed struct {
	ID          int     `json:"id"`
	RefNumber   int     `json:"ref_number"`
	StateText1  *string `json:"state_text_1"`
	StateText2  *string `json:"state_text_2"`
	StateText3  *string `json:"state_text_3"`
	StateText4  *string `json:"state_text_4"`
	StateText5  *string `json:"state_text_5"`
	StateText6  *string `json:"state_text_6"`
	StateText7  *string `json:"state_text_7"`
	StateText8  *string `json:"state_text_8"`
	StateText9  *string `json:"state_text_9"`
	StateText10 *string `json:"state_text_10"`
	StateText11 *string `json:"state_text_11"`
	StateText12 *string `json:"state_text_12"`
	StateText13 *string `json:"state_text_13"`
	StateText14 *string `json:"state_text_14"`
	StateText15 *string `json:"state_text_15"`
	StateText16 *string `json:"state_text_16"`
}

type phaseSeedFile struct {
	Phases []phaseSeed `json:"phases"`
}

type phaseSeed struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type systemPartApparatLink struct {
	SystemPartID uuid.UUID `gorm:"column:system_part_id"`
	ApparatID    uuid.UUID `gorm:"column:apparat_id"`
}

func findSeedFile(seedDir, prefix string) (string, error) {
	pattern := filepath.Join(seedDir, prefix+"_*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("seed file not found for pattern %s", pattern)
	}
	sort.Strings(matches)
	return matches[len(matches)-1], nil
}

func readSeedFile(path string, dest any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func normalizeKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func seedPhases(database *gorm.DB, seedDir string) error {
	path, err := findSeedFile(seedDir, "phases")
	if err != nil {
		return err
	}

	var payload phaseSeedFile
	if err := readSeedFile(path, &payload); err != nil {
		return err
	}

	var existing []domainProject.Phase
	if err := database.Model(&domainProject.Phase{}).
		Select("id", "name").
		Find(&existing).Error; err != nil {
		return err
	}

	existingByName := make(map[string]uuid.UUID, len(existing))
	for _, item := range existing {
		existingByName[normalizeKey(item.Name)] = item.ID
	}

	toCreate := make([]*domainProject.Phase, 0, len(payload.Phases))
	now := time.Now().UTC()
	created := 0
	for _, seed := range payload.Phases {
		name := strings.TrimSpace(seed.Name)
		if name == "" {
			continue
		}
		key := normalizeKey(name)
		if _, ok := existingByName[key]; ok {
			continue
		}
		phase := domainProject.Phase{Name: name}
		if err := phase.Base.InitForCreate(now); err != nil {
			return err
		}
		existingByName[key] = phase.ID
		toCreate = append(toCreate, &phase)
		created++
	}

	if len(toCreate) > 0 {
		if err := database.CreateInBatches(toCreate, 100).Error; err != nil {
			return err
		}
	}

	log.Printf("Seeded phases: %d new (file=%s)", created, filepath.Base(path))
	return nil
}

func seedBuildings(database *gorm.DB, seedDir string) error {
	path, err := findSeedFile(seedDir, "buildings")
	if err != nil {
		return err
	}

	var payload buildingSeedFile
	if err := readSeedFile(path, &payload); err != nil {
		return err
	}

	var existing []domainFacility.Building
	if err := database.Model(&domainFacility.Building{}).
		Select("id", "iws_code", "building_group").
		Find(&existing).Error; err != nil {
		return err
	}

	existingKeys := make(map[string]uuid.UUID, len(existing))
	for _, item := range existing {
		key := normalizeKey(item.IWSCode) + ":" + strconv.Itoa(item.BuildingGroup)
		existingKeys[key] = item.ID
	}

	toCreate := make([]*domainFacility.Building, 0, len(payload.Buildings))
	now := time.Now().UTC()
	created := 0
	for _, seed := range payload.Buildings {
		iwsCode := strings.TrimSpace(seed.IWSCode)
		if iwsCode == "" || seed.BuildingGroup == 0 {
			continue
		}
		key := normalizeKey(iwsCode) + ":" + strconv.Itoa(seed.BuildingGroup)
		if _, ok := existingKeys[key]; ok {
			continue
		}
		building := domainFacility.Building{IWSCode: iwsCode, BuildingGroup: seed.BuildingGroup}
		if err := building.Base.InitForCreate(now); err != nil {
			return err
		}
		existingKeys[key] = building.ID
		toCreate = append(toCreate, &building)
		created++
	}

	if len(toCreate) > 0 {
		if err := database.CreateInBatches(toCreate, 200).Error; err != nil {
			return err
		}
	}

	log.Printf("Seeded buildings: %d new (file=%s)", created, filepath.Base(path))
	return nil
}

func seedSystemTypes(database *gorm.DB, seedDir string) error {
	path, err := findSeedFile(seedDir, "system_types")
	if err != nil {
		return err
	}

	var payload systemTypeSeedFile
	if err := readSeedFile(path, &payload); err != nil {
		return err
	}

	var existing []domainFacility.SystemType
	if err := database.Model(&domainFacility.SystemType{}).
		Select("id", "name").
		Find(&existing).Error; err != nil {
		return err
	}

	existingByName := make(map[string]uuid.UUID, len(existing))
	for _, item := range existing {
		existingByName[normalizeKey(item.Name)] = item.ID
	}

	toCreate := make([]*domainFacility.SystemType, 0, len(payload.SystemTypes))
	now := time.Now().UTC()
	created := 0
	for _, seed := range payload.SystemTypes {
		name := strings.TrimSpace(seed.Name)
		if name == "" {
			continue
		}
		key := normalizeKey(name)
		if _, ok := existingByName[key]; ok {
			continue
		}
		systemType := domainFacility.SystemType{
			NumberMin: seed.NumberMin,
			NumberMax: seed.NumberMax,
			Name:      name,
		}
		if err := systemType.Base.InitForCreate(now); err != nil {
			return err
		}
		existingByName[key] = systemType.ID
		toCreate = append(toCreate, &systemType)
		created++
	}

	if len(toCreate) > 0 {
		if err := database.CreateInBatches(toCreate, 200).Error; err != nil {
			return err
		}
	}

	log.Printf("Seeded system types: %d new (file=%s)", created, filepath.Base(path))
	return nil
}

func seedSystemParts(database *gorm.DB, seedDir string) (map[int]uuid.UUID, error) {
	path, err := findSeedFile(seedDir, "system_parts")
	if err != nil {
		return nil, err
	}

	var payload systemPartSeedFile
	if err := readSeedFile(path, &payload); err != nil {
		return nil, err
	}

	var existing []domainFacility.SystemPart
	if err := database.Model(&domainFacility.SystemPart{}).
		Select("id", "short_name", "name").
		Find(&existing).Error; err != nil {
		return nil, err
	}

	byShort := make(map[string]uuid.UUID, len(existing))
	byName := make(map[string]uuid.UUID, len(existing))
	for _, item := range existing {
		byShort[normalizeKey(item.ShortName)] = item.ID
		byName[normalizeKey(item.Name)] = item.ID
	}

	toCreate := make([]*domainFacility.SystemPart, 0, len(payload.SystemParts))
	seedIDMap := make(map[int]uuid.UUID, len(payload.SystemParts))
	now := time.Now().UTC()
	created := 0
	for _, seed := range payload.SystemParts {
		shortName := strings.TrimSpace(seed.ShortName)
		name := strings.TrimSpace(seed.Name)
		if shortName == "" || name == "" {
			continue
		}
		keyShort := normalizeKey(shortName)
		keyName := normalizeKey(name)
		if existingID, ok := byShort[keyShort]; ok {
			seedIDMap[seed.ID] = existingID
			continue
		}
		if existingID, ok := byName[keyName]; ok {
			seedIDMap[seed.ID] = existingID
			continue
		}
		systemPart := domainFacility.SystemPart{
			ShortName:   shortName,
			Name:        name,
			Description: seed.Description,
		}
		if err := systemPart.Base.InitForCreate(now); err != nil {
			return nil, err
		}
		byShort[keyShort] = systemPart.ID
		byName[keyName] = systemPart.ID
		seedIDMap[seed.ID] = systemPart.ID
		toCreate = append(toCreate, &systemPart)
		created++
	}

	if len(toCreate) > 0 {
		if err := database.CreateInBatches(toCreate, 200).Error; err != nil {
			return nil, err
		}
	}

	log.Printf("Seeded system parts: %d new (file=%s)", created, filepath.Base(path))
	return seedIDMap, nil
}

func seedApparats(database *gorm.DB, seedDir string) (map[int]uuid.UUID, error) {
	path, err := findSeedFile(seedDir, "apparats")
	if err != nil {
		return nil, err
	}

	var payload apparatSeedFile
	if err := readSeedFile(path, &payload); err != nil {
		return nil, err
	}

	var existing []domainFacility.Apparat
	if err := database.Model(&domainFacility.Apparat{}).
		Select("id", "short_name", "name").
		Find(&existing).Error; err != nil {
		return nil, err
	}

	byShort := make(map[string]uuid.UUID, len(existing))
	byName := make(map[string]uuid.UUID, len(existing))
	for _, item := range existing {
		byShort[normalizeKey(item.ShortName)] = item.ID
		byName[normalizeKey(item.Name)] = item.ID
	}

	toCreate := make([]*domainFacility.Apparat, 0, len(payload.Apparats))
	seedIDMap := make(map[int]uuid.UUID, len(payload.Apparats))
	now := time.Now().UTC()
	created := 0
	for _, seed := range payload.Apparats {
		shortName := strings.TrimSpace(seed.ShortName)
		name := strings.TrimSpace(seed.Name)
		if shortName == "" || name == "" {
			continue
		}
		keyShort := normalizeKey(shortName)
		keyName := normalizeKey(name)
		if existingID, ok := byShort[keyShort]; ok {
			seedIDMap[seed.ID] = existingID
			continue
		}
		if existingID, ok := byName[keyName]; ok {
			seedIDMap[seed.ID] = existingID
			continue
		}
		apparat := domainFacility.Apparat{
			ShortName:   shortName,
			Name:        name,
			Description: seed.Description,
		}
		if err := apparat.Base.InitForCreate(now); err != nil {
			return nil, err
		}
		byShort[keyShort] = apparat.ID
		byName[keyName] = apparat.ID
		seedIDMap[seed.ID] = apparat.ID
		toCreate = append(toCreate, &apparat)
		created++
	}

	if len(toCreate) > 0 {
		if err := database.CreateInBatches(toCreate, 200).Error; err != nil {
			return nil, err
		}
	}

	log.Printf("Seeded apparats: %d new (file=%s)", created, filepath.Base(path))
	return seedIDMap, nil
}

func seedApparatSystemParts(database *gorm.DB, seedDir string, apparatIDs map[int]uuid.UUID, systemPartIDs map[int]uuid.UUID) error {
	path, err := findSeedFile(seedDir, "apparat_systempart")
	if err != nil {
		return err
	}

	var payload apparatSystemPartSeedFile
	if err := readSeedFile(path, &payload); err != nil {
		return err
	}

	links := make([]systemPartApparatLink, 0, len(payload.ApparatSystemPart))
	skipped := 0
	for _, seed := range payload.ApparatSystemPart {
		apparatID, okApparat := apparatIDs[seed.ApparatID]
		systemPartID, okSystemPart := systemPartIDs[seed.SystemPartID]
		if !okApparat || !okSystemPart {
			skipped++
			continue
		}
		links = append(links, systemPartApparatLink{
			SystemPartID: systemPartID,
			ApparatID:    apparatID,
		})
	}

	if len(links) > 0 {
		if err := database.Table("system_part_apparats").
			Clauses(clause.OnConflict{DoNothing: true}).
			CreateInBatches(links, 500).Error; err != nil {
			return err
		}
	}

	log.Printf("Seeded apparat-system parts: %d links (%d skipped, file=%s)", len(links), skipped, filepath.Base(path))
	return nil
}

func seedNotificationClasses(database *gorm.DB, seedDir string) error {
	path, err := findSeedFile(seedDir, "notification_classes")
	if err != nil {
		return err
	}

	var payload notificationClassSeedFile
	if err := readSeedFile(path, &payload); err != nil {
		return err
	}

	var existing []domainFacility.NotificationClass
	if err := database.Model(&domainFacility.NotificationClass{}).
		Select("id", "nc").
		Find(&existing).Error; err != nil {
		return err
	}

	existingByNc := make(map[int]uuid.UUID, len(existing))
	for _, item := range existing {
		existingByNc[item.Nc] = item.ID
	}

	toCreate := make([]*domainFacility.NotificationClass, 0, len(payload.NotificationClasses))
	now := time.Now().UTC()
	created := 0
	for _, seed := range payload.NotificationClasses {
		if strings.TrimSpace(seed.EventCategory) == "" {
			continue
		}
		if _, ok := existingByNc[seed.Nc]; ok {
			continue
		}
		nc := domainFacility.NotificationClass{
			EventCategory:        strings.TrimSpace(seed.EventCategory),
			Nc:                   seed.Nc,
			ObjectDescription:    strings.TrimSpace(seed.ObjectDescription),
			InternalDescription:  strings.TrimSpace(seed.InternalDescription),
			Meaning:              strings.TrimSpace(seed.Meaning),
			AckRequiredNotNormal: seed.AckRequiredNotNormal != 0,
			AckRequiredError:     seed.AckRequiredError != 0,
			AckRequiredNormal:    seed.AckRequiredNormal != 0,
			NormNotNormal:        seed.NormNotNormal,
			NormError:            seed.NormError,
			NormNormal:           seed.NormNormal,
		}
		if err := nc.Base.InitForCreate(now); err != nil {
			return err
		}
		existingByNc[seed.Nc] = nc.ID
		toCreate = append(toCreate, &nc)
		created++
	}

	if len(toCreate) > 0 {
		if err := database.CreateInBatches(toCreate, 200).Error; err != nil {
			return err
		}
	}

	log.Printf("Seeded notification classes: %d new (file=%s)", created, filepath.Base(path))
	return nil
}

func seedStateTexts(database *gorm.DB, seedDir string) error {
	path, err := findSeedFile(seedDir, "state_texts")
	if err != nil {
		return err
	}

	var payload stateTextSeedFile
	if err := readSeedFile(path, &payload); err != nil {
		return err
	}

	var existing []domainFacility.StateText
	if err := database.Model(&domainFacility.StateText{}).
		Select("id", "ref_number").
		Find(&existing).Error; err != nil {
		return err
	}

	existingByRef := make(map[int]uuid.UUID, len(existing))
	for _, item := range existing {
		existingByRef[item.RefNumber] = item.ID
	}

	toCreate := make([]*domainFacility.StateText, 0, len(payload.StateTexts))
	now := time.Now().UTC()
	created := 0
	for _, seed := range payload.StateTexts {
		if seed.RefNumber == 0 {
			continue
		}
		if _, ok := existingByRef[seed.RefNumber]; ok {
			continue
		}
		stateText := domainFacility.StateText{
			RefNumber:   seed.RefNumber,
			StateText1:  seed.StateText1,
			StateText2:  seed.StateText2,
			StateText3:  seed.StateText3,
			StateText4:  seed.StateText4,
			StateText5:  seed.StateText5,
			StateText6:  seed.StateText6,
			StateText7:  seed.StateText7,
			StateText8:  seed.StateText8,
			StateText9:  seed.StateText9,
			StateText10: seed.StateText10,
			StateText11: seed.StateText11,
			StateText12: seed.StateText12,
			StateText13: seed.StateText13,
			StateText14: seed.StateText14,
			StateText15: seed.StateText15,
			StateText16: seed.StateText16,
		}
		if err := stateText.Base.InitForCreate(now); err != nil {
			return err
		}
		existingByRef[seed.RefNumber] = stateText.ID
		toCreate = append(toCreate, &stateText)
		created++
	}

	if len(toCreate) > 0 {
		if err := database.CreateInBatches(toCreate, 500).Error; err != nil {
			return err
		}
	}

	log.Printf("Seeded state texts: %d new (file=%s)", created, filepath.Base(path))
	return nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load config from file: %v. Proceeding with defaults/overrides.", err)
	}

	log.Printf("Using DB Config: Type=%s, DSN=%s", cfg.DBType, cfg.DBDsn)

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database. Starting seed...")
	if err := seedPermissions(database); err != nil {
		log.Fatalf("Failed to seed permissions: %v", err)
	}

	seedDir := filepath.Join("data", "seed")
	if err := seedPhases(database, seedDir); err != nil {
		log.Fatalf("Failed to seed phases: %v", err)
	}
	if err := seedBuildings(database, seedDir); err != nil {
		log.Fatalf("Failed to seed buildings: %v", err)
	}
	if err := seedSystemTypes(database, seedDir); err != nil {
		log.Fatalf("Failed to seed system types: %v", err)
	}
	if err := seedNotificationClasses(database, seedDir); err != nil {
		log.Fatalf("Failed to seed notification classes: %v", err)
	}
	if err := seedStateTexts(database, seedDir); err != nil {
		log.Fatalf("Failed to seed state texts: %v", err)
	}

	systemPartIDs, err := seedSystemParts(database, seedDir)
	if err != nil {
		log.Fatalf("Failed to seed system parts: %v", err)
	}
	apparatIDs, err := seedApparats(database, seedDir)
	if err != nil {
		log.Fatalf("Failed to seed apparats: %v", err)
	}
	if err := seedApparatSystemParts(database, seedDir, apparatIDs, systemPartIDs); err != nil {
		log.Fatalf("Failed to seed apparat-system parts: %v", err)
	}

	log.Println("Seeding complete!")
}
