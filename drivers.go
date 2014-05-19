package object

import "sync"

type Driver interface {
	New() *Object
}

// per object per goroutine

type One2OneDriver struct {
}

func (d *One2OneDriver) New() *Object {
	calls := make(chan *Call, 128)
	obj := &Object{
		call: func(call *Call) {
			calls <- call
		},
		signals: make(map[string][]interface{}),
	}
	go func() {
		for call := range calls {
			if obj.processCall(call) { // object is dead
				return
			}
		}
	}()
	return obj
}

// one goroutine for n objects

type N2OneDriver struct {
	*Object // thread-safe is needed
	N       int
	workers []*_Worker
}

func NewN2OneDriver(n int) *N2OneDriver {
	return &N2OneDriver{
		Object: New(),
		N:      n,
	}
}

func (d *N2OneDriver) New() (obj *Object) {
	d.Call(func() {
		// select a worker
		var worker *_Worker
		for _, w := range d.workers {
			w.Call(func() {
				if w.nObjects < d.N {
					worker = w
					w.nObjects++
				}
			}).Wait()
			if worker != nil {
				break
			}
		}
		// none selected, create one
		if worker == nil {
			worker = newWorker()
			d.workers = append(d.workers, worker)
			worker.Call(func() {
				worker.nObjects++
			})
		}
		// create object
		obj = &Object{
			call: func(call *Call) {
				worker.calls <- call
			},
			signals: make(map[string][]interface{}),
		}
	}).Wait()
	return obj
}

type _Worker struct {
	*Object
	nObjects int
	calls    chan *Call
}

func newWorker() *_Worker {
	w := &_Worker{
		Object: New(),
		calls:  make(chan *Call),
	}
	go func() {
		for call := range w.calls {
			if call.object.processCall(call) { // object is dead
				w.Call(func() {
					w.nObjects--
				})
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
					runnable.object.processCall(runnable.calls[0])
					runnable.calls = runnable.calls[1:]
					runnable.lock.Unlock()
				}
			}
		}()
	}
	return d
}

type _Runnable struct {
	state  int
	lock   *sync.Mutex
	calls  []*Call
	object *Object
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
		call: func(call *Call) {
			runnable.lock.Lock()
			runnable.calls = append(runnable.calls, call)
			if runnable.state == _StateSleep {
				runnable.state = _StateReady
				d.runnables <- runnable
			}
			runnable.lock.Unlock()
		},
		signals: make(map[string][]interface{}),
	}
	runnable.object = obj
	return obj
}
