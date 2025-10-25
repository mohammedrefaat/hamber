package main

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/mohammedrefaat/hamber/services"
	"github.com/mohammedrefaat/hamber/version"
)

// @title           Hamber API Documentation
// @version         1.0
// @description     This is the API documentation for the Hamber platform
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://www.hamber-hub.com/support
// @contact.email   support@hamber-hub.com

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            test.hamber-hub.com
// @BasePath        /api

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
	color.HiMagenta("Welcome, now the service hamber is working on version " + version.VERSION)
	ss := spinner.New(spinner.CharSets[43], 10*time.Millisecond) // Build our new spinner
	ss.Start()                                                   // Start the spinner
	time.Sleep(4 * time.Second)                                  // Run for some time to simulate work
	ss.Stop()

	srv.Run()
}
