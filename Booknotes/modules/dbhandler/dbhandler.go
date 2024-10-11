package dbhandler

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

const (
	directorypath = "/home/efirlus/.config/gobooknotes" // /home/efirlus/OneDrive/obsidian/Vault/6. Calibre
	binaryfile    = "SavedId.bin"
)

type Entry struct {
	ID    uint16
	Title string
	MD5   string
}

func CompareTitle(filename string) (int, error) {
	// 일단 readbinary로 엔트리 생성
	저장된db, err := ReadBinary()
	if err != nil {
		errorMessage := fmt.Sprintf("while read binary for compare title: %v", err)
		return 0, fmt.Errorf(errorMessage)
	}

	// 바이너리 안에 든 title을 key로 하는 id값 매핑
	TitlefromBinarySet := MakeDataSet(저장된db, "title").(map[string]int)

	targetId := TitlefromBinarySet[filename]

	return targetId, nil
}

// binary에서 추출한 엔트리, 또는 obhandler로 walkdirectory한 갑을 맵 인터페이스로 생성.
// 선택지는 4개 / ID = map[uint16]bool, 바이너리 = map[string]bool, 디렉토리 = map[string]string, 제목 = map[int]string
// 출력을 위해 함수 뒤에 .(맵타입셋) 적기
func MakeDataSet(entries []Entry, 맵타입 string) interface{} {
	switch 맵타입 {
	case "ID":
		IDFromBinarySet := make(map[uint16]bool)
		for _, entry := range entries {
			IDFromBinarySet[entry.ID] = true
		}
		return IDFromBinarySet
	case "바이너리":
		MD5FromBinarySet := make(map[string]bool)
		for _, entry := range entries {
			MD5FromBinarySet[entry.MD5] = true
		}
		return MD5FromBinarySet
	case "디렉토리":
		MD5FromDirectorySet := make(map[string]string)
		for _, entry := range entries {
			MD5FromDirectorySet[entry.MD5] = entry.Title
		}
		return MD5FromDirectorySet
	case "제목":
		TitleFromBinarySet := make(map[int]string)
		for _, entry := range entries {
			TitleFromBinarySet[int(entry.ID)] = entry.Title
		}
		return TitleFromBinarySet
	// ----------------------------------------------
	// 위에건 아마도 안쓰면 삭제 예정
	case "title":
		TitleFromBinarySet := make(map[string]int)
		for _, entry := range entries {
			TitleFromBinarySet[entry.Title] = int(entry.ID)
		}
		return TitleFromBinarySet
	default:
		return nil
	}
}

// db list -f id 결과값을 숫자열로 변환하는 함수
func getListFieldId(result string) (map[uint16]bool, error) {
	lines := strings.Split(strings.TrimSpace(result), "\n")
	idMap := make(map[uint16]bool)

	for _, line := range lines[1:] {
		id, err := strconv.Atoi(line)
		if err != nil {
			errorMessage := fmt.Sprintf("while convert the type of id list to uint16: %v", err)
			return nil, fmt.Errorf(errorMessage)
		}
		idMap[uint16(id)] = true
	}

	return idMap, nil
}

// result string을 한줄씩 잘라서 숫자를 추출해서 binary에서 읽어들인 id와 맵을 짜 비교하는 함수.
// 서로에게 없는것만 남긴 2개의 []interface로 출력됨, 해당 값 뒤에 .()로 만들어 쓸것
func CompareBinaryId(result string) ([]interface{}, []interface{}, error) {
	CalibreIDset, err := getListFieldId(result)
	if err != nil {
		errorMessage := fmt.Sprintf("while extract id list and mapping these: %v", err)
		return nil, nil, fmt.Errorf(errorMessage)
	}

	저장된db, err := ReadBinary()
	if err != nil {
		errorMessage := fmt.Sprintf("while read binary for compare id: %v", err)
		return nil, nil, fmt.Errorf(errorMessage)
	}

	dbIDset := MakeDataSet(저장된db, "ID").(map[uint16]bool)

	notInBinary, notInCalibre := compareMap(dbIDset, CalibreIDset)
	return notInBinary, notInCalibre, nil
}

// -------------------------------------------------------------

func SaveIdToBinary(id int, Title string, MD5 string) error {
	저장된db, err := ReadBinary()
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("while read binary for saving id: %v", err))
	}

	추가된db := addtoBinary(저장된db, id, Title, MD5)

	err = saveBinary(추가된db)
	if err != nil {
		errorMessage := fmt.Sprintf("while save appended list to binary: %v", err)
		return fmt.Errorf(errorMessage)
	}

	return nil
}

// 책 아이디를 바이너리에서 빼는 함수
func DeleteIdFromBinary(id int) error {
	저장된db, err := ReadBinary()
	if err != nil {
		errorMessage := fmt.Sprintf("while read binary for deleting id: %v", err)
		return fmt.Errorf(errorMessage)
	}

	감소된db := substoBinary(저장된db, id)

	err = saveBinary(감소된db)
	if err != nil {
		errorMessage := fmt.Sprintf("while save reduced list to binary: %v", err)
		return fmt.Errorf(errorMessage)
	}

	return nil
}

