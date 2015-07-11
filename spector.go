package main

import (
	"bufio"
	"os"
	"fmt"
	"strconv"
	"container/list"
	"time"
	"github.com/mgutz/ansi"
)


var linearBoundaries list.List

func main() {
	maxInputValue := 1000
	terminalWidth := 130
	for boundaryIndex := 1; boundaryIndex < terminalWidth; boundaryIndex++ {
		linearBoundaries.PushBack(float64(maxInputValue) * float64(boundaryIndex)/float64(terminalWidth))
	}

	scanner := bufio.NewScanner(os.Stdin)
	var buffer list.List
	lastSampleTaken := time.Now()
	timeBetweenSamples := time.Second / 50
	for scanner.Scan() {
		var f float64
		f, _ = strconv.ParseFloat(scanner.Text(), 64)
		buffer.PushBack(f)
		if time.Since(lastSampleTaken) >= timeBetweenSamples {
			sample(buffer)
			buffer.Init()
			lastSampleTaken = time.Now()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

}
func sample(points list.List) {
	histogram := make(map[float64]int64)

	for e := points.Front(); e != nil; e = e.Next() {
		var previousBoundary float64 = 0
		for boundary := linearBoundaries.Front(); boundary != nil; boundary = boundary.Next() {
			bound := boundary.Value.(float64)
			if e.Value.(float64) > previousBoundary && e.Value.(float64) < bound {
				histogram[bound] = histogram[bound] + 1
				break
			}
			previousBoundary = bound
		}
	}
	printSample(histogram)
}

func printSample(histogram map[float64]int64) {
	var biggest int64 = 0
	var smallest int64 = 9223372036854775807
	for boundaryElement := linearBoundaries.Front(); boundaryElement != nil; boundaryElement = boundaryElement.Next() {
		boundary := boundaryElement.Value.(float64)
		freq := histogram[boundary]
		if (freq > biggest) {biggest = freq}
		if (freq > 0 && freq < smallest) {smallest = freq}
	}
	//	fmt.Println("Max was ", biggest)
	//	fmt.Println("Min was ", smallest)
	//	fmt.Println("Colour was ", colourFromNumber(300, smallest, biggest))

	for boundaryElement := linearBoundaries.Front(); boundaryElement != nil; boundaryElement = boundaryElement.Next() {
		boundary := boundaryElement.Value.(float64)
		fmt.Print(ansi.ColorCode(fmt.Sprintf(":%d", colourFromNumber(histogram[boundary], smallest, biggest))), " ", ansi.ColorCode("reset")) //█
	}
	fmt.Printf("\n")
}
func colourFromNumber(number int64, min int64, max int64) int64 {
	if max == 0 {return 0}
	return 234+(255-234)*(number-min)/(max-min) // higher contrast
//	return 234+((255-234)*number/max) // more accurate
}