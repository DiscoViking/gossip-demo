package main

import (
	"fmt"
	"github.com/stefankopieczek/gossip/log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	// Caller parameters
	caller = &endpoint{
		displayName: "Ryan",
		username:    "ryan",
		host:        "172.18.115.70",
		port:        5070,
		proxy:       "10.26.1.41",
		proxy_port:  6000,
		transport:   "UDP",
	}

	// Callee parameters
	callee = &endpoint{
		displayName: "Ryan's PC",
		username:    "stefan",
		host:        "172.18.115.70",
		port:        5060,
		proxy:       "10.26.1.21",
		proxy_port:  6010,
		transport:   "UDP",
	}
)

func main() {
	var err error
	log.SetDefaultLogLevel(log.SEVERE)
	if err != nil {
		panic(err)
	}

	err = caller.Start()
	if err != nil {
		panic(err)
	}
	err = callee.Start()
	if err != nil {
		panic(err)
	}

	f, err := os.Create("cpuprof.out")
	if err != nil {
		panic(err)
	}
	runtime.SetBlockProfileRate(0)
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	sub_num := 91000
	caller.dialog.cseq = 1
	successes := 0
	tries := 0
	thissecond := 0
	persecond := 10000
	start := time.Now()
	success := make(chan struct{}, 10)
	go func() {
		for _ = range success {
			successes++
		}
	}()
	end := time.After(20 * time.Second)
	second := time.NewTicker(1 * time.Second)
	go callee.ServeNonInvite()
	go func() {
		for {
			tries++
			thissecond++
			sub_num++
			caller.dialog.callId = fmt.Sprintf("callid%v", sub_num)
			caller.username = fmt.Sprintf("user%v", sub_num)
			caller.dialog.cseq = uint32(sub_num)
			go func() {
				err := caller.Register(callee)
				if err == nil {
					success <- struct{}{}
				} else {
					//fmt.Printf("Registration failed: %v\n", err.Error())
				}
			}()

			if thissecond == persecond {
				thissecond = 0
				<-second.C
				fmt.Printf("%v tries. %v successes.\n", tries, successes)
				fmt.Printf("Currently %v goroutines exist.\n", runtime.NumGoroutine())
				fmt.Printf("Time since start: %v\n", time.Since(start))
			}
		}
	}()
loop:
	for {
		select {
		case <-end:
			f, _ := os.Create("memprof.out")
			pprof.WriteHeapProfile(f)
			f.Close()
			f, _ = os.Create("blockprof.out")
			pprof.Lookup("block").WriteTo(f, 0)
			f.Close()
			break loop
		default:
		}
	}
}
