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
	"strings"
	"flag"
	"golang.org/x/crypto/ssh/terminal"
)

const timeBetweenSamples = time.Second
var terminalWidth uint
var maxInputValue float64 = 0
var scaleHasChanged bool = false
// ANSI colors found using https://github.com/Benvie/repl-rainbow and http://bitmote.com/index.php?post/2012/11/19/Using-ANSI-Color-Codes-to-Colorize-Your-Bash-Prompt-on-Linux
var grayScale = []uint {16, 234, 235, 236, 237, 238, 239, 240, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255}
var rainbowScale = []uint {16, 53, 90, 127, 164, 201, 165, 129, 93, 57, 21, 27, 33, 39, 45, 51, 50, 49, 48, 47, 46, 82, 118, 154, 190, 226, 220, 214, 208, 202, 196}
var heatScale = []uint {16, 17, 18, 19, 20, 21, 27, 33, 39, 45, 51, 50, 49, 48, 47, 46, 82, 118, 154, 190, 226, 220, 214, 208, 202, 196}
var colorScale = heatScale
var legend string
var newLegend string
var colorScheme string
var scaleType string
var scale = linearScale

func init() {
	flag.StringVar(&colorScheme, "color", "heat", "how to render the amplitudes (grayscale, rainbow)")
	flag.StringVar(&scaleType, "scale", "linear", "the scale to use for the x/amplitude axis (linear, logarithmic, exponential)")
	flag.Parse()

	switch colorScheme {
	case "grayscale":
		colorScale = grayScale
	case "rainbow":
		colorScale = rainbowScale
	case "heat":
		colorScale = heatScale
	default:
		fmt.Fprintln(os.Stderr, "Did not recognise color scheme specified. Must be one of grayscale, rainbow, heat")
		os.Exit(1)
	}
	switch scaleType {
	case "logarithmic":
		scale = logarithmicScale // higher resolution of higher numbers
	case "exponential":
		scale = exponentialScale // higher resolution of small numbers
	case "linear":
		scale = linearScale // uniform resolution
	default:
		fmt.Fprintln(os.Stderr, "Did not recognise scale provided. Must be one of linear, logarithmic, exponential")
		os.Exit(2)
	}
	w, _, _ := terminal.GetSize(1)
	terminalWidth = uint(w-10)
}

func main() {
	pacemakerPresentPattern, _ := regexp.Compile("PACEMAKER_PRESENT")
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
		if (pacemakerIterationSignal || (!pacemakerPresent && time.Since(lastSampleTaken) >= timeBetweenSamples)) {
			timeText := time.Time.Format(time.Now(), "15:04:05")
			if (pacemakerIterationSignal) {
				timeText = strings.Split(lineOfText, " ")[1]
			}
			histogram := sample(buffer)
			printSample(histogram, timeText)
			printScale(histogram, uint(len(timeText)))
			buffer.Init()
			lastSampleTaken = time.Now()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	fmt.Fprint(os.Stdout, "\n")
}

func linearScale(index uint) float64 {
	return float64(maxInputValue) * float64(index)/float64(terminalWidth)
}

func logarithmicScale(index uint) float64 {
	var scaleFactor = maxInputValue / math.Log2(float64(terminalWidth+1))
	var boundary float64 = scaleFactor * math.Log2(float64(index+1))
	return boundary
}

func exponentialScale(index uint) float64 {
	var scaleFactor = maxInputValue / (math.Exp2(float64(terminalWidth))-1)
	var boundary float64 = scaleFactor * (math.Exp2(float64(index+1))-1)
	return boundary
}
func sample(points list.List) map[float64]uint64 {
	histogram := make(map[float64]uint64)
	var maximumDataPoint float64 = 0
	for datapointElement := points.Front(); datapointElement != nil; datapointElement = datapointElement.Next() {
		datapoint := datapointElement.Value.(float64)
		if datapoint > maximumDataPoint {
			maximumDataPoint = datapoint
		}
		var previousBoundary float64 = 0
		for column := uint(1); column <= terminalWidth; column++ {
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

func printScale(histogram map[float64]uint64, paddingWidth uint) {
	fmt.Fprintf(os.Stderr,"%"+strconv.FormatInt(int64(paddingWidth), 10)+"s %s", "", legend)
	if (scaleHasChanged) {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("\nAdjusting scale to new maximum of %.2f", maxInputValue))
		scaleHasChanged = false
		legend = newLegend
	}
	fmt.Fprint(os.Stderr, "\r")
}

func formatScale(histogram map[float64]uint64) string {
	var scaleLabels string
	for column := uint(0); column < terminalWidth; {
		if (column - 1) % 10 == 0 {
			label := fmt.Sprintf("|%d", uint64(scale(column)))
			column += uint(len(label))
			scaleLabels += label
		} else {
			scaleLabels += " "
			column ++
		}
	}
	return scaleLabels
}

var biggest uint64 = 0
var smallest uint64 = 18446744073709551615

func printSample(histogram map[float64]uint64, timeText string) {
	// find the max and min amplitudes
	amplitudeScaleAdjusted := false
	for column := uint(1); column <= terminalWidth; column++ {
		boundary := scale(column)
		currentAmplitude := histogram[boundary]
		if (currentAmplitude > biggest) {
			biggest = currentAmplitude
			amplitudeScaleAdjusted = true
		}
		if (currentAmplitude > 0 && currentAmplitude < smallest) {
			smallest = currentAmplitude
			amplitudeScaleAdjusted = true
		}
	}
	if (amplitudeScaleAdjusted) {
		fmt.Fprintf(os.Stderr, "\nAdapting amplitude scale (shading) to suit range (%v up to %v).\n", smallest, biggest)
	}
	renderedSample := ""
	// do the plotting
	for column := uint(1); column <= terminalWidth; column++ {
		boundary := scale(column)
		number := histogram[boundary]
		renderedSample += colorizedDataPoint(number, smallest, biggest)
	}
	fmt.Fprintf(os.Stdout, "%s %s\n", timeText, renderedSample)
}

func colorFromNumber(number uint64, smallest uint64, biggest uint64) uint {
	var index uint
	if (number == 0) {
		index = 0
	} else {
		if ((biggest-smallest) > 0) {
			// it was too late and my head hurt too much to work out how to get rid of the 0.1 constant. Without it the rounding
			// down meant that the last color would only be used for the biggest numbers (a smaller band than the other colors).
			index = uint((float64(float64(len(colorScale)-1)-0.1) * float64(number-smallest) / float64(biggest-smallest)))+1
		} else {
			// not enough variation to create any spectrum, so just use last color
			index = uint(len(colorScale)-1)
		}
	}
	return colorScale[index]
}

func ansiCodeFromNumber(number uint64, smallest uint64, biggest uint64) string {
	colorNumber := colorFromNumber(number, smallest, biggest)
	return ansi.ColorCode(fmt.Sprintf("%d:%d", colorNumber, colorNumber))
}

func resetText() string {
	return ansi.ColorCode("reset")
}

func colorizedDataPoint(number uint64, smallest uint64, biggest uint64) string {
	return fmt.Sprint(ansiCodeFromNumber(number, smallest, biggest), " ", resetText()) // Using this character may help if your copy / paste does not support background formatting █
}