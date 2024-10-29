package shuffle

import (
	"math/rand/v2"
	"path/filepath"
	"sort"
)

// 파일목록과 현재 폴더 내용을 기반으로 셔플한 새 목록 생성
func ListRebuilder(본것들, 파일목록 []string) ([]string, []string) {
	resultList := []string{"#EXTM3U"}
	삭제된것들, 아직안본것들 := compareDifference(본것들, 파일목록)

	// loginfo len 삭제된것들
	resultList = append(resultList, deleteFromList(본것들, 삭제된것들)...)

	// loginfo len 아직안본것들
	shuffled := shuffleFilePaths(아직안본것들)

	// loginfo len result

	return append(resultList, shuffled...), shuffled
}

// 1. 양 목록의 비교
func compareDifference(arr1, arr2 []string) ([]string, []string) {
	inArr1NotInArr2 := []string{}
	inArr2NotInArr1 := []string{}

	existsInArr2 := make(map[string]bool)

	// Store all elements of arr2 in a map for quick lookup
	for _, value := range arr2 {
		existsInArr2[value] = true
	}

	// Find elements in arr1 that are not in arr2
	for _, value := range arr1 {
		if !existsInArr2[value] {
			inArr1NotInArr2 = append(inArr1NotInArr2, value)
		}
	}

	// Now check for elements in arr2 that are not in arr1
	existsInArr1 := make(map[string]bool)
	for _, value := range arr1 {
		existsInArr1[value] = true
	}

	for _, value := range arr2 {
		if !existsInArr1[value] {
			inArr2NotInArr1 = append(inArr2NotInArr1, value)
		}
	}

	return inArr1NotInArr2, inArr2NotInArr1
}

// 2. 구목록에서 삭제된 항목 제거
func deleteFromList(본것들, 삭제된것들 []string) []string {
	삭제대상 := make(map[string]bool, len(삭제된것들))
	for _, removed := range 삭제된것들 {
		삭제대상[removed] = true
	}

	새재생목록 := make([]string, 0, len(본것들))

	for _, url := range 본것들 {
		if !삭제대상[url] {
			새재생목록 = append(새재생목록, url)
		}
	}

	return 새재생목록
}

// 3. 목록 뒤섞기
func shuffleFilePaths(paths []string) []string {
	// Step 1: Map directories to filenames and sort filenames
	dirToFile := mapFiles(paths)

	// Step 2: Shuffle the directory keys
	keys := make([]string, 0, len(dirToFile))
	for key := range dirToFile {
		keys = append(keys, key)
	}
	shuffledKeys := shuffleKeys(keys)

	// Step 3: Reconstruct the shuffled file paths
	return reconstructPaths(dirToFile, shuffledKeys)
}

func mapFiles(paths []string) map[string][]string {
	dirToFile := make(map[string][]string)

	for _, path := range paths {
		dirBase := filepath.Dir(path)
		file := filepath.Base(path)
		dirToFile[dirBase] = append(dirToFile[dirBase], file)
	}

	// Sort filenames in each directory
	for dir := range dirToFile {
		sort.Strings(dirToFile[dir])
	}

	return dirToFile
}

func shuffleKeys(keys []string) []string {
	shuffledKeys := make([]string, len(keys))
	copy(shuffledKeys, keys)
	rand.Shuffle(len(shuffledKeys), func(i, j int) {
		shuffledKeys[i], shuffledKeys[j] = shuffledKeys[j], shuffledKeys[i]
	})

	// Ensure no consecutive keys are the same
	for i := 1; i < len(shuffledKeys); i++ {
		if shuffledKeys[i] == shuffledKeys[i-1] {
			// Swap with a non-consecutive key
			for j := i + 1; j < len(shuffledKeys); j++ {
				if shuffledKeys[j] != shuffledKeys[i-1] {
					shuffledKeys[i], shuffledKeys[j] = shuffledKeys[j], shuffledKeys[i]
					break
				}
			}
		}
	}

	return shuffledKeys
}

func reconstructPaths(dirToFile map[string][]string, shuffledKeys []string) []string {
	var shuffledPaths []string
	for _, dir := range shuffledKeys {
		for _, file := range dirToFile[dir] {
			path := filepath.Join("base", dir, file)
			shuffledPaths = append(shuffledPaths, path)
		}
	}
	return shuffledPaths
}
