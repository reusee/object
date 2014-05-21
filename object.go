package object

import "sync"

type Object struct {
	call func(_Callable)
}

var New = new(One2OneDriver).New

type _Call func()

func (c _Call) Call() {
	c()
}

func (obj *Object) Call(fun func()) {
	obj.call(_Call(fun))
}

type _SyncedCall struct {
	lock sync.Mutex
	fun  func()
}

func (c *_SyncedCall) Call() {
	c.fun()
	c.lock.Unlock()
}

func (obj *Object) SyncedCall(fun func()) {
	call := new(_SyncedCall)
	call.lock.Lock()
	call.fun = fun
	obj.call(call)
	call.lock.Lock()
}

type _FutureCall struct {
	cond *sync.Cond
	done bool
	ret  interface{}
	fun  func() interface{}
}

func (c *_FutureCall) Call() {
	c.ret = c.fun()
	c.cond.L.Lock()
	c.done = true
	c.cond.L.Unlock()
	c.cond.Broadcast()
}

func (c *_FutureCall) Get() interface{} {
	c.cond.L.Lock()
	if !c.done {
		c.cond.Wait()
	}
	c.cond.L.Unlock()
	return c.ret
}

func (obj *Object) FutureCall(fun func() interface{}) (future func() interface{}) {
	call := &_FutureCall{
		cond: sync.NewCond(new(sync.Mutex)),
		fun:  fun,
	}
	obj.call(call)
	return call.Get
}

func (obj *Object) Die() {
	obj.call(nil)
}
