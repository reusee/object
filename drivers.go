package object

import "sync"

type Driver interface {
	New() *Object
}

type _Callable interface {
	Call()
}

// per object per goroutine

type One2OneDriver struct {
}

func (d *One2OneDriver) New() *Object {
	calls := make(chan _Callable, 128)
	obj := &Object{
		call: func(call _Callable) {
			calls <- call
		},
	}
	go func() {
		for call := range calls {
			if call == nil {
				return
			}
			call.Call()
		}
	}()
	return obj
}

// one goroutine for n objects

type N2OneDriver struct {
	N       int
	workers chan *_Worker
	lock    sync.Mutex
}

func NewN2OneDriver(n int) *N2OneDriver {
	return &N2OneDriver{
		N:       n,
		workers: make(chan *_Worker),
	}
}

func (d *N2OneDriver) New() (obj *Object) {
	var worker *_Worker
	d.lock.Lock()
	select {
	case worker = <-d.workers:
	default:
		d.newWorker()
		worker = <-d.workers
	}
	d.lock.Unlock()
	obj = &Object{
		call: func(call _Callable) {
			worker.calls <- call
		},
	}
	return obj
}

type _Worker struct {
	calls chan _Callable
}

func (d *N2OneDriver) newWorker() *_Worker {
	w := &_Worker{
		calls: make(chan _Callable, d.N*128),
	}
	nObjects := 0
	go func() {
		for {
			if nObjects < d.N { // available
				select {
				case d.workers <- w:
					nObjects++
				case call := <-w.calls:
					if call == nil {
						nObjects--
					} else {
						call.Call()
					}
				}
			} else { // not available
				call := <-w.calls
				if call == nil {
					nObjects--
				} else {
					call.Call()
				}
			}
		}
	}()
	return w
}

// n:m driver

type N2MDriver struct {
	runnables chan *_Runnable
}

func NewN2MDriver(n int) *N2MDriver {
	d := &N2MDriver{
		runnables: make(chan *_Runnable, 128),
	}
	for i := 0; i < n; i++ {
		go func() {
			for runnable := range d.runnables {
				for {
					runnable.lock.Lock()
					if len(runnable.calls) == 0 {
						runnable.state = _StateSleep
						runnable.lock.Unlock()
						break
					}
					if f := runnable.calls[0]; f != nil {
						f.Call()
					}
					runnable.calls = runnable.calls[1:]
					runnable.lock.Unlock()
				}
			}
		}()
	}
	return d
}

type _Runnable struct {
	state int
	lock  *sync.Mutex
	calls []_Callable
}

const (
	_StateSleep = iota
	_StateReady
)

func (d *N2MDriver) New() *Object {
	runnable := &_Runnable{
		state: _StateSleep,
		lock:  new(sync.Mutex),
	}
	obj := &Object{
		call: func(call _Callable) {
			runnable.lock.Lock()
			runnable.calls = append(runnable.calls, call)
			if runnable.state == _StateSleep {
				runnable.state = _StateReady
				d.runnables <- runnable
			}
			runnable.lock.Unlock()
		},
	}
	return obj
}
