package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testfsnotify/modules/lg"
)

func main() {
	dir, err := DirectoryLister("/NAS4/MMD")
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(len(dir))
	printArray(dir)
}

func printArray(dir []string) {
	for _, f := range dir {
		fmt.Println(f)
	}
}

type Walker struct {
	// Track visited paths to prevent cycles
	visited map[string]bool
	// Track symlink mappings at depth 1
	symlinkMappings map[string]string
	// Root directory for the walk
	root string
	// Absolute path of root
	rootAbs string
}

func NewWalker(root string) (*Walker, error) {
	rootAbs, err := filepath.Abs(root)

	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &Walker{
		visited:         make(map[string]bool),
		symlinkMappings: make(map[string]string),
		root:            root,
		rootAbs:         rootAbs,
	}, nil
}

func (w *Walker) resolveSymlink(path string) (string, error) {
	target, err := os.Readlink(path)
	if err != nil {
		return "", fmt.Errorf("failed to read symlink: %w", err)
	}

	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}

	target, err = filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for symlink target: %w", err)
	}

	return target, nil
}

func (w *Walker) identifyRootSymlinks() error {
	entries, err := os.ReadDir(w.rootAbs)
	if err != nil {
		return fmt.Errorf("failed to read root directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			fullPath := filepath.Join(w.rootAbs, entry.Name())
			target, err := w.resolveSymlink(fullPath)
			if err != nil {
				// Log error but continue with other symlinks
				lg.Err(fmt.Sprintf("Warning: failed to resolve symlink %s", fullPath), err)
				continue
			}
			w.symlinkMappings[target] = fullPath
		}
	}

	return nil
}

func (w *Walker) walkDir(path string, info os.FileInfo, videos *[]string) error {
	// Prevent cycles by tracking visited paths
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if w.visited[absPath] {
		return nil
	}
	w.visited[absPath] = true

	// Handle symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		relPath, err := filepath.Rel(w.rootAbs, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Only follow symlinks at depth 1
		if filepath.Dir(relPath) != "." {
			return filepath.SkipDir
		}

		target, err := w.resolveSymlink(path)
		if err != nil {
			return fmt.Errorf("failed to resolve symlink: %w", err)
		}

		targetInfo, err := os.Stat(target)
		if err != nil {
			return fmt.Errorf("failed to stat symlink target: %w", err)
		}

		return w.walkDir(target, targetInfo, videos)
	}

	// Process regular files and directories
	if !info.IsDir() {
		if isVideoFile(path) {
			// Check if this file is under a symlinked directory
			for realPath, symlinkPath := range w.symlinkMappings {
				if rel, err := filepath.Rel(realPath, path); err == nil && !filepath.IsAbs(rel) {
					// Reconstruct path using symlink

					path = filepath.Join(symlinkPath, rel)

					//fmt.Println("디버그: 완성경로 -", realPath, symlinkPath, path)
					break
				}
			}

			absPath, err := filepath.Abs(path)
			if err == nil {

				relPath := strings.TrimPrefix(absPath, w.rootAbs)
				relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
				*videos = append(*videos, relPath)

				//fmt.Println("디버그: 앱스경로 -", absPath, relPath)
			}
		}
		return nil
	}

	// Read directory entries
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		info, err := entry.Info()
		if err != nil {
			lg.Err(fmt.Sprintf("Warning: failed to get info for %s", fullPath), err)
			continue
		}

		if err := w.walkDir(fullPath, info, videos); err != nil {
			lg.Err(fmt.Sprintf("Warning: error processing %s", fullPath), err)
		}
	}

	return nil
}

func DirectoryLister(root string) ([]string, error) {
	walker, err := NewWalker(root)
	if err != nil {
		return nil, fmt.Errorf("failed to create walker: %w", err)
	}

	if err := walker.identifyRootSymlinks(); err != nil {
		return nil, fmt.Errorf("failed to identify root symlinks: %w", err)
	}

	rootInfo, err := os.Stat(walker.rootAbs)
	if err != nil {
		return nil, fmt.Errorf("failed to stat root directory: %w", err)
	}

	var videos []string
	if err := walker.walkDir(walker.rootAbs, rootInfo, &videos); err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	nvideos := deletePrefix(videos, "NAS2/PMV/")

	return nvideos, nil
}

var videoExtensions = map[string]struct{}{
	".mp4": {}, ".mkv": {}, ".avi": {}, ".m4v": {}, ".mov": {}, ".wmv": {}, ".webm": {},
	".flv": {}, ".f4v": {}, ".f4p": {}, ".f4a": {}, ".f4b": {},
	".3gp": {}, ".3g2": {},
	".rmvb": {}, ".rm": {},
	".ts": {}, ".m2ts": {}, ".mts": {},
	".vob": {},
	".ogv": {}, ".ogg": {},
	".mpg": {}, ".mpeg": {}, ".mpe": {},
	".divx": {}, ".xvid": {},
	".mxf": {},
	".dv":  {},
	".asf": {},
	".qt":  {},
	".amv": {},
}

func isVideoFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, exists := videoExtensions[ext]
	return exists
}

func CountLines(file *os.File) (int, error) {
	//파일 수 세주는 함수
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	return lineCount, scanner.Err()
}

func FindTargetIndex(target string, playlist []string) (int, error) {
	for i, line := range playlist {
		if strings.Contains(line, target) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("target not found in the playlist")
}

func deletePrefix(fromDir []string, prefix string) []string {
	var remade []string
	for _, f := range fromDir {
		//fmt.Println("현재경로:", f)
		if strings.HasPrefix(f, prefix) {
			remade = append(remade, f[len(prefix):])
		} else {
			remade = append(remade, f)
		}
	}

	return remade
}
