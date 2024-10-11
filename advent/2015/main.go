package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println(notQuiteLisp2(getInput(1)))

}

func getInput(dayNumber int) string {
	filename := "day-" + strconv.Itoa(dayNumber) + "-input.txt"
	fmt.Println("file name = ", filename)
	contentByte, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	return string(contentByte)
}

func notQuiteLisp1(inp string) int {
	return 0 + strings.Count(inp, "(") - strings.Count(inp, ")")
}

func notQuiteLisp2(inp string) int {
	floor := 0
	for n, p := range inp {
		switch p {
		case 40:
			floor++
		case 41:
			floor--
		}
		if floor == -1 {
			return n + 1
		}
	}
	return 0
}
