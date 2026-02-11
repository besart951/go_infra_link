package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

const (
	NumBuildings        = 50
	CabinetsPerBuilding = 4
	SPSPerCabinet       = 3

	// Additional seed data
	NumSystemTypes              = 20
	SystemTypesPerSPS           = 1
	NumObjectDatas              = 30
	BacnetObjectsPerObjData     = 3
	NumSystemParts              = 40
	NumApparats                 = 60
	MinSystemPartsPerApparat    = 1
	MaxSystemPartsPerApparat    = 3
	MinApparatsPerObjectData    = 1
	MaxApparatsPerObjectData    = 3
	NumStateTexts               = 20
	NumNotificationClasses      = 10
	NumAlarmDefinitions         = 10
	FieldDevicesPerSpsSysType   = 1
	BacnetObjectsPerFieldDevice = 2
)

type objectDataBacnetObject struct {
	ObjectDataID   uuid.UUID `gorm:"type:uuid;column:object_data_id;primaryKey"`
	BacnetObjectID uuid.UUID `gorm:"type:uuid;column:bacnet_object_id;primaryKey"`
}

func (objectDataBacnetObject) TableName() string {
	return "object_data_bacnet_objects"
}

type objectDataApparat struct {
	ObjectDataID uuid.UUID `gorm:"type:uuid;column:object_data_id;primaryKey"`
	ApparatID    uuid.UUID `gorm:"type:uuid;column:apparat_id;primaryKey"`
}

func (objectDataApparat) TableName() string {
	return "object_data_apparats"
}

type systemPartApparat struct {
	ApparatID    uuid.UUID `gorm:"type:uuid;column:apparat_id;primaryKey"`
	SystemPartID uuid.UUID `gorm:"type:uuid;column:system_part_id;primaryKey"`
}

func (systemPartApparat) TableName() string {
	return "system_part_apparats"
}

