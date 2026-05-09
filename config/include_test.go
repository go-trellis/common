/*
Copyright © 2017 Henry Huang <hhh@rutcode.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-trellis/common.v3/config"
	"github.com/go-trellis/common.v3/utils/testutils"
)

func TestInclude_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create included config file
	includedFile := filepath.Join(tmpDir, "included.yml")
	includedContent := `
app:
  name: "included-app"
  version: "1.0.0"
database:
  host: "localhost"
  port: 5432
`
	err := os.WriteFile(includedFile, []byte(includedContent), 0644)
	testutils.Ok(t, err)

	// Create main config file with include
	// Note: In YAML, "#include" must be quoted because # starts a comment
	mainFile := filepath.Join(tmpDir, "main.yml")
	mainContent := `"#include": "included.yml"
app:
  name: "main-app"
database:
  port: 3306
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	testutils.Ok(t, err)

	// Load config
	cfg, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)
	testutils.Assert(t, cfg != nil, "config should not be nil")

	// Verify include was processed and merged
	// Included values should override main values
	appName := cfg.GetString("app.name")
	testutils.Equals(t, "included-app", appName, "app name should be from included file")

	appVersion := cfg.GetString("app.version")
	testutils.Equals(t, "1.0.0", appVersion, "app version should be from included file")

	dbHost := cfg.GetString("database.host")
	testutils.Equals(t, "localhost", dbHost, "database host should be from included file")

	dbPort := cfg.GetInt("database.port")
	testutils.Equals(t, 5432, dbPort, "database port should be from included file (included overrides main)")

	// Verify include field was removed
	includeValue := cfg.GetInterface("#include")
	testutils.Assert(t, includeValue == nil, "#include field should be removed from config")

	// Also verify legacy include field
	legacyIncludeValue := cfg.GetInterface("include")
	testutils.Assert(t, legacyIncludeValue == nil, "include field should be removed from config")
}

func TestInclude_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first included file
	file1 := filepath.Join(tmpDir, "file1.yml")
	err := os.WriteFile(file1, []byte(`
app:
  name: "file1-app"
database:
  host: "file1-host"
`), 0644)
	testutils.Ok(t, err)

	// Create second included file
	file2 := filepath.Join(tmpDir, "file2.yml")
	err = os.WriteFile(file2, []byte(`
app:
  version: "2.0.0"
database:
  port: 6379
`), 0644)
	testutils.Ok(t, err)

	// Create main config file with multiple includes
	mainFile := filepath.Join(tmpDir, "main.yml")
	mainContent := `"#include":
  - "file1.yml"
  - "file2.yml"
app:
  name: "main-app"
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	testutils.Ok(t, err)

	// Load config
	cfg, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)

	// Later includes override earlier ones, file2 values should win
	appName := cfg.GetString("app.name")
	testutils.Equals(t, "file1-app", appName, "app name should be from file1")

	appVersion := cfg.GetString("app.version")
	testutils.Equals(t, "2.0.0", appVersion, "app version should be from file2")

	dbHost := cfg.GetString("database.host")
	testutils.Equals(t, "file1-host", dbHost, "database host should be from file1")

	dbPort := cfg.GetInt("database.port")
	testutils.Equals(t, 6379, dbPort, "database port should be from file2")
}

func TestInclude_Recursive(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested included file
	nestedFile := filepath.Join(tmpDir, "nested.yml")
	err := os.WriteFile(nestedFile, []byte(`
app:
  nested: true
database:
  nested: true
`), 0644)
	testutils.Ok(t, err)

	// Create included file that includes nested
	includedFile := filepath.Join(tmpDir, "included.yml")
	err = os.WriteFile(includedFile, []byte(`"#include": "nested.yml"
app:
  name: "included-app"
database:
  host: "included-host"
`), 0644)
	testutils.Ok(t, err)

	// Create main config
	mainFile := filepath.Join(tmpDir, "main.yml")
	mainContent := `"#include": "included.yml"
app:
  main: true
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	testutils.Ok(t, err)

	// Load config
	cfg, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)

	// Verify all nested includes were processed
	testutils.Equals(t, true, cfg.GetBoolean("app.nested"), "app.nested should be true from nested file")
	testutils.Equals(t, true, cfg.GetBoolean("database.nested"), "database.nested should be true from nested file")
	testutils.Equals(t, "included-app", cfg.GetString("app.name"), "app.name should be from included file")
	testutils.Equals(t, "included-host", cfg.GetString("database.host"), "database.host should be from included file")
	testutils.Equals(t, true, cfg.GetBoolean("app.main"), "app.main should be true from main file")
}

func TestInclude_CircularReference(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file1 that includes file2
	file1 := filepath.Join(tmpDir, "file1.yml")
	err := os.WriteFile(file1, []byte(`"#include": "file2.yml"
app:
  name: "file1"
`), 0644)
	testutils.Ok(t, err)

	// Create file2 that includes file1 (circular reference)
	file2 := filepath.Join(tmpDir, "file2.yml")
	err = os.WriteFile(file2, []byte(`"#include": "file1.yml"
app:
  name: "file2"
`), 0644)
	testutils.Ok(t, err)

	// Try to load file1 - should detect circular reference
	cfg, err := config.NewConfig(file1)
	testutils.NotOk(t, err, "should return error for circular reference")
	testutils.Assert(t, cfg == nil, "config should be nil on error")
}

func TestInclude_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Create main config with non-existent include
	mainFile := filepath.Join(tmpDir, "main.yml")
	mainContent := `"#include": "nonexistent.yml"
app:
  name: "main"
`
	err := os.WriteFile(mainFile, []byte(mainContent), 0644)
	testutils.Ok(t, err)

	// Try to load - should return error
	cfg, err := config.NewConfig(mainFile)
	testutils.NotOk(t, err, "should return error for nonexistent include file")
	testutils.Assert(t, cfg == nil, "config should be nil on error")
}

func TestInclude_RelativePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "sub")
	err := os.Mkdir(subDir, 0755)
	testutils.Ok(t, err)

	// Create included file in subdirectory
	includedFile := filepath.Join(subDir, "included.yml")
	err = os.WriteFile(includedFile, []byte(`
app:
  name: "included-app"
`), 0644)
	testutils.Ok(t, err)

	// Create main config in root with relative path to include
	mainFile := filepath.Join(tmpDir, "main.yml")
	mainContent := `"#include": "sub/included.yml"
app:
  version: "1.0.0"
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	testutils.Ok(t, err)

	// Load config
	cfg, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)

	testutils.Equals(t, "included-app", cfg.GetString("app.name"), "app name should be from included file")
	testutils.Equals(t, "1.0.0", cfg.GetString("app.version"), "app version should be from main file")
}

func TestInclude_JSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create included JSON file
	includedFile := filepath.Join(tmpDir, "included.json")
	includedContent := `{
  "app": {
    "name": "included-app",
    "version": "1.0.0"
  }
}`
	err := os.WriteFile(includedFile, []byte(includedContent), 0644)
	testutils.Ok(t, err)

	// Create main JSON config with include
	mainFile := filepath.Join(tmpDir, "main.json")
	mainContent := `{
  "#include": "included.json",
  "app": {
    "name": "main-app"
  }
}`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	testutils.Ok(t, err)

	// Load config
	cfg, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)

	testutils.Equals(t, "included-app", cfg.GetString("app.name"), "app name should be from included file")
	testutils.Equals(t, "1.0.0", cfg.GetString("app.version"), "app version should be from included file")
}

func TestInclude_DeepMerge(t *testing.T) {
	tmpDir := t.TempDir()

	// Create included file with nested structure
	includedFile := filepath.Join(tmpDir, "included.yml")
	err := os.WriteFile(includedFile, []byte(`
app:
  name: "included-app"
  database:
    host: "included-host"
    port: 5432
`), 0644)
	testutils.Ok(t, err)

	// Create main config with partial nested structure
	mainFile := filepath.Join(tmpDir, "main.yml")
	mainContent := `"#include": "included.yml"
app:
  name: "main-app"
  database:
    port: 3306
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	testutils.Ok(t, err)

	// Load config
	cfg, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)

	// Deep merge: included values should override, but merge nested structures
	testutils.Equals(t, "included-app", cfg.GetString("app.name"), "app name should be from included file")
	testutils.Equals(t, "included-host", cfg.GetString("app.database.host"), "database host should be from included file")
	testutils.Equals(t, 5432, cfg.GetInt("app.database.port"), "database port should be from included file (included overrides)")
}

func TestInclude_InvalidValue(t *testing.T) {
	tmpDir := t.TempDir()

	// Create main config with invalid include value (number instead of string/array)
	mainFile := filepath.Join(tmpDir, "main.yml")
	mainContent := `"#include": 123
app:
  name: "main"
`
	err := os.WriteFile(mainFile, []byte(mainContent), 0644)
	testutils.Ok(t, err)

	// Try to load - should return error
	cfg, err := config.NewConfig(mainFile)
	testutils.NotOk(t, err, "should return error for invalid include value")
	testutils.Assert(t, cfg == nil, "config should be nil on error")
}

func TestInclude_BackwardCompatibility(t *testing.T) {
	tmpDir := t.TempDir()

	// Create included config file
	includedFile := filepath.Join(tmpDir, "included.yml")
	includedContent := `
app:
  name: "included-app"
  version: "1.0.0"
`
	err := os.WriteFile(includedFile, []byte(includedContent), 0644)
	testutils.Ok(t, err)

	// Create main config file with legacy include (without #)
	mainFile := filepath.Join(tmpDir, "main.yml")
	mainContent := `
include: "included.yml"
app:
  name: "main-app"
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	testutils.Ok(t, err)

	// Load config - should work with legacy include keyword
	cfg, err := config.NewConfig(mainFile)
	testutils.Ok(t, err)
	testutils.Assert(t, cfg != nil, "config should not be nil")

	// Verify include was processed (included values override main values)
	appName := cfg.GetString("app.name")
	testutils.Equals(t, "included-app", appName, "app name should be from included file")

	appVersion := cfg.GetString("app.version")
	testutils.Equals(t, "1.0.0", appVersion, "app version should be from included file")

	// Verify legacy include field was removed
	includeValue := cfg.GetInterface("include")
	testutils.Assert(t, includeValue == nil, "include field should be removed from config")
}
