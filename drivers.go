package object

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
	Workers []*Worker
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
		var worker *Worker
		for _, w := range d.Workers {
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
			worker = NewWorker()
			d.Workers = append(d.Workers, worker)
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

type Worker struct {
	*Object
	nObjects int
	calls    chan *Call
}

func NewWorker() *Worker {
	w := &Worker{
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
