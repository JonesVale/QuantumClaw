package billingexpr

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/vm"
)

const maxCacheSize = 256
const DefaultExprVersion = 1

func ParseExprVersion(exprStr string) (version int, body string) {
	if strings.HasPrefix(exprStr, "v1:") {
		return 1, exprStr[3:]
	}
	return DefaultExprVersion, exprStr
}

type cachedEntry struct {
	prog     *vm.Program
	usedVars map[string]bool
	version  int
}

var (
	cacheMu sync.RWMutex
	cache   = make(map[string]*cachedEntry, 64)
)

var compileEnvPrototypeV1 = map[string]interface{}{
	"p":    float64(0),
	"c":    float64(0),
	"len":  float64(0),
	"cr":   float64(0),
	"cc":   float64(0),
	"cc1h": float64(0),
	"img":  float64(0),
	"img_o": float64(0),
	"ai":   float64(0),
	"ao":   float64(0),
	"tier":                   func(string, float64) float64 { return 0 },
	"header":                 func(string) string { return "" },
	"param":                  func(string) interface{} { return nil },
	"has":                    func(interface{}, string) bool { return false },
	"hour":                   func(string) int { return 0 },
	"minute":                 func(string) int { return 0 },
	"weekday":                func(string) int { return 0 },
	"month":                  func(string) int { return 0 },
	"day":                    func(string) int { return 0 },
	"max":                    math.Max,
	"min":                    math.Min,
	"abs":                    math.Abs,
	"ceil":                   math.Ceil,
	"floor":                  math.Floor,
	"round":                  RoundNearest,
	"roundUp":                RoundUp,
	"roundDown":              RoundDown,
}

func getCompileEnv(version int) map[string]interface{} {
	switch version {
	default:
		return compileEnvPrototypeV1
	}
}

func CompileFromCache(exprStr string) (*vm.Program, error) {
	return compileFromCacheByHashInternal(exprStr, ExprHashString(exprStr))
}

func CompileFromCacheByHash(exprStr, hash string) (*vm.Program, error) {
	return compileFromCacheByHashInternal(exprStr, hash)
}

func compileFromCacheByHashInternal(exprStr, hash string) (*vm.Program, error) {
	cacheMu.RLock()
	if entry, ok := cache[hash]; ok {
		cacheMu.RUnlock()
		return entry.prog, nil
	}
	cacheMu.RUnlock()

	version, body := ParseExprVersion(exprStr)
	prog, err := expr.Compile(body, expr.Env(getCompileEnv(version)), expr.AsFloat64())
	if err != nil {
		return nil, fmt.Errorf("expr compile error: %w", err)
	}

	vars := extractUsedVars(prog)

	cacheMu.Lock()
	if len(cache) >= maxCacheSize {
		cache = make(map[string]*cachedEntry, 64)
	}
	cache[hash] = &cachedEntry{prog: prog, usedVars: vars, version: version}
	cacheMu.Unlock()

	return prog, nil
}

func ExprVersion(exprStr string) int {
	if exprStr == "" {
		return DefaultExprVersion
	}
	hash := ExprHashString(exprStr)
	cacheMu.RLock()
	if entry, ok := cache[hash]; ok {
		cacheMu.RUnlock()
		return entry.version
	}
	cacheMu.RUnlock()
	v, _ := ParseExprVersion(exprStr)
	return v
}

func extractUsedVars(prog *vm.Program) map[string]bool {
	vars := make(map[string]bool)
	node := prog.Node()
	ast.Find(node, func(n ast.Node) bool {
		if id, ok := n.(*ast.IdentifierNode); ok {
			vars[id.Value] = true
		}
		return false
	})
	return vars
}

func UsedVars(exprStr string) map[string]bool {
	if exprStr == "" {
		return nil
	}
	hash := ExprHashString(exprStr)
	cacheMu.RLock()
	if entry, ok := cache[hash]; ok {
		cacheMu.RUnlock()
		return entry.usedVars
	}
	cacheMu.RUnlock()

	if _, err := compileFromCacheByHashInternal(exprStr, hash); err != nil {
		return nil
	}
	cacheMu.RLock()
	entry, ok := cache[hash]
	cacheMu.RUnlock()
	if ok {
		return entry.usedVars
	}
	return nil
}

func InvalidateCache() {
	cacheMu.Lock()
	cache = make(map[string]*cachedEntry, 64)
	cacheMu.Unlock()
}