// walkdir로 갈취한 md5-title엔트리와 binary에 저장된 id-md5엔트리를 비교.
// 최종적으로 notinbinary, notindirectory 분류해내기
func CompareMD5(MD5fromDirectory []Entry) ([]int, []Entry, error) {

	// 일단 readbinary로 엔트리 생성
	저장된db, err := ReadBinary()
	if err != nil {
		errorMessage := fmt.Sprintf("while read binary for compare md5: %v", err)
		return nil, nil, fmt.Errorf(errorMessage)
	}

	// 두 개의 맵 제작
	MD5fromDirectorySet := MakeDataSet(MD5fromDirectory, "디렉토리").(map[string]string)
	MD5fromBinarySet := MakeDataSet(저장된db, "바이너리").(map[string]bool)

	// 컴페어 함수 돌리기
	저장돼있던노트, 수정된노트 := compareMap(MD5fromBinarySet, MD5fromDirectorySet)

	// 제목 찾기
	수정된노트엔트리, err := getValuefromMap(MD5fromDirectorySet, 수정된노트)
	if err != nil {
		errorMessage := fmt.Sprintf("while match note title with modified md5: %v", err)
		return nil, nil, fmt.Errorf(errorMessage)
	}

	// 아이디 찾기
	지워지진않은아이디 := FindLostEntryIDs(저장된db, 저장돼있던노트)

	return 지워지진않은아이디, 수정된노트엔트리, nil
}

// 바이너리 파일을 읽어서 엔트리 스트럭트로 구성
func ReadBinary() ([]Entry, error) {

	filename := filepath.Join(directorypath, binaryfile)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// 바이너리가 없으면 생성
		file, err := os.Create(filename)
		if err != nil {
			errorMessage := fmt.Sprintf("while create a new binary file:  %v", err)
			return nil, fmt.Errorf(errorMessage)
		}
		defer file.Close()
	}

	file, err := os.Open(filename) // 있으면 오픈
	if err != nil {
		errorMessage := fmt.Sprintf("while open a existing binary file:  %v", err)
		return nil, fmt.Errorf(errorMessage)
	}
	defer file.Close()

	var binarylist []Entry
	for {
		var entry Entry

		// 아이디 읽기
		err := binary.Read(file, binary.LittleEndian, &entry.ID)
		if err == io.EOF {
			break
		}
		if err != nil {
			errorMessage := fmt.Sprintf("while get the single id in binary:  %v", err)
			return nil, fmt.Errorf(errorMessage)
		}

		// title 길이를 읽기
		var titleLen uint16
		err = binary.Read(file, binary.LittleEndian, &titleLen)
		if err != nil {
			errorMessage := fmt.Sprintf("while get the space for the single title in binary:  %v", err)
			return nil, fmt.Errorf(errorMessage)
		}

		// 제목 읽기
		titleBytes := make([]byte, titleLen)
		_, err = file.Read(titleBytes)
		if err != nil {
			errorMessage := fmt.Sprintf("while get the single title in binary:  %v", err)
			return nil, fmt.Errorf(errorMessage)
		}
		entry.Title = string(titleBytes)

		// md5 길이를 읽기
		var md5Len uint16
		err = binary.Read(file, binary.LittleEndian, &md5Len)
		if err != nil {
			errorMessage := fmt.Sprintf("while get the space for the single md5 in binary:  %v", err)
			return nil, fmt.Errorf(errorMessage)
		}

		// md5 읽기
		md5Bytes := make([]byte, md5Len)
		_, err = file.Read(md5Bytes)
		if err != nil {
			errorMessage := fmt.Sprintf("while get the single md5 in binary:  %v", err)
			return nil, fmt.Errorf(errorMessage)
		}
		entry.MD5 = string(md5Bytes)

		binarylist = append(binarylist, entry)
	}
	return binarylist, nil
}

// 순서에 맞추어 엔트리 하나를 추가하는 함수
func addtoBinary(existEntries []Entry, id int, title, md5 string) []Entry {
	uid := uint16(id)
	for _, entry := range existEntries {
		if entry.ID == uid {
			return existEntries // Entry with the same ID already exists, return unchanged
		}
	}
	newBinary := Entry{
		ID:    uid,
		Title: title,
		MD5:   md5,
	}
	return append(existEntries, newBinary) // Append the new Entry if it doesn't exist
}

// 순서에 맞게 엔트리 하나를 빼는 함수
func substoBinary(entries []Entry, id int) []Entry {
	var result []Entry
	for _, entry := range entries {
		if entry.ID != uint16(id) {
			result = append(result, entry) // Include all entries except the one with the given ID
		}
	}
	return result
}

