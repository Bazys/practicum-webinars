package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"text/template"
)

const wrapperTemplate = `
func {{.Name}}WithTiming(args ...interface{}) (result []interface{}, err error) {
    startTime := time.Now()
    defer func() {
        log.Printf("{{.Name}} took %s", time.Since(startTime))
    }()
    return {{.Name}}(args...), nil
}
`

func main() {
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, "your_go_file.go", nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.New("wrapper").Parse(wrapperTemplate))

	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if ok {
			err := tmpl.Execute(os.Stdout, fn)
			if err != nil {
				log.Fatal(err)
			}
		}
		return true
	})
}
