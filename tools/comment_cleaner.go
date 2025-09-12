//go:build tools
// +build tools

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func isKeptCommentText(text string) bool {
	t := strings.TrimSpace(text)
	if strings.HasPrefix(t, "// @") { // Swagger annotations
		return true
	}
	if strings.HasPrefix(t, "//go:build") || strings.HasPrefix(t, "// +build") { // build tags
		return true
	}
	// Preserve security and lint directives
	if strings.Contains(t, "#nosec") || strings.Contains(t, "nolint") {
		return true
	}
	return false
}

func keepGroup(g *ast.CommentGroup) bool {
	if g == nil {
		return false
	}
	for _, c := range g.List {
		if isKeptCommentText(c.Text) {
			return true
		}
	}
	return false
}

func cleanGoFile(root, path string) error {
	lower := strings.ToLower(path)
	if strings.Contains(lower, string(os.PathSeparator)+"vendor"+string(os.PathSeparator)) {
		return nil
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("abs root: %w", err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("abs path: %w", err)
	}
	if absPath != absRoot && !strings.HasPrefix(absPath, absRoot+string(os.PathSeparator)) {
		return fmt.Errorf("ruta fuera de la raíz: %s", absPath)
	}
	// Skip symlinks
	if fi, err := os.Lstat(absPath); err == nil && (fi.Mode()&os.ModeSymlink) != 0 {
		return nil
	}

	fset := token.NewFileSet()
	// #nosec G304 -- path validado contra root, no symlinks, y extensión .go controlada
	src, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("leer %s: %w", path, err)
	}
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsear %s: %w", path, err)
	}
	kept := map[*ast.CommentGroup]bool{}
	for _, cg := range file.Comments {
		if keepGroup(cg) {
			kept[cg] = true
		}
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Doc != nil && !kept[x.Doc] {
				x.Doc = nil
			}
		case *ast.FuncDecl:
			if x.Doc != nil && !kept[x.Doc] {
				x.Doc = nil
			}
		case *ast.TypeSpec:
			if x.Doc != nil && !kept[x.Doc] {
				x.Doc = nil
			}
		case *ast.Field:
			if x.Doc != nil && !kept[x.Doc] {
				x.Doc = nil
			}
			if x.Comment != nil && !kept[x.Comment] {
				x.Comment = nil
			}
		}
		return true
	})
	filtered := make([]*ast.CommentGroup, 0, len(file.Comments))
	for _, cg := range file.Comments {
		if kept[cg] {
			filtered = append(filtered, cg)
		}
	}
	file.Comments = filtered
	var b strings.Builder
	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	if err := cfg.Fprint(&b, fset, file); err != nil {
		return fmt.Errorf("imprimir %s: %w", path, err)
	}
	if err := os.WriteFile(absPath, []byte(b.String()), 0o600); err != nil {
		return fmt.Errorf("escribir %s: %w", path, err)
	}
	return nil
}

func main() {
	var root string
	flag.StringVar(&root, "root", ".", "Directorio raíz del repositorio")
	flag.Parse()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "tmp" || name == "swagger" || name == "cover" || name == "logging_cov" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		return cleanGoFile(root, path)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
