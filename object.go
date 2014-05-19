package object

import "sync"

type Object struct {
	call    func(*Call)
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
	object   *Object
	what     int
	signal   string
	fun      interface{}
	doneCond *sync.Cond
	done     bool
	arg      []interface{}
	ret      interface{}
}

var New = new(One2OneDriver).New

func (obj *Object) processCall(call *Call) (exit bool) {
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
			for i, fun := range obj.signals[call.signal] {
				if fun == nil {
					continue
				}
				ret := fun.(func(interface{}) bool)(call.arg[0])
				if !ret {
					obj.signals[call.signal][i] = nil
				}
			}
		} else {
			for i, fun := range obj.signals[call.signal] {
				if fun == nil {
					continue
				}
				ret := fun.(func() bool)()
				if !ret {
					obj.signals[call.signal][i] = nil
				}
			}
		}
	}
	call.doneCond.L.Lock()
	call.done = true
	call.doneCond.L.Unlock()
	call.doneCond.Broadcast()
	condPool.Put(call.doneCond)
	if call.what == _Die {
		return true
	}
	return false
}

func (obj *Object) Call(fun interface{}) *Call {
	call := &Call{
		object:   obj,
		what:     _Call,
		fun:      fun,
		doneCond: condPool.Get().(*sync.Cond),
	}
	obj.call(call)
	return call
}

func (obj *Object) Die() *Call {
	call := &Call{
		object:   obj,
		what:     _Die,
		doneCond: condPool.Get().(*sync.Cond),
	}
	obj.call(call)
	return call
}

func (obj *Object) Connect(signal string, fun interface{}) *Call {
	call := &Call{
		object:   obj,
		what:     _Connect,
		signal:   signal,
		fun:      fun,
		doneCond: condPool.Get().(*sync.Cond),
	}
	obj.call(call)
	return call
}

func (obj *Object) Emit(signal string, arg ...interface{}) *Call {
	call := &Call{
		object:   obj,
		what:     _Emit,
		signal:   signal,
		doneCond: condPool.Get().(*sync.Cond),
		arg:      arg,
	}
	obj.call(call)
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
