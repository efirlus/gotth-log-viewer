package commander

import (
	"fmt"
	"os/exec"
	"strconv"
)

func commandBuilder(additional []string) []string {
	basic := []string{"calibredb", "--with-library=http://localhost:8080", "--username", "efirlus", "--password", "<f:/home/efirlus/calpass/calpass>"}
	return append(basic, additional...)
}

func Listdb() []string {
	addit := []string{"list", "-f", "id, title", "--separator", "=-=-="}
	return commandBuilder(addit)
}

func ShowMD(id int) []string {
	sid := strconv.Itoa(id)
	addit := []string{"show_metadata", sid}
	return commandBuilder(addit)
}

func SetMD(id, keyval string) []string {
	addit := []string{"set_metadata", id, "-f", keyval}
	return commandBuilder(addit)
}

func CommandExec(comm []string) (string, error) {
	cmd := exec.Command(comm[0], comm[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ERROR: %v", err)
	}

	return string(output), nil
}
