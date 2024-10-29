package indexing

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 목표 디렉토리를 훑으며 파일목록을 반환
func DirectoryLister(targetDir string) ([]string, error) {
	폴더 := make([]string, 0)
	iter := len(targetDir)

	err := filepath.Walk(targetDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && isVideoFile(path) {
				폴더 = append(폴더, path[iter:])
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	// loginfo len(폴더) parsed
	return 폴더, nil
}

func isVideoFile(path string) bool {
	//동영상인지 체크하는 함수
	extensions := []string{".mp4", ".mkv", ".avi", ".m4v", ".mov", ".wmv", ".webm"}
	for _, ext := range extensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
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
