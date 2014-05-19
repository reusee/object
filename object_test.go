package object

import (
	"fmt"
	"sync"
	"testing"
)

// object creation

type testObject struct {
	*Object
	i int
}

func testNewDefaultObject() *Object {
	return New()
}

var testOG4NDriver = NewOneGoroutineForNObjects(32)

func testNewOG4NObject() *Object {
	return NewWithDriver(testOG4NDriver)
}

// tests

func TestDefaultCall(t *testing.T) {
	testCall(t, testNewDefaultObject())
}

func TestOG4NCall(t *testing.T) {
	testCall(t, testNewOG4NObject())
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
	testCallGet(t, testNewDefaultObject())
}

func TestOG4NCallGet(t *testing.T) {
	testCallGet(t, testNewOG4NObject())
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
	test1to1Signal(t, testNewDefaultObject())
}

func TestOG4N1to1Signal(t *testing.T) {
	test1to1Signal(t, testNewDefaultObject())
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
	test1toNSignal(t, testNewDefaultObject())
}

func TestOG4N1toNSignal(t *testing.T) {
	test1toNSignal(t, testNewOG4NObject())
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
	testNto1Signal(t, testNewDefaultObject())
}

func TestOG4NNto1Signal(t *testing.T) {
	testNto1Signal(t, testNewOG4NObject())
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
	testArgumentedSignal(t, testNewDefaultObject())
}

func TestOG4NArgumentedSignal(t *testing.T) {
	testArgumentedSignal(t, testNewOG4NObject())
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
	testOneshotSignal(t, testNewDefaultObject())
}

func TestOG4NOneshotSignal(t *testing.T) {
	testOneshotSignal(t, testNewOG4NObject())
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
	testReturnValue(t, testNewDefaultObject())
}

func TestOG4NReturnValue(t *testing.T) {
	testReturnValue(t, testNewOG4NObject())
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
	benchCall(b, testNewDefaultObject())
}

func BenchmarkOG4NCall(b *testing.B) {
	benchCall(b, testNewOG4NObject())
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
	benchCallNoWait(b, testNewDefaultObject())
}

func BenchmarkOG4NCallNoWait(b *testing.B) {
	benchCallNoWait(b, testNewOG4NObject())
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
	benchEmit(b, testNewDefaultObject())
}

func BenchmarkOG4NEmit(b *testing.B) {
	benchEmit(b, testNewOG4NObject())
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
	benchArgumentedEmit(b, testNewDefaultObject())
}

func BenchmarkOG4NArgumentedEmit(b *testing.B) {
	benchArgumentedEmit(b, testNewOG4NObject())
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
