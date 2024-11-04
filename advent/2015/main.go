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
	fmt.Println(ProbablyAFireHazard2(getInput(6)))
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

func ProbablyAFireHazard1(inp string) int {
	var lightMap [1000][1000]bool
	lines := strings.Split(inp, "\n")
	lighted := 0

	for _, line := range lines {
		re := regexp.MustCompile(`^(.*) (\d*),(\d*) through (\d*),(\d*)$`)
		match := re.FindAllStringSubmatch(line, -1)

		wStart, _ := strconv.Atoi(match[0][2])
		wEnd, _ := strconv.Atoi(match[0][4])
		hStart, _ := strconv.Atoi(match[0][3])
		hEnd, _ := strconv.Atoi(match[0][5])

		for w := wStart; w <= wEnd; w++ {
			for h := hStart; h <= hEnd; h++ {
				switch match[0][1] {
				case "turn on":
					lightMap[w][h] = true
				case "turn off":
					lightMap[w][h] = false
				case "toggle":
					lightMap[w][h] = !lightMap[w][h]
				}
			}
		}
	}

	for w := 0; w <= 999; w++ {
		for h := 0; h <= 999; h++ {
			if lightMap[w][h] {
				lighted++
			}
		}
	}

	return lighted
}

func ProbablyAFireHazard2(inp string) int {
	var lightMap [1000][1000]int
	lines := strings.Split(inp, "\n")
	lighted := 0

	for _, line := range lines {
		re := regexp.MustCompile(`^(.*) (\d*),(\d*) through (\d*),(\d*)$`)
		match := re.FindAllStringSubmatch(line, -1)

		wStart, _ := strconv.Atoi(match[0][2])
		wEnd, _ := strconv.Atoi(match[0][4])
		hStart, _ := strconv.Atoi(match[0][3])
		hEnd, _ := strconv.Atoi(match[0][5])

		for w := wStart; w <= wEnd; w++ {
			for h := hStart; h <= hEnd; h++ {
				switch match[0][1] {
				case "turn on":
					lightMap[w][h]++
				case "turn off":
					if lightMap[w][h] != 0 {
						lightMap[w][h]--
					}
				case "toggle":
					lightMap[w][h] += 2
				}
			}
		}
	}

	for w := 0; w <= 999; w++ {
		for h := 0; h <= 999; h++ {
			lighted += lightMap[w][h]
		}
	}

	return lighted
}

// 5일차
func doesntHeHaveInternElvesForThis1(line string) bool {
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

func doesntHeHaveInternElvesForThis2(line string) bool {
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
