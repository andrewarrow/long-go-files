package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.go> <num_files>\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	numFiles, err := strconv.Atoi(os.Args[2])
	if err != nil || numFiles <= 0 {
		fmt.Fprintf(os.Stderr, "Error: num_files must be a positive integer\n")
		os.Exit(1)
	}

	if err := splitGoFile(inputFile, numFiles); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func splitGoFile(inputFile string, numFiles int) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, inputFile, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %v", err)
	}

	packageName := node.Name.Name
	imports := extractImports(node)
	typeDecls := extractTypeDecls(node)
	functions := extractFunctions(node)

	if len(functions) == 0 {
		return fmt.Errorf("no functions found in file")
	}

	baseFileName := strings.TrimSuffix(filepath.Base(inputFile), ".go")
	outputDir := filepath.Dir(inputFile)

	existingFiles, err := getExistingFiles(outputDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	funcsPerFile := len(functions) / numFiles
	if len(functions)%numFiles != 0 {
		funcsPerFile++
	}

	usedNames := make(map[string]bool)

	for i := 0; i < numFiles; i++ {
		start := i * funcsPerFile
		end := start + funcsPerFile
		if end > len(functions) {
			end = len(functions)
		}
		if start >= len(functions) {
			break
		}

		var typesForFile []*ast.GenDecl
		if i == 0 {
			typesForFile = typeDecls
		}

		suffix := generateUniqueFilenameSuffix(functions[start:end], i == 0, baseFileName, existingFiles, usedNames)
		usedNames[suffix] = true
		outputFile := filepath.Join(outputDir, fmt.Sprintf("%s_%s.go", baseFileName, suffix))
		
		requiredImports := analyzeRequiredImports(functions[start:end], typesForFile, imports)
		
		if err := writeGoFile(outputFile, packageName, requiredImports, typesForFile, functions[start:end], fset); err != nil {
			return fmt.Errorf("failed to write %s: %v", outputFile, err)
		}
		fmt.Printf("Created: %s\n", outputFile)
	}

	return nil
}

func extractImports(node *ast.File) []*ast.ImportSpec {
	var imports []*ast.ImportSpec
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					imports = append(imports, importSpec)
				}
			}
		}
	}
	return imports
}

func extractTypeDecls(node *ast.File) []*ast.GenDecl {
	var typeDecls []*ast.GenDecl
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			typeDecls = append(typeDecls, genDecl)
		}
	}
	return typeDecls
}

func extractFunctions(node *ast.File) []*ast.FuncDecl {
	var functions []*ast.FuncDecl
	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			functions = append(functions, funcDecl)
		}
	}
	return functions
}

func writeGoFile(filename, packageName string, imports []*ast.ImportSpec, typeDecls []*ast.GenDecl, functions []*ast.FuncDecl, fset *token.FileSet) error {
	file := &ast.File{
		Name: &ast.Ident{Name: packageName},
	}

	if len(imports) > 0 {
		genDecl := &ast.GenDecl{
			Tok: token.IMPORT,
		}
		for _, imp := range imports {
			genDecl.Specs = append(genDecl.Specs, imp)
		}
		file.Decls = append(file.Decls, genDecl)
	}

	for _, typeDecl := range typeDecls {
		file.Decls = append(file.Decls, typeDecl)
	}

	for _, fn := range functions {
		file.Decls = append(file.Decls, fn)
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return format.Node(f, fset, file)
}

func getExistingFiles(dir string) (map[string]bool, error) {
	files := make(map[string]bool)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return files, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			files[entry.Name()] = true
		}
	}
	return files, nil
}

