package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/UniversityRadioYork/baps3d/comm"
	"github.com/UniversityRadioYork/baps3d/console"
	"github.com/UniversityRadioYork/baps3d/list"
	"github.com/UniversityRadioYork/baps3d/netsrv"
)

func main() {
	var wg sync.WaitGroup

	rootLog := log.New(os.Stderr, "[root] ", log.LstdFlags)

	lst := list.New()
	lstCon, rootClient := comm.NewController(lst)
	wg.Add(1)
	go func() {
		lstCon.Run()
		wg.Done()
	}()

	netLog := log.New(os.Stderr, "[net] ", log.LstdFlags)
	netClient, err := rootClient.Copy()
	if err != nil {
		rootLog.Println("couldn't create network client:", err)
		return
	}
	netSrv := netsrv.New(netLog, "localhost:1357", netClient, lst)
	wg.Add(1)
	go func() {
		netSrv.Run()
		wg.Done()
	}()

	consoleClient, err := rootClient.Copy()
	if err != nil {
		rootLog.Println("couldn't create console client:", err)
		return
	}
	consoleBf, consoleBfClient := comm.NewBifrost(consoleClient, lst)
	wg.Add(1)
	go func() {
		consoleBf.Run()
		wg.Done()
	}()

	console, err := console.New(consoleBfClient)
	if err != nil {
		rootLog.Println("couldn't bring up console:", err)
		return
	}

	wg.Add(2)
	go func() {
		console.RunRx()
		wg.Done()
	}()
	go func() {
		console.RunTx()
		if err = console.Close(); err != nil {
			fmt.Println(err)
		}
		consoleClient.Shutdown()
		wg.Done()
	}()

	for range rootClient.Rx {
	}

	wg.Wait()
	rootLog.Println("It's now safe to turn off your baps3d.")
}