// 엔트리를 바이너리 파일로 저장하는 함수
func saveBinary(binarylist []Entry) error {
	filename := filepath.Join(directorypath, binaryfile)
	file, err := os.Create(filename)
	if err != nil {
		errorMessage := fmt.Sprintf("while create a new binary: %v", err)
		return fmt.Errorf(errorMessage)
	}
	defer file.Close()

	for _, entry := range binarylist {
		// Write the ID
		err := binary.Write(file, binary.LittleEndian, entry.ID)
		if err != nil {
			errorMessage := fmt.Sprintf("while write a single id to binary: %v", err)
			return fmt.Errorf(errorMessage)
		}

		// Write the title length
		titleBytes := []byte(entry.Title)
		titleLen := uint16(len(titleBytes))
		err = binary.Write(file, binary.LittleEndian, titleLen)
		if err != nil {
			errorMessage := fmt.Sprintf("while make a space for a single title to binary: %v", err)
			return fmt.Errorf(errorMessage)
		}

		// Write the title
		_, err = file.Write(titleBytes)
		if err != nil {
			errorMessage := fmt.Sprintf("while write a single title to binary: %v", err)
			return fmt.Errorf(errorMessage)
		}

		// Write the md5 length
		md5Bytes := []byte(entry.MD5)
		md5Len := uint16(len(md5Bytes))
		err = binary.Write(file, binary.LittleEndian, md5Len)
		if err != nil {
			errorMessage := fmt.Sprintf("while make a space for a single md5 to binary: %v", err)
			return fmt.Errorf(errorMessage)
		}

		// Write the md5
		_, err = file.Write(md5Bytes)
		if err != nil {
			errorMessage := fmt.Sprintf("while write a single md5 to binary: %v", err)
			return fmt.Errorf(errorMessage)
		}
	}

	return nil
}

// 어떤 타입이던 비교 가능한 맵비교 함수.
// 입력은 반드시 맵이어야 하며, 둘의 key 타입만큼은 똑같아야 함.
// 출력은 자동적으로 key 타입의 어레이값으로 나옴
func compareMap(바이너리맵 interface{}, 비교대상맵 interface{}) ([]interface{}, []interface{}) {
	바이너리에없는것 := []interface{}{}
	비교대상에없는것 := []interface{}{}

	// Use reflect to check if inputs are maps and iterate over them
	바이너리맵값 := reflect.ValueOf(바이너리맵)
	비교대상맵값 := reflect.ValueOf(비교대상맵)

	// Ensure both inputs are maps
	if 바이너리맵값.Kind() != reflect.Map || 비교대상맵값.Kind() != reflect.Map {
		fmt.Println("Both arguments must be maps.")
		return nil, nil
	}

	// Check map types
	if 바이너리맵값.Type().Key() != 비교대상맵값.Type().Key() {
		fmt.Println("Both maps must have the same key type.")
		return nil, nil
	}

	// Compare map1 with map2
	for _, key := range 바이너리맵값.MapKeys() {
		if !비교대상맵값.MapIndex(key).IsValid() {
			비교대상에없는것 = append(비교대상에없는것, key.Interface())
		}
	}

	// Compare map2 with map1
	for _, key := range 비교대상맵값.MapKeys() {
		if !바이너리맵값.MapIndex(key).IsValid() {
			바이너리에없는것 = append(바이너리에없는것, key.Interface())
		}
	}

	return 비교대상에없는것, 바이너리에없는것
}

// 맵, 그리고 어레이를 넣으면 어레이를 키로 하는 맵의 값이 출력됨
func getValuefromMap(맵 map[string]string, 어레이 interface{}) ([]Entry, error) {
	var entries []Entry
	keys := 어레이.([]interface{})

	for _, key := range keys {
		// Assert that the key is of type string
		strKey, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("key type mismatch: expected string, got %T", key)
		}

		// Check if the key exists in the map
		if value, found := 맵[strKey]; found {
			var entry Entry

			entry.MD5 = strKey
			entry.Title = value

			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// Function to find the entry IDs corresponding to lost MD5s
func FindLostEntryIDs(entries []Entry, lostMd5 interface{}) []int {
	lostIDs := []int{}
	lmd5 := lostMd5.([]interface{})
	md5ToID := make(map[string]int)

	// Create a map of md5 to entry id
	for _, e := range entries {
		md5ToID[e.MD5] = int(e.ID)
	}

	// Compare lostMd5 with the md5 in the entries
	for _, md5 := range lmd5 {
		strmd5, ok := md5.(string)
		if !ok {
			fmt.Printf("key type mismatch: expected string, got %T", md5)
			return nil
		}
		if id, found := md5ToID[strmd5]; found {
			lostIDs = append(lostIDs, id)
		}
	}

	return lostIDs
}

// 맵[아이디]제목 만드는 함수
func TitletoID() (map[int]string, error) {
	저장된db, err := ReadBinary()
	if err != nil {
		errorMessage := fmt.Sprintf("while read binary for compare md5: %v", err)
		return nil, fmt.Errorf(errorMessage)
	}

	result := MakeDataSet(저장된db, "제목").(map[int]string)

	return result, nil
}
