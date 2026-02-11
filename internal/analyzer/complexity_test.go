package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestCalcComplexity_EmptyFunction(t *testing.T) {
	src := `package main
func empty() {
}`

	complexity := parseAndCalcComplexity(t, src)
	if complexity != 1 {
		t.Errorf("expected complexity 1, got %d", complexity)
	}
}

func TestCalcComplexity_IfStatement(t *testing.T) {
	src := `package main
func withIf(x int) {
	if x > 0 {
		return
	}
}`

	complexity := parseAndCalcComplexity(t, src)
	if complexity != 2 {
		t.Errorf("expected complexity 2, got %d", complexity)
	}
}

func TestCalcComplexity_MultipleIf(t *testing.T) {
	src := `package main
func multipleIf(x, y int) {
	if x > 0 {
		return
	}
	if y > 0 {
		return
	}
}`

	complexity := parseAndCalcComplexity(t, src)
	if complexity != 3 {
		t.Errorf("expected complexity 3 (1 base + 2 if), got %d", complexity)
	}
}

func TestCalcComplexity_NestedIf(t *testing.T) {
	src := `package main
func nestedIf(x, y int) {
	if x > 0 {
		if y > 0 {
			return
		}
	}
}`

	complexity := parseAndCalcComplexity(t, src)
	if complexity != 3 {
		t.Errorf("expected complexity 3 (1 base + 2 if), got %d", complexity)
	}
}

func TestCalcComplexity_ForLoop(t *testing.T) {
	src := `package main
func withFor() {
	for i := 0; i < 10; i++ {
		println(i)
	}
}`

	complexity := parseAndCalcComplexity(t, src)
	if complexity != 2 {
		t.Errorf("expected complexity 2, got %d", complexity)
	}
}

func TestCalcComplexity_RangeLoop(t *testing.T) {
	src := `package main
func withRange(items []int) {
	for _, item := range items {
		println(item)
	}
}`

	complexity := parseAndCalcComplexity(t, src)
	if complexity != 2 {
		t.Errorf("expected complexity 2, got %d", complexity)
	}
}

func TestCalcComplexity_SwitchStatement(t *testing.T) {
	src := `package main
func withSwitch(x int) {
	switch x {
	case 1:
		return
	case 2:
		return
	case 3:
		return
	default:
		return
	}
}`

	// Each case (except default if it has no condition) adds 1
	complexity := parseAndCalcComplexity(t, src)
	if complexity != 4 {
		t.Errorf("expected complexity 4 (1 base + 3 cases), got %d", complexity)
	}
}

func TestCalcComplexity_TypeSwitchStatement(t *testing.T) {
	src := `package main
func withTypeSwitch(x interface{}) {
	switch x.(type) {
	case int:
		return
	case string:
		return
	}
}`

	complexity := parseAndCalcComplexity(t, src)
	if complexity != 4 {
		t.Errorf("expected complexity 4 (1 base + 1 type switch + 2 cases), got %d", complexity)
	}
}

func TestCalcComplexity_SelectStatement(t *testing.T) {
	src := `package main
func withSelect(ch1, ch2 chan int) {
	select {
	case x := <-ch1:
		println(x)
	case y := <-ch2:
		println(y)
	}
}`

	complexity := parseAndCalcComplexity(t, src)
	if complexity != 4 {
		t.Errorf("expected complexity 4 (1 base + 1 select + 2 cases), got %d", complexity)
	}
}

func TestCalcComplexity_LogicalAnd(t *testing.T) {
	src := `package main
func withAnd(x, y bool) bool {
	return x && y
}`

	complexity := parseAndCalcComplexity(t, src)
	if complexity != 2 {
		t.Errorf("expected complexity 2 (1 base + 1 &&), got %d", complexity)
	}
}

func TestCalcComplexity_LogicalOr(t *testing.T) {
	src := `package main
func withOr(x, y bool) bool {
	return x || y
}`

	complexity := parseAndCalcComplexity(t, src)
	if complexity != 2 {
		t.Errorf("expected complexity 2 (1 base + 1 ||), got %d", complexity)
	}
}

func TestCalcComplexity_MultipleLogicalOperators(t *testing.T) {
	src := `package main
func complex(a, b, c, d bool) bool {
	return (a && b) || (c && d)
}`

	complexity := parseAndCalcComplexity(t, src)
	if complexity != 4 {
		t.Errorf("expected complexity 4 (1 base + 3 logical ops), got %d", complexity)
	}
}

