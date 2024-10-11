package obhandler

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	dbhandler "booknotes/modules/dbhandler"
)

// 미래의 편의성을 위해 폴더 경로를 몽땅 모아놓음
const (
	directorypath = "/home/efirlus/OneDrive/obsidian/Vault/6. Calibre" // /home/efirlus/OneDrive/obsidian/Vault/6. Calibre
	CoverDir      = "/home/efirlus/OneDrive/obsidian/Vault/images"
)

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
	ID          string
	Comments    string
	Status      string
}

// md5 스트럭트
type TitleMD5 struct {
	ID    uint16
	Title string
	MD5   string
}

// 노트폴더의 모든 파일의 파일명(경로포함)을 값으로 가지는 md5 키 맵을 생성
func GetAllMD5s() ([]TitleMD5, error) {
	TitleMD5Structs := []TitleMD5{}
	directory := filepath.Join(directorypath)

	// Walk the directory
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errorMessage := fmt.Sprintf("while walking inside of the note directory: %v", err)
			return fmt.Errorf(errorMessage)
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get the MD5 of the file
		md5Hash, err := GetMD5(path[len(directorypath)+1 : len(path)-3])
		if err != nil {
			errorMessage := fmt.Sprintf("while calculate md5 of a note: %v", err)
			return fmt.Errorf(errorMessage)
		}

		// Append the MD5 hash to the map
		TitleMD5Structs = append(TitleMD5Structs, md5Hash)
		return nil
	})
	if err != nil {
		errorMessage := fmt.Sprintf("while approach to the note directory: %v", err)
		return nil, fmt.Errorf(errorMessage)
	}

	return TitleMD5Structs, nil
}

// 노트의 md5 해시값을 받아오는 함수
func GetMD5(filetitle string) (TitleMD5, error) {

	TitleMD5Struct := TitleMD5{}
	filename := filetitle + ".md"
	filepath := filepath.Join(directorypath, filename)

	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		errorMessage := fmt.Sprintf("while open a singlie note to get md5: %v", err)
		return TitleMD5{}, fmt.Errorf(errorMessage)
	}
	defer file.Close()

	// Create a new hash interface to write to
	hash := md5.New()

	// First, hash the filename
	if _, err := io.WriteString(hash, filetitle); err != nil {
		errorMessage := fmt.Sprintf("while hash the file name: %v", err)
		return TitleMD5{}, fmt.Errorf(errorMessage)
	}

	// Copy the file's content into the hash
	if _, err := io.Copy(hash, file); err != nil {
		errorMessage := fmt.Sprintf("while hash the file contents: %v", err)
		return TitleMD5{}, fmt.Errorf(errorMessage)
	}

	// Get the final MD5 checksum in bytes
	hashInBytes := hash.Sum(nil)[:16]

	// Convert the checksum bytes to a string
	md5String := hex.EncodeToString(hashInBytes)

	TitleMD5Struct.Title = filetitle
	TitleMD5Struct.MD5 = md5String

	return TitleMD5Struct, nil
}

// 책을 추가한다는 건 표에 추가, 결과물 파싱, 마크다운 생성 저장, 바이너리 저장이 필요.
// 바이너리는 dbhandle, 나머지는 여기서
func AddBook(result string, BookId int) (string, error) {

	// 일단 raw메타데이타를 파싱
	parsedResult, err := parseResult(result)
	if err != nil {
		errorMessage := fmt.Sprintf("while parse the metadata from raw data of calibredb: %v", err)
		return "", fmt.Errorf(errorMessage)
	}

	// 북노트 저장
	_, err = generateMarkdown(parsedResult, BookId)
	if err != nil {
		errorMessage := fmt.Sprintf("while generate a booknote based on the metadata: %v", err)
		return "", fmt.Errorf(errorMessage)
	}

	return parsedResult.Title, nil
}

