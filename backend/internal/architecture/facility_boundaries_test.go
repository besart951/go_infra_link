package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var cleanDomainDirectories = []string{
	"internal/application/facilityjobs",
	"internal/application/hierarchydelete",
	"internal/application/hierarchyrestore",
	"internal/application/fielddeviceimport",
	"internal/domain/facility/fielddevice",
	"internal/domain/facility/hierarchy",
	"internal/domain/facility/objectdata",
}

var focusedCapabilityFiles = []string{
	"internal/application/facilityjobs/steps.go",
	"internal/service/facility/field_device_multi_create.go",
	"internal/service/facility/field_device_bulk_update.go",
	"internal/service/facility/field_device_specification_writer.go",
	"internal/wire/field_device_bulk_jobs.go",
	"internal/wire/facility_delete_jobs.go",
	"internal/repository/facilitysql/hierarchy_delete.go",
	"internal/repository/facilitysql/lifecycle_mutation.go",
	"internal/repository/facilitysql/lifecycle_visibility.go",
	"internal/repository/historysql/restore_chunks.go",
	"internal/wire/history_restore_jobs.go",
	"internal/application/fielddeviceimport/service.go",
	"internal/infrastructure/importing/excelize_reader.go",
	"internal/infrastructure/importing/archive_reader.go",
	"internal/repository/importsql/staging.go",
	"internal/service/facility/field_device_import.go",
	"internal/wire/field_device_import.go",
	"internal/handler/facility/import.go",
}

func TestFacilityDomainCapabilitiesRemainPersistenceAgnostic(t *testing.T) {
	root := backendRoot(t)
	for _, relative := range cleanDomainDirectories {
		files, err := filepath.Glob(filepath.Join(root, relative, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			assertPersistenceAgnostic(t, file)
		}
	}
}

func TestNewFacilityCapabilitiesKeepFocusedFunctions(t *testing.T) {
	root := backendRoot(t)
	for _, relative := range focusedCapabilityFiles {
		assertFocusedFunctions(t, filepath.Join(root, relative))
	}
}

func assertPersistenceAgnostic(t *testing.T, path string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(contents)
	if strings.Contains(text, "gorm.io/gorm") || strings.Contains(text, "gorm:\"") {
		t.Errorf("domain capability contains GORM coupling: %s", path)
	}
	if strings.Contains(text, "Repository[") {
		t.Errorf("domain capability extends generic repository: %s", path)
	}
}

func assertFocusedFunctions(t *testing.T, path string) {
	t.Helper()
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		if countParameters(function.Type.Params) > 3 {
			t.Errorf("%s has more than three parameters: %s", path, function.Name.Name)
		}
		lines := set.Position(function.End()).Line - set.Position(function.Pos()).Line + 1
		if lines > 35 {
			t.Errorf("%s has %d lines: %s", path, lines, function.Name.Name)
		}
	}
}

func countParameters(fields *ast.FieldList) int {
	if fields == nil {
		return 0
	}
	count := 0
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			count++
		} else {
			count += len(field.Names)
		}
	}
	return count
}

func backendRoot(t *testing.T) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(current), "..", ".."))
}
