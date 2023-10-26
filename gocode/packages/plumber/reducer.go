package plumber

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type ReducerWork func() error

type Reducer interface {
	SetDbgOut(w io.Writer)
	Launch(label string, work ReducerWork) error
	Close() error
}

type unscalableReducer struct {
	limit          int
	maxSleep       time.Duration
	mux            sync.Mutex
	wg             sync.WaitGroup
	done           chan struct{}
	nextId         int64
	queue          map[int64]string
	waitTimes      map[int64]time.Duration
	durations      map[int64]time.Duration
	durationLabels map[int64]string
	meanTime       time.Duration
	dbgOut         io.Writer
}

func (usRed *unscalableReducer) SetDbgOut(w io.Writer) { usRed.dbgOut = w }

func (usRed *unscalableReducer) printDbg(s string) {
	if usRed.dbgOut != nil {
		fmt.Fprintf(usRed.dbgOut, "%s %s\n", time.Now().Format("2006-01-02T15:04:05.000"), s)
	}
}

func (usRed *unscalableReducer) newDuration(label string, id, start, end int64) {
	usRed.mux.Lock()
	ld := usRed.limit * 2
	if ld < 20 {
		ld = 20
	}
	if len(usRed.durations) >= ld {
		min := usRed.nextId
		for id, _ := range usRed.durations {
			if id < min {
				min = id
			}
		}
		delete(usRed.durations, id)
	}
	usRed.durations[id] = time.Duration(end - start)
	total := time.Duration(0)
	for _, duration := range usRed.durations {
		total += duration
	}
	meanTime := total / time.Duration(len(usRed.durations))
	if meanTime > usRed.meanTime {
		fmt.Fprintf(io.Discard, "meantime %d\n", meanTime)
	}
	usRed.meanTime = meanTime
	delete(usRed.queue, id)
	delete(usRed.waitTimes, id)
	usRed.wg.Done()
	usRed.mux.Unlock()
	if usRed.dbgOut != nil {
		usRed.printDbg(fmt.Sprintf("%-12s id %d label %s duration %d", "newDuration", id, label, end-start))
	}
}

func (usRed *unscalableReducer) waitPosition(id int64) int {
	position := 1
	for otherId, _ := range usRed.queue {
		if otherId < id {
			position++
		}
	}
	if position <= usRed.limit {
		return 0
	}
	return position - usRed.limit
}

func (usRed *unscalableReducer) waitTime(id int64) time.Duration {
	meanTime := usRed.meanTime
	wt := time.Duration(meanTime) * time.Duration(usRed.waitPosition(id))
	if wt > usRed.maxSleep {
		wt = usRed.maxSleep
	}
	fmt.Fprintf(io.Discard, "waitTime %d mean %d pos %d for %d %s\n", wt/1e6, meanTime/1e6, usRed.waitPosition(id), id, usRed.queue[id])
	return wt
}

func (usRed *unscalableReducer) doLaunch(label string, id int64, work ReducerWork) error {
	usRed.wg.Add(1)
	usRed.printDbg(fmt.Sprintf("%-12s id %d label %s", "doLaunch", id, label))
	usRed.mux.Unlock()
	start := time.Now().UnixNano()
	err := work()
	end := time.Now().UnixNano()
	usRed.newDuration(label, id, start, end)
	return err
}

func (usRed *unscalableReducer) waitAndLaunch(label string, id int64, work ReducerWork) error {
	for true {
		wt, ok := usRed.waitTimes[id]
		if !ok {
			wt = usRed.waitTime(id)
			usRed.waitTimes[id] = wt
		}
		usRed.mux.Unlock()
		select {
		case <-usRed.done:
			wt2, _ := usRed.waitTimes[id]
			return fmt.Errorf("waiting %d ms I/O %d %s aborted", wt2/1e6, id, label)
		case <-time.After(wt):
			usRed.mux.Lock()
			p := usRed.waitPosition(id)
			usRed.printDbg(fmt.Sprintf("%-12s id %d label %s after %d pos %d queue %v", "After", id, label, wt/1e6, p, usRed.queue))
			return usRed.doLaunch(label, id, work)
		}
	}
	panic("logic")
}

