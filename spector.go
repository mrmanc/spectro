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
	"regexp"
)

var maxInputValue float64 = 0
var terminalWidth int = 130
var useColour bool = true
var scaleHasChanged bool = false
// ANSI colours found using https://github.com/Benvie/repl-rainbow and http://bitmote.com/index.php?post/2012/11/19/Using-ANSI-Color-Codes-to-Colorize-Your-Bash-Prompt-on-Linux
var rainbow = []int64 {16, 53, 90, 127, 164, 201, 165, 129, 93, 57, 21, 27, 33, 39, 45, 51, 50, 49, 48, 47, 46, 82, 118, 154, 190, 226, 220, 214, 208, 202, 196}
var legend string
var newLegend string
var timeBetweenSamples = time.Second / 15

func main() {
	pacemakerPresentPattern, _ := regexp.Compile("\\[PACEMAKER_PRESENT\\]")
	pacemakerIterationPattern, _ := regexp.Compile("PACEMAKER_ITERATION")
	numberPattern, _ := regexp.Compile("[0-9.]+$")

	scanner := bufio.NewScanner(os.Stdin)
	var buffer list.List

	lastSampleTaken := time.Now()
	pacemakerPresent := false
	
	for scanner.Scan() {
		lineOfText := scanner.Text()
		if (pacemakerPresentPattern.MatchString(lineOfText)) {
			pacemakerPresent = true
		}
		pacemakerIterationSignal := pacemakerIterationPattern.MatchString(lineOfText)
		if (!pacemakerIterationSignal && numberPattern.MatchString(lineOfText)) {
			var f float64
			numberText := numberPattern.FindString(lineOfText)
			f, _ = strconv.ParseFloat(numberText, 64)
			buffer.PushBack(f)
		}
		if (!pacemakerPresent && time.Since(lastSampleTaken) >= timeBetweenSamples) || pacemakerIterationSignal {
			histogram := sample(buffer)
			printSample(histogram)
			printScale(histogram)
			buffer.Init()
			lastSampleTaken = time.Now()
			if pacemakerPresent {
				time.Sleep(time.Second)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	fmt.Fprint(os.Stdout, "\n")

}
func scale(index int) float64 {
	return linearScale(index) // uniform resolution
//	return logarithmicScale(index) // higher resolution of higher numbers
//	return exponentialScale(index) // higher resolution of small numbers
}
func linearScale(index int) float64 {
	return float64(maxInputValue) * float64(index)/float64(terminalWidth)
}
func logarithmicScale(index int) float64 {
	var scaleFactor = maxInputValue / math.Log2(float64(terminalWidth+1))
	var boundary float64 = scaleFactor * math.Log2(float64(index+1))
	return boundary
}
func exponentialScale(index int) float64 {
	var scaleFactor = maxInputValue / (math.Exp2(float64(terminalWidth))-1)
	var boundary float64 = scaleFactor * (math.Exp2(float64(index+1))-1)
	return boundary
}
func sample(points list.List) map[float64]int64 {
	histogram := make(map[float64]int64)
	var maximumDataPoint float64 = 0
	for datapointElement := points.Front(); datapointElement != nil; datapointElement = datapointElement.Next() {
		datapoint := datapointElement.Value.(float64)
		if datapoint > maximumDataPoint {
			maximumDataPoint = datapoint
		}
		var previousBoundary float64 = 0
		for column := 1; column <= terminalWidth; column++ {
			boundary := scale(column)
			if datapoint > previousBoundary && datapoint < boundary {
				histogram[boundary] = histogram[boundary] + 1
				break
			}
			previousBoundary = boundary
		}
	}
	if maximumDataPoint > maxInputValue {
		maxInputValue = maximumDataPoint * 1.2
		scaleHasChanged = true
		newLegend = formatScale(histogram)
	}
	return histogram
}

func printScale(histogram map[float64]int64) {
	fmt.Fprint(os.Stderr, legend)
	if (scaleHasChanged) {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("\nAdjusting scale to new maximum of %.2f", maxInputValue))
		scaleHasChanged = false
		legend = newLegend
	}
	fmt.Fprint(os.Stderr, "\r")
}

func formatScale(histogram map[float64]int64) string {
	var scaleLabels string
	for column := 0; column < terminalWidth; {
		if (column - 1) % 10 == 0 {
			label := fmt.Sprintf("|%d", int64(scale(column)))
			column += len(label)
			scaleLabels += label
		} else {
			scaleLabels += " "
			column ++
		}
	}
	return scaleLabels
}


func printSample(histogram map[float64]int64) {

	// find the max and min amplitudes
	var biggest int64 = 0
	var smallest int64 = 9223372036854775807
	for column := 1; column <= terminalWidth; column++ {
		boundary := scale(column)
		freq := histogram[boundary]
		if (freq > biggest) {biggest = freq}
		if (freq > 0 && freq < smallest) {smallest = freq}
	}
	// do the plotting
	for column := 1; column <= terminalWidth; column++ {
		boundary := scale(column)
		number := histogram[boundary]
		fmt.Fprint(os.Stdout, colorizedDataPoint(number, smallest, biggest))
//		fmt.Fprint(os.Stdout, colorizedDataPoint(int64(i), int64(0), int64(terminalWidth))) // tests a gradient
	}
	fmt.Fprint(os.Stdout, "\n")
}
func greyscaleFromNumber(number int64, smallest int64, biggest int64) int64 {
	if (biggest - smallest == 0) {return 234}
	return 234+(255-234)*(number-smallest)/(biggest-smallest) // higher contrast
//	return 234+((255-234)*number/biggest) // more accurate
}
func rainbowFromNumber(number int64, smallest int64, biggest int64) int64 {
	var index int64
	if (number == 0) {
		index = 0
	} else {
		if ((biggest-smallest) > 0) {
			// it was too late and my head hurt too much to work out how to get rid of the 0.1 constant. Without it the rounding
			// down meant that the last colour would only be used for the biggest numbers (a smaller band than the other colours).
			index = int64((float64(float64(len(rainbow))-0.1) * float64(number-smallest) / float64(biggest-smallest)))
		} else {
			// not enough variation to create any spectrum, so just use last colour
			index = int64(len(rainbow)-1)
		}
	}
	return rainbow[index]
}
func ansiCodeFromNumber(number int64, smallest int64, biggest int64) string {
	var colorNumber int64
	if useColour {
		colorNumber = rainbowFromNumber(number, smallest, biggest)
	} else {
		colorNumber = greyscaleFromNumber(number, smallest, biggest)
	}
	return ansi.ColorCode(fmt.Sprintf("%d:%d", colorNumber, colorNumber))
}
func resetText() string {
	return ansi.ColorCode("reset")
}
func colorizedDataPoint(number int64, smallest int64, biggest int64) string {
	return fmt.Sprint(ansiCodeFromNumber(number, smallest, biggest), " ", resetText()) //█
}