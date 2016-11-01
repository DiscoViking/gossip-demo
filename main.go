package main

import (
	"flag"
	"time"
	"gossip-demo/demo"
	"github.com/stefankopieczek/gossip/log"
)

var (
	// Caller parameters
	caller = &demo.EndPoint{
		DisplayName: "Test caller",
		UserName:    "ted",
		Host:        "127.0.0.1",
		Port:        5061,
		Transport:   "UDP",
	}

	// Callee parameters
	callee = &demo.EndPoint{
		DisplayName: "Test Service",
		UserName:    "service",
		Host:        "127.0.0.1",
		Port:        5060,
		Transport:   "UDP",
	}
)



func TestCall() {
	log.SetDefaultLogLevel(log.DEBUG)
	err := caller.Start()
	if err != nil {
		log.Warn("Failed to start caller: %v", err)
		return
	}
	caller.Invite(callee)

}


func TestServer() {
	log.SetDefaultLogLevel(log.DEBUG)
	err := callee.Start()
	if err != nil {
		log.Warn("Failed to start caller: %v", err)
		return
	}

	// Receive an incoming call.
	callee.ServeInvite()

	<-time.After(2 * time.Second)

	// Send the BYE
	callee.Bye(caller)
}

/*Usage: In one shell window run "go run main.go" without arguments
  to start server.

  In another shell window, run "go run main.go -call 1" to do a test call
  */

func main() {
	var doCall = 0
	flag.IntVar(&doCall, "call",0, "do a test call")
	flag.Parse()
	if (doCall > 0) {
		TestCall()
	} else {
		TestServer()
	}
}
