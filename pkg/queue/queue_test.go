package queue

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestStructPointer(t *testing.T) {
	type StringStruct struct {
		Contents string
	}
	q := New[*StringStruct]()
	if q.Length() != 0 {
		t.Error("empty queue length not 0")
	}
	if q.Front() != nil {
		t.Error("There should be nil")
	}
	firstTestString := "first"
	secondTestString := "second"
	thirdTestString := "third"
	insertionOne := &StringStruct{Contents: secondTestString}
	insertionTwo := &StringStruct{Contents: thirdTestString}
	insertionThree := &StringStruct{Contents: firstTestString}
	q.Append(insertionOne)
	q.Append(insertionTwo)
	q.Append(insertionThree)
	insertionOne.Contents = firstTestString
	insertionTwo.Contents = secondTestString
	insertionThree.Contents = thirdTestString
	if compare := strings.Compare(q.Pop().Contents, firstTestString); compare != 0 {
		t.Error("content ", compare, " is not expected: ", firstTestString)
	}
	if compare := strings.Compare(q.Pop().Contents, secondTestString); compare != 0 {
		t.Error("content ", compare, " is not expected: ", secondTestString)
	}
	if compare := strings.Compare(q.Pop().Contents, thirdTestString); compare != 0 {
		t.Error("content ", compare, " is not expected: ", thirdTestString)
	}
}

func TestQueueSimple(t *testing.T) {
	q := New[int]()

	for i := 0; i < minQueueLen; i++ {
		q.Append(i)
	}
	for i := 0; i < minQueueLen; i++ {
		x := q.Pop()
		if x != i {
			t.Error("remove", i, "had value", x)
		}
	}
}

func TestQueueSimplePrepend(t *testing.T) {
	q := New[int]()

	for i := 0; i < minQueueLen; i++ {
		q.Prepend(i)
	}
	for i := minQueueLen - 1; i >= 0; i-- {
		x := q.Pop()
		if x != i {
			t.Error("remove", i, "had value", x)
		}
	}
}

func TestQueueManual(t *testing.T) {
	q := New[int]()

	q.Append(1)
	q.Append(2)
	q.Prepend(4)

	if q.Pop() != 4 {
		t.Error("Invalid element")
	}

	q.Prepend(3)

	if q.Pop() != 3 {
		t.Error("Invalid element")
	}

	if q.Pop() != 1 {
		t.Error("Invalid element")
	}

	if q.Pop() != 2 {
		t.Error("Invalid element")
	}
}

func TestQueueWrapping(t *testing.T) {
	q := New[int]()

	for i := 0; i < minQueueLen; i++ {
		q.Append(i)
	}
	for i := 0; i < 3; i++ {
		q.Pop()
		q.Append(minQueueLen + i)
	}

	for i := 0; i < minQueueLen; i++ {
		q.Pop()
	}
}

func TestQueueWrappingPrepend(t *testing.T) {
	q := New[int]()

	for i := 0; i < minQueueLen; i++ {
		q.Prepend(i)
	}
	for i := 0; i < 3; i++ {
		q.Pop()
		q.Prepend(minQueueLen + i)
	}

	for i := 0; i < minQueueLen; i++ {
		q.Pop()
	}
}

func TestQueueThreadSafety(t *testing.T) {
	q := New[int]()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		for i := 0; i < 10000; i++ {
			q.Append(i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			if q.Pop() != i {
				t.Errorf("Invalid returned index: %d", i)
				wg.Done()
				return
			}
		}
		wg.Done()
	}()

	wg.Wait()
}

