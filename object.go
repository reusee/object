package object

import "sync"

type Object struct {
	call func(func())
}

var New = new(One2OneDriver).New

func (obj *Object) Call(fun func()) {
	obj.call(fun)
}

func (obj *Object) SyncedCall(fun func()) {
	var lock sync.Mutex
	lock.Lock()
	obj.call(func() {
		fun()
		lock.Unlock()
	})
	lock.Lock()
}

func (obj *Object) FutureCall(fun func() interface{}) (future func() interface{}) {
	cond := sync.NewCond(new(sync.Mutex))
	var done bool
	var ret interface{}
	obj.call(func() {
		ret = fun()
		cond.L.Lock()
		done = true
		cond.L.Unlock()
		cond.Broadcast()
	})
	future = func() interface{} {
		cond.L.Lock()
		if !done {
			cond.Wait()
		}
		cond.L.Unlock()
		return ret
	}
	return
}

func (obj *Object) Die() {
	obj.call(nil)
}
