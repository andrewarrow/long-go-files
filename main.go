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

	funcsPerFile := len(functions) / numFiles
	if len(functions)%numFiles != 0 {
		funcsPerFile++
	}

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

		suffix := generateFilenameSuffix(functions[start:end], i == 0)
		outputFile := filepath.Join(outputDir, fmt.Sprintf("%s_%s.go", baseFileName, suffix))
		if err := writeGoFile(outputFile, packageName, imports, typesForFile, functions[start:end], fset); err != nil {
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