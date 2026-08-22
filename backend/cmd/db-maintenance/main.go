package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/config"
	"github.com/besart951/go_infra_link/backend/internal/db"
)

func main() {
	backup := flag.Bool("backup-verified", false, "confirm a verified PostgreSQL backup")
	release := flag.Bool("compatible-release-delivered", false, "confirm a compatible release was delivered")
	stopped := flag.Bool("applications-stopped", false, "confirm all API and worker instances are stopped")
	legacyIdle := flag.String("legacy-idle-since", "", "RFC3339 timestamp of the last legacy request")
	flag.Parse()

	idleSince, err := time.Parse(time.RFC3339, *legacyIdle)
	if err != nil {
		exit("legacy-idle-since must be RFC3339", err)
	}
	cfg, err := config.Load()
	if err != nil {
		exit("config load", err)
	}
	database, err := db.Connect(cfg.DBConfig)
	if err != nil {
		exit("database connect", err)
	}
	sqlDB, _ := database.DB()
	if sqlDB != nil {
		defer sqlDB.Close()
	}
	err = db.ApplyFacilityContractMigration(database, db.FacilityContractOptions{
		CompatibleReleaseDelivered: *release,
		ApplicationsStopped:         *stopped,
		LegacyIdleSince:            idleSince,
		BackupVerified:             *backup,
	})
	if err != nil {
		exit("facility contract migration", err)
	}
}

func exit(message string, err error) {
	_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", message, err)
	os.Exit(1)
}