func (usRed *unscalableReducer) Launch(label string, work ReducerWork) error {
	usRed.mux.Lock()
	id := usRed.nextId
	usRed.queue[id] = label
	usRed.printDbg(fmt.Sprintf("%-12s id %d label %s", "Launch", id, label))
	usRed.nextId++
	if len(usRed.queue) < usRed.limit {
		return usRed.doLaunch(label, id, work)
	}
	return usRed.waitAndLaunch(label, id, work)
}

func (usRed *unscalableReducer) Close() error {
	usRed.mux.Lock()
	max := int64(-1)
	min := int64(9223372036854775807)
	for id := range usRed.queue {
		if id < min {
			min = id
		}
		if id > max {
			max = id
		}
	}
	if min > max {
		min = max
	}
	cur := min
	for cur < max {
		for id := range usRed.queue {
			if id == cur {
				wt, _ := usRed.waitTimes[id]
				fmt.Fprintf(os.Stderr, "queue wt %d %d %s\n", wt/1e6, id, usRed.queue[id])
				break
			}
		}
		cur++
	}

	usRed.mux.Unlock()
	close(usRed.done)
	usRed.wg.Wait()
	return nil
}

func NewUsReducer(limit int, maxSleep time.Duration) Reducer {
	if limit == 0 {
		limit = 10
	}
	if maxSleep == 0 {
		maxSleep = time.Duration(100) * time.Duration(time.Second)
	}
	red := &unscalableReducer{
		limit:     limit,
		maxSleep:  maxSleep,
		queue:     map[int64]string{},
		waitTimes: map[int64]time.Duration{},
		durations: map[int64]time.Duration{},
		done:      make(chan struct{}),
	}
	return red
}

type srqe struct {
	id       int64
	label    string
	callback chan struct{}
}

type scalableReducer struct {
	limit      int
	mux        sync.Mutex
	wg         sync.WaitGroup
	done       chan struct{}
	isDone     bool
	nextId     int64
	request    chan srqe
	available  chan srqe
	actives    map[int64]srqe
	terminated chan srqe
	dbgOut     io.Writer
}

func (red *scalableReducer) SetDbgOut(w io.Writer) { red.dbgOut = w }

func (red *scalableReducer) printDbg(s string) {
	if red.dbgOut != nil {
		fmt.Fprintf(red.dbgOut, "%s %s\n", time.Now().Format("2006-01-02T15:04:05.000"), s)
	}
}

func (red *scalableReducer) Launch(label string, work ReducerWork) error {
	red.printDbg(fmt.Sprintf("%-12s label %s", "Launch 0", label))
	red.mux.Lock()
	id := red.nextId
	red.nextId++
	red.wg.Add(1)
	red.printDbg(fmt.Sprintf("%-12s id %d label %s", "Launch 1", id, label))
	red.mux.Unlock()
	qe := srqe{
		id:       id,
		label:    label,
		callback: make(chan struct{}),
	}
	red.printDbg(fmt.Sprintf("%-12s requested %d %s", "Launch 2", id, label))
	red.request <- qe
	red.printDbg(fmt.Sprintf("%-12s loop %d %s", "Launch 3", id, label))
	<-qe.callback
	red.printDbg(fmt.Sprintf("%-12s loop %d %s", "Launch 4", id, label))
	var err error
	if !red.isDone {
		err = work()
	} else {
		err = fmt.Errorf("waiting I/O aborted for %s (%d)", label, id)
	}
	red.printDbg(fmt.Sprintf("%-12s loop %d %s err %v", "Launch 5", id, label, err))
	red.terminated <- qe
	red.printDbg(fmt.Sprintf("%-12s loop %d %s", "Launch 6", id, label))
	<-qe.callback
	red.printDbg(fmt.Sprintf("%-12s loop %d %s", "Launch 7", id, label))
	close(qe.callback)
	red.wg.Done()
	return err
}

type srScheduler struct {
	red        *scalableReducer
	available  bool
	empty      bool
	done       bool
	terminated bool
}

func (srs *srScheduler) update() {
	srs.available = len(srs.red.actives) < srs.red.limit
	srs.empty = len(srs.red.actives) == 0
}

