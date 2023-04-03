// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2023-2025 Intel Corporation

package etimers

import (
	"strings"
	"sync"
	"time"

	"github.com/pktgen/go-pktgen/internal/tlog"
)

// etimers is a package to handle timers for the performance monitor tool.
// A number of timers needed to be handled in a consistent way and from a single
// go routine.

// EventTimers to process when timer expires
type EventTimers struct {
	lock     sync.Mutex
	timo     time.Duration
	maxSteps int
	step     int
	list     map[string]*EventAction
	action   chan *EventAction
	ticker   *time.Ticker
	ticks    uint64
}

// EventAction information
type EventAction struct {
	Name    string
	Action  string
	Routine func(step int, ticks uint64)
}

type EventTimerOption func(*EventTimers)

func WithTimeout(t time.Duration) EventTimerOption {
	return func(e *EventTimers) {
		e.timo = t * time.Second
		e.ticker = time.NewTicker(t * time.Second)
	}
}

func WithSteps(s int) EventTimerOption {
	return func(e *EventTimers) {
		e.maxSteps = s
	}
}

// New use to setup timers to callback handlers
func New(options ...EventTimerOption) *EventTimers {

	et := &EventTimers{
		// Create the first map holding all event actions
		list:     make(map[string]*EventAction, 0),
		timo:     time.Second,
		maxSteps: 1,
		// Create the channel and ticker instances
		action: make(chan *EventAction, 16),
	}

	// Process the option function calls
	for _, f := range options {
		f(et)
	}
	et.ticker = time.NewTicker(et.timo / time.Duration(et.maxSteps))

	return et
}

// Start to handle timeouts with cview
func (et *EventTimers) Start() {

	// Create a go routine that handles all of the timer events
	go func() {

		// Loop forever until a quit message is received over the channel
	ForLoop:
		for {
			select {
			case event := <-et.action:
				tlog.DebugPrintf("EventAction: %s --> %s\n", event.Name, event.Action)

				// When a timer expires then execute the actions attached to the events.
				et.doAction(event)

				if strings.ToLower(event.Action) == "quit" {
					break ForLoop
				}

			// Process the timer ticks and call the timeout routine handler.
			case <-et.ticker.C:
				et.doTimeout()
			}
		}
	}()
}

func (et *EventTimers) doTimeout() {

	// Lock the timer while processing a timeout event
	et.lock.Lock()
	defer et.lock.Unlock()

	// Bump the step counter, which is passed to the action routines as a time
	// reference like value normally 0-4 or 0-8 steps. Each step is 1/4 or 1/8
	// of a second.
	if et.step >= et.maxSteps {
		et.step = 0
	}

	// Call all of the actions when for this timer event
	for _, a := range et.list {
		tlog.DebugPrintf("Call Action: %s\n", a.Name)
		a.Routine(et.step, et.ticks)
	}

	// bump the steps and ticks processed as a type of tick counter.
	et.step++
	et.ticks++
}

// Process an action when a timer event happens
func (et *EventTimers) doAction(a *EventAction) {

	et.lock.Lock()
	defer et.lock.Unlock()

	// Handle the action for a given event
	switch strings.ToLower(a.Action) {
	case "add":
		// Add a new action to the list atomically
		tlog.DebugPrintf("Add Action: %+v\n", a)
		et.list[a.Name] = a
		a.Routine(-1, et.ticks) // Call the action routine on initial add action

	case "remove":
		// Remove an action atomically
		tlog.DebugPrintf("Remove Action: %s\n", a.Name)
		if _, ok := et.list[a.Name]; ok {
			tlog.DebugPrintf("Removed: %s\n", a.Name)
			delete(et.list, a.Name)
		}

	case "quit":
		// Quit the event timer go routine
		close(et.action)
		et.ticker.Stop()
	}
}

// Add to the list of timers
func (et *EventTimers) Add(name string, f func(step int, ticks uint64)) {

	et.lock.Lock()
	defer et.lock.Unlock()

	// Add a timer event by passing it to the timer go routine
	ea := &EventAction{Name: name, Action: "Add", Routine: f}

	et.action <- ea
}

// Remove to the list of timers
func (et *EventTimers) Remove(name string) {

	et.lock.Lock()
	defer et.lock.Unlock()

	// Remove the timer event action by sending it to the go routine
	ea := &EventAction{Name: name, Action: "Remove"}

	et.action <- ea
}

// Stop the timers
func (et *EventTimers) Stop() {

	et.lock.Lock()
	defer et.lock.Unlock()

	// Force the timer routine to quit
	ea := &EventAction{Action: "quit"}

	et.action <- ea

	et.ticker.Stop()
}
