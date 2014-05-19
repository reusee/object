package object

import "reflect"

type Driver interface {
	Drive(obj *Object)
}

// per object per goroutine

type PerObjectPerGoroutine struct {
}

func (d *PerObjectPerGoroutine) Drive(obj *Object) {
	go func() {
		for call := range obj.calls {
			if obj.processCall(call) {
				break
			}
		}
	}()
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

type Worker struct {
	Objects   []*Object
	Cases     []reflect.SelectCase
	NewObject chan *Object
}

func NewWorker() *Worker {
	newObjChan := make(chan *Object)
	w := &Worker{
		NewObject: newObjChan,
		Cases: []reflect.SelectCase{
			{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(newObjChan)},
		},
	}
	go func() {
		for {
			n, recv, ok := reflect.Select(w.Cases)
			switch {
			// new object
			case n == 0:
				obj := recv.Interface().(*Object)
				w.Objects = append(w.Objects, obj)
				w.Cases = append(w.Cases, reflect.SelectCase{
					Dir: reflect.SelectRecv, Chan: reflect.ValueOf(obj.calls),
				})

			// object call
			default:
				if !ok { // object is dead
					w.Objects = append(w.Objects[:n-1], w.Objects[n:]...)
					w.Cases = append(w.Cases[:n], w.Cases[n+1:]...)
				} else { // process call
					w.Objects[n-1].processCall(recv.Interface().(*Call))
				}
			}
		}
	}()
	return w
}

func (d *OneGoroutineForNObjects) Drive(obj *Object) {
	d.Call(func() {
		// select a worker
		var worker *Worker
		for _, w := range d.Workers {
			if len(w.Objects) < d.N {
				worker = w
				break
			}
		}
		// none selected, create one
		if worker == nil {
			worker = NewWorker()
			d.Workers = append(d.Workers, worker)
		}
		// add to worker
		worker.NewObject <- obj
	})
}