func generateUniqueFilenameSuffix(functions []*ast.FuncDecl, hasTypes bool, baseFileName string, existingFiles map[string]bool, usedNames map[string]bool) string {
	baseSuffix := generateFilenameSuffix(functions, hasTypes)
	suffix := baseSuffix
	
	filename := fmt.Sprintf("%s_%s.go", baseFileName, suffix)
	if !existingFiles[filename] && !usedNames[suffix] {
		return suffix
	}
	
	alternatives := []string{"core", "main", "base", "util", "helper", "common", "shared", "extra", "misc", "other", "new", "alt", "impl", "logic", "work", "task", "ops", "flow", "step", "part", "unit", "chunk", "block", "piece", "item", "elem", "node", "link", "path", "route", "view", "ctrl", "model", "data", "info", "meta", "config", "setup", "init", "boot", "start", "launch", "run", "exec", "proc", "action", "event", "state", "change", "update", "modify", "edit", "fix", "patch", "clean", "clear", "reset", "fresh", "quick", "fast", "slow", "temp", "local", "remote", "public", "private", "secure", "safe", "simple", "basic", "advanced", "custom", "special", "unique", "single", "multi", "batch", "group", "list", "set", "map", "tree", "graph", "queue", "stack", "buffer", "cache", "store", "load", "save", "fetch", "send", "recv", "sync", "async", "wait", "done", "ready", "active", "idle", "busy", "free", "open", "close", "lock", "unlock", "check", "test", "verify", "valid", "error", "warn", "debug", "trace", "log", "print", "show", "hide", "render", "draw", "build", "make", "craft", "forge", "shape", "form", "mold", "cast", "press", "push", "pull", "move", "shift", "slide", "jump", "skip", "next", "prev", "first", "last", "top", "bottom", "left", "right", "center", "middle", "inner", "outer", "upper", "lower", "high", "low", "big", "small", "large", "tiny", "huge", "mini", "full", "empty", "blank", "void", "null", "zero", "one", "two", "three", "many", "few", "some", "all", "none", "auto", "manual", "smart", "dumb", "cool", "warm", "hot", "cold", "wet", "dry", "soft", "hard", "light", "dark", "bright", "dim", "loud", "quiet", "fast", "slow", "old", "young", "rich", "poor", "true", "false", "good", "bad", "nice", "ugly", "clean", "dirty", "pure", "mixed", "solid", "liquid", "gas", "fire", "water", "earth", "air", "wood", "metal", "stone", "glass", "paper", "cloth", "rope", "wire", "pipe", "tube", "box", "bag", "cup", "bowl", "plate", "knife", "fork", "spoon", "tool", "gear", "part", "chip", "disk", "tape", "card", "key", "lock", "door", "window", "wall", "floor", "roof", "room", "house", "city", "town", "road", "street", "bridge", "river", "lake", "sea", "ocean", "mountain", "hill", "tree", "flower", "grass", "leaf", "seed", "fruit", "root", "branch", "trunk", "bark", "wood", "forest", "field", "farm", "garden", "park", "zoo", "museum", "school", "library", "store", "shop", "market", "bank", "office", "factory", "lab", "studio", "theater", "cinema", "restaurant", "cafe", "hotel", "hospital", "church", "temple", "castle", "tower", "bridge", "tunnel", "cave", "valley", "desert", "island", "beach", "shore", "coast", "port", "harbor", "dock", "ship", "boat", "plane", "train", "car", "bike", "truck", "bus", "taxi", "rocket", "satellite", "star", "moon", "sun", "planet", "comet", "meteor", "galaxy", "universe", "space", "time", "year", "month", "week", "day", "hour", "minute", "second", "moment", "instant", "flash", "spark", "flame", "smoke", "cloud", "rain", "snow", "ice", "frost", "dew", "mist", "fog", "wind", "storm", "thunder", "lightning", "rainbow", "shadow", "light", "beam", "ray", "wave", "sound", "music", "song", "voice", "word", "letter", "number", "symbol", "sign", "mark", "dot", "line", "curve", "circle", "square", "triangle", "diamond", "heart", "star", "cross", "arrow", "spiral", "wave", "grid", "mesh", "net", "web", "thread", "string", "rope", "chain", "link", "bond", "tie", "knot", "loop", "ring", "band", "strip", "belt", "strap", "handle", "grip", "hold", "grasp", "touch", "feel", "sense", "see", "look", "watch", "view", "scan", "search", "find", "seek", "hunt", "track", "follow", "lead", "guide", "direct", "point", "aim", "target", "goal", "end", "finish", "complete", "done", "over", "final", "last", "stop", "halt", "pause", "break", "rest", "sleep", "wake", "rise", "fall", "drop", "lift", "raise", "lower", "turn", "spin", "rotate", "twist", "bend", "fold", "open", "close", "shut", "seal", "break", "crack", "split", "join", "merge", "unite", "divide", "separate", "cut", "slice", "chop", "tear", "rip", "grab", "catch", "throw", "toss", "cast", "shoot", "fire", "blast", "explode", "crash", "smash", "hit", "strike", "punch", "kick", "stomp", "step", "walk", "run", "jog", "sprint", "rush", "hurry", "slow", "crawl", "climb", "jump", "leap", "hop", "skip", "dance", "swim", "dive", "float", "fly", "soar", "glide", "drift", "flow", "stream", "rush", "pour", "drip", "leak", "spill", "splash", "spray", "mist", "dust", "powder", "grain", "sand", "dirt", "mud", "clay", "rock", "stone", "pebble", "boulder", "crystal", "gem", "gold", "silver", "copper", "iron", "steel", "brass", "bronze", "tin", "lead", "zinc", "chrome", "nickel", "cobalt", "carbon", "silicon", "oxygen", "hydrogen", "nitrogen", "helium", "neon", "argon", "mercury", "venus", "mars", "jupiter", "saturn", "uranus", "neptune", "pluto"}
	
	for _, alt := range alternatives {
		testSuffix := alt
		filename = fmt.Sprintf("%s_%s.go", baseFileName, testSuffix)
		if !existingFiles[filename] && !usedNames[testSuffix] {
			return testSuffix
		}
	}
	
	for i := 1; i <= 999; i++ {
		testSuffix := fmt.Sprintf("%s%d", baseSuffix, i)
		filename = fmt.Sprintf("%s_%s.go", baseFileName, testSuffix)
		if !existingFiles[filename] && !usedNames[testSuffix] {
			return testSuffix
		}
	}
	
	return fmt.Sprintf("%s%d", baseSuffix, 999)
}

