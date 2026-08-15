package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
)

func main() {
	output := filepath.FromSlash("../frontend/src/lib/api/generated/permissions.ts")
	if len(os.Args) == 2 {
		output = os.Args[1]
	}

	permissions := make([]string, 0)
	for _, definition := range domainUser.CanonicalPermissionDefinitions() {
		permissions = append(permissions, definition.Name)
	}
	sort.Strings(permissions)

	var content strings.Builder
	content.WriteString("/* eslint-disable */\n")
	content.WriteString("// This file is generated from backend/internal/domain/user/permission_catalog.go. Do not edit manually.\n\n")
	content.WriteString("export const PERMISSIONS = [\n")
	for _, permission := range permissions {
		fmt.Fprintf(&content, "  '%s',\n", permission)
	}
	content.WriteString("] as const;\n\n")
	content.WriteString("export type PermissionName = (typeof PERMISSIONS)[number];\n")
	content.WriteString("export type ProjectPermissionName = Exclude<Extract<PermissionName, `project.${string}`>, 'project.create' | 'project.listAll'>;\n")

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(output, []byte(content.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
