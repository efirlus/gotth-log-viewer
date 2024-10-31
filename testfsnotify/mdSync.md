package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	gd "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"

	"testfsnotify/modules/lg"
)

// 각종 파일, 디렉토리 경로
const (
	MDdir   = "/home/efirlus/OneDrive/obsidian/Vault/6. Calibre"
	LogPath = "/home/efirlus/goproject/Logs/app.log"
)

// 메타데이타 스트럭트
type BookMetadata struct {
	Title       string
	Series      []rune
	SeriesIndex rune
	Cover       []rune
	Genre       []rune
	Authors     []rune
	Chapter     []rune
	PubDate     []rune
	Completion  bool
	Rating      []rune
	Timestamp   []rune
	ID          []rune
	Comments    []rune
	Status      string
}

// 0. 이벤트핸들러는 이 함수만 호출
func main() {
	lg.NewLogger("mdSyncExec", LogPath)
	lg.Info("===== initiated =====")

	filename := os.Args[1]

	// 1. 파일 읽기
	runes, err := getNoteContents(filename)
	if err != nil {
		lg.Err("파일 읽기 실패", err)
	}

	// 빈 메타데이터 스트럭트 선언
	// 메모리 절약을 위해 포인터 사용
	var bookdata BookMetadata

	// 풀패스에서 노트제목 추출
	bookdata.Title = strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	lg.Info(fmt.Sprintf("[[%s]] 처리 시작", bookdata.Title))

	//2. 노트 3등분
	// 일단 \n***\n 으로 나누는 코멘트
	cRunes := splitByPattern(runes, []rune{10, 42, 42, 42, 10})

	// 그 다음은 앞부분에서 [!abstract 으로 한번 더 나눔
	mdRunes := splitByPattern(runes, []rune{91, 33, 97, 98, 115, 116, 114, 97, 99, 116})

	// 3. 핵심적인 메타데이터 파트 추출 3단계
	// a. 프론트매터 읽기
	frontmatterRune := splitByRuneValue(mdRunes[0], 10)
	readRuneFrontmatter(frontmatterRune, &bookdata)

	// b. 메타데이터 읽기
	metadataRunes := splitByRuneValue(mdRunes[1], 10)
	readRuneMetadata(metadataRunes, &bookdata)

	// c. 코멘트 읽기
	readRuneComments(cRunes[1], &bookdata)
	lg.Info(fmt.Sprintf("[[%s]]의 메타데이터 파싱 성공", bookdata.Title))

	// 4. 읽어들인 모든 정보를 인수로 만들어서 바로 커맨드
	arglist := buildArgs(bookdata)
	for a := range arglist {
		cmd := setMD(string(bookdata.ID), arglist[a])
		_, err := commandExec(cmd)
		if err != nil {
			lg.Err(fmt.Sprintf("'%s' 입력 실패", arglist[a]), err)
		}
	}

	lg.Info("메타데이터 갱신 완료")
}

