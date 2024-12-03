package main

import (
	"os"
)

var workingDIR *string

func main() {
	DIR, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	workingDIR = &DIR

	Backtests()

}
