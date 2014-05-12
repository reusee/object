package object

//TODO receive channels

import (
	"reflect"
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

type Call struct {
	what     int
	signal   string
	fun      interface{}
	doneCond *sync.Cond
	done     bool
	args     []interface{}
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
				call.fun.(func())()
			case _Connect:
				obj.signals[call.signal] = append(obj.signals[call.signal], call.fun)
			case _Emit:
				if len(call.args) > 0 {
					var argValues []reflect.Value
					for _, arg := range call.args {
						argValues = append(argValues, reflect.ValueOf(arg))
					}
					for _, fun := range obj.signals[call.signal] {
						reflect.ValueOf(fun).Call(argValues)
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

func (obj *Object) Connect(signal string, fun interface{}) *Call {
	call := &Call{
		what:     _Connect,
		signal:   signal,
		fun:      fun,
		doneCond: sync.NewCond(new(sync.Mutex)),
	}
	obj.calls <- call
	return call
}

func (obj *Object) Emit(signal string, args ...interface{}) *Call {
	call := &Call{
		what:     _Emit,
		signal:   signal,
		doneCond: sync.NewCond(new(sync.Mutex)),
		args:     args,
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
