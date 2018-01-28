package main

// This file contains the implementation of a map of strings and booleans that are updated
// by the servers critical components.  Clients can also request that if any of the modules goes into
// a false condition that they will be informed using a channel and when the entire collection
// goes into a true condition that they will also be informed.
//
// This allows for a server to for example maintain an overall check on whether any of its critical
// modules are down and when they are all alive.

import (
	"sync"
	"time"
)

type Modules struct{}

type catalog struct {
	listeners []chan bool
	m         map[string]bool
	sync.Mutex
}

var (
	modules = catalog{
		m: map[string]bool{},
	}
	modulesUpdateC = make(chan struct{})
)

func (*Modules) AddListener(listener chan bool) {
	modules.Lock()
	defer modules.Unlock()
	modules.listeners = append(modules.listeners, listener)
}

func (*Modules) SetModule(module string, up bool) {
	modules.Lock()
	defer modules.Unlock()
	modules.m[module] = up
	select {
	case modulesUpdateC <- struct{}{}:
	default:
	}
}

func (*Modules) doUpdate() {
	modules.Lock()
	defer modules.Unlock()

	// Is the sever entirely up or not
	up := true
	for _, v := range modules.m {
		if v != true {
			up = false
			break
		}
	}
	// Tell everyone what the collective state is for the server
	for i, listener := range modules.listeners {
		func() {
			defer func() {
				// A send to a closed channel will panic and so if a
				// panic does occur we remove the listener
				if r := recover(); r != nil {
					modules.listeners = append(modules.listeners[:i], modules.listeners[i+1:]...)
				}
			}()
			select {
			case <-time.After(20 * time.Millisecond):
			case listener <- up:
			}
		}()
	}
}

func initModuleTracking(quitC <-chan struct{}) {
	go func() {
		internalCheck := time.Duration(5 * time.Second)
		modules := &Modules{}
		for {
			select {

			case <-time.After(internalCheck):
				modules.doUpdate()
			case <-modulesUpdateC:
				modules.doUpdate()
			case <-quitC:
				return
			}
		}
	}()
}
