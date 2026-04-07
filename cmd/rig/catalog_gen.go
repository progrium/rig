package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
	"tractor.dev/toolkit-go/engine/cli"
)

func genCatalogCmd() *cli.Command {
	return &cli.Command{
		Usage: "gen <dir>",
		Short: "Generate catalog for project in DIR",
		Args:  cli.ExactArgs(1),
		Run:   genCatalog,
	}
}

func genCatalog(ctx *cli.Context, args []string) {
	dir := args[0]

	// Find all import paths recursively
	imports := findImports(dir)
	imports = append(imports, ".")

	cfg := &packages.Config{
		Mode:  packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedName | packages.NeedModule, //packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
		Fset:  token.NewFileSet(),
		Dir:   dir,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg, imports...)
	if err != nil {
		log.Fatalf("Failed to load packages: %v", err)
	}

	structs := findStructs(pkgs)

	err = generateSourceFile(structs, findFactories(pkgs), filepath.Join(dir, "meta.gen.go"))
	if err != nil {
		log.Fatalf("Failed to generate source file: %v", err)
	}
}

// StructInfo holds information about a struct with its package path
type StructInfo struct {
	PkgPath string
	PkgName string
	Name    string
	Typed   bool // todo: better handling of type params
}

// findImports finds all imported packages recursively from the main package
func findImports(dir string) []string {
	mainFile := filepath.Join(dir, "main.go")
	fset := token.NewFileSet()

	// Parse the main.go file to find imports
	node, err := parser.ParseFile(fset, mainFile, nil, parser.ImportsOnly)
	if err != nil {
		log.Fatalf("Failed to parse main.go: %v", err)
	}

	var imports []string
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, importPath)
	}

	// Recursively find imports of imported packages
	// return findAllImports(imports)
	return imports
}

// findAllImports recursively finds all imported packages
func findAllImports(imports []string) []string {
	var allImports = make(map[string]struct{})
	var toVisit = imports

	for len(toVisit) > 0 {
		pkgPath := toVisit[0]
		toVisit = toVisit[1:]

		if _, visited := allImports[pkgPath]; visited {
			continue
		}
		allImports[pkgPath] = struct{}{}

		cfg := &packages.Config{Mode: packages.NeedImports}
		pkgs, err := packages.Load(cfg, pkgPath)
		if err != nil {
			log.Printf("Failed to load package: %v", err)
			continue
		}

		for _, pkg := range pkgs {
			for imp := range pkg.Imports {
				if _, visited := allImports[imp]; !visited {
					toVisit = append(toVisit, imp)
				}
			}
		}
	}

	// Convert the map to a slice of package paths
	var result []string
	for imp := range allImports {
		result = append(result, imp)
	}
	return result
}

func findFactories(pkgs []*packages.Package) (structs []StructInfo) {

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				// Find function declarations
				if fn, ok := n.(*ast.FuncDecl); ok {
					// Check if the function is exported and not a method
					if fn.Name.IsExported() && fn.Recv == nil {
						// Ensure Params and Results are not nil before accessing them
						params := fn.Type.Params
						results := fn.Type.Results
						if (params == nil || len(params.List) == 0) && (results != nil && len(results.List) == 1) {

							// Check if the return type is *node.Raw for node factories
							if starExpr, ok := results.List[0].Type.(*ast.StarExpr); ok {
								if expr, ok := starExpr.X.(*ast.SelectorExpr); ok {
									if expr.Sel.Name == "Raw" && expr.X.(*ast.Ident).Name == "node" {
										pkgPath := pkg.PkgPath
										if pkg.Name == "main" {
											pkgPath = "main"
										}
										structs = append(structs, StructInfo{
											PkgPath: pkgPath,
											PkgName: pkg.Name,
											Name:    fn.Name.Name,
										})
										return true
									}
								}
							}

							// any other single return, no argument functions in main
							// will be registered to potentially be used as component factories
							if pkg.Name == "main" {
								pkgPath := pkg.PkgPath
								if pkg.Name == "main" {
									pkgPath = "main"
								}
								structs = append(structs, StructInfo{
									PkgPath: pkgPath,
									PkgName: pkg.Name,
									Name:    fn.Name.Name,
								})
							}
						}
					}
				}
				return true
			})
		}
	}
	return
}

