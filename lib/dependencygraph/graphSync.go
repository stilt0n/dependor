package dependencygraph

import (
	"dependor/lib/config"
	"dependor/lib/tokenizer"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

type SingleThreadedGraph struct {
	tokens   map[string]*tokenizer.FileToken
	config   *config.Config
	rootPath string
	edgeList map[string][]string
}

// Supports single optional rootPath argument. Uses "." by default.
func NewSync(rootPath ...string) *SingleThreadedGraph {
	rp := "."
	if len(rootPath) > 0 {
		rp = rootPath[0]
	}
	configPath := rp + "/dependor.json"
	if _, err := os.Stat(configPath); rp != "." && err != nil {
		panic(fmt.Sprintf("Root path does not exist or does not have a valid dependor.json config file. To use default config, omit rootPath arg. See error %s\n", err))
	}

	cfg, err := config.ReadConfig(rp + "/dependor.json")
	if err != nil {
		fmt.Println("WARN: No dependor.json file was found so the default config is being used.")
	}

	return &SingleThreadedGraph{
		config:   cfg,
		rootPath: rp,
		tokens:   make(map[string]*tokenizer.FileToken, 0),
	}
}

func (graph *SingleThreadedGraph) ParseGraph() (map[string][]string, error) {
	err := graph.walk()
	if err != nil {
		return nil, err
	}
	graph.resolveImportExtensions()
	graph.createIndexMaps()
	graph.parseTokens()
	if graph.edgeList == nil {
		return nil, errors.New("parse tokens failed with nil edgeList")
	}
	return graph.edgeList, nil
}

// Walks file tree from root path and populates tokenizedFiles
func (graph *SingleThreadedGraph) walk() error {
	searchableExtensions := regexp.MustCompile(`(\.js|\.jsx|\.ts|\.tsx)$`)
	err := filepath.WalkDir(graph.rootPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("There was an error accessing path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && graph.config.ShouldIgnore(path) {
			fmt.Printf("Ignoring directory %q\n", info.Name())
			return filepath.SkipDir
		}

		if searchableExtensions.MatchString(info.Name()) {
			graph.readImports(path)
		}
		return nil
	})

	return err
}

func (graph *SingleThreadedGraph) parseTokens() {
	graph.edgeList = make(map[string][]string, len(graph.tokens))
	for _, tk := range graph.tokens {
		edges := make([]string, 0)
		for importPath, importIdents := range tk.Imports {
			if isIndexFile(importPath) {
				edges = append(edges, graph.resolveIndexImport(importPath, importIdents)...)
			} else {
				edges = append(edges, importPath)
			}
		}
		graph.edgeList[tk.FilePath] = edges
	}
}

func (graph *SingleThreadedGraph) readImports(filePath string) {
	tk, err := tokenizer.NewTokenizerFromFile(filePath)
	if err != nil {
		return
	}
	tokenizedFile := tk.TokenizeImports()
	graph.tokens[tokenizedFile.FilePath] = &tokenizedFile
}

func (graph *SingleThreadedGraph) resolveImportExtensions() {
	for _, tk := range graph.tokens {
		updatedImports := make(map[string][]string, 0)
		for originalPath, idents := range tk.Imports {
			updatedPath := withExtension(graph.tokens, originalPath)
			updatedImports[updatedPath] = idents
		}
		tk.Imports = updatedImports

		if len(tk.ReExports) == 0 {
			continue
		}

		// ReExports aren't needed for withExtension to work so they
		// can be safely overwritten in-place
		for i, originalPath := range tk.ReExports {
			tk.ReExports[i] = withExtension(graph.tokens, originalPath)
		}
	}
}

func (graph *SingleThreadedGraph) resolveIndexImport(pth string, idents []string) []string {
	resolvedPaths := make([]string, 0)
	for _, ident := range idents {
		resolved, ok := graph.tokens[pth].ReExportMap[ident]
		if !ok {
			continue
		}
		resolvedPaths = append(resolvedPaths, resolved)
	}
	return resolvedPaths
}

func (graph *SingleThreadedGraph) createIndexMaps() {
	for _, tk := range graph.tokens {
		if !isIndexFile(tk.FilePath) {
			continue
		}

		indexMap := make(map[string]string, 0)
		for _, reExportPath := range tk.ReExports {
			reExportFileNode, ok := graph.tokens[reExportPath]
			if !ok {
				continue
			}
			for _, export := range reExportFileNode.Exports {
				indexMap[export] = reExportPath
			}
		}

		for _, export := range tk.Exports {
			indexMap[export] = tk.FilePath
		}

		tk.ReExportMap = indexMap
	}
}

func withExtension(pathMap map[string]*tokenizer.FileToken, path string) string {
	extensions := []string{
		".js",
		".ts",
		".jsx",
		".tsx",
		"/index.js",
		"/index.ts",
		"/index.jsx",
		"/index.tsx",
	}

	for _, extension := range extensions {
		if _, ok := pathMap[path+extension]; ok {
			return path + extension
		}
	}

	return path
}

var indexFilePattern = regexp.MustCompile("index.(js|ts|jsx|tsx)$")

func isIndexFile(filePath string) bool {
	return indexFilePattern.MatchString(filePath)
}