package cv

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

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
	// Regex pattern to find [[some phrase]]
	re := regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)

	// Replace each occurrence with the HTML link
	result := re.ReplaceAllStringFunc(comments, func(match string) string {
		// Extract the phrase inside [[ ]]
		phrase := re.FindStringSubmatch(match)[1]
		// Generate the link
		link, err := linkGenerator(phrase)
		if err != nil {
			return ""
		}
		// Return the replacement string
		return fmt.Sprintf(`<a href="%s">%s</a>`, link, phrase)
	})

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
			return "", err
		}
		return fmt.Sprintf("calibre://book-details/_hex_-426f6f6b/%d", id), nil

	case false:
		hexcode := hex.EncodeToString([]byte(phrase))
		return fmt.Sprintf("calibre://show-note/Book/authors/hex_%s", hexcode), nil
	}

	return "", fmt.Errorf("cannot generate address")
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
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "> [id:: ") {
			rawID := line[7 : len(line)-1]

			id, err = strconv.Atoi(strings.TrimSpace(rawID))
			if err != nil {
				return 0, err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	return id, nil
}

func emptizer(vars ...interface{}) {
	for _, v := range vars {
		val := reflect.ValueOf(v)
		if val.Kind() != reflect.Ptr {
			fmt.Printf("Warning: Cannot empty non-pointer value of type %T\n", v)
			continue
		}
		val = val.Elem() // Dereference the pointer
		switch val.Kind() {
		case reflect.String:
			val.SetString("")
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val.SetInt(0)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val.SetUint(0)
		case reflect.Float32, reflect.Float64:
			val.SetFloat(0)
		case reflect.Bool:
			val.SetBool(false)
		case reflect.Slice, reflect.Map:
			val.Set(reflect.Zero(val.Type()))
		case reflect.Struct:
			for i := 0; i < val.NumField(); i++ {
				field := val.Field(i)
				if field.CanSet() {
					field.Set(reflect.Zero(field.Type()))
				}
			}
		default:
			fmt.Printf("Warning: Unsupported type %T\n", v)
		}
	}
}
