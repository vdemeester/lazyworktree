package services

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/chmouel/lazyworktree/internal/models"
)

// StatusFile represents a file entry from git status.
type StatusFile = models.StatusFile

// StatusTreeNode represents a node in the status file tree (directory or file).
type StatusTreeNode struct {
	Path        string            // Full path (e.g., "internal/app" or "internal/app/app.go")
	File        *StatusFile       // nil for directories
	Children    []*StatusTreeNode // nil for files
	Compression int               // Number of compressed path segments (e.g., "a/b" = 1)
	Depth       int               // Cached depth for rendering
}

// StatusService manages status tree state.
type StatusService struct {
	Tree          *StatusTreeNode
	TreeFlat      []*StatusTreeNode
	CollapsedDirs map[string]bool
	Index         int
}

// NewStatusService creates a new StatusService.
func NewStatusService() *StatusService {
	return &StatusService{
		CollapsedDirs: make(map[string]bool),
	}
}

// BuildStatusTree builds a tree structure from a flat list of files.
// Files are grouped by directory, with directories sorted before files.
func BuildStatusTree(files []StatusFile) *StatusTreeNode {
	if len(files) == 0 {
		return &StatusTreeNode{Path: "", Children: nil}
	}

	root := &StatusTreeNode{Path: "", Children: make([]*StatusTreeNode, 0)}
	childrenByPath := make(map[string]*StatusTreeNode)

	for i := range files {
		file := &files[i]
		parts := strings.Split(file.Filename, "/")

		current := root
		for j := range parts {
			isFile := j == len(parts)-1
			pathSoFar := strings.Join(parts[:j+1], "/")

			if existing, ok := childrenByPath[pathSoFar]; ok {
				current = existing
				continue
			}

			var newNode *StatusTreeNode
			if isFile {
				newNode = &StatusTreeNode{
					Path: pathSoFar,
					File: file,
				}
			} else {
				newNode = &StatusTreeNode{
					Path:     pathSoFar,
					Children: make([]*StatusTreeNode, 0),
				}
			}
			current.Children = append(current.Children, newNode)
			childrenByPath[pathSoFar] = newNode
			current = newNode
		}
	}

	SortStatusTree(root)
	CompressStatusTree(root)
	return root
}

// SortStatusTree sorts tree nodes: directories first, then alphabetically.
func SortStatusTree(node *StatusTreeNode) {
	if node == nil || node.Children == nil {
		return
	}

	sort.Slice(node.Children, func(i, j int) bool {
		iIsDir := node.Children[i].File == nil
		jIsDir := node.Children[j].File == nil
		if iIsDir != jIsDir {
			return iIsDir // directories first
		}
		return node.Children[i].Path < node.Children[j].Path
	})

	for _, child := range node.Children {
		SortStatusTree(child)
	}
}

// CompressStatusTree squashes single-child directory chains (e.g., a/b/c becomes one node).
func CompressStatusTree(node *StatusTreeNode) {
	if node == nil {
		return
	}

	for _, child := range node.Children {
		CompressStatusTree(child)
	}

	// Compress children that are single-child directories
	for i, child := range node.Children {
		for child.File == nil && len(child.Children) == 1 && child.Children[0].File == nil {
			grandchild := child.Children[0]
			grandchild.Compression = child.Compression + 1
			node.Children[i] = grandchild
			child = grandchild
		}
	}
}

// FlattenStatusTree returns visible nodes respecting collapsed state.
func FlattenStatusTree(node *StatusTreeNode, collapsed map[string]bool, depth int) []*StatusTreeNode {
	if node == nil {
		return nil
	}

	result := make([]*StatusTreeNode, 0)

	// Skip root node itself but process its children
	if node.Path != "" {
		nodeCopy := *node
		nodeCopy.Depth = depth
		result = append(result, &nodeCopy)

		// If collapsed, don't include children
		if collapsed[node.Path] {
			return result
		}
	}

	if node.Children != nil {
		childDepth := depth
		if node.Path != "" {
			childDepth = depth + 1
		}
		for _, child := range node.Children {
			result = append(result, FlattenStatusTree(child, collapsed, childDepth)...)
		}
	}

	return result
}

// IsDir returns true if this node is a directory.
func (n *StatusTreeNode) IsDir() bool {
	return n.File == nil
}

// Name returns the display name for this node.
func (n *StatusTreeNode) Name() string {
	return filepath.Base(n.Path)
}

// CollectFiles recursively collects all StatusFile pointers from this node and its children.
func (n *StatusTreeNode) CollectFiles() []*StatusFile {
	var files []*StatusFile
	if n.File != nil {
		files = append(files, n.File)
	}
	for _, child := range n.Children {
		files = append(files, child.CollectFiles()...)
	}
	return files
}

// RebuildFlat rebuilds the flattened tree representation.
func (s *StatusService) RebuildFlat() {
	if s.CollapsedDirs == nil {
		s.CollapsedDirs = make(map[string]bool)
	}
	s.TreeFlat = FlattenStatusTree(s.Tree, s.CollapsedDirs, 0)
}

// ToggleCollapse toggles a directory collapse state and rebuilds the flat list.
func (s *StatusService) ToggleCollapse(path string) {
	if path == "" {
		return
	}
	if s.CollapsedDirs == nil {
		s.CollapsedDirs = make(map[string]bool)
	}
	s.CollapsedDirs[path] = !s.CollapsedDirs[path]
	s.RebuildFlat()
}

// SelectedPath returns the path of the currently selected node.
func (s *StatusService) SelectedPath() string {
	if s.Index >= 0 && s.Index < len(s.TreeFlat) {
		return s.TreeFlat[s.Index].Path
	}
	return ""
}

// RestoreSelection sets Index based on the provided path if it exists.
func (s *StatusService) RestoreSelection(path string) {
	if path == "" {
		return
	}
	for i, node := range s.TreeFlat {
		if node.Path == path {
			s.Index = i
			return
		}
	}
}

// ClampIndex ensures Index is within the valid range for the flat list.
func (s *StatusService) ClampIndex() {
	if s.Index < 0 {
		s.Index = 0
	}
	if len(s.TreeFlat) > 0 && s.Index >= len(s.TreeFlat) {
		s.Index = len(s.TreeFlat) - 1
	}
	if len(s.TreeFlat) == 0 {
		s.Index = 0
	}
}
