package analyzer

import (
	"go/ast"
	"go/token"
	"path/filepath"
)

type FunctionComplexity struct {
	File       string
	Name       string
	Line       int
	Complexity int
}

func analyzeComplexity(fset *token.FileSet, file *ast.File, path string) []FunctionComplexity {
	var results []FunctionComplexity

	ast.Inspect(file, func(n ast.Node) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			name := fn.Name.Name
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				name = receiverName(fn.Recv.List[0].Type) + "." + name
			}
			complexity := calcComplexity(fn.Body)
			pos := fset.Position(fn.Pos())
			results = append(results, FunctionComplexity{
				File:       filepath.Base(path),
				Name:       name,
				Line:       pos.Line,
				Complexity: complexity,
			})
		}
		return true
	})

	return results
}

func calcComplexity(body *ast.BlockStmt) int {
	if body == nil {
		return 1
	}

	complexity := 1

	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt:
			complexity++
		case *ast.ForStmt:
			complexity++
		case *ast.RangeStmt:
			complexity++
		case *ast.CaseClause:
			if node.List != nil {
				complexity++
			}
		case *ast.CommClause:
			if node.Comm != nil {
				complexity++
			}
		case *ast.BinaryExpr:
			if node.Op == token.LAND || node.Op == token.LOR {
				complexity++
			}
		case *ast.SelectStmt:
			complexity++
		case *ast.TypeSwitchStmt:
			complexity++
		case *ast.FuncLit:
			return false
		}
		return true
	})

	return complexity
}

func receiverName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverName(t.X)
	case *ast.IndexExpr:
		return receiverName(t.X)
	default:
		return "?"
	}
}
