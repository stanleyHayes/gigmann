// Package architecture holds tests that enforce structural rules.
package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestHexagonalBoundaries enforces that the domain core and application layer
// remain free of framework/infrastructure imports (ports & adapters rule).
func TestHexagonalBoundaries(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller path")
	}
	// backend/ root is two levels up from internal/architecture/.
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))

	rules := map[string][]string{
		"internal/core": {
			"/internal/adapters",
			"github.com/go-chi",
			"github.com/jackc/pgx",
			"github.com/redis",
		},
		"internal/app": {
			"/internal/adapters",
			"github.com/go-chi",
		},
	}

	for rel, forbidden := range rules {
		dir := filepath.Join(root, rel)
		walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			fset := token.NewFileSet()
			f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if perr != nil {
				return perr
			}
			for _, imp := range f.Imports {
				p := strings.Trim(imp.Path.Value, `"`)
				for _, bad := range forbidden {
					if strings.Contains(p, bad) {
						t.Errorf("hexagonal violation: %s imports %q (%s)", rel, p, filepath.Base(path))
					}
				}
			}
			return nil
		})
		if walkErr != nil {
			t.Fatalf("walk %s: %v", rel, walkErr)
		}
	}
}
