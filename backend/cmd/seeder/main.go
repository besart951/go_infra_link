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
	NumBuildings        = 1000
	CabinetsPerBuilding = 10
	SPSPerCabinet       = 10
)

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

	var (
		buildings       []facility.Building
		controlCabinets []facility.ControlCabinet
		spsControllers  []facility.SPSController
	)

	// Generate data in memory
	log.Println("Generating data structures...")

	start := time.Now()

	for i := 0; i < NumBuildings; i++ {
		bID := uuid.New()
		iwsCode := fmt.Sprintf("%04d", rand.Intn(10000)) // Random 4 digits

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
				// GA Device <= 10 chars. "GA-%05d" = 3+5 = 8 chars.
				gaDevice := fmt.Sprintf("GA-%05d", rand.Intn(99999))

				spsControllers = append(spsControllers, facility.SPSController{
					Base: domain.Base{
						ID:        sID,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					ControlCabinetID: cID,
					GADevice:         ptr(gaDevice),
					DeviceName:       fmt.Sprintf("SPS-%s-%d", cabNr, k),
					IPAddress:        ptr(fmt.Sprintf("192.168.%d.%d", rand.Intn(255), rand.Intn(255))),
				})
			}
		}
	}

	log.Printf("Generation took %v", time.Since(start))
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

	log.Println("Seeding complete!")
}

func ptr(s string) *string {
	return &s
}
