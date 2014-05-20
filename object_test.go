package object

import (
	"sync"
	"testing"
	"time"
)

// object creation

type testObject struct {
	*Object
	i int
}

var testN2One = NewN2OneDriver(32)

var testN2M = NewN2MDriver(32)

// tests

func TestDefaultCall(t *testing.T) {
	testCall(t, New())
}

func TestN2OneCall(t *testing.T) {
	testCall(t, testN2One.New())
}

func TestN2MCall(t *testing.T) {
	testCall(t, testN2M.New())
}

func testCall(t *testing.T, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	n := 10000
	wg := new(sync.WaitGroup)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			obj.Call(func() {
				obj.i++
			}).Wait()
			wg.Done()
		}()
	}
	wg.Wait()
	if obj.i != n {
		t.Fail()
	}
}

func TestDefaultCallGet(t *testing.T) {
	testCallGet(t, New())
}

func TestN2OneCallGet(t *testing.T) {
	testCallGet(t, testN2One.New())
}

func TestN2MCallGet(t *testing.T) {
	testCallGet(t, testN2M.New())
}

func testCallGet(t *testing.T, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	call := obj.Call(func() interface{} {
		return obj.i
	})
	if call.Get().(int) != obj.i {
		t.Fail()
	}
}

func TestDefaultReturnValue(t *testing.T) {
	testReturnValue(t, New())
}

func TestN2OneReturnValue(t *testing.T) {
	testReturnValue(t, testN2One.New())
}

func TestN2MReturnValue(t *testing.T) {
	testReturnValue(t, testN2M.New())
}

func testReturnValue(t *testing.T, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	var ret int
	obj.Call(func() {
		ret = obj.i
	}).Wait()
	if ret != obj.i {
		t.Fail()
	}
}

// benchmarks

func BenchmarkDefaultCall(b *testing.B) {
	benchCall(b, New())
}

func BenchmarkN2OneCall(b *testing.B) {
	benchCall(b, testN2One.New())
}

func BenchmarkN2MCall(b *testing.B) {
	benchCall(b, testN2M.New())
}

func benchCall(b *testing.B, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj.Call(func() {}).Wait()
	}
}

func BenchmarkDefaultCallNoWait(b *testing.B) {
	benchCallNoWait(b, New())
}

func BenchmarkN2OneCallNoWait(b *testing.B) {
	benchCallNoWait(b, testN2One.New())
}

func BenchmarkN2MCallNoWait(b *testing.B) {
	benchCallNoWait(b, testN2M.New())
}

func benchCallNoWait(b *testing.B, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj.Call(func() {})
	}
}

func BenchmarkDefaultLongtimeCall(b *testing.B) {
	benchLongtimeCall(b, New)
}

func BenchmarkN2OneLongtimeCall(b *testing.B) {
	benchLongtimeCall(b, testN2One.New)
}

func BenchmarkN2MLongtimeCall(b *testing.B) {
	benchLongtimeCall(b, testN2M.New)
}

func benchLongtimeCall(b *testing.B, ctor func() *Object) {
	wg := new(sync.WaitGroup)
	wg.Add(b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := ctor()
		obj.Call(func() {
			time.Sleep(time.Microsecond * 200)
			wg.Done()
		})
	}
	wg.Wait()
}
