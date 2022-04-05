package doExcel

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	debug *bool
)

// call flag.Parse() in main
func init() {
	//enable log file
	if path := os.Getenv(LOG_PATH); path != "" {
		l := &lumberjack.Logger{
			Filename: path,
			MaxSize:  1,    // megabytes
			MaxAge:   7,    //days
			Compress: true, // disabled by default
		}

		l.Rotate()
		log.SetOutput(l)
	}
	//enable debug
	debug = flag.Bool("d", false, "Run in debug mode")
}
func Debug(format string, args ...interface{}) {
	if *debug {
		Log(format, args...)
	}
}
func Log(format string, args ...interface{}) {
	log.Printf("%s\n", fmt.Sprintf(format, args...))
}
