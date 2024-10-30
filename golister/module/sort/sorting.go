package shuffle

import (
	"math/rand/v2"
	"path/filepath"
	"regexp"
	"sort"
)

// 1. 파일목록과 현재 폴더 내용을 기반으로 셔플한 새 목록 생성
func ListRebuilder(본것들, 파일목록 []string, mod string) ([]string, []string) {
	var resultList []string
	var shuffled []string
	// a. compare
	삭제된것들, 아직안본것들 := compareDifference(본것들, 파일목록)

	// b. delete
	resultList = append(resultList, deleteFromList(본것들, 삭제된것들)...)

	// c. shuffle
	switch mod {
	case "MMD":
		shuffled = randomDirectoryOrderFile(아직안본것들)
	case "PMV", "Fancam":
		shuffled = totalRandom(아직안본것들)
	case "AV":
		shuffled = serialRandom(아직안본것들)
	default: // 셔플 안함
		shuffled = 아직안본것들
	}

	return append(resultList, shuffled...), shuffled
}

// 1.a. 양 목록의 비교
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

// 1.b. 구목록에서 삭제된 항목 제거
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

/*
// Helper struct to track directory state
type dirState struct {
	name       string
	files      []string
	fileIndex  int
	totalFiles int
}

// 1.c. 셔플 함수 본체
func randomDirectoryOrderFile(paths []string) []string {
	// Step 1: Map files to directories
	dirToFile := mapFiles(paths)
	lg.Debug("디버그", "받은 파일 갯수", len(paths))

	// Step 2: Get directory keys
	var keys []string
	for dir := range dirToFile {
		keys = append(keys, dir)
	}

	// 키 셔플러, 이건 그냥 rand-shuffle이니까 중요치 않음
	shuffledKeys := make([]string, len(keys))
	copy(shuffledKeys, keys)
	rand.Shuffle(len(shuffledKeys), func(i, j int) {
		shuffledKeys[i], shuffledKeys[j] = shuffledKeys[j], shuffledKeys[i]
	})

	// Step 3: Reconstruct paths with balanced distribution
	return reconstructPaths(dirToFile, shuffledKeys)
}

// 1.c. s1 - 파일의 매핑
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

// 1.c.s3 - 키가 뒤섞인 상황을 기준으로 하는 발류 셔플러
func reconstructPaths(dirToFile map[string][]string, shuffledKeys []string) []string {
	// Initialize directory states
	var dirs []*dirState
	for _, key := range shuffledKeys {
		dirs = append(dirs, &dirState{
			name:       key,
			files:      dirToFile[key],
			fileIndex:  0,
			totalFiles: len(dirToFile[key]),
		})
	}

	var result []string
	totalFiles := getTotalFiles(dirToFile)

	for len(result) < totalFiles {
		// Calculate remaining files for each directory
		activeDirs := make([]*dirState, 0)
		for _, dir := range dirs {
			if dir.fileIndex < dir.totalFiles {
				activeDirs = append(activeDirs, dir)
			}
		}

		if len(activeDirs) == 0 {
			break
		}

		// Sort directories by remaining files ratio (descending)
		sort.Slice(activeDirs, func(i, j int) bool {
			remainingI := float64(activeDirs[i].totalFiles-activeDirs[i].fileIndex) / float64(activeDirs[i].totalFiles)
			remainingJ := float64(activeDirs[j].totalFiles-activeDirs[j].fileIndex) / float64(activeDirs[j].totalFiles)
			return remainingI > remainingJ
		})

		// Select next directory, avoiding consecutive usage if possible
		selectedIndex := 0
		if len(result) > 0 { // Only check for consecutive dirs if we have already added files
			lastDir := result[len(result)-1][:strings.Index(result[len(result)-1], "/")]
			// Try to find a different directory than the last used one
			for i, dir := range activeDirs {
				if dir.name != lastDir {
					selectedIndex = i
					break
				}
			}
		}

		selectedDir := activeDirs[selectedIndex]

		// Add the next file from selected directory
		path := filepath.Join(selectedDir.name, selectedDir.files[selectedDir.fileIndex])
		result = append(result, path)
		selectedDir.fileIndex++
	}

	return result
}
*/

// Directory state structure
type dirState struct {
	name       string
	files      []string
	fileIndex  int
	totalFiles int
}

func reconstructPaths(dirToFile map[string][]string, shuffledKeys []string) []string {
	// Initialize directory states
	var dirs []*dirState
	for _, key := range shuffledKeys {
		dirs = append(dirs, &dirState{
			name:       key,
			files:      dirToFile[key],
			fileIndex:  0,
			totalFiles: len(dirToFile[key]),
		})
	}

	var result []string
	totalFiles := getTotalFiles(dirToFile)

	for len(result) < totalFiles {
		// Calculate remaining files for each directory
		activeDirs := make([]*dirState, 0)
		for _, dir := range dirs {
			if dir.fileIndex < dir.totalFiles {
				activeDirs = append(activeDirs, dir)
			}
		}

		if len(activeDirs) == 0 {
			break
		}

		// Sort directories by remaining files ratio (descending)
		sort.Slice(activeDirs, func(i, j int) bool {
			remainingI := float64(activeDirs[i].totalFiles-activeDirs[i].fileIndex) / float64(activeDirs[i].totalFiles)
			remainingJ := float64(activeDirs[j].totalFiles-activeDirs[j].fileIndex) / float64(activeDirs[j].totalFiles)
			return remainingI > remainingJ
		})

		// Select next directory, avoiding consecutive usage if possible
		selectedIndex := 0
		if len(result) > 0 {
			// Use filepath.Dir to safely extract the directory part
			lastPath := result[len(result)-1]
			lastDir := filepath.Dir(lastPath)

			// Try to find a different directory than the last used one
			for i, dir := range activeDirs {
				if dir.name != lastDir {
					selectedIndex = i
					break
				}
			}
		}

		selectedDir := activeDirs[selectedIndex]

		// Add the next file from selected directory
		path := filepath.Join(selectedDir.name, selectedDir.files[selectedDir.fileIndex])
		result = append(result, path)
		selectedDir.fileIndex++
	}

	return result
}

