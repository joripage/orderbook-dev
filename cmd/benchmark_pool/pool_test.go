package benchmarkpool

import (
	"fmt"
	"sync"
	"testing"
)

type OrderTest struct{ ID int }

func TestPool_ChangeData(t *testing.T) {
	orderPool := &sync.Pool{
		New: func() any { return &OrderTest{} },
	}

	o1 := orderPool.Get().(*OrderTest)
	o2 := orderPool.Get().(*OrderTest)

	fmt.Println(o1 == o2) // luôn false với use-case đúng
	o1.ID = 5
	fmt.Println(o2.ID) // 0, không bị ảnh hưởng

	// lấy object
	o3 := orderPool.Get().(*OrderTest)
	o3.ID = 5
	fmt.Println("Got o3:", o3.ID)

	// trả lại pool
	orderPool.Put(o3)

	// ép chạy GC
	// runtime.GC()
	// time.Sleep(time.Second)

	// lấy lại object
	o4 := orderPool.Get().(*OrderTest)
	fmt.Println("Got o4:", o4.ID) // Có thể = 0 nếu o1 bị GC dọn
}