// 1. 노트 풀패스를 받아서 열어서 그 내용을 전부 []rune으로 반환
func getNoteContents(filename string) ([]rune, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("파일 열기 실패: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var runes []rune

	for {
		r, _, err := reader.ReadRune()
		if err == io.EOF {
			break // End of file reached, break the loop
		}
		if err != nil {
			return nil, fmt.Errorf("파일 읽기 실패: %v", err) // Return other errors
		}
		runes = append(runes, r) // Append the rune to the slice
	}

	return runes, nil // Return the rune slice and no error
}

// 2. 확실히 rune slice가 더 처리가 쉬움
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

func splitByRuneValue(target []rune, splitRune rune) map[int][]rune {
	result := make(map[int][]rune)
	chunk := []rune{}
	chunkIndex := 1

	for _, val := range target {
		if val == splitRune {

			result[chunkIndex] = chunk
			chunkIndex++

			chunk = []rune{} // Start a new chunk after the split
		} else {
			chunk = append(chunk, val)
		}
	}

	// Add the final chunk if it has any elements
	result[chunkIndex] = chunk

	return result
}

// 3.a. 프론트 매터 읽기
func readRuneFrontmatter(frontmatterRune map[int][]rune, bookdata *BookMetadata) {
	// 딱 2가지만 읽으면 됨
	for i, r := range frontmatterRune {
		// 3째 줄에 이모지
		if i == 3 {
			if slices.Contains(r, 128215) {
				bookdata.Status = "읽음"
			} else if slices.Contains(r, 128216) {
				bookdata.Status = "안읽음"
			} else if slices.Contains(r, 128278) {
				bookdata.Status = "대기중"
			} else if slices.Contains(r, 128214) {
				bookdata.Status = "읽는중"
			} else if slices.Contains(r, 128213) {
				bookdata.Status = "읽다맘"
			}
		}

		// 혹시 '시리즈'라는 글자가 보이는 줄이 있을 때 시리즈 추출
		if containsSubsequence(r, []rune{49884, 47532, 51592, 51032}) {
			bookdata.Series = r[2:slices.Index(r, 93)]
			bookdata.SeriesIndex = r[len(r)-2]
		}
	}
}

// 이거도 ㄷ위에 3개랑 같이 카피해놓기
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

// 3.b. 메타데이터 추출, 어려울 거 없고 라인마다 정해져 있어서 추출 쉬움
// 해당 라인 내에선 :를 찾고, ]를 찾으면 쉬움, 그렇게 인덱스 2개 찾은 다음
// 커팅해버리고, 포인터 스트럭트에 때려박으면 됨
func readRuneMetadata(metadataRune map[int][]rune, bookdata *BookMetadata) {
	bookdata.Authors = metadataRune[5][slices.Index(metadataRune[5], 58)+5 : len(metadataRune[5])-3]
	bookdata.Genre = metadataRune[4][slices.Index(metadataRune[4], 58)+3 : len(metadataRune[4])-1]
	bookdata.Chapter = metadataRune[6][slices.Index(metadataRune[6], 58)+3 : len(metadataRune[6])-1]
	bookdata.Completion = checkCompletionBool(metadataRune[11][slices.Index(metadataRune[11], 58)+3 : len(metadataRune[11])-1])
	bookdata.Rating = convertRatingTenToFive(metadataRune[12][slices.Index(metadataRune[12], 58)+3 : len(metadataRune[12])-1])
	bookdata.ID = metadataRune[18][slices.Index(metadataRune[18], 58)+3 : len(metadataRune[18])-1]
}

// 3.b.ㄱ. 완, 연, 2개의 문자를 찾아 불리언으로 변환
func checkCompletionBool(r []rune) bool {
	if len(r) < 2 {
		return false
	}
	switch r[0] {
	case 50756:
		return true
	case 50672:
		return false
	default:
		return false
	}
}

// 3.b.ㄴ. rune 코드로 된 숫자를 바로 계산해서 바로 룬으로 묶어줌
func convertRatingTenToFive(input []rune) []rune {
	// Convert string to integer
	num, _ := strconv.Atoi(string(input))

	// Perform the division and rounding
	result := int(math.Round(float64(num) / 2))

	// Convert the result back to a slice of ASCII values
	output := []rune(strconv.Itoa(result))

	return output
}

// 3.c 코멘트 처리 함수
func readRuneComments(commentsRune []rune, bookdata *BookMetadata) {
	// ㄱ. 마크다운식 위키링크를 html 링크로 변환
	linkedComments, err := commentLinker(commentsRune)
	if err != nil {
		lg.Err("코멘트 링크 처리 실패", err)
	}

	// ㄴ. 골드마크 패키지로 마크다운화
	// 옵션에 unsafe링크 허가를 넣어줘야 위에서 변환한 링크를 받을 수 있음
	md := gd.New(gd.WithRendererOptions(html.WithUnsafe()))
	var buf bytes.Buffer
	// 이 부분 뿌듯한 부분, 룬을 바로 바이트로 전환
	if err := md.Convert(runesToBytes(linkedComments), &buf); err != nil {
		lg.Err("코멘트 html화 실패", err)
	}

	// 그 뒤 바이트 버퍼를 바로 룬으로 전환
	bookdata.Comments = bytesBufferToRunes(&buf)
}

// 3.c.ㄱ. 위키링크를 html로 변환하는 함수
func commentLinker(comments []rune) ([]rune, error) {
	// 정규식으로 처리함, 사실 정규식이 2배 빠름
	re := regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)
	var err error

	// 리플레이스 펑션, 난 이게 아직도 익숙하지가 않다
	result := re.ReplaceAllStringFunc(string(comments), func(match string) string {
		if err != nil {
			return match
		}
		// 3.c.ㄱ.i - 매치된 걸 클리너로 처리, 정확히는 / 경로처리자와 | 닉네임처리자를 분리
		phrase := phraseCleaner([]rune(re.FindStringSubmatch(match)[1]))

		// 원명으로 링크를 만들고, 닉네임이 있을 땐 닉네임을 링크문으로 지정
		return linkGenerator(phrase)
	})

	if err != nil {
		return nil, err
	}

	return []rune(result), nil
}

// 3.c.ㄱ.i - 프레이즈를 3등분, directory/filename|phrase
// 맵으로 반환
func phraseCleaner(phrase []rune) map[string][]rune {
	start := slices.Index(phrase, 47)
	end := slices.Index(phrase, 124)

	// Default to an empty map with three parts: left, middle, right
	result := map[string][]rune{
		"directory": {},
		"filename":  {},
		"phrase":    {},
	}

	// 디렉토리가 없으면 비움
	if start == -1 {
		start = 0
	} else {
		// Elements before '/' are in the 'left' part
		result["directory"] = phrase[:start]
		start += 1 // Start after '/'
	}

	// 닉네임이 없으면 비움
	if end == -1 {
		end = len(phrase)
	} else {
		// Elements after '|' are in the 'right' part
		result["phrase"] = phrase[end+1:]
	}

	// Elements between '/' and '|' (or adjusted boundaries) are in the 'middle'
	result["filename"] = phrase[start:end]

	return result
}

