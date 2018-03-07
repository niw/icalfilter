package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/niw/icalfilter"
)

var (
	input  = flag.String("input", "", "input file")
	output = flag.String("output", "", "output file")
	months = flag.Int("months", 3, "months to keep")
)

func main() {
	flag.Parse()

	if *input == "" || *input == "-" {
		*input = "/dev/stdin"
	}

	c, err := icalfilter.Open(*input)
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	before := time.Now().AddDate(0, -*months, 0)
	err = c.FilterBefore(before)
	if err != nil {
		log.Fatalln(err)
	}

	result := c.String()

	if *output == "" || *output == "-" {
		fmt.Print(result)
	} else {
		f, err := os.Create(*output)
		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()
		fmt.Fprint(f, result)
	}
}