func TestCalcComplexity_ComplexFunction(t *testing.T) {
	src := `package main
func veryComplex(x int, items []string) int {
	if x < 0 {
		return -1
	}
	
	for i := 0; i < 10; i++ {
		if i == x {
			return i
		}
	}
	
	for _, item := range items {
		switch item {
		case "foo":
			return 1
		case "bar":
			return 2
		}
	}
	
	return 0
}`

	// 1 base + 1 if + 1 for + 1 if + 1 range + 2 cases = 7
	complexity := parseAndCalcComplexity(t, src)
	if complexity != 7 {
		t.Errorf("expected complexity 7, got %d", complexity)
	}
}

func TestCalcComplexity_FunctionLiteral(t *testing.T) {
	src := `package main
func withClosure(x int) {
	if x > 0 {
		fn := func() {
			if x > 10 {
				println("big")
			}
		}
		fn()
	}
}`

	// Should only count outer if, not the if inside the closure
	complexity := parseAndCalcComplexity(t, src)
	if complexity != 2 {
		t.Errorf("expected complexity 2 (closure should not count), got %d", complexity)
	}
}

func TestAnalyzeComplexity_SimpleFunction(t *testing.T) {
	src := `package main

func simple() {
	println("hello")
}

func withIf(x int) {
	if x > 0 {
		println("positive")
	}
}
`

	results := parseAndAnalyzeComplexity(t, src, "test.go")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Check first function
	if results[0].Name != "simple" {
		t.Errorf("expected name 'simple', got '%s'", results[0].Name)
	}
	if results[0].Complexity != 1 {
		t.Errorf("expected complexity 1, got %d", results[0].Complexity)
	}

	// Check second function
	if results[1].Name != "withIf" {
		t.Errorf("expected name 'withIf', got '%s'", results[1].Name)
	}
	if results[1].Complexity != 2 {
		t.Errorf("expected complexity 2, got %d", results[1].Complexity)
	}
}

func TestAnalyzeComplexity_MethodWithReceiver(t *testing.T) {
	src := `package main

type Counter struct {
	count int
}

func (c Counter) Increment() {
	c.count++
}

func (c *Counter) IncrementPtr() {
	c.count++
}
`

	results := parseAndAnalyzeComplexity(t, src, "test.go")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Check value receiver
	if results[0].Name != "Counter.Increment" {
		t.Errorf("expected name 'Counter.Increment', got '%s'", results[0].Name)
	}

	// Check pointer receiver
	if results[1].Name != "Counter.IncrementPtr" {
		t.Errorf("expected name 'Counter.IncrementPtr', got '%s'", results[1].Name)
	}
}

func TestAnalyzeComplexity_GenericReceiver(t *testing.T) {
	src := `package main

type Container[T any] struct {
	value T
}

func (c Container[T]) Get() T {
	return c.value
}
`

	results := parseAndAnalyzeComplexity(t, src, "test.go")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	// Check generic receiver - should extract base type name
	if results[0].Name != "Container.Get" {
		t.Errorf("expected name 'Container.Get', got '%s'", results[0].Name)
	}
}

func TestReceiverName_Ident(t *testing.T) {
	src := `package main
type Foo struct{}
func (f Foo) Bar() {}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	var receiverExpr ast.Expr
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Recv != nil {
			receiverExpr = fn.Recv.List[0].Type
			return false
		}
		return true
	})

	if receiverExpr == nil {
		t.Fatal("no receiver found")
	}

	name := receiverName(receiverExpr)
	if name != "Foo" {
		t.Errorf("expected 'Foo', got '%s'", name)
	}
}

func TestReceiverName_StarExpr(t *testing.T) {
	src := `package main
type Foo struct{}
func (f *Foo) Bar() {}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	var receiverExpr ast.Expr
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Recv != nil {
			receiverExpr = fn.Recv.List[0].Type
			return false
		}
		return true
	})

	if receiverExpr == nil {
		t.Fatal("no receiver found")
	}

	name := receiverName(receiverExpr)
	if name != "Foo" {
		t.Errorf("expected 'Foo', got '%s'", name)
	}
}

// Helper function to parse source and calculate complexity
func parseAndCalcComplexity(t *testing.T, src string) int {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	var body *ast.BlockStmt
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			body = fn.Body
			return false
		}
		return true
	})

	if body == nil {
		t.Fatal("no function body found")
	}

	return calcComplexity(body)
}

// Helper function to parse source and analyze complexity
func parseAndAnalyzeComplexity(t *testing.T, src, path string) []FunctionComplexity {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	return analyzeComplexity(fset, file, path)
}
