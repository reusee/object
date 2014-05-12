package object

import "sync"

type Object struct {
	calls   chan *Call
	signals map[string][]func()
}

const (
	_Call = iota
	_Die
	_Connect
	_Emit
)

type Call struct {
	what     int
	signal   string
	fun      func()
	doneCond *sync.Cond
	done     bool
}

func New() *Object {
	obj := &Object{
		calls:   make(chan *Call, 128),
		signals: make(map[string][]func()),
	}
	go func() {
		for call := range obj.calls {
			switch call.what {
			case _Call:
				call.fun()
			case _Connect:
				obj.signals[call.signal] = append(obj.signals[call.signal], call.fun)
			case _Emit:
				for _, fun := range obj.signals[call.signal] {
					fun()
				}
			}
			call.doneCond.L.Lock()
			call.done = true
			call.doneCond.L.Unlock()
			call.doneCond.Broadcast()
			if call.what == _Die {
				break
			}
		}
	}()
	return obj
}

func (obj *Object) Call(fun func()) *Call {
	call := &Call{
		what:     _Call,
		fun:      fun,
		doneCond: sync.NewCond(new(sync.Mutex)),
	}
	obj.calls <- call
	return call
}

func (obj *Object) Die() *Call {
	call := &Call{
		what:     _Die,
		doneCond: sync.NewCond(new(sync.Mutex)),
	}
	obj.calls <- call
	return call
}

func (obj *Object) Connect(signal string, fun func()) *Call {
	call := &Call{
		what:     _Connect,
		signal:   signal,
		fun:      fun,
		doneCond: sync.NewCond(new(sync.Mutex)),
	}
	obj.calls <- call
	return call
}

func (obj *Object) Emit(signal string) *Call {
	call := &Call{
		what:     _Emit,
		signal:   signal,
		doneCond: sync.NewCond(new(sync.Mutex)),
	}
	obj.calls <- call
	return call
}

func (call *Call) Wait() {
	call.doneCond.L.Lock()
	if !call.done {
		call.doneCond.Wait()
	}
	call.doneCond.L.Unlock()
}
