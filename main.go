package main

import (
	"flag"
	"github.com/andrepxx/go-dsp-guitar/controller"
)

/*
 * The entry point of our program.
 */
func main() {
	numChannels := flag.Int("channels", 0, "Number of channels for batch processing")
	flag.Parse()
	cn := controller.CreateController()
	cn.Operate(*numChannels)
}
