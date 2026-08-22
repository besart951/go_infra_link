package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	facilitybenchmark "github.com/besart951/go_infra_link/backend/internal/benchmark/facility"
)

func main() {
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "benchmark PostgreSQL DSN")
	output := flag.String("output", "benchmark-results", "report directory")
	fieldDevices := flag.Int64("field-devices", facilitybenchmark.FieldDeviceCount, "deterministic field-device rows (50000..5000000)")
	smoke := flag.Bool("smoke", false, "run reduced query-shape sample profile")
	scenario := flag.String("scenario", "", "measure one exact scenario name")
	flag.Parse()
	if flag.NArg() != 1 || *dsn == "" {
		fail("usage: facility-benchmark -dsn <url> [seed|analyze|measure|report]")
	}
	options := commandOptions{dsn: *dsn, output: *output, fieldDevices: *fieldDevices, smoke: *smoke, scenario: *scenario}
	if err := run(context.Background(), options, flag.Arg(0)); err != nil {
		fail(err.Error())
	}
}

type commandOptions struct {
	dsn          string
	output       string
	fieldDevices int64
	smoke        bool
	scenario     string
}

func run(ctx context.Context, options commandOptions, command string) error {
	if command == "report" {
		return renderExistingReport(options.output)
	}
	database, err := facilitybenchmark.Open(ctx, options.dsn)
	if err != nil {
		return err
	}
	defer database.Close(ctx)
	if err := database.SetFieldDeviceCount(options.fieldDevices); err != nil {
		return err
	}
	if options.smoke {
		database.Samples = facilitybenchmark.SmokeSampleProfile()
	}
	switch command {
	case "seed":
		return database.Seed(ctx)
	case "analyze":
		return database.Analyze(ctx)
	case "measure":
		return measure(ctx, database, options)
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func measure(ctx context.Context, database *facilitybenchmark.Database, options commandOptions) error {
	report, err := database.Measure(ctx, facilitybenchmark.MeasureOptions{Scenario: options.scenario})
	if err != nil {
		return err
	}
	if err := facilitybenchmark.WriteReport(options.output, report); err != nil {
		return err
	}
	if !facilitybenchmark.Passed(report) {
		return fmt.Errorf("benchmark gate failed; inspect %s", options.output)
	}
	return nil
}

func renderExistingReport(output string) error {
	report, err := facilitybenchmark.ReadReport(filepath.Join(output, "facility-benchmark.json"))
	if err != nil {
		return err
	}
	return facilitybenchmark.WriteReport(output, report)
}

func fail(message string) {
	_, _ = fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
