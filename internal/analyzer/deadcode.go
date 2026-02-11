package analyzer

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
	"unicode"
)

type DeadFunction struct {
	File string
	Name string
	Line int
}

func analyzeDeadCode(fset *token.FileSet, files []*ast.File) []DeadFunction {
	type funcInfo struct {
		file string
		name string
		line int
	}

	exported := make(map[string]funcInfo)
	called := make(map[string]bool)

	for _, f := range files {
		pos := fset.Position(f.Pos())
		relPath := filepath.Base(pos.Filename)

		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				name := node.Name.Name

				if name == "main" || name == "init" ||
					strings.HasPrefix(name, "Test") ||
					strings.HasPrefix(name, "Benchmark") ||
					strings.HasPrefix(name, "Example") {
					return true
				}

				if !isExported(name) {
					return true
				}

				key := name
				if node.Recv != nil && len(node.Recv.List) > 0 {
					key = receiverName(node.Recv.List[0].Type) + "." + name
				}

				exported[key] = funcInfo{
					file: relPath,
					name: key,
					line: fset.Position(node.Pos()).Line,
				}

			case *ast.CallExpr:
				switch fn := node.Fun.(type) {
				case *ast.Ident:
					called[fn.Name] = true
				case *ast.SelectorExpr:
					called[fn.Sel.Name] = true
					if ident, ok := fn.X.(*ast.Ident); ok {
						called[ident.Name+"."+fn.Sel.Name] = true
					}
				}
			}
			return true
		})
	}

	var dead []DeadFunction
	for key, info := range exported {
		if !called[key] {
			parts := strings.Split(key, ".")
			simpleName := parts[len(parts)-1]
			if !called[simpleName] {
				dead = append(dead, DeadFunction{
					File: info.file,
					Name: info.name,
					Line: info.line,
				})
			}
		}
	}

	return dead
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return unicode.IsUpper(rune(name[0]))
}
