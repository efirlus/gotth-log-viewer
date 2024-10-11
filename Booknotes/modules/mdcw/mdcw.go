package mdcw

// 원드라이브로 동기화된 마크다운파일의 내용을
// 칼리버 데이터베이스에 반영하는 핸들러

import (
	cmdr "booknotes/modules/commander"
	"booknotes/modules/cv"
	"booknotes/modules/lg"
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	gd "github.com/yuin/goldmark"
)

// 메타데이타 스트럭트
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

// 0. 이벤트핸들러는 이 함수만 호출
func MDCreateWrite(filename string) {
	// 1. 이벤트가 뜬 파일의 Metadata를 읽어들인다
	BookMD, err := getBookInfo(filename)
	if err != nil {
		lg.Err("cannot get information from markdown", err)
	}

	// 2. 아규먼트 생성
	arglist := buildArgs(BookMD)

	for _, arg := range arglist {
		exc := cmdr.SetMD(BookMD.ID, arg)
		_, err := cmdr.CommandExec(exc)
		if err != nil {
			lg.Err("cannot execute set_metadata command", err)
		}

		exc = nil
	}

	BookMD = BookMetadata{}
	arglist = nil
}

// 1. 옵시디언으로 수정된 북노트에서 메타데이타를 추출해내는 함수
func getBookInfo(filename string) (BookMetadata, error) {
	var bookdata BookMetadata
	bookdata.Title = getFileName(filename)

	file, err := os.Open(filename)
	if err != nil {
		return BookMetadata{}, fmt.Errorf("노트 파일 열기 실패: %v", err)
	}
	defer file.Close()

	var totalContents string // 내용 전체를 담을 버퍼

	scanner := bufio.NewScanner(file) // 한줄읽기
	for scanner.Scan() {
		line := scanner.Text()
		totalContents += line + "\n" // 읽으면서 버퍼에 담기

		// 시리즈 - 인덱스
		if strings.Contains(line, "[[") && strings.Contains(line, "시리즈의") {
			// Handle series info
			seriesPart := strings.Split(line, "시리즈의")
			bookdata.Series = strings.Trim(seriesPart[0], "[] ")
			index, _ := strconv.Atoi(strings.TrimSpace(seriesPart[1])[:1])
			bookdata.SeriesIndex = index
		}

		// 이하 각각 내용 캐치
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
			completion := cv.YesNo(line[21 : len(line)-1])
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

	// 상태 태그 - regex로
	regextag := regexp.MustCompile(`\btags:\s*\n\s*- .*?\n`).FindAllString(totalContents, -1)
	rawTag := regextag[0]

	if strings.Contains(rawTag, "📗") {
		bookdata.Status = "읽음"
	} else if strings.Contains(rawTag, "📘") {
		bookdata.Status = "안읽음"
	} else if strings.Contains(rawTag, "🔖") {
		bookdata.Status = "대기중"
	} else if strings.Contains(rawTag, "📖") {
		bookdata.Status = "읽는중"
	} else if strings.Contains(rawTag, "📕") {
		bookdata.Status = "읽다맘"
	}

	// 코멘트 파트는 미리 담아놓은 버퍼에서 strings.cut으로
	separator := "***"
	_, rawComments, _ := strings.Cut(totalContents, separator)

	// 그 다음에 html화
	bookdata.Comments = convertToHTML(rawComments)

	// scanner 닫고
	if err := scanner.Err(); err != nil {
		return BookMetadata{}, fmt.Errorf("노트 파일 스캔 실패: %v", err)
	}

	regextag = nil
	rawTag = ""
	rawComments = ""

	return bookdata, nil
}

// a. 파일명 -> 제목
func getFileName(fullPath string) string {
	// Get the base name of the file (filename with extension)
	fileWithExt := filepath.Base(fullPath)
	// Remove the extension from the file name
	fileName := strings.TrimSuffix(fileWithExt, filepath.Ext(fileWithExt))
	return fileName
}

// c. 커멘트를 html화 - goldmark 이용
func convertToHTML(multilineString string) string {
	// 1. 마크다운 변환
	var buf bytes.Buffer
	if err := gd.Convert([]byte(multilineString), &buf); err != nil {
		lg.Err("코멘트 html화 실패", err)
	}

	// 2. comment link generate
	linked, err := commentLinker(multilineString)
	if err != nil {
		lg.Warn("코멘트 내 링크 생성 실패")
		linked = multilineString
	}

	// Remove <p></p> and <p> </p>
	cleanedResult := strings.ReplaceAll(linked, "<p></p>", "")
	cleanedResult = strings.ReplaceAll(cleanedResult, "<p> </p>", "")
	cleanedResult = strings.ReplaceAll(cleanedResult, "<p>　</p>", "")

	// Join all paragraphs and wrap in a div
	htmlContent := "<div>\n" + cleanedResult + "\n</div>"
	cleanedResult = ""
	linked = ""
	return htmlContent
}

// c-1. commentLinker replaces occurrences of [[some phrase]] with <a href=somelink>some phrase</a>
func commentLinker(comments string) (string, error) {
	// Regex pattern to find [[some phrase]]
	re := regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)
	var err error

	// Replace each occurrence with the HTML link
	result := re.ReplaceAllStringFunc(comments, func(match string) string {
		if err != nil {
			return match
		}
		// Extract the phrase inside [[ ]]
		phrase := re.FindStringSubmatch(match)[1]
		// Generate the link
		var link string
		link, err = linkGenerator(phrase)
		if err != nil {
			return "[no link generated]"
		}
		// Return the replacement string
		return fmt.Sprintf(`<a href="%s">%s</a>`, link, phrase)
	})

	if err != nil {
		return "", err
	}

	return result, nil
}

