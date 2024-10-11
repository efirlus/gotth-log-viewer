package cv

// 모든 패키지에서 공통적으로 호출해야하는 것들 모아놓음

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// 각종 파일, 디렉토리 경로
const (
	MDdir    = "/home/efirlus/OneDrive/obsidian/Vault/6. Calibre"
	CoverDir = "/home/efirlus/OneDrive/obsidian/Vault/images"
	Bookdir  = "/NAS/samba/Book"
	DBfile   = "/NAS/samba/Book/metadata.db"
	LogPath  = "/home/efirlus/goproject/Logs/app.log"
)

// 문자열의 대답을 bool 타입으로, 혹은 그 반대로 전환하는 함수.
// 기본적으로 no일 땐 아예 그랩이 안되고 yes일 땐 Yes로 뜸.
// Completion 필드 값 찾을 때 쓰는 함수
func YesNo(input interface{}) interface{} {
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

// c-1. commentLinker replaces occurrences of [[some phrase]] with <a href=somelink>some phrase</a>
func CommentLinker(comments string) (string, error) {
	var err error
	var link string
	// Regex pattern to find [[some phrase]]
	re := regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)

	// Replace each occurrence with the HTML link
	result := re.ReplaceAllStringFunc(comments, func(match string) string {
		// Extract the phrase inside [[ ]]
		phrase := re.FindStringSubmatch(match)[1]
		// Generate the link
		link, err = linkGenerator(phrase)

		// Return the replacement string
		return fmt.Sprintf(`<a href="%s">%s</a>`, link, phrase)
	})

	if err != nil {
		return "", fmt.Errorf("링크에 정규식 적용 실패: %v", err)
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
			return "", fmt.Errorf("md파일 확인 실패: %v", err)
		}
		return fmt.Sprintf("calibre://book-details/_hex_-426f6f6b/%d", id), nil

	case false:
		hexcode := hex.EncodeToString([]byte(phrase))
		return fmt.Sprintf("calibre://show-note/Book/authors/hex_%s", hexcode), nil
	}

	return "", fmt.Errorf("주소 생성 실패")
}

// c-2-1. 링크 생성을 위한 파일 확인
func fileExists(filename string) bool {
	fullPath := filepath.Join(MDdir, filename)
	_, err := os.Stat(fullPath)
	return err == nil
}

// c-2-2. id 추출
func grabIDofMD(filename string) (int, error) {
	filePath := filepath.Join(MDdir, filename)
	var id int

	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("md파일 열기 실패: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "> [id:: ") {
			rawID := line[7 : len(line)-1]

			id, err = strconv.Atoi(strings.TrimSpace(rawID))
			if err != nil {
				return 0, fmt.Errorf("id 숫자 추출 실패: %v", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("md파일 스캔 실패: %v", err)
	}

	return id, nil
}