// db의 데이터를 추출해 BookMetadata로 저장
func parseResult(resultString string) (BookMetadata, error) {
	result := BookMetadata{}

	// 일단 shell 텍스트를 줄 단위로 커팅
	lines := strings.Split(resultString, "\n")

	// 정규식으로 내용 추출
	re := regexp.MustCompile(`^\s*(.+?)\s*:\s*(.+?)$`)
	htmlRe := regexp.MustCompile(`<.*?>`) // comments part only

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
			switch key {
			case "Title":
				result.Title = value
			case "Series":
				result.Series = value
			case "Cover":
				result.Cover = value
			case "장르":
				result.Genre = value
			case "Author(s)":
				result.Authors = strings.Split(value, " [")[0]
			case "화수":
				fmt.Sscanf(value, "%d", &result.Chapter)
			case "PubDate":
				result.PubDate = value
			case "완결":
				result.Completion = yesno(value).(bool)
			case "Rating":
				fmt.Sscanf(value, "%d", &result.Rating)
			case "Timestamp":
				result.Timestamp = value
			case "ID":
				result.ID = value
			case "Comments":
				currentKey = "Comments"
				commentsBuilder.WriteString(value + "\n")
			case "상태":
				result.Status = value
			}
		} else if currentKey == "Comments" {
			// 뭔진 몰라도 더 추가
			commentsBuilder.WriteString(strings.TrimSpace(line) + "\n")
		}
	}

	// 마지막까지 코멘트
	if currentKey == "Comments" {
		result.Comments = strings.TrimSpace(commentsBuilder.String())
	}

	result.Comments = ReverseCommentLinker(result.Comments)

	// Remove HTML tags from the Comments field
	result.Comments = htmlRe.ReplaceAllString(result.Comments, "")

	// 최종 결과
	return result, nil
}

// 문자열의 대답을 bool 타입으로, 혹은 그 반대로 전환하는 함수.
// 기본적으로 no일 땐 아예 그랩이 안되고 yes일 땐 Yes로 뜸.
// Completion 필드 값 찾을 때 쓰는 함수
func yesno(input interface{}) interface{} {
	switch v := input.(type) {
	case string:
		// Convert the input string to lowercase to make the function case-insensitive
		lowerInput := strings.ToLower(v)

		// Return true if the input is "yes", false if "no"
		if lowerInput == "yes" || lowerInput == "완결" {
			return true
		} else if lowerInput == "no" || lowerInput == "연재중" {
			return false
		}
		// Return nil if the input string is neither "yes" nor "no"
		return false
	case bool:
		// Return "Completed" if true, "Not Yet" if false
		if v {
			return "완결"
		} else {
			return "연재중"
		}
	default:
		// Return nil if the input type is neither string nor bool
		return nil
	}
}

