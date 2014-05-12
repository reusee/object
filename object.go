package object

//TODO receive channels

import (
	"sync"
)

type Object struct {
	calls   chan *Call
	signals map[string][]interface{}
}

const (
	_Call = iota
	_Die
	_Connect
	_Emit
)

var condPool = sync.Pool{
	New: func() interface{} {
		return sync.NewCond(new(sync.Mutex))
	},
}

type Call struct {
	what     int
	signal   string
	fun      interface{}
	doneCond *sync.Cond
	done     bool
	arg      []interface{}
	ret      interface{}
}

func New() *Object {
	obj := &Object{
		calls:   make(chan *Call, 128),
		signals: make(map[string][]interface{}),
	}
	go func() {
		for call := range obj.calls {
			switch call.what {
			case _Call:
				switch f := call.fun.(type) {
				case func() interface{}:
					call.ret = f()
				case func():
					f()
				default:
					panic("wrong closure type")
				}
			case _Connect:
				obj.signals[call.signal] = append(obj.signals[call.signal], call.fun)
			case _Emit:
				if len(call.arg) > 0 {
					for _, fun := range obj.signals[call.signal] {
						fun.(func(interface{}))(call.arg[0])
					}
				} else {
					for _, fun := range obj.signals[call.signal] {
						fun.(func())()
					}
				}
			}
			call.doneCond.L.Lock()
			call.done = true
			call.doneCond.L.Unlock()
			call.doneCond.Broadcast()
			condPool.Put(call.doneCond)
			if call.what == _Die {
				break
			}
		}
	}()
	return obj
}

func (obj *Object) Call(fun interface{}) *Call {
	call := &Call{
		what:     _Call,
		fun:      fun,
		doneCond: condPool.Get().(*sync.Cond),
	}
	obj.calls <- call
	return call
}

func (obj *Object) Die() *Call {
	call := &Call{
		what:     _Die,
		doneCond: condPool.Get().(*sync.Cond),
	}
	obj.calls <- call
	return call
}

func (obj *Object) Connect(signal string, fun interface{}) *Call {
	call := &Call{
		what:     _Connect,
		signal:   signal,
		fun:      fun,
		doneCond: condPool.Get().(*sync.Cond),
	}
	obj.calls <- call
	return call
}

func (obj *Object) Emit(signal string, arg ...interface{}) *Call {
	call := &Call{
		what:     _Emit,
		signal:   signal,
		doneCond: condPool.Get().(*sync.Cond),
		arg:      arg,
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

func (call *Call) Get() interface{} {
	call.Wait()
	return call.ret
}
