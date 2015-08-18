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
var reverseScale = reverseExponentialScale

func init() {
	flag.StringVar(&colorScheme, "color", "heat", "how to render the amplitudes (grayscale, rainbow)")
	flag.StringVar(&scaleType, "scale", "linear", "the scale to use for the x/amplitude axis (linear, logarithmic, exponential)")
	flag.Float64Var(&maxInputValue, "max", 0, "allows you to specify the expected maximum value to avoid rendering interruptions")
	flag.Uint64Var(&biggest, "max-amp", 0, "allows you to specify the expected maximum amplitude value (i.e. frequency, which depends on the width of the summarisation buckets) to avoid rendering interruptions")
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
		reverseScale = reverseLogarithmicScale
	case "exponential":
		scale = exponentialScale // higher resolution of small numbers
		reverseScale = reverseExponentialScale
	case "linear":
		scale = linearScale // uniform resolution
		reverseScale = reverseLinearScale
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
	maximumDataPointInSample := float64(0)

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
			if (f > maximumDataPointInSample) {maximumDataPointInSample = f}
			buffer.PushBack(f)
		}
		if (pacemakerIterationSignal || (!pacemakerPresent && time.Since(lastSampleTaken) >= timeBetweenSamples)) {
			timeText := time.Time.Format(time.Now(), "15:04:05")
			if (pacemakerIterationSignal) {
				timeText = strings.Split(lineOfText, " ")[1]
			}
			histogram := sample(buffer, maximumDataPointInSample)
			legend = checkLegendInitialised(legend)
			printSample(histogram, timeText)
			printScale(histogram, uint(len(timeText)))
			buffer.Init()
			maximumDataPointInSample = 0
			lastSampleTaken = time.Now()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	fmt.Fprint(os.Stdout, "\n")
}
func checkLegendInitialised(legend string) string {
	if len(legend) == 0 {
		return formatScale(maxInputValue)
	}
	return legend
}
func linearScale(index uint, maxInputValue float64, terminalWidth uint) float64 {
	return float64(maxInputValue) * float64(index)/float64(terminalWidth)
}
func reverseLinearScale(number float64, maxInputValue float64, terminalWidth uint) uint {
	return uint(number * float64(terminalWidth) / float64(maxInputValue))
}


func logarithmicScale(index uint, maxInputValue float64, terminalWidth uint) float64 {
	var scaleFactor = maxInputValue / math.Log10(float64(terminalWidth+1))
	return scaleFactor * math.Log10(float64(index+1))
}

func reverseLogarithmicScale(number float64, maxInputValue float64, terminalWidth uint) uint {
	var scaleFactor = maxInputValue / math.Log10(float64(terminalWidth+1))
	return uint((math.Pow(10, number / scaleFactor)-1) + 0.5)
}

func exponentialScale(index uint, maxInputValue float64, terminalWidth uint) float64 {
	var xFactor = math.Log10(maxInputValue + 1) / float64(terminalWidth)
	return math.Pow(10, float64(index) * xFactor) - 1
}

func reverseExponentialScale(number float64, maxInputValue float64, terminalWidth uint) uint {
	var xFactor = math.Log10(maxInputValue + 1) / float64(terminalWidth)
	return uint(math.Floor((math.Log10(number+1)/xFactor)+0.5))
}
func sample(points list.List, maximumDataPointInSample float64) map[float64]uint64 {
	if maximumDataPointInSample > maxInputValue {
		maxInputValue = maximumDataPointInSample * 1.2
		scaleHasChanged = true
		newLegend = formatScale(maxInputValue)
	}
	histogram := make(map[float64]uint64)
	for datapointElement := points.Front(); datapointElement != nil; datapointElement = datapointElement.Next() {
		dataPoint := datapointElement.Value.(float64)
		column := reverseScale(dataPoint, maxInputValue, terminalWidth)
		histogram[scale(column, maxInputValue, terminalWidth)] += 1
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

func formatScale(maxInputValue float64) string {
	var scaleLabels string
	for column := uint(0); column < terminalWidth; {
		label := fmt.Sprintf("|%d", uint64(scale(column, maxInputValue, terminalWidth)))
		if (column - 1) % 10 == 0 && column + uint(len(label)) < terminalWidth {
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

func printSample(histogram map[float64]uint64, timeText string) {
	// find the max and min amplitudes
	amplitudeScaleAdjusted := false
	for column := uint(1); column <= terminalWidth; column++ {
		boundary := scale(column, maxInputValue, terminalWidth)
		currentAmplitude := histogram[boundary]
		if (currentAmplitude > biggest) {
			biggest = uint64(float64(currentAmplitude) * 1.1)
			amplitudeScaleAdjusted = true
		}
	}
	if (amplitudeScaleAdjusted) {
		fmt.Fprintf(os.Stderr, "\nAdjusting amplitude scale (shading) to suit maximum of %v.\n", biggest)
	}
	renderedSample := ""
	// do the plotting
	for column := uint(1); column <= terminalWidth; column++ {
		boundary := scale(column, maxInputValue, terminalWidth)
		number := histogram[boundary]
		renderedSample += colorizedDataPoint(number, biggest)
	}
	fmt.Fprintf(os.Stdout, "%s %s\n", timeText, renderedSample)
}

func colorFromNumber(number uint64, biggest uint64) uint {
	var index uint
	if (number == 0) {
		index = 0
	} else {
		// it was too late and my head hurt too much to work out how to get rid of the 0.1 constant. Without it the rounding
		// down meant that the last color would only be used for the biggest numbers (a smaller band than the other colors).
		index = uint((float64(float64(len(colorScale)-1)-0.1) * float64(number) / float64(biggest)))+1
	}
	return colorScale[index]
}

func ansiCodeFromNumber(number uint64, biggest uint64) string {
	colorNumber := colorFromNumber(number, biggest)
	return ansi.ColorCode(fmt.Sprintf("%d:%d", colorNumber, colorNumber))
}

func resetText() string {
	return ansi.ColorCode("reset")
}

func colorizedDataPoint(number uint64, biggest uint64) string {
	return fmt.Sprint(ansiCodeFromNumber(number, biggest), " ", resetText()) // Using this character may help if your copy / paste does not support background formatting █
}