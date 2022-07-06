package poll

import "sync"

type pollingHandlerState struct {
	mut                sync.RWMutex
	numErrors          int
	numAlarmsWithError int
	isRunning          bool
}

func (state *pollingHandlerState) incrementErrors() {
	state.mut.Lock()
	state.numErrors++
	state.mut.Unlock()
}

func (state *pollingHandlerState) incrementAlarmsWithError() {
	state.mut.Lock()
	state.numAlarmsWithError++
	state.mut.Unlock()
}

func (state *pollingHandlerState) getNumErrors() int {
	state.mut.RLock()
	defer state.mut.RUnlock()

	return state.numErrors
}

func (state *pollingHandlerState) getNumAlarmsWithError() int {
	state.mut.RLock()
	defer state.mut.RUnlock()

	return state.numAlarmsWithError
}

func (state *pollingHandlerState) resetNumErrors() {
	state.mut.Lock()
	state.numAlarmsWithError = 0
	state.numErrors = 0
	state.mut.Unlock()
}

func (state *pollingHandlerState) setIsRunning() {
	state.mut.Lock()
	state.isRunning = true
	state.mut.Unlock()
}

func (state *pollingHandlerState) setIsStopped() {
	state.mut.Lock()
	state.isRunning = false
	state.mut.Unlock()
}

// IsRunning returns true if the main processLoop is running
func (state *pollingHandlerState) IsRunning() bool {
	state.mut.RLock()
	defer state.mut.RUnlock()

	return state.isRunning
}
