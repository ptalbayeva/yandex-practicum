package linter

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer анализатор
var Analyzer = &analysis.Analyzer{
	Name: "customlinter",
	Doc:  "reports panic and restricted calls to os.Exit or log.Fatal",
	Run:  run,
}

// run запуск анализатора
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
			// Проверка panic
			case *ast.Ident:
				if fun.Name == "panic" {
					pass.Reportf(fun.Pos(), "panic found")
				}

			// Проверка вызовов вида os.Exit или log.Fatal
			case *ast.SelectorExpr:
				if pkg, ok := fun.X.(*ast.Ident); ok {
					if (pkg.Name == "os" && fun.Sel.Name == "Exit") ||
						(pkg.Name == "log" && fun.Sel.Name == "Fatal") {

						// Проверяем контекст: запрещено везде, кроме func main() в package main
						if !isMainInMain(pass, n) {
							pass.Reportf(fun.Pos(), "direct call to %s.%s is prohibited outside of main function in main package", pkg.Name, fun.Sel.Name)
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

// isMainInMain Проверка запуска main только внутри main
func isMainInMain(pass *analysis.Pass, n ast.Node) bool {
	if pass.Pkg.Name() != "main" {
		return false
	}

	for _, file := range pass.Files {
		if n.Pos() >= file.Pos() && n.End() <= file.End() {
			var inMainFunc bool
			ast.Inspect(file, func(node ast.Node) bool {
				if fn, ok := node.(*ast.FuncDecl); ok {
					if n.Pos() >= fn.Pos() && n.End() <= fn.End() {
						inMainFunc = fn.Name.Name == "main"
					}
				}
				return true
			})
			return inMainFunc
		}
	}
	return false
}