func main() {
	rand.Seed(time.Now().UnixNano())

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

	var (
		buildings           []facility.Building
		controlCabinets     []facility.ControlCabinet
		spsControllers      []facility.SPSController
		systemTypes         []facility.SystemType
		systemParts         []facility.SystemPart
		apparats            []facility.Apparat
		spsSysTypes         []facility.SPSControllerSystemType
		stateTexts          []facility.StateText
		notificationClasses []facility.NotificationClass
		alarmDefinitions    []facility.AlarmDefinition
		specifications      []facility.Specification
		fieldDevices        []facility.FieldDevice
		objectDatas         []facility.ObjectData
		bacnetObjects       []facility.BacnetObject
		fieldDeviceBos      []facility.BacnetObject
		objDataLinks        []objectDataBacnetObject
		objDataApparatLinks []objectDataApparat
		sysPartApparatLinks []systemPartApparat
	)

	// Generate data in memory
	log.Println("Generating data structures...")

	start := time.Now()

	for i := 0; i < NumBuildings; i++ {
		bID := uuid.New()
		iwsCode := fmt.Sprintf("%04d", i) // Unique 4 digits within seed batch

		buildings = append(buildings, facility.Building{
			Base: domain.Base{
				ID:        bID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			IWSCode:       iwsCode,
			BuildingGroup: rand.Intn(10) + 1,
		})

		for j := 0; j < CabinetsPerBuilding; j++ {
			cID := uuid.New()
			// Ensure cabNr is <= 11 chars. "C%04d-%02d" = 1+4+1+2 = 8 chars. Safe.
			cabNr := fmt.Sprintf("C%s-%02d", iwsCode, j)
			netA := i % 256
			netB := j % 256
			vlan := fmt.Sprintf("%d", i+1)

			controlCabinets = append(controlCabinets, facility.ControlCabinet{
				Base: domain.Base{
					ID:        cID,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				ControlCabinetNr: ptr(cabNr),
				BuildingID:       bID,
			})

			for k := 0; k < SPSPerCabinet; k++ {
				sID := uuid.New()
				gaDevice := gaDeviceCode(k)
				ip := fmt.Sprintf("10.%d.%d.%d", netA, netB, k+2)
				gateway := fmt.Sprintf("10.%d.%d.1", netA, netB)
				subnet := "255.255.255.0"

				spsControllers = append(spsControllers, facility.SPSController{
					Base: domain.Base{
						ID:        sID,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					ControlCabinetID: cID,
					GADevice:         ptr(gaDevice),
					DeviceName:       fmt.Sprintf("SPS-%s-%d", cabNr, k),
					IPAddress:        ptr(ip),
					Subnet:           ptr(subnet),
					Gateway:          ptr(gateway),
					Vlan:             ptr(vlan),
				})
			}
		}
	}

	// System types
	for i := 0; i < NumSystemTypes; i++ {
		min := i*10 + 1
		max := (i + 1) * 10
		systemTypes = append(systemTypes, facility.SystemType{
			Base: domain.Base{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			NumberMin: min,
			NumberMax: max,
			Name:      fmt.Sprintf("SystemType-%02d", i+1),
		})
	}

	// System parts
	for i := 0; i < NumSystemParts; i++ {
		id := uuid.New()
		suffix := id.String()[:8]
		desc := fmt.Sprintf("SystemPart seeded %s", suffix)
		systemParts = append(systemParts, facility.SystemPart{
			Base: domain.Base{
				ID:        id,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			ShortName:   fmt.Sprintf("SP-%03d-%s", i+1, suffix),
			Name:        fmt.Sprintf("SystemPart-%03d-%s", i+1, suffix),
			Description: &desc,
		})
	}

	// Apparats
	for i := 0; i < NumApparats; i++ {
		id := uuid.New()
		suffix := id.String()[:8]
		desc := fmt.Sprintf("Apparat seeded %s", suffix)
		apparats = append(apparats, facility.Apparat{
			Base: domain.Base{
				ID:        id,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			ShortName:   fmt.Sprintf("A-%03d-%s", i+1, suffix),
			Name:        fmt.Sprintf("Apparat-%03d-%s", i+1, suffix),
			Description: &desc,
		})
	}

	// State texts
	for i := 0; i < NumStateTexts; i++ {
		st1 := fmt.Sprintf("StateText-%03d-A", i+1)
		st2 := fmt.Sprintf("StateText-%03d-B", i+1)
		stateTexts = append(stateTexts, facility.StateText{
			Base: domain.Base{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			RefNumber:  i + 1,
			StateText1: &st1,
			StateText2: &st2,
		})
	}

	// Notification classes
	for i := 0; i < NumNotificationClasses; i++ {
		notificationClasses = append(notificationClasses, facility.NotificationClass{
			Base: domain.Base{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			EventCategory:        "event",
			Nc:                   100 + i,
			ObjectDescription:    fmt.Sprintf("NC-%03d", i+1),
			InternalDescription:  fmt.Sprintf("Notification Class %03d", i+1),
			Meaning:              "seeded",
			AckRequiredNotNormal: i%2 == 0,
			AckRequiredError:     i%3 == 0,
			AckRequiredNormal:    i%4 == 0,
			NormNotNormal:        rand.Intn(3),
			NormError:            rand.Intn(3),
			NormNormal:           rand.Intn(3),
		})
	}

	// Alarm definitions
	for i := 0; i < NumAlarmDefinitions; i++ {
		note := fmt.Sprintf("Alarm note %03d", i+1)
		alarmDefinitions = append(alarmDefinitions, facility.AlarmDefinition{
			Base: domain.Base{
				ID:        uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      fmt.Sprintf("Alarm Definition %03d", i+1),
			AlarmNote: &note,
		})
	}

	// Link SystemPart <-> Apparat
	if len(systemParts) > 0 {
		for i := 0; i < len(apparats); i++ {
			ap := apparats[i]
			n := MinSystemPartsPerApparat
			if MaxSystemPartsPerApparat > MinSystemPartsPerApparat {
				n += rand.Intn(MaxSystemPartsPerApparat - MinSystemPartsPerApparat + 1)
			}

			seen := make(map[uuid.UUID]struct{}, n)
			for j := 0; j < n; j++ {
				sp := systemParts[rand.Intn(len(systemParts))]
				if _, ok := seen[sp.ID]; ok {
					j--
					continue
				}
				seen[sp.ID] = struct{}{}
				sysPartApparatLinks = append(sysPartApparatLinks, systemPartApparat{ApparatID: ap.ID, SystemPartID: sp.ID})
			}
		}
	}

	// Build lookup map for apparat -> system parts
	apparatToSystemParts := make(map[uuid.UUID][]uuid.UUID, len(apparats))
	for _, link := range sysPartApparatLinks {
		apparatToSystemParts[link.ApparatID] = append(apparatToSystemParts[link.ApparatID], link.SystemPartID)
	}

	// SPS controller system types (linking SPS controllers to system types)
	if len(systemTypes) > 0 {
		for i := 0; i < len(spsControllers); i++ {
			sps := spsControllers[i]
			for j := 0; j < SystemTypesPerSPS; j++ {
				st := systemTypes[(i+j)%len(systemTypes)]
				n := (j + 1)
				doc := fmt.Sprintf("DOC-%s-%d", sps.DeviceName, n)
				spsSysTypes = append(spsSysTypes, facility.SPSControllerSystemType{
					Base: domain.Base{
						ID:        uuid.New(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Number:          &n,
					DocumentName:    &doc,
					SPSControllerID: sps.ID,
					SystemTypeID:    st.ID,
				})
			}
		}
	}

	// Object data templates + bacnet objects, linked via join table
	softwareTypes := []facility.BacnetSoftwareType{
		facility.BacnetSoftwareTypeAI,
		facility.BacnetSoftwareTypeAO,
		facility.BacnetSoftwareTypeAV,
		facility.BacnetSoftwareTypeBI,
		facility.BacnetSoftwareTypeBO,
		facility.BacnetSoftwareTypeBV,
		facility.BacnetSoftwareTypeMI,
		facility.BacnetSoftwareTypeMO,
		facility.BacnetSoftwareTypeMV,
		facility.BacnetSoftwareTypeCA,
		facility.BacnetSoftwareTypeEE,
		facility.BacnetSoftwareTypeLP,
		facility.BacnetSoftwareTypeNC,
		facility.BacnetSoftwareTypeSC,
		facility.BacnetSoftwareTypeTL,
	}

	// Field devices + specifications + field device bacnet objects
	if len(spsSysTypes) > 0 && len(apparats) > 0 && len(systemParts) > 0 {
		for i := 0; i < len(spsSysTypes); i++ {
			spsType := spsSysTypes[i]
			for j := 0; j < FieldDevicesPerSpsSysType; j++ {
				fdID := uuid.New()
				apparat := apparats[rand.Intn(len(apparats))]
				systemPartIDs := apparatToSystemParts[apparat.ID]
				spID := systemParts[rand.Intn(len(systemParts))].ID
				if len(systemPartIDs) > 0 {
					spID = systemPartIDs[rand.Intn(len(systemPartIDs))]
				}

				bmk := fmt.Sprintf("BMK-%03d-%02d", i+1, j+1)
				desc := fmt.Sprintf("Field device %s", bmk)
				spec := facility.Specification{
					Base: domain.Base{
						ID:        uuid.New(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					SpecificationSupplier: ptr("Supplier"),
					SpecificationBrand:    ptr("Brand"),
					SpecificationType:     ptr("Type"),
				}
				specifications = append(specifications, spec)

				fieldDevices = append(fieldDevices, facility.FieldDevice{
					Base: domain.Base{
						ID:        fdID,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					BMK:                       &bmk,
					Description:               &desc,
					ApparatNr:                 rand.Intn(99) + 1,
					SPSControllerSystemTypeID: spsType.ID,
					SystemPartID:              spID,
					SpecificationID:           &spec.ID,
					ApparatID:                 apparat.ID,
				})

				for k := 0; k < BacnetObjectsPerFieldDevice; k++ {
					boID := uuid.New()
					textFix := fmt.Sprintf("FD%03d-%02d-%s", i+1, k+1, boID.String()[:6])
					descBo := fmt.Sprintf("Bacnet object FD %s", bmk)

					var stID *uuid.UUID
					if len(stateTexts) > 0 {
						id := stateTexts[rand.Intn(len(stateTexts))].ID
						stID = &id
					}
					var ncID *uuid.UUID
					if len(notificationClasses) > 0 {
						id := notificationClasses[rand.Intn(len(notificationClasses))].ID
						ncID = &id
					}
					var adID *uuid.UUID
					if len(alarmDefinitions) > 0 {
						id := alarmDefinitions[rand.Intn(len(alarmDefinitions))].ID
						adID = &id
					}

					fieldDeviceBos = append(fieldDeviceBos, facility.BacnetObject{
						Base: domain.Base{
							ID:        boID,
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						TextFix:             textFix,
						Description:         &descBo,
						GMSVisible:          rand.Intn(2) == 0,
						Optional:            rand.Intn(3) == 0,
						SoftwareType:        softwareTypes[rand.Intn(len(softwareTypes))],
						SoftwareNumber:      uint16(rand.Intn(65000) + 1),
						HardwareType:        facility.BacnetHardwareType(""),
						HardwareQuantity:    0,
						FieldDeviceID:       &fdID,
						SoftwareReferenceID: nil,
						StateTextID:         stID,
						NotificationClassID: ncID,
						AlarmDefinitionID:   adID,
					})
				}
			}
		}
	}

	for i := 0; i < NumObjectDatas; i++ {
		odID := uuid.New()
		objectDatas = append(objectDatas, facility.ObjectData{
			Base: domain.Base{
				ID:        odID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Description: fmt.Sprintf("ObjectData-%03d", i+1),
			Version:     fmt.Sprintf("%d.%d", 1+(i/50), (i%50)+1),
			IsActive:    true,
			ProjectID:   nil,
		})

		// Link ObjectData <-> Apparats
		if len(apparats) > 0 {
			n := MinApparatsPerObjectData
			if MaxApparatsPerObjectData > MinApparatsPerObjectData {
				n += rand.Intn(MaxApparatsPerObjectData - MinApparatsPerObjectData + 1)
			}
			seen := make(map[uuid.UUID]struct{}, n)
			for j := 0; j < n; j++ {
				ap := apparats[rand.Intn(len(apparats))]
				if _, ok := seen[ap.ID]; ok {
					j--
					continue
				}
				seen[ap.ID] = struct{}{}
				objDataApparatLinks = append(objDataApparatLinks, objectDataApparat{ObjectDataID: odID, ApparatID: ap.ID})
			}
		}

		for j := 0; j < BacnetObjectsPerObjData; j++ {
			boID := uuid.New()
			textFix := fmt.Sprintf("OD%03d-%02d-%s", i+1, j+1, boID.String()[:8])
			desc := fmt.Sprintf("Bacnet object %d/%d for OD %d", j+1, BacnetObjectsPerObjData, i+1)

			bo := facility.BacnetObject{
				Base: domain.Base{
					ID:        boID,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				TextFix:             textFix,
				Description:         &desc,
				GMSVisible:          rand.Intn(2) == 0,
				Optional:            rand.Intn(3) == 0,
				TextIndividual:      nil,
				SoftwareType:        softwareTypes[rand.Intn(len(softwareTypes))],
				SoftwareNumber:      uint16(rand.Intn(65000) + 1),
				HardwareType:        facility.BacnetHardwareType(""),
				HardwareQuantity:    0,
				FieldDeviceID:       nil,
				SoftwareReferenceID: nil,
				StateTextID:         nil,
				NotificationClassID: nil,
				AlarmDefinitionID:   nil,
			}

			bacnetObjects = append(bacnetObjects, bo)
			objDataLinks = append(objDataLinks, objectDataBacnetObject{ObjectDataID: odID, BacnetObjectID: boID})
		}
	}

	log.Printf("Generation took %v", time.Since(start))

	log.Printf("Inserting %d SystemTypes...", len(systemTypes))
	if err := database.CreateInBatches(systemTypes, 1000).Error; err != nil {
		log.Fatalf("Error inserting system types: %v", err)
	}

	log.Printf("Inserting %d SystemParts...", len(systemParts))
	if err := database.CreateInBatches(systemParts, 1000).Error; err != nil {
		log.Fatalf("Error inserting system parts: %v", err)
	}

	log.Printf("Inserting %d Apparats...", len(apparats))
	if err := database.CreateInBatches(apparats, 1000).Error; err != nil {
		log.Fatalf("Error inserting apparats: %v", err)
	}

	log.Printf("Inserting %d Buildings...", len(buildings))
	if err := database.CreateInBatches(buildings, 1000).Error; err != nil {
		log.Fatalf("Error inserting buildings: %v", err)
	}

	log.Printf("Inserting %d Control Cabinets...", len(controlCabinets))
	if err := database.CreateInBatches(controlCabinets, 1000).Error; err != nil {
		log.Fatalf("Error inserting control cabinets: %v", err)
	}

	log.Printf("Inserting %d SPS Controllers...", len(spsControllers))
	if err := database.CreateInBatches(spsControllers, 1000).Error; err != nil {
		log.Fatalf("Error inserting sps controllers: %v", err)
	}

	log.Printf("Inserting %d SPS Controller System Types...", len(spsSysTypes))
	if err := database.CreateInBatches(spsSysTypes, 1000).Error; err != nil {
		log.Fatalf("Error inserting sps controller system types: %v", err)
	}

	log.Printf("Inserting %d State Texts...", len(stateTexts))
	if err := database.CreateInBatches(stateTexts, 1000).Error; err != nil {
		log.Fatalf("Error inserting state texts: %v", err)
	}

	log.Printf("Inserting %d Notification Classes...", len(notificationClasses))
	if err := database.CreateInBatches(notificationClasses, 1000).Error; err != nil {
		log.Fatalf("Error inserting notification classes: %v", err)
	}

	log.Printf("Inserting %d Alarm Definitions...", len(alarmDefinitions))
	if err := database.CreateInBatches(alarmDefinitions, 1000).Error; err != nil {
		log.Fatalf("Error inserting alarm definitions: %v", err)
	}

	log.Printf("Inserting %d Specifications...", len(specifications))
	if err := database.CreateInBatches(specifications, 1000).Error; err != nil {
		log.Fatalf("Error inserting specifications: %v", err)
	}

	log.Printf("Inserting %d Field Devices...", len(fieldDevices))
	if err := database.CreateInBatches(fieldDevices, 1000).Error; err != nil {
		log.Fatalf("Error inserting field devices: %v", err)
	}

	allBos := append(bacnetObjects, fieldDeviceBos...)
	log.Printf("Inserting %d Bacnet Objects...", len(allBos))
	if err := database.CreateInBatches(allBos, 1000).Error; err != nil {
		log.Fatalf("Error inserting bacnet objects: %v", err)
	}

	log.Printf("Inserting %d Object Datas...", len(objectDatas))
	if err := database.CreateInBatches(objectDatas, 1000).Error; err != nil {
		log.Fatalf("Error inserting object datas: %v", err)
	}

	log.Printf("Linking %d SystemPart <-> Apparat rows...", len(sysPartApparatLinks))
	if err := database.CreateInBatches(sysPartApparatLinks, 2000).Error; err != nil {
		log.Fatalf("Error inserting system_part_apparats links: %v", err)
	}

	log.Printf("Linking %d ObjectData <-> Apparat rows...", len(objDataApparatLinks))
	if err := database.CreateInBatches(objDataApparatLinks, 2000).Error; err != nil {
		log.Fatalf("Error inserting object_data_apparats links: %v", err)
	}

	log.Printf("Linking %d ObjectData <-> BacnetObject rows...", len(objDataLinks))
	if err := database.CreateInBatches(objDataLinks, 2000).Error; err != nil {
		log.Fatalf("Error inserting object_data_bacnet_objects links: %v", err)
	}

	log.Println("Seeding complete!")
}

func ptr(s string) *string {
	return &s
}

func gaDeviceCode(index int) string {
	// Base-26 encoding to 3 uppercase letters.
	first := rune('A' + (index % 26))
	second := rune('A' + ((index / 26) % 26))
	third := rune('A' + ((index / (26 * 26)) % 26))
	return string([]rune{third, second, first})
}
