package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"strconv"
)

func main() {
	fmt.Println("[PACEMAKER_PRESENT]")
	timePattern, _ := regexp.Compile("[012][0-9]:[0-5][0-9]:[0-5][0-9]")
	scanner := bufio.NewScanner(os.Stdin)
	var lastStart int
	firstLine := true
	for scanner.Scan() {
		var line string
		line = scanner.Text()
		if (timePattern.MatchString(line)) {
			timeText := timePattern.FindString(line)
			currentSeconds := secondsFromString(timeText)
			if (firstLine) {
				lastStart = currentSeconds
				firstLine = false
			}
			for currentSeconds != lastStart  {
				fmt.Fprintf(os.Stdout, "PACEMAKER_ITERATION:%d\n", lastStart)
				lastStart = (lastStart + 1) % (60*60*24)
			}
			fmt.Println(line)
		} else {
			fmt.Println("did not match")
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}
func secondsFromString(time string) int {
	bits := strings.Split(time, ":")
	hour, _ := strconv.Atoi(bits[0])
	min, _ := strconv.Atoi(bits[1])
	sec, _ := strconv.Atoi(bits[2])
	return hour * 60 * 60 + min * 60 + sec
}