package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"

	"testfsnotify/modules/lg"
)

const (
	MDdir    = "/home/efirlus/OneDrive/obsidian/Vault/6. Calibre" //
	CoverDir = "/home/efirlus/OneDrive/obsidian/Vault/images"
	Bookdir  = "/NAS/samba/Book"
	LogPath  = "/home/efirlus/goproject/Logs/app.log"
)

func main() {
	lg.NewLogger("dbSync", LogPath)
	lg.Info("===== Initiated =====")

	// 1. 리스트 추출
	listRune, _ := commandExec(listdb())
	calibreDBMap := processEntries(splitByRuneValue(listRune, 10))

	// 2. 디렉토리 추출
	mdList := grabListofMD()

	// 3. 둘 비교
	notInMD, notInDB := findMissingItems(calibreDBMap, mdList)

	// 4. 지울 건 지우고
	if len(notInDB) > 0 {
		lg.Info(fmt.Sprintf("삭제할 문서들: %v", notInDB))

		for _, title := range notInDB {
			fullpath := filepath.Join(MDdir, title+".md")
			err := os.Remove(fullpath)
			if err != nil {
				lg.Err(fmt.Sprintf("cannot delete note: %s", title), err)
			}
		}
	} else {
		lg.Info("삭제할 문서 없음")
	}

	// 5. 만들건 만들기
	if len(notInMD) > 0 {
		lg.Info(fmt.Sprintf("추가할 문서들: %v", notInMD))

		for k := range notInMD {
			runeres, _ := commandExec(showMD(k))
			bookMD, commentsMap := runeMetadata(splitByRuneValue(runeres, 10))

			err := generateMarkdown(bookMD, commentsMap, k)
			if err != nil {
				lg.Err("마크다운 생성 실패", err)
			}
		}
	} else {
		lg.Info("추가할 문서 없음")
	}
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

// 2. directory list 생성
func grabListofMD() []string {
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

// 5. cmd로 읽어들인 메타데이터 해석
func runeMetadata(runeMap map[int][]rune) (map[string][]rune, map[int][]rune) {
	bookMD := make(map[string][]rune)
	commentsMap := make(map[int][]rune)

	for i, r := range runeMap {
		// ' : ' 패턴을 확인한 뒤 앞뒤로 커팅해서 맵으로 반환
		if containsSubsequence(r, []rune{32, 58, 32}) {
			mapLine := splitByPattern(r, []rune{32, 58, 32})
			mapPrefix := splitByPattern(mapLine[0], []rune{32, 32})

			// 만약 저 패턴이 있는데 코멘트면 코맨트 맵 맨 첫줄
			if string(mapPrefix[0]) == "Comments" {
				commentsMap[0] = mapLine[1]
			} else {
				bookMD[string(mapPrefix[0])] = mapLine[1]
			}
		} else {
			// 패턴 없는 건 싹 코멘트
			commentsMap[i] = r
		}
	}

	return bookMD, commentsMap
}

func containsSubsequence(slice, subsequence []rune) bool {
	if len(subsequence) > len(slice) {
		return false
	}
	for i := 0; i <= len(slice)-len(subsequence); i++ {
		if slices.Equal(slice[i:i+len(subsequence)], subsequence) {
			return true
		}
	}
	return false
}

// 3. 특정 아이디를 받아 그 책의 메타데이타를 개별 문서로 만들어주는 함수
func generateMarkdown(result map[string][]rune, comments map[int][]rune, id int) error {

	// Process the Authors field
	authors := string(splitByPattern(result["Author(s)"], []rune{32, 91})[0])

	// Copying cover Inage
	cover, err := copyAndRenameCover(id)
	if err != nil {
		return fmt.Errorf("커버 파일 확보 실패: %v", err)
	}

	// Convert the Rating to a 10-point scale
	var rating int
	if string(result["Rating"]) != "" {
		rawRating, err := strconv.Atoi(string(result["Rating"]))
		if err != nil {
			return fmt.Errorf("점수 반환 실패: %v", err)
		}
		rating = rawRating * 2
	} else {
		rating = 0
	}

	//시간 포멧 정리
	formattedTimestamp, err := formatTime(string(result["Timestamp"]), "2006-01-02")
	if err != nil {
		if string(result["Timestamp"]) == "" {
			formattedTimestamp = ""
		} else {
			return fmt.Errorf("등록일 계산 실패: %v", err)
		}
	}

	formattedPubdate, err := formatTime(string(result["Published"]), "2006-01-02")
	if err != nil {
		if string(result["Published"]) == "" {
			formattedTimestamp = ""
		} else {
			return fmt.Errorf("출간일 계산 실패: %v", err)
		}
	}

	// Map the Status to an emoji
	var emoji string
	switch string(result["상태"]) {
	case "안읽음":
		emoji = "📘"
	case "읽는중":
		emoji = "📖"
	case "읽음":
		emoji = "📗"
	case "읽다맘":
		emoji = "📕"
	case "대기중":
		emoji = "🔖"
	}

	// 컴플리션 체크
	completion := checkCompletionString(result["완결"])

	// Remove HTML tags from the Comments field
	// Step 1: Extract the keys
	keys := make([]int, 0, len(comments))
	for key := range comments {
		keys = append(keys, key)
	}
	// Step 2: Sort the keys
	sort.Ints(keys)
	// Step 3: Append the arrays in the order of the sorted keys
	var concatComments []rune
	for _, key := range keys {
		concatComments = append(concatComments, comments[key]...)
	}

	mdComments := commentMarkdownizer(string(concatComments))

	// Get the current timestamp for the Created field
	created := time.Now().Format("2006-01-02 15:04")

	// Generate the Markdown content
	var markdownBuilder strings.Builder
	markdownBuilder.WriteString(fmt.Sprintf(`---
tags:
 - %s
Created: %s
Category: "[[책]]"
---`, emoji, created))

	// 시리즈 추가에 대한 조건문 추가
	if s, exists := result["Series"]; exists {
		sMap := splitByPattern(s, []rune{32, 35})

		series := string(sMap[0])
		seriesIndex := string(sMap[1])
		markdownBuilder.WriteString(fmt.Sprintf(`

[[%s]] 시리즈의 %s부
`, series, seriesIndex))

		series = ""
		seriesIndex = ""
	}

	var genre []rune
	if g, exists := result["장르"]; exists {
		genre = g
	} else if g, exists := result["Tags"]; exists {
		genre = g
	}

	markdownBuilder.WriteString(fmt.Sprintf(`

![thumbnail|150](%s)

> [!even-columns] 책 정보
>
>> [!abstract] 개요
>>
>> - [장르:: [[웹소설]]]
>> - [분야:: %s]
>> - [작가:: [[%s]]]
>> - [화수:: %s]
>> - [출간일:: %s]
>
>> [!bookinfo] 읽기
>>
>> - [연재상태:: %s]
>> - [점수:: %d]
>> - [등록일:: [[%s]]]
>> - [읽기 시작한 날:: ]
>> - [다 읽은 날:: ]

> [!metadata]- Calibre Link
> [id:: %d]

***
%s
`, cover, string(genre), authors, string(result["화수"]), formattedPubdate, completion, rating, formattedTimestamp, id, mdComments))

	// 빌드한 마크다운을 스트링으로
	markdown := markdownBuilder.String()

	// 결과값을 줄 별로 나눔
	lines := strings.Split(markdown, "\n")

	// Create the output file
	notefilepath := filepath.Join(MDdir, string(result["Title"])+".md")

	outputFile, err := os.Create(notefilepath)
	if err != nil {
		return fmt.Errorf("빈 마크다운 문서 생성 실패: %v", err)
	}
	defer outputFile.Close()

	// 버퍼를 가동해서 아웃풋 파일을 메모리에 얹음
	writer := bufio.NewWriter(outputFile)

	for _, line := range lines {

		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("마크다운 내용 채우기 실패: %v", err)
		}
	}

	// Flush the writer to ensure all content is written to the file
	writer.Flush()

	return nil
}

// a. CopyAndRenameCover copies the cover image from the source path to the destination path,
// renaming it using the MD5 hash of the source path.
func copyAndRenameCover(bookID int) (string, error) {
	// Construct the glob pattern to find the correct directory
	globPattern := fmt.Sprintf("%s/*/* (%d)/cover.*", Bookdir, bookID)

	// Find the matching file path using the glob pattern
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return "", fmt.Errorf("도서 폴더 접근 실패: %v", err)
	}
	if len(matches) == 0 { // no cover page for the book
		return "", nil //fmt.Errorf("no cover file found for ID: %d", bookID)
	}

	// Use the first match (assuming there's only one match)
	sourcePath := matches[0]

	// Open the source file to binary hashing
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("커버 이미지 열기 실패: %v", err)
	}
	defer sourceFile.Close()

	// Generate the MD5 hash for the new filename based on the content
	hash := md5.New()
	if _, err := io.Copy(hash, sourceFile); err != nil {
		return "", fmt.Errorf("커버 이미지 읽기 실패: %v", err)
	}
	extension := filepath.Ext(sourcePath)
	newFileName := hex.EncodeToString(hash.Sum(nil))[:32] + extension
	destinationPath := filepath.Join(CoverDir, newFileName)

	// Re-Open the source file to copy
	sourceFile.Seek(0, io.SeekStart)

	// Create the destination file
	destFile, err := os.Create(destinationPath)
	if err != nil {
		return "", fmt.Errorf("이미지 복사 개시 실패: %v", err)
	}
	defer destFile.Close()

	// Copy the contents of the source file to the destination file
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return "", fmt.Errorf("이미지 복사 실패: %v", err)
	}

	return newFileName, nil
}

