package linter_test

import (
	"testing"

	"github.com/yandex-practicum/shorten-url/cmd/linter"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestLinterInMem(t *testing.T) {
	files := map[string]string{
		"a/a.go": `
			package a
			import ("os"; "log")

			func check() {
				panic("stop")       // want "panic found"
				os.Exit(1)          // want "direct call to os.Exit is prohibited"
				log.Fatal("err")    // want "direct call to log.Fatal is prohibited"
			}
		`,
		"main/main.go": `
			package main
			import ("os"; "log")

			func main() {
				os.Exit(0)     // success (main в пакете main)
				log.Fatal("!") // success
				panic("no")    // want "panic found"
			}
		`,
	}

	dir, cleanup, err := analysistest.WriteFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	analysistest.Run(t, dir, linter.Analyzer, "a", "main")
}
