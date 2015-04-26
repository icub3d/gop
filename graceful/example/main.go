// Copyright (c) 2015 Joshua Marsh. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file in the root of the repository or at
// https://raw.githubusercontent.com/icub3d/gop/master/LICENSE.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/icub3d/gop/graceful"
)

func main() {

	// Listen for the SIGTERM.
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Print("got SIGHUP, shutting down.")
		graceful.Close()
	}()

	// Start the server.
	fmt.Println("Using PID:", os.Getpid())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Sleep for a bit so we can open some connections and send the
		// signal.
		time.Sleep(10 * time.Second)
		log.Printf("%v %v", r.Method, r.URL)
		fmt.Fprintln(w, r.Method, r.URL)
	})
	log.Println(graceful.ListenAndServe(":8080", nil))
	// At this point, try opening a few connection in another
	// terminal. Then in another, send a TERM kignal.
	// For example, in terminal one:
	//
	//   > go run main.go
	//   Using PID: 22191
	//
	// Then in terminal two:
	//
	//   > curl localhost:8080
	//
	// Then in terminal three:
	//
	//   > kill 22191
	//
	// Note that terminal two still gets a response and terminal one
	// remains open until the response is sent.
}