func generateFilenameSuffix(functions []*ast.FuncDecl, hasTypes bool) string {
	if hasTypes {
		return "types"
	}
	
	if len(functions) == 0 {
		return "empty"
	}
	
	funcNames := make([]string, 0, len(functions))
	for _, fn := range functions {
		if fn.Name != nil {
			funcNames = append(funcNames, strings.ToLower(fn.Name.Name))
		}
	}
	
	if len(funcNames) == 0 {
		return "funcs"
	}
	
	keywords := []string{"parse", "generate", "create", "process", "handle", "convert", "build", "render", "execute", "validate", "format", "transform", "encode", "decode", "load", "save", "write", "read", "get", "set", "add", "remove", "update", "delete", "find", "search", "filter", "sort", "merge", "split", "join", "copy", "move", "clean", "init", "start", "stop", "run", "exec", "call", "invoke", "apply", "map", "reduce", "collect", "stream", "buffer", "cache", "store", "fetch", "send", "receive", "upload", "download", "compress", "extract", "zip", "unzip", "backup", "restore", "sync", "async", "wait", "notify", "signal", "lock", "unlock", "open", "close", "connect", "disconnect", "bind", "unbind", "listen", "serve", "request", "response", "query", "insert", "select", "count", "sum", "max", "min", "avg", "group", "order", "limit", "offset", "page", "chunk", "batch", "parallel", "serial", "concurrent", "thread", "worker", "job", "task", "queue", "stack", "list", "array", "map", "set", "tree", "graph", "node", "edge", "path", "route", "url", "uri", "link", "ref", "ptr", "addr", "size", "len", "cap", "empty", "full", "contains", "exists", "valid", "invalid", "ok", "error", "warn", "info", "debug", "trace", "log", "print", "show", "display", "render", "draw", "paint", "color", "style", "theme", "skin", "layout", "align", "position", "size", "resize", "scale", "zoom", "pan", "scroll", "drag", "drop", "click", "hover", "focus", "blur", "select", "deselect", "toggle", "switch", "enable", "disable", "show", "hide", "visible", "invisible", "active", "inactive", "on", "off", "true", "false", "yes", "no", "ok", "cancel", "submit", "reset", "clear", "clean", "flush", "purge", "refresh", "reload", "restart", "resume", "pause", "play", "record", "replay", "undo", "redo", "cut", "copy", "paste", "clone", "duplicate", "mirror", "reflect", "invert", "reverse", "flip", "rotate", "shift", "move", "slide", "fade", "animate", "transition", "effect", "filter", "mask", "overlay", "background", "foreground", "layer", "depth", "level", "priority", "weight", "rank", "score", "rate", "ratio", "percent", "fraction", "decimal", "integer", "float", "double", "string", "char", "byte", "bit", "word", "line", "paragraph", "section", "chapter", "page", "document", "file", "folder", "directory", "path", "name", "title", "label", "tag", "attr", "prop", "field", "column", "row", "cell", "table", "grid", "matrix", "vector", "point", "coord", "pos", "loc", "place", "spot", "area", "region", "zone", "space", "room", "box", "container", "wrapper", "holder", "frame", "border", "edge", "corner", "side", "top", "bottom", "left", "right", "center", "middle", "inner", "outer", "inside", "outside", "before", "after", "first", "last", "next", "prev", "current", "new", "old", "temp", "tmp", "backup", "orig", "copy", "clone", "draft", "final", "test", "demo", "sample", "example", "template", "pattern", "model", "schema", "struct", "class", "type", "kind", "sort", "category", "group", "team", "user", "admin", "guest", "public", "private", "secure", "safe", "unsafe", "danger", "risk", "warn", "alert", "notice", "message", "text", "content", "data", "info", "meta", "config", "setting", "option", "param", "arg", "value", "result", "output", "input", "source", "target", "dest", "from", "to", "via", "through", "by", "with", "without", "using", "based", "upon", "over", "under", "above", "below", "between", "among", "within", "outside", "beyond", "across", "along", "around", "through", "during", "while", "until", "since", "before", "after", "when", "where", "what", "who", "why", "how", "which", "whose", "whom", "that", "this", "these", "those", "such", "same", "other", "another", "each", "every", "all", "any", "some", "none", "nothing", "something", "anything", "everything", "somewhere", "anywhere", "everywhere", "nowhere", "someone", "anyone", "everyone", "no one"}
	
	for _, keyword := range keywords {
		for _, name := range funcNames {
			if strings.Contains(name, keyword) {
				return keyword
			}
		}
	}
	
	if len(funcNames[0]) > 0 {
		return strings.ToLower(string(funcNames[0][0]))
	}
	
	return "funcs"
}

