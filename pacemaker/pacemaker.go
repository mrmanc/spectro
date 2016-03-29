package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var nowait bool
var secondsBetweenSamples int

func init() {
	flag.BoolVar(&nowait, "nowait", false, "whether to pause between plotting samples")
	flag.IntVar(&secondsBetweenSamples, "sample-period-s", 1, "controls the minimum amount of seconds between samples")
	flag.Parse()
}
func main() {
	fmt.Println("[PACEMAKER_PRESENT]")
	timePattern, _ := regexp.Compile("[012][0-9]:[0-5][0-9]:[0-5][0-9]")
	scanner := bufio.NewScanner(os.Stdin)
	var lastStart int
	firstLine := true
	for scanner.Scan() {
		line := scanner.Text()
		if timePattern.MatchString(line) {
			timeText := timePattern.FindString(line)
			currentSeconds := secondsFromString(timeText)
			if firstLine {
				lastStart = currentSeconds - currentSeconds % secondsBetweenSamples
				firstLine = false
			}
			if currentSeconds < (lastStart-secondsBetweenSamples) {
				lastStart -= secondsBetweenSamples
			}
			for currentSeconds > lastStart {
				fmt.Fprintf(os.Stdout, "PACEMAKER_ITERATION %s\n", stringFromSeconds(lastStart))
				lastStart = (lastStart + secondsBetweenSamples) % (60 * 60 * 24)
				if !nowait {
					time.Sleep(time.Second * time.Duration(secondsBetweenSamples))
				}
			}
			fmt.Println(line)
		} else {
			fmt.Printf("Could not find a time within the line of text: %s\n", line)
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
	return hour*60*60 + min*60 + sec
}
func stringFromSeconds(totalSeconds int) string {
	hours := (totalSeconds - (totalSeconds % 3600)) / 3600
	minutes := (totalSeconds % 3600 - totalSeconds % 60) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}