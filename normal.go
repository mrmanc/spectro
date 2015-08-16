package main

import "fmt"
import "math/rand"

func main() {
	for true {
//		fmt.Println(int(rand.Intn(1000)))
		fmt.Println(int(rand.NormFloat64()*100)+475)
	}
}