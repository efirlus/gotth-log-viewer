package dbpr

// db의 변화를 파악한 뒤에 가동할 실질적인 라이브러리
// 일단 list -f로 목록을 받아서 directory와 비교한 후
// 없는거 지우고, 새로 생긴거 만드는 과정을 2개로 엮음
// 리스트 확인 - 지우기 한세트
// 만들기 한세트
// 나눈 이유는 ignoreUntil 때문

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"

	"booknotes/modules/commander"
	"booknotes/modules/cv"
	"booknotes/modules/lg"
)

type Entry struct {
	ID    int
	Title string
}

// 기본적인 메타데이타 스트럭트
type BookMetadata struct {
	Title       string
	Series      string
	SeriesIndex int
	Cover       string
	Genre       string
	Authors     string
	Chapter     int
	PubDate     string
	Completion  bool
	Rating      int
	Timestamp   string
	ID          int
	Comments    string
	Status      string
}

// 1. 리스트 확인, 없어진 거 삭제
func DBhandleFirst() map[int]string {
	// a. db 에서 리스트를 추출, 매핑
	cmd := commander.Listdb()
	rawres, err := commander.CommandExec(cmd)
	if err != nil {
		lg.Err("cannot execute list -f command", err)
	}
	lines := strings.Split(strings.TrimSpace(rawres), "\n")
	dbMap := processEntries(lines)

	// b. 디렉토리 리스트 추출
	mdList := grabListofMD()

	// c. 비교
	notInMD, notInDB := findMissingItems(dbMap, mdList)

	// d. 파일 삭제
	for _, title := range notInDB {
		fullpath := filepath.Join(cv.MDdir, title+".md")
		err := os.Remove(fullpath)
		if err != nil {
			lg.Err(fmt.Sprintf("cannot delete note: %s", title), err)
		}
	}

	cmd = nil
	rawres = ""
	lines = nil
	dbMap = nil
	mdList = nil
	notInDB = nil
	err = nil
	// ignore, create를 위해 fsn으로 return
	return notInMD
}

// a. dblist를 매핑
func processEntries(lines []string) map[string]int {
	results := make(map[string]int)

	for _, line := range lines {
		parts := strings.SplitN(line, "=-=-=", 2)
		if len(parts) != 2 {
			continue // Skip invalid lines
		}

		// Trim spaces from both parts
		idPart := strings.TrimSpace(parts[0])
		title := strings.TrimSpace(parts[1])

		// Extract the ID string
		idStr := strings.TrimSpace(idPart)

		id, err := strconv.Atoi(idStr)
		if err != nil {
			lg.Warn(fmt.Sprintf("%v번 아이디 - %v 인식 불가", idPart, title))
			continue // Skip lines with invalid ID
		}

		results[title] = id

		parts = nil
		idPart = ""
		title = ""
		idStr = ""
		id = 0
	}
	return results
}

// b. directory list 생성
func grabListofMD() []string {
	var dirList []string
	files, err := os.ReadDir(cv.MDdir)
	if err != nil {
		lg.Fatal("md 디렉토리 목록 읽기 실패", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".md" {
			dirList = append(dirList, strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())))
		}
	}

	files = nil
	return dirList
}

// c. 맵과 directory 리스트 비교
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

	sliceSet = nil
	return resultMap, missingStrings
}

// 2. db의 데이터를 추출해 BookMetadata로 저장
func ParseResult(resultString string) (BookMetadata, error) {
	result := BookMetadata{}

	// 일단 shell 텍스트를 줄 단위로 커팅
	lines := strings.Split(resultString, "\n")

	// 정규식으로 내용 추출
	re := regexp.MustCompile(`^\s*(.+?)\s*:\s*(.+?)$`)

	// 코멘트 자르기가 이렇게 어렵다 ㅅㅂ;
	var currentKey string
	var commentsBuilder strings.Builder

	// 커팅된 줄 단위로 루프 돌리기
	for _, line := range lines {
		// 정규식은 일단 re에 식을 저장하고, 원하는 스트링에 match하는 문법을 따름
		match := re.FindStringSubmatch(line)

		// shell은 한 줄마다 줄머리에 필드가 선언됨, 그걸 잘라내서 key, value로 나누는 것
		if len(match) == 3 {
			key, value := match[1], strings.TrimSpace(match[2])

			// 일단 코멘트 파트부터
			if currentKey == "Comments" {
				result.Comments = strings.TrimSpace(commentsBuilder.String())
				currentKey = ""
				commentsBuilder.Reset()
			}

			// 키 별 스위치로 value 저장
			var err error
			switch key {
			case "Title":
				result.Title = value
			case "Series":
				result.Series = value
			case "Cover":
				result.Cover = value
			case "장르":
				result.Genre = value
			case "Tags":
				result.Genre = value
			case "Author(s)":
				result.Authors = strings.Split(value, " [")[0]
			case "화수":
				result.Chapter, err = strconv.Atoi(value)
				if err != nil {
					lg.Err("화수 읽기 실패", err)
				}
			case "PubDate":
				result.PubDate = value
			case "완결":
				completionValue := cv.YesNo(value)
				if completionBool, ok := completionValue.(bool); ok {
					result.Completion = completionBool
				} else {
					err = fmt.Errorf("완결 여부 확인 실패")
				}
				if err != nil {
					lg.Err("화수 읽기 실패", err)
				}
			case "Rating":
				result.Rating, err = strconv.Atoi(value)
				if err != nil {
					lg.Err("점수 읽기 실패", err)
				}
			case "Timestamp":
				result.Timestamp = value
			case "ID":
				result.ID, err = strconv.Atoi(value)
				if err != nil {
					lg.Err("아이디 읽기 실패", err)
				}
			case "Comments":
				currentKey = "Comments"
				commentsBuilder.WriteString(value + "\n")
			case "상태":
				result.Status = value
			default:
				lg.Err("unknown key", fmt.Errorf("%v", key))
			}
		} else if currentKey == "Comments" {
			// 뭔진 몰라도 더 추가
			commentsBuilder.WriteString(strings.TrimSpace(line) + "\n")
		}

		match = nil
	}

	// 마지막까지 코멘트
	if currentKey == "Comments" {
		result.Comments = strings.TrimSpace(commentsBuilder.String())
	}

	if result.Title == "" {
		return result, errors.New("missing required field : title")
	}

	lines = nil
	// 최종 결과
	return result, nil
}

