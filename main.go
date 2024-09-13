package main

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/mohammedrefaat/humber/services"
	"github.com/mohammedrefaat/humber/version"
)

func main() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			fmt.Errorf("recover %v", r)
		}
	}()
	srv, err := services.NewServer()
	if err != nil {
		color.Yellow(err.Error())
		log.Fatal(err)
	}
	color.HiMagenta("Welcome, now the service Humber is working on version " + version.VERSION)
	ss := spinner.New(spinner.CharSets[43], 10*time.Millisecond) // Build our new spinner
	ss.Start()                                                   // Start the spinner
	time.Sleep(4 * time.Second)                                  // Run for some time to simulate work
	ss.Stop()

	srv.Run()
}
