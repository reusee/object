package object

import (
	"sync"
	"testing"
)

func TestAll(t *testing.T) {
	drivers := []Driver{
		new(One2OneDriver),
		NewN2OneDriver(32),
		NewN2MDriver(128),
	}
	tests := []func(*testing.T, Driver){
		testCall,
		testSyncedCall,
		testFutureCall,
	}
	for _, test := range tests {
		for _, driver := range drivers {
			test(t, driver)
		}
	}
}

// tests

type testObject struct {
	*Object
	i int
}

func testCall(t *testing.T, d Driver) {
	obj := &testObject{
		Object: d.New(),
	}
	defer obj.Die()
	n := 102400
	for i := 0; i < n; i++ {
		obj.Call(func() {
			obj.i++
		})
	}
	wait := make(chan bool)
	obj.Call(func() {
		close(wait)
	})
	<-wait
	if obj.i != n {
		t.Fatal("Call")
	}
}

func testSyncedCall(t *testing.T, d Driver) {
	obj := &testObject{
		Object: d.New(),
	}
	defer obj.Die()
	n := 102400
	for i := 0; i < n; i++ {
		obj.SyncedCall(func() {
			obj.i++
		})
	}
	if obj.i != n {
		t.Fatal("SyncedCall")
	}
}

func testFutureCall(t *testing.T, d Driver) {
	obj := d.New()
	val := obj.FutureCall(func() interface{} {
		return "foobar"
	})
	n := 1024
	wg := new(sync.WaitGroup)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			if val().(string) != "foobar" {
				t.Fatal("FutureCall")
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

// benchmarks

var benchNto1Driver = NewN2OneDriver(32)
var benchNtoMDriver = NewN2MDriver(128)

func Benchmark1to1Call(b *testing.B) {
	benchCall(b, New())
}

func BenchmarkNto1Call(b *testing.B) {
	benchCall(b, benchNto1Driver.New())
}

func BenchmarkNtoMCall(b *testing.B) {
	benchCall(b, benchNtoMDriver.New())
}

func benchCall(b *testing.B, obj *Object) {
	for i := 0; i < b.N; i++ {
		obj.Call(func() {})
	}
}

func Benchmark1to1SyncedCall(b *testing.B) {
	benchSyncedCall(b, New())
}

func BenchmarkNto1SyncedCall(b *testing.B) {
	benchSyncedCall(b, benchNto1Driver.New())
}

func BenchmarkNtoMSyncedCall(b *testing.B) {
	benchSyncedCall(b, benchNtoMDriver.New())
}

func benchSyncedCall(b *testing.B, obj *Object) {
	for i := 0; i < b.N; i++ {
		obj.SyncedCall(func() {})
	}
}

func Benchmark1to1FutureCall(b *testing.B) {
	benchFutureCall(b, New())
}

func BenchmarkNto1FutureCall(b *testing.B) {
	benchFutureCall(b, benchNto1Driver.New())
}

func BenchmarkNtoMFutureCall(b *testing.B) {
	benchFutureCall(b, benchNtoMDriver.New())
}

func benchFutureCall(b *testing.B, obj *Object) {
	for i := 0; i < b.N; i++ {
		val := obj.FutureCall(func() interface{} {
			return true
		})
		_ = val().(bool)
	}
}