func TestQueueThreadSafety2(t *testing.T) {
	q := New[int]()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		for i := 0; i < 10000; i++ {
			q.Append(i)
			q.Prepend(i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 20000; i++ {
			q.Pop()
		}
		wg.Done()
	}()

	wg.Wait()
}

func TestQueueThreadSafety3(t *testing.T) {
	q := New[int]()

	var wg sync.WaitGroup

	wg.Add(10000)

	for i := 0; i < 5000; i++ {
		go func() {
			q.Append(i)
			wg.Done()
		}()
	}

	for i := 0; i < 5000; i++ {
		go func() {
			q.Pop()
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestQueueLength(t *testing.T) {
	q := New[int]()

	if q.Length() != 0 {
		t.Error("empty queue length not 0")
	}

	for i := 0; i < 1000; i++ {
		q.Append(i)
		if q.Length() != i+1 {
			t.Error("adding: queue with", i, "elements has length", q.Length())
		}
	}
	for i := 0; i < 1000; i++ {
		q.Pop()
		if q.Length() != 1000-i-1 {
			t.Error("removing: queue with", 1000-i-i, "elements has length", q.Length())
		}
	}
}

func TestQueueBlocking(t *testing.T) {
	q := New[int]()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		q.Append(1)
		time.Sleep(1 * time.Second)
		q.Append(2)
		wg.Done()
	}()

	item := q.Pop()
	if item != 1 {
		t.Error("Returned invalid 1 element")
	}
	item2 := q.Pop()
	if item2 != 2 {
		t.Error("Returned invalid 2 element")
	}

	wg.Wait()
}

func assertPanics(t *testing.T, name string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: didn't panic as expected", name)
		}
	}()

	f()
}

func TestFront(t *testing.T) {
	q := New[int]()

	q.Append(1)
	q.Append(2)

	if q.Front() != 1 {
		t.Error("There should be 1 on front")
	}

	q.Pop()

	if q.Front() != 2 {
		t.Error("There should be 2 on front")
	}
}

func TestFrontEmpty(t *testing.T) {
	q := New[int]()

	if q.Front() != 0 {
		t.Error("There should be nil")
	}

	q.Append(1)
	q.Append(2)
	q.Prepend(3)
	q.Pop()
	q.Pop()
	q.Pop()

	if q.Front() != 0 {
		t.Error("There should be nil")
	}
}

func TestBackEmpty(t *testing.T) {
	q := New[int]()

	if q.Back() != 0 {
		t.Error("There should be nil")
	}

	q.Append(1)
	q.Append(2)
	q.Prepend(3)
	q.Pop()
	q.Pop()
	q.Pop()

	if q.Back() != 0 {
		t.Error("There should be nil")
	}
}

func TestBack(t *testing.T) {
	q := New[int]()

	q.Append(1)
	q.Append(2)

	if q.Back() != 2 {
		t.Errorf("There should be 2 on back, there is %v", q.Back())
	}

	q.Pop()

	if q.Back() != 2 {
		t.Errorf("There should be 2 on back, there is %v", q.Back())
	}
}

func TestRemove(t *testing.T) {
	q := New[int]()

	q.Append(1)
	q.Append(2)
	q.Append(3)
	q.Remove(2)

	if q.Length() != 2 {
		t.Errorf("Queue length should be 2, it is %d", q.Length())
	}

	p := q.Pop()
	if p != 1 {
		t.Errorf("There should be 1 on pop, there is %v", p)
	}

	p = q.Pop()
	if p != 3 {
		t.Errorf("There should be 3 on pop, there is %v", p)
	}
}

func TestTestQueueClean(t *testing.T) {
	q := New[int]()

	q.Append(4)
	q.Append(6)
	q.Clean()

	q.Append(1)
	q.Append(2)
	q.Append(3)
	q.Remove(2)

	if q.Length() != 2 {
		t.Errorf("Queue length should be 2, it is %d", q.Length())
	}

	p := q.Pop()
	if p != 1 {
		t.Errorf("There should be 1 on pop, there is %v", p)
	}

	p = q.Pop()
	if p != 3 {
		t.Errorf("There should be 3 on pop, there is %v", p)
	}
}

func TestTestQueueClean2(t *testing.T) {
	q := New[int]()

	for i := 0; i < 50; i++ {
		q.Append(i)
	}

	q.Clean()

	for i := 0; i < 50; i++ {
		q.Append(i)
	}
}

// General warning: Go's benchmark utility (go test -bench .) increases the number of
// iterations until the benchmarks take a reasonable amount of time to run; memory usage
// is *NOT* considered. On my machine, these benchmarks hit around ~1GB before they've had
// enough, but if you have less than that available and start swapping, then all bets are off.

func BenchmarkQueueSerial(b *testing.B) {
	q := New[int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Append(i)
	}
	for i := 0; i < b.N; i++ {
		q.Pop()
	}
}

func BenchmarkQueueTickTock(b *testing.B) {
	q := New[int]()
	for i := 0; i < b.N; i++ {
		q.Append(i)
		q.Pop()
	}
}
