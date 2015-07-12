package main

import (
	"bufio"
	"os"
	"fmt"
	"strconv"
	"container/list"
	"time"
	"github.com/mgutz/ansi"
	"math"
)


var maxInputValue float64 = 1000
var terminalWidth int16 = 130

func main() {

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
func scale(index int16) float64 {
//	return linearScale(index)
//	return logarithmicScale(index)
	return exponentialScale(index)
}
func linearScale(index int16) float64 {
	return float64(maxInputValue) * float64(index)/float64(terminalWidth)
}
func logarithmicScale(index int16) float64 {
	var scaleFactor = maxInputValue / math.Log2(float64(terminalWidth+1))
	var boundary float64 = scaleFactor * math.Log2(float64(index+1))
	return boundary
}
func exponentialScale(index int16) float64 {
	var scaleFactor = maxInputValue / (math.Exp2(float64(terminalWidth))-1)
	var boundary float64 = scaleFactor * (math.Exp2(float64(index+1))-1)
	return boundary
}
func sample(points list.List) {
	histogram := make(map[float64]int64)
	var maximumDataPoint float64 = 0
	for datapointElement := points.Front(); datapointElement != nil; datapointElement = datapointElement.Next() {
		datapoint := datapointElement.Value.(float64)
		if datapoint > maximumDataPoint {
			maximumDataPoint = datapoint
		}
		var previousBoundary float64 = 0
		for i := int16(1); i <= terminalWidth; i++ {
			boundary := scale(i)
			if datapoint > previousBoundary && datapoint < boundary {
				histogram[boundary] = histogram[boundary] + 1
				break
			}
			previousBoundary = boundary
		}
	}
	if maximumDataPoint > maxInputValue {
		maxInputValue = maximumDataPoint * 1.2
		fmt.Println("Adjusting scale to new maximum of ", maxInputValue)
	}
	printSample(histogram)
}

func printSample(histogram map[float64]int64) {

	// find the max and min amplitudes
	var biggest int64 = 0
	var smallest int64 = 9223372036854775807
	for i := int16(1); i <= terminalWidth; i++ {
		boundary := scale(i)

		freq := histogram[boundary]
		if (freq > biggest) {biggest = freq}
		if (freq > 0 && freq < smallest) {smallest = freq}
	}
	//	fmt.Println("Max was ", biggest)
	//	fmt.Println("Min was ", smallest)
	//	fmt.Println("Colour was ", colourFromNumber(300, smallest, biggest))

	// do the plotting
	for i := int16(1); i <= terminalWidth; i++ {
		boundary := scale(i)
		number := histogram[boundary]
		fmt.Print(colorizedDataPoint(number, smallest, biggest))
	}
	fmt.Printf("\n")
}
func greyscaleFromNumber(number int64, smallest int64, biggest int64) int64 {
	if (biggest - smallest == 0) {return 234}
	return 234+(255-234)*(number-smallest)/(biggest-smallest) // higher contrast
	//	return 234+((255-234)*number/max) // more accurate
}
func greyscaleAnsiCodeFromNumber(number int64, smallest int64, biggest int64) string {
	color := greyscaleFromNumber(number, smallest, biggest)
	return ansi.ColorCode(fmt.Sprintf("%d:%d", color, color))
}
func resetText() string {
	return ansi.ColorCode("reset")
}
func colorizedDataPoint(number int64, smallest int64, biggest int64) string {
	return fmt.Sprint(greyscaleAnsiCodeFromNumber(number, smallest, biggest), "█", resetText())
}