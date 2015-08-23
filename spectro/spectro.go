package main

import (
	"bufio"
	"container/list"
	"flag"
	"fmt"
	"github.com/mgutz/ansi"
	"golang.org/x/crypto/ssh/terminal"
	"math"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var msBetweenSamples uint

// ANSI colors found using https://github.com/Benvie/repl-rainbow and http://bitmote.com/index.php?post/2012/11/19/Using-ANSI-Color-Codes-to-Colorize-Your-Bash-Prompt-on-Linux
var maximumValue float64
var maximumMagnitude uint64
var grayScale = []uint{16, 234, 235, 236, 237, 238, 239, 240, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255}
var rainbowScale = []uint{16, 53, 90, 127, 164, 201, 165, 129, 93, 57, 21, 27, 33, 39, 45, 51, 50, 49, 48, 47, 46, 82, 118, 154, 190, 226, 220, 214, 208, 202, 196}
var heatScale = []uint{16, 17, 18, 19, 20, 21, 27, 33, 39, 45, 51, 50, 49, 48, 47, 46, 82, 118, 154, 190, 226, 220, 214, 208, 202, 196}
var colorScale = heatScale
var colorScheme string
var scaleType string
var scale = linearScale
var reverseScale = reverseExponentialScale

func init() {
	flag.StringVar(&colorScheme, "color", "heat", "how to render the magnitudes (grayscale, rainbow)")
	flag.StringVar(&scaleType, "scale", "linear", "the scale to use for the x/value axis (linear, logarithmic, exponential)")
	flag.Float64Var(&maximumValue, "maximum", 0, "allows you to specify the expected maximum value to avoid rendering interruptions")
	flag.Uint64Var(&maximumMagnitude, "magnitude", 0, "allows you to specify the expected maximum magnitude value (i.e. frequency, which depends on the width of the summarisation buckets) to avoid rendering interruptions")
	flag.UintVar(&msBetweenSamples, "sample-period-ms", 1000, "controls the minimum amount of milliseconds between samples")
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

}


func main() {

	var c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
		fmt.Fprintf(os.Stdout, "\n")
	}()

	pacemakerPresentPattern, _ := regexp.Compile("PACEMAKER_PRESENT")
	pacemakerIterationPattern, _ := regexp.Compile("PACEMAKER_ITERATION")
	numberPattern, _ := regexp.Compile("[0-9.]+$")

	scanner := bufio.NewScanner(os.Stdin)
	var buffer list.List
	var legend string
	var scaleHasChanged = false
	w, _, _ := terminal.GetSize(1)
	terminalWidth := uint(w - 10)

	lastSampleTaken := time.Now()
	pacemakerPresent := false

	for scanner.Scan() {
		lineOfText := scanner.Text()
		if pacemakerPresentPattern.MatchString(lineOfText) {
			pacemakerPresent = true
		}
		pacemakerIterationSignal := pacemakerIterationPattern.MatchString(lineOfText)
		if !pacemakerIterationSignal && numberPattern.MatchString(lineOfText) {
			var dataPoint float64
			numberText := numberPattern.FindString(lineOfText)
			dataPoint, _ = strconv.ParseFloat(numberText, 64)
			if dataPoint > maximumValue {
				maximumValue = 1.2 * dataPoint
				scaleHasChanged = true
			}
			buffer.PushBack(dataPoint)
		}
		if pacemakerIterationSignal || (!pacemakerPresent && time.Since(lastSampleTaken) >= time.Millisecond*time.Duration(msBetweenSamples)) {
			timeText := time.Time.Format(time.Now(), "15:04:05")
			if pacemakerIterationSignal {
				timeText = strings.Split(lineOfText, " ")[1]
			}
			histogram, newMaximumMagnitude, maximumMagnitudeHasChanged := sample(buffer, maximumValue, terminalWidth, maximumMagnitude)
			if maximumMagnitudeHasChanged {
				maximumMagnitude = newMaximumMagnitude
			} // otherwise scope means it is forgotten each time
			legend = updateLegendAndNotifyIfScaleHasChanged(legend, maximumValue, scaleHasChanged, terminalWidth)
			printSample(histogram, timeText, maximumValue, terminalWidth, maximumMagnitude, maximumMagnitudeHasChanged)
			printScale(histogram, uint(len(timeText)), legend)
			// reset for next sample
			scaleHasChanged = false
			buffer.Init()
			lastSampleTaken = time.Now()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
	fmt.Fprint(os.Stdout, "\n")
}

func updateLegendAndNotifyIfScaleHasChanged(legend string, maximumValue float64, scaleHasChanged bool, terminalWidth uint) string {
	if len(legend) == 0 {
		return formatScale(maximumValue, terminalWidth) // used when max is set
	} else if scaleHasChanged {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("\nAdjusting value scale to fit new maximum of %.2f", maximumValue))
		return formatScale(maximumValue, terminalWidth)
	}
	return legend
}

func linearScale(index uint, maximumValue float64, terminalWidth uint) float64 {
	return float64(maximumValue) * float64(index) / float64(terminalWidth)
}

func reverseLinearScale(number float64, maximumValue float64, terminalWidth uint) uint {
	return uint(number * float64(terminalWidth) / float64(maximumValue))
}

func logarithmicScale(index uint, maximumValue float64, terminalWidth uint) float64 {
	var scaleFactor = maximumValue / math.Log10(float64(terminalWidth+1))
	return scaleFactor * math.Log10(float64(index+1))
}

func reverseLogarithmicScale(number float64, maximumValue float64, terminalWidth uint) uint {
	var scaleFactor = maximumValue / math.Log10(float64(terminalWidth+1))
	return uint((math.Pow(10, number/scaleFactor) - 1) + 0.5)
}

func exponentialScale(index uint, maximumValue float64, terminalWidth uint) float64 {
	var xFactor = math.Log10(maximumValue+1) / float64(terminalWidth)
	return math.Pow(10, float64(index)*xFactor) - 1
}

func reverseExponentialScale(number float64, maximumValue float64, terminalWidth uint) uint {
	var xFactor = math.Log10(maximumValue+1) / float64(terminalWidth)
	return uint(math.Floor((math.Log10(number+1) / xFactor) + 0.5))
}
func sample(points list.List, maximumValue float64, terminalWidth uint, maximumMagnitude uint64) (map[float64]uint64, uint64, bool) {
	histogram := make(map[float64]uint64)
	maximumMagnitudeHasChanged := false
	for datapointElement := points.Front(); datapointElement != nil; datapointElement = datapointElement.Next() {
		dataPoint := datapointElement.Value.(float64)
		column := reverseScale(dataPoint, maximumValue, terminalWidth)
		histogram[scale(column, maximumValue, terminalWidth)]++
		if histogram[scale(column, maximumValue, terminalWidth)] > maximumMagnitude {
			maximumMagnitude = histogram[scale(column, maximumValue, terminalWidth)]
			maximumMagnitudeHasChanged = true
		}
	}
	return histogram, maximumMagnitude, maximumMagnitudeHasChanged
}

func printScale(histogram map[float64]uint64, paddingWidth uint, legend string) {
	fmt.Fprintf(os.Stderr, "%"+strconv.FormatInt(int64(paddingWidth), 10)+"s %s\r", "", legend)
}

func formatScale(maximumValue float64, terminalWidth uint) string {
	var scaleLabels string
	for column := uint(0); column < terminalWidth; {
		label := fmt.Sprintf("|%.2f", scale(column, maximumValue, terminalWidth))
		if (column-1)%10 == 0 && column+uint(len(label)) < terminalWidth {
			column += uint(len(label))
			scaleLabels += label
		} else {
			scaleLabels += " "
			column++
		}
	}
	return scaleLabels
}

func printSample(histogram map[float64]uint64, timeText string, maximumValue float64, terminalWidth uint, maximumMagnitude uint64, maximumMagnitudeHasChanged bool) {
	if maximumMagnitudeHasChanged {
		wholeLine := fmt.Sprintf("Adjusting magnitude (color) scale to suit maximum of %v.", maximumMagnitude)
		fmt.Fprintf(os.Stderr, "%-"+strconv.FormatUint(uint64(terminalWidth+3), 10)+"s\n", wholeLine)
	}
	renderedSample := ""
	// do the plotting
	for column := uint(1); column <= terminalWidth; column++ {
		boundary := scale(column, maximumValue, terminalWidth)
		number := histogram[boundary]
		renderedSample += colorizedDataPoint(number, maximumMagnitude)
	}
	fmt.Fprintf(os.Stdout, "%s %s\n", timeText, renderedSample)
}

func colorFromNumber(number uint64, biggest uint64) uint {
	var index uint
	if number == 0 {
		index = 0
	} else {
		// it was too late and my head hurt too much to work out how to get rid of the 0.1 constant. Without it the rounding
		// down meant that the last color would only be used for the biggest numbers (a smaller band than the other colors).
		index = uint((float64(float64(len(colorScale)-1)-0.1) * float64(number) / float64(biggest))) + 1
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
