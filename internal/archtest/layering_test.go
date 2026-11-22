// Package archtest_test enforces the layering rule: services, controllers,
// DTOs and the domain never depend on persistence code. Repositories are the
// only place that may know about database/sql, Bun or any ORM, which is what
// lets the persistence library be replaced by rewriting repositories alone.
package archtest_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const module = "github.com/letenk/golang-authentication"

// repoRoot is relative to this package directory (internal/archtest).
const repoRoot = "../.."

// ormAndPersistence lists import paths (exact or prefix) that only
// repositories and configs/database may use.
var ormAndPersistence = []string{
	"database/sql",
	module + "/configs/database",
	"github.com/stephenafamo/bob",
	"github.com/aarondl/opt",
	"gorm.io/",
	"github.com/jackc/pgx",
	"github.com/uptrace/bun",
	"github.com/jmoiron/sqlx",
}

// findViolations parses every non-test .go file directly in dir and returns
// "file: import" for each import equal to, or prefixed by, a forbidden entry.
func findViolations(dir string, forbidden []string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var violations []string
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil, err
		}

		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return nil, err
			}
			for _, f := range forbidden {
				if importPath == f || strings.HasPrefix(importPath, strings.TrimSuffix(f, "/")+"/") {
					violations = append(violations, path+": "+importPath)
					break
				}
			}
		}
	}
	return violations, nil
}

// findPersistenceTags parses every non-test .go file directly in dir and
// returns "file: Struct.Field tag key" for each struct field whose tag uses one
// of keys. Struct tags need no import, so findViolations cannot see them.
func findPersistenceTags(dir string, keys []string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var found []string
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, err
		}

		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts := spec.(*ast.TypeSpec)
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range st.Fields.List {
					if field.Tag == nil {
						continue
					}
					raw, err := strconv.Unquote(field.Tag.Value)
					if err != nil {
						return nil, err
					}
					tag := reflect.StructTag(raw)
					for _, key := range keys {
						if _, ok := tag.Lookup(key); ok {
							found = append(found, path+": "+ts.Name.Name+"."+fieldName(field)+" tag "+key)
						}
					}
				}
			}
		}
	}
	return found, nil
}

func fieldName(f *ast.Field) string {
	if len(f.Names) > 0 {
		return f.Names[0].Name
	}
	return "(embedded)"
}

// globDirs expands patterns relative to the repo root and fails the test when
// nothing matches, so moving a folder cannot silently disable the check.
func globDirs(t *testing.T, patterns ...string) []string {
	t.Helper()
	var dirs []string
	for _, p := range patterns {
		matches, err := filepath.Glob(filepath.Join(repoRoot, p))
		require.NoError(t, err)
		require.NotEmpty(t, matches, "pattern %q matched no directory; update archtest if the layout changed", p)
		dirs = append(dirs, matches...)
	}
	return dirs
}

func TestFindViolations_DetectsLeakyFixture(t *testing.T) {
	v, err := findViolations("testdata/leaky", ormAndPersistence)

	require.NoError(t, err)
	assert.Len(t, v, 2, "database/sql and bun; fmt is allowed: %v", v)
}

func TestLayering_ServiceControllerDTOHaveNoPersistenceImports(t *testing.T) {
	dirs := globDirs(t,
		"internal/applications/*/service",
		"internal/applications/*/controller",
		"internal/applications/*/dto",
	)

	for _, dir := range dirs {
		v, err := findViolations(dir, ormAndPersistence)
		require.NoError(t, err)
		assert.Empty(t, v, "%s must depend on repository interfaces and internal/domain only", dir)
	}
}

func TestLayering_DomainImportsOnlyStdlib(t *testing.T) {
	forbidden := append([]string{module + "/"}, ormAndPersistence...)

	for _, dir := range globDirs(t, "internal/domain") {
		v, err := findViolations(dir, forbidden)
		require.NoError(t, err)
		assert.Empty(t, v, "internal/domain must import only the standard library")
	}
}

// persistenceTagKeys are struct tag keys that bind a struct to an ORM or SQL
// mapper. Domain structs must not carry them: the repository owns its own row
// structs and maps them to the domain.
var persistenceTagKeys = []string{"bun", "db", "gorm"}

func TestFindPersistenceTags_DetectsTaggedFixture(t *testing.T) {
	v, err := findPersistenceTags("testdata/taggeddomain", persistenceTagKeys)

	require.NoError(t, err)
	assert.Len(t, v, 2, "bun and db tags; json-only and untagged fields are allowed: %v", v)
}

func TestLayering_DomainHasNoPersistenceTags(t *testing.T) {
	for _, dir := range globDirs(t, "internal/domain") {
		v, err := findPersistenceTags(dir, persistenceTagKeys)
		require.NoError(t, err)
		assert.Empty(t, v, "internal/domain structs must not carry persistence tags; map from a repository row struct instead")
	}
}
