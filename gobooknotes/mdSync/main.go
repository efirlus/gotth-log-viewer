package main

import (
	"fmt"
	"mdsync/lg"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func main() {
	d := "/home/efirlus/OneDrive/obsidian/Vault/6. Calibre"

	// 1. 리스트 추출
	listRune, _ := commandExec(listdb())
	calibreDBMap := processEntries(splitByRuneValue(listRune, 10))
	fmt.Println("-------------- db 추출 맵 -----------------")
	fmt.Println(calibreDBMap)

	// 2. 디렉토리 추출
	mdList := grabListofMD(d)
	fmt.Println("---------- directory 추출 ----------------")
	fmt.Println(mdList)

	// 3. 둘 비교
	notInMD, notInDB := findMissingItems(calibreDBMap, mdList)

	fmt.Println("--------------- md 없는 것 ---------------")
	fmt.Println(notInMD)
	fmt.Println("--------------- db 없는 것 ---------------")
	fmt.Println(notInDB)
}

// 1. dblist를 매핑
func processEntries(lines map[int][]rune) map[string]int {
	results := make(map[string]int)

	for i, line := range lines {
		if i == 0 {
			continue
		}

		idTitle := splitByPattern(line, []rune{61, 45, 61, 45, 61}) // "=-=-="

		if len(idTitle) > 1 {
			id := slices.DeleteFunc(idTitle[0], func(r rune) bool { // 아이디에서 ' ' 삭제
				return r == 32
			})
			intid, err := strconv.Atoi(string(id))
			if err != nil {
				lg.Err("아이디 정수화 실패", err)
			}
			// 키는 string title, 값은 int id
			results[string(idTitle[1])] = intid
		}
	}
	return results
}

// 라인브레이크용
func splitByRuneValue(target []rune, splitRune rune) map[int][]rune {
	result := make(map[int][]rune)
	chunk := []rune{}
	chunkIndex := 1

	for _, val := range target {
		if val == splitRune {
			if len(chunk) > 0 {
				result[chunkIndex] = chunk
				chunkIndex++
			}
			chunk = []rune{} // Start a new chunk after the split
		} else {
			chunk = append(chunk, val)
		}
	}

	// Add the final chunk if it has any elements
	if len(chunk) > 0 {
		result[chunkIndex] = chunk
	}

	return result
}

func splitByPattern(slice, pattern []rune) map[int][]rune {
	resmap := make(map[int][]rune)
	start := 0
	for i := 0; i <= len(slice)-len(pattern); i++ {
		if slices.Equal(slice[i:i+len(pattern)], pattern) {
			if i > start {
				resmap[0] = slice[start:i]
			}
			start = i + len(pattern)
			i = start - 1 // -1 because the loop will increment i
		}
	}
	if start < len(slice) {
		resmap[1] = slice[start:]
	}
	return resmap
}

func grabListofMD(MDdir string) []string {
	var dirList []string
	files, err := os.ReadDir(MDdir)
	if err != nil {
		lg.Fatal("md 디렉토리 목록 읽기 실패", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".md" {
			dirList = append(dirList, strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())))
		}
	}

	return dirList
}

// 3. 맵과 directory 리스트 비교
func findMissingItems(m map[string]int, slice []string) (map[int]string, []string) {
	// Create a set from the slice for efficient lookup
	sliceSet := make(map[string]bool)
	resultMap := make(map[int]string)
	for _, item := range slice {
		sliceSet[item] = true
	}

	// Find IDs in the map that are not in the slice
	for key, id := range m {
		if _, exists := sliceSet[key]; !exists {
			resultMap[id] = key
		}
	}

	// Find strings in the slice that are not in the map
	missingStrings := []string{}
	for _, item := range slice {
		if _, exists := m[item]; !exists {
			missingStrings = append(missingStrings, item)
		}
	}

	return resultMap, missingStrings
}

// 0. 기본 커맨드 빌더
func commandBuilder(additional []string) []string {
	basic := []string{"calibredb", "--with-library=http://localhost:8080", "--username", "efirlus", "--password", "<f:/home/efirlus/calpass/calpass>"}
	return append(basic, additional...)
}

// 1-a. 리스트 커맨드
func listdb() []string {
	addit := []string{"list", "-f", "id, title", "--separator", "=-=-="}
	return commandBuilder(addit)
}

// 2. 커맨드 실행 (에러 반환)
func commandExec(comm []string) ([]rune, error) {
	cmd := exec.Command(comm[0], comm[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("cannot exec command [%v]: %v", comm, err)
	}

	return []rune(string(output)), nil
}