// b. 항상 문제가 되는 시간 문자열의 포멧을 관리하는 함수
// 기존 시간 문자열, 그리고 원하는 포멧을 각각 적어넣으면 됨
// 예를 들어 "2020-03-05T21:22:30", "2000-03-04 23:11"
func formatTime(timestamp, targetFormat string) (string, error) {
	//시간 문자열 받기
	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "", fmt.Errorf("시간 해석 실패: %v", err)
	}

	formattedTime := parsedTime.Format(targetFormat)

	return formattedTime, nil
}

// Y가 있으면 완결, 나머진 연재중
func checkCompletionString(r []rune) string {
	if len(r) == 0 {
		return "알 수 없음"
	}
	switch r[0] {
	case 89:
		return "완결"
	case 78:
		return "연재중"
	default:
		return "알 수 없음"
	}
}

// c. 코멘트 html을 마크다운으로 변환
func commentMarkdownizer(html string) (markdown string) {
	// b. 나머지 html 변환
	converter := md.NewConverter("", true, nil)

	tempmarkdown, err := converter.ConvertString(html)
	if err != nil {
		lg.Err("마크다운 문법 생성 실패", err)
	}

	// c. []() 링크를 전부 [[]]로 재변환
	markdown = reverseCommentLinker(replaceBrackets(tempmarkdown))

	return markdown
}

func replaceBrackets(content string) string {
	// Replace escaped brackets with regular brackets
	content = strings.ReplaceAll(content, `\[\[`, `[[`)
	content = strings.ReplaceAll(content, `\]\]`, `]]`)
	return content
}

// c-a. ReverseCommentLinker replaces <a href="...">some phrase</a> with [[some phrase]]
func reverseCommentLinker(comments string) string {
	// Regex to match <a href="...">some phrase</a>
	re := regexp.MustCompile(`\[(.+?)\]\(.+?\)`)

	// Replace each occurrence with [[some phrase]]
	result := re.ReplaceAllString(comments, `[[${1}]]`)

	return result
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

// 1-b. 쇼메타 커맨드
func showMD(id int) []string {
	sid := strconv.Itoa(id)
	addit := []string{"show_metadata", sid}
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
