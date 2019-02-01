package main

import (
	"flag"
	"fmt"
	"github.com/andrepxx/go-dsp-guitar/controller"
)

/*
 * The entry point of our program.
 */
func main() {
	numChannels := flag.Int("channels", 0, "Number of channels for batch processing")
	versionFlag := flag.Bool("version", false, "Print version information and exit")
	flag.Parse()

	/*
	 * Print version information or start the actual application.
	 */
	if *versionFlag {
		msg, err := controller.Version()

		/*
		 * If an error occured, print error message, otherwise
		 * print version information.
		 */
		if err != nil {
			msg = err.Error()
		}

		fmt.Printf("%s\n", msg)
	} else {
		cn := controller.CreateController()
		cn.Operate(*numChannels)
	}

}
