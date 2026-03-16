package linter

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "customlinter",
	Doc:  "reports panic and restricted calls to os.Exit or log.Fatal",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		// для тестов пропускаем проверки
		if strings.HasSuffix(filename, "_test.go") {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			switch fun := call.Fun.(type) {
			case *ast.Ident:
				// Проверка на panic
				obj := pass.TypesInfo.ObjectOf(fun)
				if obj == nil {
					return true
				}
				if b, ok := obj.(*types.Builtin); ok && b.Name() == "panic" {
					pass.Reportf(fun.Pos(), "panic found)")
				}

			case *ast.SelectorExpr:
				// Проверка на os.Exit / log.Fatal
				selObj := pass.TypesInfo.ObjectOf(fun.Sel)
				if selObj == nil {
					return true
				}

				pkg := selObj.Pkg()
				if pkg == nil {
					return true
				}

				pkgName := pkg.Name()
				selName := selObj.Name()

				isOsExit := pkgName == "os" && selName == "Exit"
				isLogFatal := pkgName == "log" && selName == "Fatal"

				if !isOsExit && !isLogFatal {
					return true
				}

				// Проверка на main()
				if pass.Pkg.Name() != "main" {
					reportForbidden(pass, fun, pkgName, selName)
					return true
				}

				if !isInsideMainFunc(pass, file, call) {
					reportForbidden(pass, fun, pkgName, selName)
				}
			}

			return true
		})
	}
	return nil, nil
}

func reportForbidden(pass *analysis.Pass, sel *ast.SelectorExpr, pkgName, selName string) {
	pass.Reportf(sel.Pos(),
		"direct call to %s.%s() is prohibited",
		pkgName, selName)
}

// isInsideMainFunc проверяет использование main()
func isInsideMainFunc(pass *analysis.Pass, file *ast.File, call ast.Node) bool {
	var found bool

	ast.Inspect(file, func(n ast.Node) bool {
		if found {
			return false
		}

		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if fn.Name.Name != "main" || fn.Recv != nil {
			return true
		}

		if call.Pos() >= fn.Body.Pos() && call.End() <= fn.Body.End() {
			found = true
			return false
		}

		return true
	})

	return found
}
