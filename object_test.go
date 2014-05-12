package object

import (
	"fmt"
	"sync"
	"testing"
)

type testObject struct {
	*Object
	i int
}

func TestCall(t *testing.T) {
	obj := &testObject{
		Object: New(),
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

func TestCallGet(t *testing.T) {
	obj := &testObject{
		Object: New(),
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

func Test1to1Signal(t *testing.T) {
	obj := &testObject{
		Object: New(),
	}
	defer func() {
		obj.Die().Wait()
	}()

	n := 10000
	for i := 0; i < n; i++ {
		obj.Connect(fmt.Sprintf("sig-%d", i), func() {
			obj.i++
		})
	}
	for i := 0; i < n; i++ {
		obj.Emit(fmt.Sprintf("sig-%d", i)).Wait()
	}
	if obj.i != n {
		t.Fail()
	}
}

func Test1toNSignal(t *testing.T) {
	obj := &testObject{
		Object: New(),
	}
	defer func() {
		obj.Die().Wait()
	}()

	n := 10000
	for i := 0; i < n; i++ {
		obj.Connect("signal", func() {
			obj.i++
		})
	}
	obj.Emit("signal").Wait()
	if obj.i != n {
		t.Fail()
	}
}

func TestNto1Signal(t *testing.T) {
	obj := &testObject{
		Object: New(),
	}
	defer func() {
		obj.Die().Wait()
	}()

	obj.Connect("signal", func() {
		obj.i++
	})
	n := 10000
	for i := 0; i < n; i++ {
		obj.Emit("signal").Wait()
	}
	if obj.i != n {
		t.Fail()
	}
}

func TestArgumentedSiganl(t *testing.T) {
	obj := &testObject{
		Object: New(),
	}
	defer func() {
		obj.Die().Wait()
	}()

	obj.Connect("signal", func(i interface{}) {
		obj.i += i.(int)
	})
	n := 10000
	for i := 0; i < n; i++ {
		obj.Emit("signal", 1).Wait()
	}
	if obj.i != n {
		t.Fail()
	}
}

func TestReturnValue(t *testing.T) {
	obj := &testObject{
		Object: New(),
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

func BenchmarkCall(b *testing.B) {
	obj := &testObject{
		Object: New(),
	}
	defer func() {
		obj.Die().Wait()
	}()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		obj.Call(func() {}).Wait()
	}
}

func BenchmarkEmit(b *testing.B) {
	obj := &testObject{
		Object: New(),
	}
	defer func() {
		obj.Die().Wait()
	}()
	obj.Connect("signal", func() {})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		obj.Emit("signal").Wait()
	}
}

func BenchmarkArgumentedEmit(b *testing.B) {
	obj := &testObject{
		Object: New(),
	}
	defer func() {
		obj.Die().Wait()
	}()
	obj.Connect("signal", func(b interface{}) {})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		obj.Emit("signal", true).Wait()
	}
}

func BenchmarkBaseline(b *testing.B) {
	for i := 0; i < b.N; i++ {
		func() {
		}()
	}
}