func mapFiles(paths []string) map[string][]string {
	dirToFile := make(map[string][]string)

	for _, path := range paths {
		// Use filepath.Clean to normalize the path
		cleanPath := filepath.Clean(path)
		dirBase := filepath.Dir(cleanPath)
		file := filepath.Base(cleanPath)
		dirToFile[dirBase] = append(dirToFile[dirBase], file)
	}

	// Sort filenames in each directory
	for dir := range dirToFile {
		sort.Strings(dirToFile[dir])
	}

	return dirToFile
}

func randomDirectoryOrderFile(paths []string) []string {
	// Clean input paths
	cleanPaths := make([]string, len(paths))
	for i, path := range paths {
		cleanPaths[i] = filepath.Clean(path)
	}

	// Step 1: Map files to directories
	dirToFile := mapFiles(cleanPaths)

	// Step 2: Get directory keys
	var keys []string
	for dir := range dirToFile {
		keys = append(keys, dir)
	}

	// Shuffle keys
	shuffledKeys := make([]string, len(keys))
	copy(shuffledKeys, keys)
	rand.Shuffle(len(shuffledKeys), func(i, j int) {
		shuffledKeys[i], shuffledKeys[j] = shuffledKeys[j], shuffledKeys[i]
	})

	// Step 3: Reconstruct paths with balanced distribution
	return reconstructPaths(dirToFile, shuffledKeys)
}

// 파일 양에 상관없는 이븐한 배치를 위한 함수
func getTotalFiles(dirToFile map[string][]string) int {
	total := 0
	for _, files := range dirToFile {
		total += len(files)
	}
	return total
}

// New implementation - totalRandom
func totalRandom(paths []string) []string {
	if len(paths) == 0 {
		return paths
	}

	// Create a copy to avoid modifying original slice
	result := make([]string, len(paths))
	copy(result, paths)

	// First do a complete shuffle
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	// Fix any consecutive directories
	for i := 1; i < len(result); i++ {
		currentDir := filepath.Dir(result[i])
		prevDir := filepath.Dir(result[i-1])

		if currentDir == prevDir {
			// Find next non-consecutive directory to swap with
			for j := i + 1; j < len(result); j++ {
				nextDir := filepath.Dir(result[j])
				// Check if swap won't create new consecutive dirs
				isValidSwap := nextDir != prevDir &&
					(j == len(result)-1 || nextDir != filepath.Dir(result[j+1])) &&
					(i == len(result)-1 || nextDir != filepath.Dir(result[i+1]))

				if isValidSwap {
					result[i], result[j] = result[j], result[i]
					break
				}
			}
		}
	}

	return result
}

func serialRandom(playlist []string) []string {
	// Helper function to extract key and id from filename
	getKeyAndID := func(filename string) (string, string, bool) {
		// Extract just the base filename without directory
		base := filepath.Base(filename)

		// Regular expression to match the pattern
		re := regexp.MustCompile(`^.*?([a-zA-Z]{2,6})-([0-9]{2,5}).*$`)
		matches := re.FindStringSubmatch(base)

		if len(matches) < 3 {
			return base, "", false // Return full filename and indicate no pattern match
		}

		return matches[1], matches[2], true // Return key, numeric ID, and indicate pattern match
	}

	// Separate files into pattern-matching and non-matching groups
	groups := make(map[string][]string)   // For pattern-matching files
	nonMatchingFiles := make([]string, 0) // For files that don't match the pattern

	for _, file := range playlist {
		key, id, matched := getKeyAndID(file)
		if matched {
			groupKey := key + "-" + id
			groups[groupKey] = append(groups[groupKey], file)
		} else {
			nonMatchingFiles = append(nonMatchingFiles, file)
		}
	}

	// Create a slice of unique group keys for randomization
	groupKeys := make([]string, 0, len(groups))
	for k := range groups {
		groupKeys = append(groupKeys, k)
	}

	// Shuffle both the group keys and non-matching files
	rand.Shuffle(len(groupKeys), func(i, j int) {
		groupKeys[i], groupKeys[j] = groupKeys[j], groupKeys[i]
	})
	rand.Shuffle(len(nonMatchingFiles), func(i, j int) {
		nonMatchingFiles[i], nonMatchingFiles[j] = nonMatchingFiles[j], nonMatchingFiles[i]
	})

	// Build the final shuffled playlist
	result := make([]string, 0, len(playlist))

	// Randomly mix pattern-matching groups and non-matching files
	i, j := 0, 0 // indices for groupKeys and nonMatchingFiles
	for i < len(groupKeys) || j < len(nonMatchingFiles) {
		// Randomly choose whether to add a group or a non-matching file
		if j >= len(nonMatchingFiles) || (i < len(groupKeys) && rand.Float32() < 0.5) {
			// Add a group
			if i < len(groupKeys) {
				result = append(result, groups[groupKeys[i]]...)
				i++
			}
		} else {
			// Add a non-matching file
			if j < len(nonMatchingFiles) {
				result = append(result, nonMatchingFiles[j])
				j++
			}
		}
	}

	return result
}
