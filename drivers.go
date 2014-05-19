package object

//TODO m goroutine for n object driver

type Driver interface {
	New() *Object
}

// per object per goroutine

type PerObjectPerGoroutine struct {
}

func (d *PerObjectPerGoroutine) New() *Object {
	obj := &Object{
		calls:   make(chan *Call, 128),
		signals: make(map[string][]interface{}),
	}
	go func() {
		for call := range obj.calls {
			if obj.processCall(call) { // object is dead
				return
			}
		}
	}()
	return obj
}

// one goroutine for n objects

type OneGoroutineForNObjects struct {
	*Object // thread-safe is needed
	N       int
	Workers []*Worker
}

func NewOneGoroutineForNObjects(n int) *OneGoroutineForNObjects {
	return &OneGoroutineForNObjects{
		Object: New(),
		N:      n,
	}
}

func (d *OneGoroutineForNObjects) New() (obj *Object) {
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
			calls:   worker.calls,
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
