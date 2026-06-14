package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

// Each fixture has a function named per the table with an if and a loop, so a
// working analyzer should detect the function and count at least one branch
// (complexity >= 2). Loose on exact counts to stay robust across heuristics.

const goFixture = `package main

func handle(x int) int {
	if x > 0 {
		return 1
	}
	for i := 0; i < x; i++ {
		println(i)
	}
	return 0
}
`

const pyFixture = `def handle(x):
    if x > 0:
        return 1
    for i in range(x):
        print(i)
    return 0
`

const tsFixture = `export function handle(x: number): number {
  if (x > 0) {
    return 1;
  }
  for (let i = 0; i < x; i++) {
    console.log(i);
  }
  return 0;
}
`

const rsFixture = `pub fn handle(x: i32) -> i32 {
    if x > 0 {
        return 1;
    }
    for i in 0..x {
        println!("{}", i);
    }
    0
}
`

const javaFixture = `class H {
    int handle(int x) {
        if (x > 0) {
            return 1;
        }
        for (int i = 0; i < x; i++) {
            System.out.println(i);
        }
        return 0;
    }
}
`

const rbFixture = `def handle(x)
  if x > 0
    return 1
  end
  for i in 0..x
    puts i
  end
  0
end
`

const phpFixture = `<?php
function handle($x) {
    if ($x > 0) {
        return 1;
    }
    for ($i = 0; $i < $x; $i++) {
        echo $i;
    }
    return 0;
}
`

const csFixture = `class H {
    int Handle(int x) {
        if (x > 0) {
            return 1;
        }
        for (int i = 0; i < x; i++) {
            System.Console.WriteLine(i);
        }
        return 0;
    }
}
`

func TestLanguageAnalyzers_DetectComplexity(t *testing.T) {
	tests := []struct {
		lang Language
		file string
		src  string
		fn   string
	}{
		{LangGo, "h.go", goFixture, "handle"},
		{LangPython, "h.py", pyFixture, "handle"},
		{LangTypeScript, "h.ts", tsFixture, "handle"},
		{LangRust, "h.rs", rsFixture, "handle"},
		{LangJava, "H.java", javaFixture, "handle"},
		{LangRuby, "h.rb", rbFixture, "handle"},
		{LangPHP, "h.php", phpFixture, "handle"},
		{LangCSharp, "H.cs", csFixture, "Handle"},
	}

	for _, tt := range tests {
		t.Run(string(tt.lang), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.file)
			if err := os.WriteFile(path, []byte(tt.src), 0o644); err != nil {
				t.Fatal(err)
			}

			funcs, count := NewLanguageAnalyzer(tt.lang).AnalyzeComplexity([]string{path})
			if count == 0 {
				t.Fatalf("%s: detected no functions", tt.lang)
			}

			var found *FunctionComplexity
			for i := range funcs {
				if funcs[i].Name == tt.fn {
					found = &funcs[i]
					break
				}
			}
			if found == nil {
				t.Fatalf("%s: function %q not detected; got %+v", tt.lang, tt.fn, funcs)
			}
			if found.Complexity < 2 {
				t.Errorf("%s: %s complexity = %d, want >= 2", tt.lang, tt.fn, found.Complexity)
			}
		})
	}
}