// c-2. 링크 생성
func linkGenerator(phrase string) (string, error) {

	filename := phrase + ".md"

	tf := fileExists(filename)

	switch tf {
	case true:
		id, err := grabIDofMD(filename)
		if err != nil {
			return "", fmt.Errorf("아이디 확인 실패: %v", err)
		}
		return fmt.Sprintf("calibre://book-details/_hex_-426f6f6b/%d", id), nil

	case false:
		hexcode := hex.EncodeToString([]byte(phrase))
		return fmt.Sprintf("calibre://show-note/Book/authors/hex_%s", hexcode), nil
	}

	filename = ""
	return "", fmt.Errorf("링크 주소 생성 실패")
}

// c-2-1. 링크 생성을 위한 파일 확인
func fileExists(filename string) bool {
	fullPath := filepath.Join(cv.MDdir, filename)
	_, err := os.Stat(fullPath)
	return err == nil
}

// c-2-2. id 추출
func grabIDofMD(filename string) (int, error) {
	filePath := filepath.Join(cv.MDdir, filename)
	var id int

	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("노트 파일 열기 실패: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "> [id:: ") {
			rawID := line[7 : len(line)-1]

			id, err = strconv.Atoi(strings.TrimSpace(rawID))
			if err != nil {
				return 0, fmt.Errorf("노트 id 추출 실패: %v", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("노트 파일 스캔 실패: %v", err)
	}

	return id, nil
}

// 2. set-metadata 인수 빌더, 변형 예정
func buildArgs(metadata BookMetadata) []string {
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
		args = append(args, "#genre:"+metadata.Genre)
	}
	if metadata.Authors != "" {
		args = append(args, "authors:"+metadata.Authors)
	}
	if metadata.Chapter != 0 {
		args = append(args, "#chapter:"+strconv.Itoa(metadata.Chapter))
	}
	if metadata.PubDate != "" {
		args = append(args, "pubdate:"+metadata.PubDate)
	}
	if metadata.Completion {
		args = append(args, "#complete:true")
	} else {
		args = append(args, "#complete:false")
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
		args = append(args, "#status:"+metadata.Status)
	}

	return args
}