func (srs *srScheduler) handleRequest(client string, qe srqe) {
	srs.red.mux.Lock()
	srs.red.printDbg(fmt.Sprintf("%-12s request %d %s %d / %d", client, qe.id, qe.label, len(srs.red.actives), srs.red.limit))
	srs.red.actives[qe.id] = qe
	srs.update()
	srs.red.mux.Unlock()
	qe.callback <- struct{}{}
}

func (srs *srScheduler) handleTerminated(client string, qe srqe) {
	srs.red.mux.Lock()
	srs.red.printDbg(fmt.Sprintf("%-12s terminated %d %s %d / %d", client, qe.id, qe.label, len(srs.red.actives), srs.red.limit))
	delete(srs.red.actives, qe.id)
	srs.update()
	srs.red.mux.Unlock()
	qe.callback <- struct{}{}

}
func (srs *srScheduler) handleAvailable() {
	for {
		srs.red.printDbg(fmt.Sprintf("%-12s loop", "handleAvailable"))
		select {
		case <-srs.red.done:
			srs.red.printDbg(fmt.Sprintf("%-12s done", "handleAvailable"))
			srs.done = true
			return
		case qe := <-srs.red.request:
			srs.handleRequest("handleAvailable", qe)
		case qe := <-srs.red.terminated:
			srs.handleTerminated("handleAvailable", qe)
		}
		if !srs.available {
			return
		}
	}
}

func (srs *srScheduler) handleDrainAvailable() {
	for {
		srs.red.printDbg(fmt.Sprintf("%-12s loop", "handleDrainAvailable"))
		select {
		case qe := <-srs.red.request:
			srs.handleRequest("handleDrainAvailable", qe)
		case qe := <-srs.red.terminated:
			srs.handleTerminated("handleDrainAvailable", qe)
		}
		if !srs.available || srs.empty {
			return
		}
	}
}

func (srs *srScheduler) handleQueuing() {
	for {
		srs.red.printDbg(fmt.Sprintf("%-12s loop", "handleQueuing"))
		select {
		case <-srs.red.done:
			srs.red.printDbg(fmt.Sprintf("%-12s done", "handleQueuing"))
			srs.done = true
			return
		case qe := <-srs.red.terminated:
			srs.handleTerminated("handleQueuing", qe)
		}
		if srs.available {
			return
		}
	}
}

func (srs *srScheduler) handleDrainQueuing() {
	for {
		srs.red.printDbg(fmt.Sprintf("%-12s loop", "handleDrainQueuing"))
		select {
		case qe := <-srs.red.terminated:
			srs.handleTerminated("handleDrainQueuing", qe)
		}
		if srs.available {
			return
		}
	}
}

func (red *scalableReducer) schedule() {
	srs := srScheduler{
		red:  red,
		done: false,
	}
	srs.red.mux.Lock()
	srs.update()
	srs.red.mux.Unlock()
	for !srs.terminated {
		srs.red.printDbg(fmt.Sprintf("%-12s loop", "schedule"))
		if srs.available && !srs.done {
			srs.handleAvailable()
			continue
		}
		if srs.available && srs.done && !srs.empty {
			srs.handleDrainAvailable()
			continue
		}
		if !srs.available && !srs.done {
			srs.handleQueuing()
			continue
		}
		if !srs.available && srs.done {
			srs.handleDrainQueuing()
			continue
		}
		srs.terminated = true
	}
	srs.red.printDbg(fmt.Sprintf("%-12s terminated", "schedule"))
	red.wg.Done()
}

func (red *scalableReducer) Close() error {
	red.printDbg(fmt.Sprintf("%-12s 1", "close"))
	red.mux.Lock()
	red.isDone = true
	red.mux.Unlock()
	close(red.done)
	red.wg.Wait()
	red.printDbg(fmt.Sprintf("%-12s 2", "close"))
	return nil
}

func NewReducer(limit int, maxSleep time.Duration) Reducer {
	if limit == 0 {
		limit = 10
	}
	red := &scalableReducer{
		limit:      limit,
		done:       make(chan struct{}),
		request:    make(chan srqe),
		actives:    map[int64]srqe{},
		terminated: make(chan srqe),
	}
	red.wg.Add(1)
	go red.schedule()
	return red
}
