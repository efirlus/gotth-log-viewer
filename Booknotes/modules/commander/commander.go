package commander

import (
	"fmt"
	"os/exec"
	"strconv"
)

// 0. 기본 커맨드 빌더
func commandBuilder(additional []string) []string {
	basic := []string{"calibredb", "--with-library=http://localhost:8080", "--username", "efirlus", "--password", "<f:/home/efirlus/calpass/calpass>"}
	return append(basic, additional...)
}

// 1-a. 리스트 커맨드
func Listdb() []string {
	addit := []string{"list", "-f", "id, title", "--separator", "=-=-="}
	return commandBuilder(addit)
}

// 1-b. 쇼메타 커맨드
func ShowMD(id int) []string {
	sid := strconv.Itoa(id)
	addit := []string{"show_metadata", sid}
	return commandBuilder(addit)
}

// 1-c. 셋메타 커맨드
func SetMD(id, keyval string) []string {
	addit := []string{"set_metadata", id, "-f", keyval}
	return commandBuilder(addit)
}

// 2. 커맨드 실행 (에러 반환)
func CommandExec(comm []string) (string, error) {
	cmd := exec.Command(comm[0], comm[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cannot exec command [%v]: %v", comm, err)
	}

	return string(output), nil
}