// 3.c.ㄱ.ii - 링크 생성
// 3등분 된 것 중 directory는 무쓸모, filename을 링크로, phrase를 링크장식문으로
func linkGenerator(phMap map[string][]rune) string {
	// map[string][]rune
	// if "phrase" == nil -> "filename"
	filename := string(phMap["filename"])
	phrase := string(phMap["phrase"])

	if len(phMap["phrase"]) == 0 {
		phrase = string(phMap["filename"])
	}

	var link string
	lg.Info(fmt.Sprintf("이번 링크 대상 = %s", filename))
	switch fileExists(filename) {
	case true:
		id, err := catchIDofLink(filename)
		if err != nil {
			lg.Err("링크 대상 아이디 확인 실패", err)
			return fmt.Sprintf(`<a>%s</a>`, phrase)
		}
		link = fmt.Sprintf("calibre://book-details/_hex_-426f6f6b/%s", string(id))
	case false:
		hexcode := hex.EncodeToString([]byte(filename))
		link = fmt.Sprintf("calibre://show-note/Book/authors/hex_%s", hexcode)
	}

	return fmt.Sprintf(`<a href="%s">%s</a>`, link, phrase)
}

// 3.c.ㄱ.ii-1. 링크 생성을 위한 파일 확인
func fileExists(phrase string) bool {
	fullPath := filepath.Join(MDdir, phrase+".md")
	_, err := os.Stat(fullPath)
	return err == nil
}

// 3.c.ㄱ.ii-2. id 추출
func catchIDofLink(phrase string) ([]rune, error) {
	filename := filepath.Join(MDdir, phrase+".md")
	rs, err := getNoteContents(filename)
	if err != nil {
		return nil, fmt.Errorf("링크 아이디 값을 찾기 위해 파일 읽기 실패: %v", err)
	}
	tr := splitByPattern(rs, []rune{62, 32, 91, 105, 100, 58, 58, 32})[1]
	return tr[:slices.Index(tr, 93)], nil
}

// 룬을 바이트로
func runesToBytes(rs []rune) []byte {
	size := 0
	for _, r := range rs {
		size += utf8.RuneLen(r)
	}

	bs := make([]byte, size)

	count := 0
	for _, r := range rs {
		count += utf8.EncodeRune(bs[count:], r)
	}

	return bs
}

// 바이트 버퍼를 룬으로
func bytesBufferToRunes(buf *bytes.Buffer) []rune {

	estimated := utf8.RuneCount(buf.Bytes())
	rs := make([]rune, 0, estimated)

	for {
		r, size, err := buf.ReadRune()
		if err != nil {
			break
		}
		if r == utf8.RuneError && size == 1 {
			continue // Skip invalid UTF-8
		}
		rs = append(rs, r)
	}

	return rs
}

// 4. set-metadata 인수 빌더
func buildArgs(metadata BookMetadata) []string {
	var args []string

	if metadata.Title != "" {
		args = append(args, "title:"+metadata.Title)
	}
	if metadata.Series != nil {
		args = append(args, "series:"+string(metadata.Series))
	}
	if metadata.SeriesIndex != 0 {
		args = append(args, "seriesindex:"+string(metadata.SeriesIndex))
	}
	if metadata.Genre != nil {
		args = append(args, "#genre:"+string(metadata.Genre))
	}
	if metadata.Authors != nil {
		args = append(args, "authors:"+string(metadata.Authors))
	}
	if metadata.Chapter != nil {
		args = append(args, "#chapter:"+string(metadata.Chapter))
	}
	if metadata.Completion {
		args = append(args, "#complete:true")
	} else {
		args = append(args, "#complete:false")
	}
	if metadata.Rating != nil {
		args = append(args, "rating:"+string(metadata.Rating))
	}
	if metadata.Comments != nil {
		args = append(args, "comments:"+string(metadata.Comments))
	}
	if metadata.Status != "" {
		args = append(args, "#status:"+metadata.Status)
	}

	return args
}

// 0. 기본 커맨드 빌더
func commandBuilder(additional []string) []string {
	basic := []string{"calibredb", "--with-library=http://localhost:8080", "--username", "efirlus", "--password", "<f:/home/efirlus/calpass/calpass>"}
	return append(basic, additional...)
}

// 1-c. 셋메타 커맨드
func setMD(id, keyval string) []string {
	addit := []string{"set_metadata", id, "-f", keyval}
	return commandBuilder(addit)
}

// 2. 커맨드 실행 (에러 반환)
func commandExec(comm []string) (string, error) {
	cmd := exec.Command(comm[0], comm[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cannot exec command [%v]: %v", comm, err)
	}

	return string(output), nil
}
