package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func main() {
	fmt.Println(doesntHeHaveInternElvesForThis(getInput(5)))
	//fmt.Println("is he nice? : ", naughtyOrNice2("qjhvhtzxzqqjkmpb"))
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

func doesntHeHaveInternElvesForThis(inp string) int {
	goodKids := 0
	lines := strings.Split(inp, "\n")
	for _, line := range lines {
		if naughtyOrNice2(line) {
			goodKids++
		}
	}
	return goodKids
}

func naughtyOrNice1(line string) bool {
	var before rune

	re := regexp.MustCompile(`[aeiou]`)
	match := re.FindAllStringSubmatch(line, -1)

	r := []rune(line)
	hasSec := slices.ContainsFunc(r, func(n rune) bool {
		answer := false
		if before == n {
			answer = true
		} else {
			before = n
		}
		return answer
	})

	hasNaughty, _ := regexp.MatchString(`ab|cd|pq|xy`, line)

	if len(match) >= 3 && hasSec && !hasNaughty {
		return true
	}
	return false
}

func naughtyOrNice2(line string) bool {
	r := []rune(line)
	var hasOoO bool
	var hasDup bool

	for i := range len(r) - 2 {
		if r[i] == r[i+2] {
			hasOoO = true
		}

		if strings.Contains(line[i+2:], string(r[i:i+2])) {
			hasDup = true
		}
	}

	if hasOoO && hasDup {
		return true
	}
	return false
}

// 4일차
func adventCoin(inp string) string {
	h := md5.New()
	io.WriteString(h, inp)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func theIdealStockingStuffer(inp, pref string) int {
	coinage := "abcdef"
	key := 1
	for !strings.HasPrefix(coinage, pref) {
		key++
		coinage = adventCoin(inp + strconv.Itoa(key))
	}
	return key
}

// 3일차
func perfectlySphericalHousesInAVacuum(inp string) int {
	// up = 94, right = 62, down = 118, left = 60
	mapForSanta := make(map[string]int)
	mapForSanta["0/0"] = 1
	currentNS := 0
	currentEW := 0
	direction := []rune(inp)
	for _, coordinate := range direction {
		switch coordinate {
		case 94:
			currentNS++
		case 62:
			currentEW++
		case 118:
			currentNS--
		case 60:
			currentEW--
		}

		currentCoord := fmt.Sprintf("%d/%d", currentNS, currentEW)
		mapForSanta[currentCoord]++
	}

	return len(mapForSanta)
}

func perfectlySphericalHousesInAVacuum2(inp string) int {
	// up = 94, right = 62, down = 118, left = 60
	mapForPresent := make(map[string]int)
	mapForPresent["0/0"] = 1
	santaCurrentNS, santaCurrentEW, robotCurrentNS, robotCurrentEW := 0, 0, 0, 0
	direction := []rune(inp)
	for i, coordinate := range direction {
		if i%2 == 0 {
			switch coordinate {
			case 94:
				santaCurrentNS++
			case 62:
				santaCurrentEW++
			case 118:
				santaCurrentNS--
			case 60:
				santaCurrentEW--
			}
			currentCoord := fmt.Sprintf("%d/%d", santaCurrentNS, santaCurrentEW)
			mapForPresent[currentCoord]++
		} else {
			switch coordinate {
			case 94:
				robotCurrentNS++
			case 62:
				robotCurrentEW++
			case 118:
				robotCurrentNS--
			case 60:
				robotCurrentEW--
			}
			currentCoord := fmt.Sprintf("%d/%d", robotCurrentNS, robotCurrentEW)
			mapForPresent[currentCoord]++
		}
	}

	return len(mapForPresent)
}

// 2일차
func iWasToldThereWouldBeNoMath(inp string) map[int][]int {
	lines := strings.Split(inp, "\n")
	re := regexp.MustCompile(`(\d*)x(\d*)x(\d*)`)

	boxes := make(map[int][]int)
	for i, line := range lines {
		match := re.FindAllStringSubmatch(line, -1)
		var box []int

		l, _ := strconv.Atoi(match[0][1])
		w, _ := strconv.Atoi(match[0][2])
		h, _ := strconv.Atoi(match[0][3])
		box = append(box, l, w, h)
		slices.Sort(box)
		boxes[i] = box
	}

	return boxes
}

func forWrappingPaper(boxes map[int][]int) int {
	var total int
	for _, box := range boxes {
		l := box[0]
		w := box[1]
		h := box[2]

		total = total + (l*w*3 + w*h*2 + l*h*2)
	}
	return total
}

func forRibbon(boxes map[int][]int) int {
	var total int

	for _, box := range boxes {
		total = total + (box[0]*2 + box[1]*2) + (box[0] * box[1] * box[2])
	}
	return total
}

// 1일차
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