func analyzeRequiredImports(functions []*ast.FuncDecl, typeDecls []*ast.GenDecl, allImports []*ast.ImportSpec) []*ast.ImportSpec {
	usedIdentifiers := make(map[string]bool)
	
	for _, fn := range functions {
		ast.Inspect(fn, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.SelectorExpr:
				if ident, ok := node.X.(*ast.Ident); ok {
					usedIdentifiers[ident.Name] = true
				}
			case *ast.Ident:
				usedIdentifiers[node.Name] = true
			}
			return true
		})
	}
	
	for _, typeDecl := range typeDecls {
		ast.Inspect(typeDecl, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.SelectorExpr:
				if ident, ok := node.X.(*ast.Ident); ok {
					usedIdentifiers[ident.Name] = true
				}
			case *ast.Ident:
				usedIdentifiers[node.Name] = true
			}
			return true
		})
	}
	
	var requiredImports []*ast.ImportSpec
	for _, imp := range allImports {
		var importName string
		if imp.Name != nil {
			importName = imp.Name.Name
		} else {
			importPath := strings.Trim(imp.Path.Value, `"`)
			parts := strings.Split(importPath, "/")
			importName = parts[len(parts)-1]
		}
		
		if usedIdentifiers[importName] {
			requiredImports = append(requiredImports, imp)
		}
	}
	
	return requiredImports
}