// 특정 아이디를 받아 그 책의 메타데이타를 개별 문서로 만들어주는 함수
func generateMarkdown(result BookMetadata, id int) (string, error) {

	// Process the Authors field
	authors := strings.Split(result.Authors, " [")[0]

	// Copying Cover Inage
	Cover, err := copyAndRenameCover(id)
	if err != nil {
		errorMessage := fmt.Sprintf("while copy the cover of a book: %v", err)
		return "", fmt.Errorf(errorMessage)
	}

	// Convert the Rating to a 10-point scale
	rating := result.Rating * 2

	//시간 포멧 정리
	formattedTimestamp, err := formatTime(result.Timestamp, "2006-01-02")
	if err != nil {
		if result.Timestamp == "" {
			formattedTimestamp = ""
		} else {
			errorMessage := fmt.Sprintf("while build specific time format for specify registered date: %v", err)
			return "", fmt.Errorf(errorMessage)
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
	completion := yesno(result.Completion).(string)

	// Remove HTML tags from the Comments field
	re := regexp.MustCompile(`<.*?>`)
	comments := re.ReplaceAllString(result.Comments, "")
	comments = strings.ReplaceAll(comments, "\n", "\n\n")

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

	// 조건문 추가
	if result.Series != "" {
		re := regexp.MustCompile(`(.*) #(\d+)$`)
		matches := re.FindStringSubmatch(result.Series)
		if len(matches) > 0 {
			series := strings.TrimSpace(matches[1])
			seriesIndex := strings.TrimSpace(matches[2])
			markdownBuilder.WriteString(fmt.Sprintf(`

[[%s]] 시리즈의 %s부
`, series, seriesIndex))
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
`, Cover, result.Genre, authors, result.Chapter, result.PubDate, completion, rating, formattedTimestamp, id, comments))

	// 빌드한 마크다운을 스트링으로
	markdown := markdownBuilder.String()

	// 결과값을 줄 별로 나눔
	lines := strings.Split(markdown, "\n")

	// Create the output file
	notefilepath := filepath.Join(directorypath, result.Title)
	outputFile, err := os.Create(notefilepath + ".md")
	if err != nil {
		errorMessage := fmt.Sprintf("while create an empty booknote: %v", err)
		return "", fmt.Errorf(errorMessage)
	}
	defer outputFile.Close()

	// 버퍼를 가동해서 아웃풋 파일을 메모리에 얹음
	writer := bufio.NewWriter(outputFile)

	for _, line := range lines {

		if _, err := writer.WriteString(line + "\n"); err != nil {
			errorMessage := fmt.Sprintf("while write a note contents: %v", err)
			return "", fmt.Errorf(errorMessage)
		}
	}

	// Flush the writer to ensure all content is written to the file
	writer.Flush()
	if err != nil {
		errorMessage := fmt.Sprintf("while flush the note contents buffer: %v", err)
		return "", fmt.Errorf(errorMessage)
	}

	return authors, nil
}

// CopyAndRenameCover copies the cover image from the source path to the destination path,
// renaming it using the MD5 hash of the source path.
func copyAndRenameCover(bookID int) (string, error) {
	// Construct the glob pattern to find the correct directory
	globPattern := fmt.Sprintf("/NAS/samba/Book/*/* (%d)/cover.*", bookID)

	// Find the matching file path using the glob pattern
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		errorMessage := fmt.Sprintf("while approach to the database directory: %v", err)
		return "", fmt.Errorf(errorMessage)
	}
	if len(matches) == 0 { // no cover page for the book
		return "", nil //fmt.Errorf("no cover file found for ID: %d", bookID)
	}

	// Use the first match (assuming there's only one match)
	sourcePath := matches[0]

	// Open the source file to binary hashing
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		errorMessage := fmt.Sprintf("while open the cover image: %v", err)
		return "", fmt.Errorf(errorMessage)
	}
	defer sourceFile.Close()

	// Generate the MD5 hash for the new filename based on the content
	hash := md5.New()
	if _, err := io.Copy(hash, sourceFile); err != nil {
		errorMessage := fmt.Sprintf("while read a cover image: %v", err)
		return "", fmt.Errorf(errorMessage)
	}
	extension := filepath.Ext(sourcePath)
	newFileName := hex.EncodeToString(hash.Sum(nil))[:32] + extension
	destinationPath := filepath.Join(CoverDir, newFileName)

	// Re-Open the source file to copy
	sourceFile.Seek(0, io.SeekStart)

	// Create the destination file
	destFile, err := os.Create(destinationPath)
	if err != nil {
		errorMessage := fmt.Sprintf("while create a empty image: %v", err)
		return "", fmt.Errorf(errorMessage)
	}
	defer destFile.Close()

	// Copy the contents of the source file to the destination file
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		errorMessage := fmt.Sprintf("while copy a cover image: %v", err)
		return "", fmt.Errorf(errorMessage)
	}

	return newFileName, nil
}

// 항상 문제가 되는 시간 문자열의 포멧을 관리하는 함수
// 기존 시간 문자열, 그리고 원하는 포멧을 각각 적어넣으면 됨
// 예를 들어 "2020-03-05T21:22:30", "2000-03-04 23:11"
func formatTime(timestamp, targetFormat string) (string, error) {
	//시간 문자열 받기
	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		errorMessage := fmt.Sprintf("while decode a timestamp: %v", err)
		return "", fmt.Errorf(errorMessage)
	}

	formattedTime := parsedTime.Format(targetFormat)

	return formattedTime, nil
}

// 제목을 받아서 해당하는 북노트를 삭제하는 함수
func DeleteBook(rawTitle string) error {
	// 파일명을 선언
	fiiename := rawTitle + ".md"
	filepath := filepath.Join(directorypath, fiiename)
	// 노트의 존재를 체크
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// 이미 그 노트가 없으면 그냥 쓰루
		return nil
	}

	err := os.Remove(filepath) // 있으면 삭제
	if err != nil {
		errorMessage := fmt.Sprintf("while delete a single note: %v", err)
		return fmt.Errorf(errorMessage)
	}

	return nil
}

// 옵시디언으로 수정된 북노트에서 메타데이타를 추출해내는 함수
func GetBookInfo(filename string) (BookMetadata, error) {
	var bookdata BookMetadata
	bookdata.Title = filename
	notefilepath := filepath.Join(directorypath, filename)

	file, err := os.Open(notefilepath + ".md")
	if err != nil {
		errorMessage := fmt.Sprintf("while open a booknote to get metadata: %v", err)
		return BookMetadata{}, fmt.Errorf(errorMessage)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var inTags, inBody, inComments bool
	var commentsBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// tags를 발견하면 intag 상태를 가동하고 컨티뉴
		if strings.HasPrefix(line, "tags:") {
			inTags = true
			continue
		}

		if inTags { // 인태그가 가동됐을 때 이모지를 발견하면 상태값 출력
			line = strings.TrimSpace(line)
			if strings.Contains(line, "📗") {
				bookdata.Status = "읽음"
			} else if strings.Contains(line, "📘") {
				bookdata.Status = "안읽음"
			} else if strings.Contains(line, "🔖") {
				bookdata.Status = "대기중"
			} else if strings.Contains(line, "📖") {
				bookdata.Status = "읽는중"
			} else if strings.Contains(line, "📕") {
				bookdata.Status = "읽다맘"
			}

			inTags = false // 상태를 출력한 뒤 바로 인태그 모드 종료
		}

		// Handle transition from frontmatter to body
		if line == "---" && !inBody {
			inBody = true
			continue
		}

		// Handle transition from body to comments
		if line == "***" {
			inComments = true
			continue
		}

		// Handle body content
		if inBody && !inComments {
			if strings.Contains(line, "[[") && strings.Contains(line, "시리즈의") {
				// Handle series info
				seriesPart := strings.Split(line, "시리즈의")
				bookdata.Series = strings.Trim(seriesPart[0], "[] ")
				index, _ := strconv.Atoi(strings.TrimSpace(seriesPart[1])[:1])
				bookdata.SeriesIndex = index
			}

			if strings.HasPrefix(line, ">> - [분야:: ") {
				bookdata.Genre = line[15 : len(line)-1]
			}

			if strings.HasPrefix(line, ">> - [작가:: ") {
				bookdata.Authors = line[17 : len(line)-3]
			}

			if strings.HasPrefix(line, ">> - [화수:: ") {
				bookdata.Chapter, _ = strconv.Atoi(line[15 : len(line)-1])
			}

			if strings.HasPrefix(line, ">> - [연재상태:: ") {
				completion := yesno(line[21 : len(line)-1])
				bookdata.Completion = completion.(bool)
			}

			if strings.HasPrefix(line, ">> - [점수:: ") {
				score, _ := strconv.Atoi(line[15 : len(line)-1])
				bookdata.Rating = score / 2
			}

			if strings.HasPrefix(line, ">> - [등록일:: ") {
				bookdata.Timestamp = line[20 : len(line)-3]
			}

			if strings.HasPrefix(line, "> [id:: ") {
				bookdata.ID = line[7 : len(line)-1]
			}
		}

		// Handle comments content
		if inComments {
			commentsBuilder.WriteString("<p>")
			commentsBuilder.WriteString(line)
			commentsBuilder.WriteString("</p>")
		}
	}

	completeComments := commentEditor(commentsBuilder.String()) // 여러번 빈 줄에 새겨지는 <p> 태그 삭제
	bookdata.Comments = "<div>" + completeComments + "</div>"

	if err := scanner.Err(); err != nil {
		errorMessage := fmt.Sprintf("while read the contents of the note: %v", err)
		return BookMetadata{}, fmt.Errorf(errorMessage)
	}

	return bookdata, nil
}

// 코멘트의 태그를 정리해서 빈 줄이 무한생성되지 않게 만들어주는 함수
func commentEditor(comment string) string {
	var completeComments string

	tempComments := strings.ReplaceAll(comment, "<p></p>", "")
	linkedComments, err := CommentLinker(tempComments)
	if err != nil {
		errorMessage := fmt.Sprintf("Error occurs on: %v", err)
		return errorMessage
	}
	completeComments = strings.ReplaceAll(linkedComments, "<p> </p>", "")
	return completeComments
}

// CommentLinker replaces occurrences of [[some phrase]] with <a href=somelink>some phrase</a>
func CommentLinker(comments string) (string, error) {
	// Regex pattern to find [[some phrase]]
	re := regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)

	// Replace each occurrence with the HTML link
	result := re.ReplaceAllStringFunc(comments, func(match string) string {
		// Extract the phrase inside [[ ]]
		phrase := re.FindStringSubmatch(match)[1]
		// Generate the link
		link, err := LinkGenerator(phrase)
		if err != nil {
			return ""
		}
		// Return the replacement string
		return fmt.Sprintf(`<a href="%s">%s</a>`, link, phrase)
	})

	return result, nil
}

// LinkGenerator creates the link for the given phrase
func LinkGenerator(phrase string) (string, error) {
	제목세트, err := dbhandler.TitletoID()
	if err != nil {
		errorMessage := fmt.Sprintf("Error occurs on: %v", err)
		return "", fmt.Errorf(errorMessage)
	}

	for id, title := range 제목세트 {
		ctitle := strings.ReplaceAll(title, " ", "")
		cphrase := strings.ReplaceAll(phrase, " ", "")
		if strings.EqualFold(cphrase, ctitle) {
			return fmt.Sprintf("calibre://book-details/_hex_-426f6f6b/%d", id), nil
		}
	}
	// if not, make it utf8 hex build
	hexcode := hex.EncodeToString([]byte(phrase))
	return fmt.Sprintf("calibre://show-note/Book/authors/hex_%s", hexcode), nil
}

// ReverseCommentLinker replaces <a href="...">some phrase</a> with [[some phrase]]
func ReverseCommentLinker(comments string) string {
	// Regex to match <a href="...">some phrase</a>
	re := regexp.MustCompile(`<a\s+href="[^"]*">([^<]+)</a>`)

	// Replace each occurrence with [[some phrase]]
	result := re.ReplaceAllString(comments, `[[${1}]]`)

	return result
}

func BuildArgs(metadata BookMetadata) []string {
	var args []string

	if metadata.Title != "" {
		args = append(args, "title:"+metadata.Title)
	}
	if metadata.Series != "" {
		args = append(args, "series:"+metadata.Series)
	}
	if metadata.SeriesIndex != 0 {
		args = append(args, "seriesindex:"+strconv.Itoa(metadata.SeriesIndex))
	}
	if metadata.Cover != "" {
		args = append(args, "cover:"+metadata.Cover)
	}
	if metadata.Genre != "" {
		args = append(args, "genre:"+metadata.Genre)
	}
	if metadata.Authors != "" {
		args = append(args, "authors:"+metadata.Authors)
	}
	if metadata.Chapter != 0 {
		args = append(args, "chapter:"+strconv.Itoa(metadata.Chapter))
	}
	if metadata.PubDate != "" {
		args = append(args, "pubdate:"+metadata.PubDate)
	}
	if metadata.Completion {
		args = append(args, "completion:true")
	} else {
		args = append(args, "completion:false")
	}
	if metadata.Rating != 0 {
		args = append(args, "rating:"+strconv.Itoa(metadata.Rating))
	}
	if metadata.Timestamp != "" {
		args = append(args, "timestamp:"+metadata.Timestamp)
	}
	if metadata.ID != "" {
		args = append(args, "id:"+metadata.ID)
	}
	if metadata.Comments != "" {
		args = append(args, "comments:"+metadata.Comments)
	}
	if metadata.Status != "" {
		args = append(args, "status:"+metadata.Status)
	}

	return args
}
