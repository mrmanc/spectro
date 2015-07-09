package main

import (
	"bufio"
	"os"
	"fmt"
	"strconv"
	"container/list"
	"time"
)

func main() {

	scanner := bufio.NewScanner(os.Stdin)
	var buffer list.List
	lastSampleTaken := time.Now()
	timeBetweenSamples := time.Second
	for scanner.Scan() {
		var f float64
		f, _ = strconv.ParseFloat(scanner.Text(), 64)
		fmt.Println(f)
		buffer.PushBack(f)
		if time.Since(lastSampleTaken) >= timeBetweenSamples {
			fmt.Println("time")
			lastSampleTaken = time.Now()
			sample(buffer)
			buffer.Init()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

}
func sample(points list.List) {
//	var histogram map[float64]float64
	for e := points.Front(); e != nil; e = e.Next() {
		fmt.Println("point was %f", e.Value)
	}
}