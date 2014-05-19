package object

import (
	"fmt"
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

// tests

func TestDefaultCall(t *testing.T) {
	testCall(t, New())
}

func TestN2OneCall(t *testing.T) {
	testCall(t, testN2One.New())
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

func TestDefault1to1Signal(t *testing.T) {
	test1to1Signal(t, New())
}

func TestN2One1to1Signal(t *testing.T) {
	test1to1Signal(t, New())
}

func test1to1Signal(t *testing.T, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	n := 10000
	for i := 0; i < n; i++ {
		obj.Connect(fmt.Sprintf("sig-%d", i), func() bool {
			obj.i++
			return true
		})
	}
	for i := 0; i < n; i++ {
		obj.Emit(fmt.Sprintf("sig-%d", i)).Wait()
	}
	if obj.i != n {
		t.Fail()
	}
}

func TestDefault1toNSignal(t *testing.T) {
	test1toNSignal(t, New())
}

func TestN2One1toNSignal(t *testing.T) {
	test1toNSignal(t, testN2One.New())
}

func test1toNSignal(t *testing.T, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	n := 10000
	for i := 0; i < n; i++ {
		obj.Connect("signal", func() bool {
			obj.i++
			return true
		})
	}
	obj.Emit("signal").Wait()
	if obj.i != n {
		t.Fail()
	}
}

func TestDefaultNto1Signal(t *testing.T) {
	testNto1Signal(t, New())
}

func TestN2OneNto1Signal(t *testing.T) {
	testNto1Signal(t, testN2One.New())
}

func testNto1Signal(t *testing.T, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	obj.Connect("signal", func() bool {
		obj.i++
		return true
	})
	n := 10000
	for i := 0; i < n; i++ {
		obj.Emit("signal").Wait()
	}
	if obj.i != n {
		t.Fail()
	}
}

func TestDefaultArgumentedSignal(t *testing.T) {
	testArgumentedSignal(t, New())
}

func TestN2OneArgumentedSignal(t *testing.T) {
	testArgumentedSignal(t, testN2One.New())
}

func testArgumentedSignal(t *testing.T, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	obj.Connect("signal", func(i interface{}) bool {
		obj.i += i.(int)
		return true
	})
	n := 10000
	for i := 0; i < n; i++ {
		obj.Emit("signal", 1).Wait()
	}
	if obj.i != n {
		t.Fail()
	}
}

func TestDefaultOneshotSignal(t *testing.T) {
	testOneshotSignal(t, New())
}

func TestN2OneOneshotSignal(t *testing.T) {
	testOneshotSignal(t, testN2One.New())
}

func testOneshotSignal(t *testing.T, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	obj.Connect("signal", func(i interface{}) bool {
		obj.i += i.(int)
		return false
	})
	obj.Emit("signal", 8).Wait()
	obj.Emit("signal", 10).Wait()
	if obj.i != 8 {
		t.Fail()
	}
}

func TestDefaultReturnValue(t *testing.T) {
	testReturnValue(t, New())
}

func TestN2OneReturnValue(t *testing.T) {
	testReturnValue(t, testN2One.New())
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

func BenchmarkDefaultEmit(b *testing.B) {
	benchEmit(b, New())
}

func BenchmarkN2OneEmit(b *testing.B) {
	benchEmit(b, testN2One.New())
}

func benchEmit(b *testing.B, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	obj.Connect("signal", func() bool {
		return true
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj.Emit("signal").Wait()
	}
}

func BenchmarkDefaultArgumentedEmit(b *testing.B) {
	benchArgumentedEmit(b, New())
}

func BenchmarkN2OneArgumentedEmit(b *testing.B) {
	benchArgumentedEmit(b, testN2One.New())
}

func benchArgumentedEmit(b *testing.B, o *Object) {
	obj := &testObject{
		Object: o,
	}
	defer func() {
		obj.Die().Wait()
	}()
	obj.Connect("signal", func(b interface{}) bool {
		return true
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj.Emit("signal", true).Wait()
	}
}

func BenchmarkDefaultLongtimeCall(b *testing.B) {
	benchLongtimeCall(b, New)
}

func BenchmarkN2OneLongtimeCall(b *testing.B) {
	benchLongtimeCall(b, testN2One.New)
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