// findStructs finds all struct definitions in the given package paths
func findStructs(pkgs []*packages.Package) []StructInfo {
	var structs []StructInfo

	// Iterate through all loaded packages
	for _, pkg := range pkgs {
		if strings.HasPrefix(pkg.PkgPath, "internal/") {
			continue
		}
		// Iterate through each file in the package
		for _, file := range pkg.Syntax {
			// Find struct definitions within the file
			ast.Inspect(file, func(n ast.Node) bool {
				// Check if the node is a type specification
				if typeSpec, ok := n.(*ast.TypeSpec); ok {
					// Check if the type is a struct type and the name is exported (starts with an uppercase letter)
					if _, ok := typeSpec.Type.(*ast.StructType); ok && isExported(typeSpec.Name.Name) {
						pkgPath := pkg.PkgPath
						if pkg.Name == "main" {
							pkgPath = "main"
						}
						structs = append(structs, StructInfo{
							PkgPath: pkgPath,
							PkgName: pkg.Name,
							Name:    typeSpec.Name.Name,
							Typed:   typeSpec.TypeParams != nil,
						})
					}
				}
				return true
			})
		}
	}

	return structs
}

// isExported checks if a given name is exported (starts with an uppercase letter)
func isExported(name string) bool {
	if name == "" {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}

func generateSourceFile(structs, factories []StructInfo, outputPath string) error {
	// Template for the generated Go source file
	const tmpl = `// Code generated; DO NOT EDIT.
package main

import (
	"reflect"
{{range .Imports}}	"{{.}}"
{{end}}
)

func init() {
	meta.Components = map[string]reflect.Type{
{{range .Structs}}		"{{.PkgPath}}.{{.Name}}": reflect.TypeOf((*{{if ne .PkgName "main"}}{{.PkgName}}.{{end}}{{.Name}}{{if .Typed}}[any]{{end}})(nil)).Elem(),
{{end}}	}
	meta.Factories = map[string]reflect.Value{
{{range .Factories}}		"{{.PkgPath}}.{{.Name}}": reflect.ValueOf({{if ne .PkgName "main"}}{{.PkgName}}.{{end}}{{.Name}}),
{{end}}	}
}
`

	// Collect unique package imports
	importSet := make(map[string]string)
	for _, s := range structs {
		importSet[s.PkgPath] = s.PkgPath
	}
	importSet["github.com/progrium/rig/pkg/meta"] = "github.com/progrium/rig/pkg/meta"
	delete(importSet, "main")

	// Prepare data for the template
	data := struct {
		Imports []string
		Structs []struct {
			PkgPath string
			PkgName string
			Name    string
			Typed   bool
		}
		Factories []struct {
			PkgPath string
			PkgName string
			Name    string
		}
	}{
		Imports: make([]string, 0, len(importSet)),
		Structs: make([]struct {
			PkgPath string
			PkgName string
			Name    string
			Typed   bool
		}, len(structs)),
		Factories: make([]struct {
			PkgPath string
			PkgName string
			Name    string
		}, len(factories)),
	}

	// Fill in import paths and struct information
	i := 0
	for pkgPath := range importSet {
		data.Imports = append(data.Imports, pkgPath)
	}
	for j, s := range structs {
		data.Structs[j] = struct {
			PkgPath string
			PkgName string
			Name    string
			Typed   bool
		}{
			PkgPath: s.PkgPath,
			PkgName: s.PkgName,
			Name:    s.Name,
			Typed:   s.Typed,
		}
		i++
	}
	for j, s := range factories {
		data.Factories[j] = struct {
			PkgPath string
			PkgName string
			Name    string
		}{
			PkgPath: s.PkgPath,
			PkgName: s.PkgName,
			Name:    s.Name,
		}
		i++
	}

	// Parse and execute the template
	t := template.Must(template.New("source").Parse(tmpl))
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}
