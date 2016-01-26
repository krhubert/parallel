package parallel

import (
	"math/rand"
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

var atomicInt int32 = 0

type s struct {
	val int
}

func work(microseconds int) interface{} {
	time.Sleep(time.Duration(microseconds) * time.Microsecond)
	return &s{int(atomic.AddInt32(&atomicInt, 1))}
}

func workRandom(microseconds int) interface{} {
	if microseconds != 0 {
		time.Sleep(time.Duration(rand.Intn(microseconds)) * time.Microsecond)
	}
	return &s{int(atomic.AddInt32(&atomicInt, 1))}
}

var tests = []struct {
	worker       int // number of worker
	n            int // number of data to feed
	microseconds int // argument passed to f function
	fn           func(int) interface{}
}{
	{1, 1, 0, work},
	{1, 2, 0, work},
	{1, 5, 0, work},
	{1, 20, 0, work},
	{1, 100, 0, work},
	{1, 1, 100, work},
	{1, 2, 100, work},
	{1, 5, 100, work},
	{1, 20, 100, work},
	{1, 100, 100, work},
	{1, 1, 1000, work},
	{1, 2, 1000, work},
	{1, 5, 1000, work},
	{1, 20, 1000, work},
	{1, 100, 1000, work},

	{5, 1, 0, work},
	{5, 2, 0, work},
	{5, 5, 0, work},
	{5, 20, 0, work},
	{5, 100, 0, work},
	{5, 1, 100, work},
	{5, 2, 100, work},
	{5, 5, 100, work},
	{5, 20, 100, work},
	{5, 100, 100, work},
	{5, 1, 1000, work},
	{5, 2, 1000, work},
	{5, 5, 1000, work},
	{5, 20, 1000, work},
	{5, 100, 1000, work},

	{20, 1, 0, work},
	{20, 2, 0, work},
	{20, 5, 0, work},
	{20, 20, 0, work},
	{20, 100, 0, work},
	{20, 1, 100, work},
	{20, 2, 100, work},
	{20, 5, 100, work},
	{20, 20, 100, work},
	{20, 100, 100, work},
	{20, 1, 1000, work},
	{20, 2, 1000, work},
	{20, 5, 1000, work},
	{20, 20, 1000, work},
	{20, 100, 1000, work},

	{100, 1, 0, work},
	{100, 2, 0, work},
	{100, 5, 0, work},
	{100, 20, 0, work},
	{100, 100, 0, work},
	{100, 1, 100, work},
	{100, 2, 100, work},
	{100, 5, 100, work},
	{100, 20, 100, work},
	{100, 100, 100, work},
	{100, 1, 1000, work},
	{100, 2, 1000, work},
	{100, 5, 1000, work},
	{100, 20, 1000, work},
	{100, 100, 1000, work},

	{1, 1, 0, workRandom},
	{1, 2, 0, workRandom},
	{1, 5, 0, workRandom},
	{1, 20, 0, workRandom},
	{1, 100, 0, workRandom},
	{1, 1, 100, workRandom},
	{1, 2, 100, workRandom},
	{1, 5, 100, workRandom},
	{1, 20, 100, workRandom},
	{1, 100, 100, workRandom},
	{1, 1, 1000, workRandom},
	{1, 2, 1000, workRandom},
	{1, 5, 1000, workRandom},
	{1, 20, 1000, workRandom},
	{1, 100, 1000, workRandom},

	{5, 1, 0, workRandom},
	{5, 2, 0, workRandom},
	{5, 5, 0, workRandom},
	{5, 20, 0, workRandom},
	{5, 100, 0, workRandom},
	{5, 1, 100, workRandom},
	{5, 2, 100, workRandom},
	{5, 5, 100, workRandom},
	{5, 20, 100, workRandom},
	{5, 100, 100, workRandom},
	{5, 1, 1000, workRandom},
	{5, 2, 1000, workRandom},
	{5, 5, 1000, workRandom},
	{5, 20, 1000, workRandom},
	{5, 100, 1000, workRandom},

	{20, 1, 0, workRandom},
	{20, 2, 0, workRandom},
	{20, 5, 0, workRandom},
	{20, 20, 0, workRandom},
	{20, 100, 0, workRandom},
	{20, 1, 100, workRandom},
	{20, 2, 100, workRandom},
	{20, 5, 100, workRandom},
	{20, 20, 100, workRandom},
	{20, 100, 100, workRandom},
	{20, 1, 1000, workRandom},
	{20, 2, 1000, workRandom},
	{20, 5, 1000, workRandom},
	{20, 20, 1000, workRandom},
	{20, 100, 1000, workRandom},

	{100, 1, 0, workRandom},
	{100, 2, 0, workRandom},
	{100, 5, 0, workRandom},
	{100, 20, 0, workRandom},
	{100, 100, 0, workRandom},
	{100, 1, 100, workRandom},
	{100, 2, 100, workRandom},
	{100, 5, 100, workRandom},
	{100, 20, 100, workRandom},
	{100, 100, 100, workRandom},
	{100, 1, 1000, workRandom},
	{100, 2, 1000, workRandom},
	{100, 5, 1000, workRandom},
	{100, 20, 1000, workRandom},
	{100, 100, 1000, workRandom},
}

func TestParallel(t *testing.T) {
	for _, tt := range tests {
		atomic.StoreInt32(&atomicInt, 0)
		task, err := NewTask(tt.worker, tt.fn)
		if err != nil {
			t.Fatalf("test(worker = %d, n = %d, m = %d, fn = %s) failed: %s", tt.worker, tt.n, tt.microseconds, funcName(tt.fn), err)
		}
		for i := 0; i < tt.n; i++ {
			task.Feed(tt.microseconds)
		}
		data, err := task.Run()
		if err != nil {
			t.Fatalf("test(worker = %d, n = %d, m = %d, fn = %s) failed: %s", tt.worker, tt.n, tt.microseconds, funcName(tt.fn), err)
		}
		if len(data) != tt.n {
			t.Fatalf("test(worker = %d, n = %d, m = %d, fn = %s) failed: invalid data lenght got = %d", tt.worker, tt.n, tt.microseconds, funcName(tt.fn), len(data))
		}
		for i := 0; i < tt.n; i++ {
			if data[i] == nil {
				t.Fatalf("test(worker = %d, n = %d, m = %d, fn = %s) failed: data[%d] == nil", tt.worker, tt.n, tt.microseconds, funcName(tt.fn), i)
			}
		}

	int_loop:
		for n := 1; n < int(atomicInt); n++ {
			for i := 0; i < len(data); i++ {
				if data[i].(*s).val == n {
					continue int_loop
				}
			}
			t.Fatalf("test(worker = %d, n = %d, m = %d, fn = %s) failed: value %d not found in result array", tt.worker, tt.n, tt.microseconds, funcName(tt.fn), n)
		}
	}
}

func funcName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
