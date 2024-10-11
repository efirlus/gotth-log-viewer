package bin

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	directorypath = "/home/efirlus/.config/gobooknotes" // /home/efirlus/OneDrive/obsidian/Vault/6. Calibre
	binaryfil     = "SavedId.bin"
)

type Entry struct {
	ID    uint16
	Title string
}

// 바이너리 파일을 읽어서 엔트리 스트럭트로 구성
func ReadBinary() ([]Entry, error) {

	filename := filepath.Join(directorypath, binaryfil)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// 바이너리가 없으면 생성
		mdDir := "test/MD"
		entries := FirstWalkDir(mdDir)

		err := SaveBinary(entries, filename)
		if err != nil {
			return nil, err
		}
		return entries, nil
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

		binarylist = append(binarylist, entry)
	}
	return binarylist, nil
}

// 엔트리를 바이너리 파일로 저장하는 함수
func SaveBinary(binarylist []Entry) error {
	filename := filepath.Join(directorypath, binaryfil)

	// 임시로 input에 binDir, 나중엔 const로 처리 예정
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
	}

	return nil
}

// md 디렉토리 한 번 긁기
func FirstWalkDir(mdDir string) []Entry {
	var entries []Entry

	files, err := os.ReadDir(mdDir)
	if err != nil {
		fmt.Errorf("error reading directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".md" {
			continue
		}

		var ent Entry

		filePath := filepath.Join(mdDir, file.Name())

		f, err := os.Open(filePath)
		if err != nil {
			fmt.Errorf("error opening file %s: %v", file.Name(), err)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "> [id:: ") {
				rawID := line[7 : len(line)-1]

				id, err := strconv.Atoi(strings.TrimSpace(rawID))
				if err != nil {
					fmt.Errorf("error converting id to int in file %s: %v", file.Name(), err)
				}
				ent.ID = uint16(id)

				ent.Title = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

				entries = append(entries, ent)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Errorf("error scanning file %s: %v", file.Name(), err)
		}
	}

	return entries
}

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
