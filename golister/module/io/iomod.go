package iomod

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	conf "golister/module/conf"
	lg "golister/module/lg"
)

// m3u을 백업, dpl을 읽고 두 파일을 처리한 뒤 재생목록을 반환
func BuildMediaList(paths *conf.PathConfig) []string {
	원본파일 := paths.MediaDirectory + "/" + paths.ModName + ".m3u"

	// m3u를 읽어들여서 m3u변수로 저장
	m3u, err := fileInitiate(원본파일, paths.Backup)
	if err != nil {
		lg.Err(fmt.Sprintf("[ %s ] 백업 실패", 원본파일), err)
		return nil
	}

	// dpl 파일 체크
	마지막시청, err := whereIReadBefore(paths.IndexFile)
	if err != nil {
		lg.Err("dpl 읽기 실패", err)
		return nil
	}

	// index number 얻어내기
	indexnumber, err := indexOfMark(마지막시청, m3u)
	if err != nil {
		lg.Err("인덱스 확인 실패", err)
		return nil
	}

	재생목록 := m3u[:indexnumber+1]
	//loginfo "indexnumber 번째 파일 마지막시청"

	return 재생목록
}

// 각 리스트를 재생성하기 전에 원래 걸 백업해두고, 내용은 변수화 하는 함수
func fileInitiate(원본파일경로, 백업파일경로 string) ([]string, error) {
	in, err := os.Open(원본파일경로)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no original file to backup: %v", err)
		}
		return nil, err
	}
	defer in.Close()

	//대상 만들기
	out, err := os.Create(백업파일경로)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	fmt.Println("백업파일 새로 생성")

	// io.copy - io.sync보다 백만배 훌륭한 bufio writer
	var lines []string

	writer := bufio.NewWriter(out)
	scanner := bufio.NewScanner(in)

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)

		if _, err := writer.WriteString(line + "\n"); err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Ensure everything is flushed to the backup file
	if err := writer.Flush(); err != nil {
		return nil, err
	}

	return lines, nil
}

// dpl인덱스에 해당하는 값을 기준으로
func whereIReadBefore(filepath string) (string, error) {
	//open the dpl file
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var targetPath string

	// regex trimming for prefix
	r := regexp.MustCompile(`^playname=[a-zA-Z]:\\[^\\]+\\`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if r.MatchString(line) {
			// Extract the path part
			targetPath = r.ReplaceAllString(line, "")
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	if targetPath == "" {
		return "", fmt.Errorf("no matching line found")
	}

	return targetPath, nil
}

func indexOfMark(target string, playlist []string) (int, error) {
	for i, line := range playlist {
		if strings.Contains(line, target) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("target not found in the playlist")
}

// 재생목록 파일 생성함수
func CreatePlayList(result []string, mobres []string, dirpath, modname string) error {
	playList := dirpath + "/" + modname + ".m3u"
	mobilePlayList := dirpath + ".m3u"
	var erc int

	err := writeLines(result, playList, "")
	if err != nil {
		erc++
		return fmt.Errorf("정규 플레이리스트 생성 실패: %v", err)
	}

	err = writeLines(mobres, mobilePlayList, modname)
	if err != nil {
		erc++
		return fmt.Errorf("모바일 플레이리스트 생성 실패: %v", err)
	}

	if erc == 0 {
		lg.Info(fmt.Sprintf("%s 목록 전부 생성 완료", modname))
	}

	return nil
}

func writeLines(res []string, path string, mod string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	for _, line := range res {
		if mod != "" {
			_, err := fmt.Fprintln(w, mod+"/"+line)
			if err != nil {
				return err
			}
		} else {
			_, err := fmt.Fprintln(w, line)
			if err != nil {
				return err
			}
		}
	}

	return w.Flush()
}