// 3. 특정 아이디를 받아 그 책의 메타데이타를 개별 문서로 만들어주는 함수
func GenerateMarkdown(result BookMetadata) (string, error) {

	// Process the Authors field
	authors := strings.Split(result.Authors, " [")[0]

	// Copying cover Inage
	cover, err := copyAndRenameCover(result.ID)
	if err != nil {
		return "", fmt.Errorf("커버 파일 확보 실패: %v", err)
	}

	// Convert the Rating to a 10-point scale
	rating := result.Rating * 2

	//시간 포멧 정리
	formattedTimestamp, err := formatTime(result.Timestamp, "2006-01-02")
	if err != nil {
		if result.Timestamp == "" {
			formattedTimestamp = ""
		} else {
			return "", fmt.Errorf("while build specific time format for specify registered date: %v", err)
		}
	}

	// Map the Status to an emoji
	var emoji string
	switch result.Status {
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
	completion := cv.YesNo(result.Completion).(string)

	// Remove HTML tags from the Comments field
	comments := commentMarkdownizer(result.Comments)

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
	if result.Series != "" {
		re := regexp.MustCompile(`(.*) #(\d+)$`)
		matches := re.FindStringSubmatch(result.Series)
		if len(matches) > 0 {
			series := strings.TrimSpace(matches[1])
			seriesIndex := strings.TrimSpace(matches[2])
			markdownBuilder.WriteString(fmt.Sprintf(`

[[%s]] 시리즈의 %s부
`, series, seriesIndex))

			series = ""
			seriesIndex = ""
		}
	}

	markdownBuilder.WriteString(fmt.Sprintf(`

![thumbnail|150](%s)

> [!even-olumns] 책 정보
>
>> [!abstract] 개요
>>
>> - [장르:: [[웹소설]]]
>> - [분야:: %s]
>> - [작가:: [[%s]]]
>> - [화수:: %d]
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
`, cover, result.Genre, authors, result.Chapter, result.PubDate, completion, rating, formattedTimestamp, result.ID, comments))

	// 빌드한 마크다운을 스트링으로
	markdown := markdownBuilder.String()

	// 결과값을 줄 별로 나눔
	lines := strings.Split(markdown, "\n")

	// Create the output file
	notefilepath := filepath.Join(cv.MDdir, result.Title+".md")
	outputFile, err := os.Create(notefilepath)
	if err != nil {
		return "", fmt.Errorf("빈 마크다운 문서 생성 실패: %v", err)
	}
	defer outputFile.Close()

	// 버퍼를 가동해서 아웃풋 파일을 메모리에 얹음
	writer := bufio.NewWriter(outputFile)

	for _, line := range lines {

		if _, err := writer.WriteString(line + "\n"); err != nil {
			return "", fmt.Errorf("마크다운 내용 채우기 실패: %v", err)
		}
	}

	// Flush the writer to ensure all content is written to the file
	writer.Flush()

	cover = ""
	rating = 0
	formattedTimestamp = ""
	emoji = ""
	completion = ""
	comments = ""
	created = ""
	lines = nil
	markdown = ""

	return authors, nil
}

// a. CopyAndRenameCover copies the cover image from the source path to the destination path,
// renaming it using the MD5 hash of the source path.
func copyAndRenameCover(bookID int) (string, error) {
	// Construct the glob pattern to find the correct directory
	globPattern := fmt.Sprintf("%s/*/* (%d)/cover.*", cv.Bookdir, bookID)

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
	destinationPath := filepath.Join(cv.CoverDir, newFileName)

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

	globPattern = ""
	matches = nil
	hash = nil
	extension = ""
	newFileName = ""
	destinationPath = ""
	sourceFile = nil
	destFile = nil
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

// c. 코멘트 html을 마크다운으로 변환
func commentMarkdownizer(html string) (markdown string) {
	// html -> markdown 주의점
	//     [[]]를 링크로 처리 못하고 \[\[\]\]로 처리함
	//     그러니까 미리 이걸 <a href> 처리하고 나서 돌려야 함
	//     그 다음에 []() 형식으로 만들어진 링크들을 전부 [[]]로 변환

	// a. <a href> linker
	linked, err := cv.CommentLinker(html)
	if err != nil {
		lg.Err("임시 html 링크 생성 실패", err)
	}

	// b. 나머지 html 변환
	converter := md.NewConverter("", true, nil)

	tempmarkdown, err := converter.ConvertString(linked)
	if err != nil {
		lg.Err("마크다운 문법 생성 실패", err)
	}

	// c. []() 링크를 전부 [[]]로 재변환
	markdown = reverseCommentLinker(tempmarkdown)

	linked = ""
	converter = nil
	tempmarkdown = ""
	return markdown
}

// c-a. ReverseCommentLinker replaces <a href="...">some phrase</a> with [[some phrase]]
func reverseCommentLinker(comments string) string {
	// Regex to match <a href="...">some phrase</a>
	re := regexp.MustCompile(`\[(.+?)\]\(.+?\)`)

	// Replace each occurrence with [[some phrase]]
	result := re.ReplaceAllString(comments, `[[${1}]]`)

	return result
}
