package cherryTimeWheel

import (
	"runtime"
	"testing"
	"time"
)

// bucketNum 复刻 Go map 的 B 计算（负载因子 6.5 = 13/2），返回桶数。
func bucketNum(hint int) uint64 {
	if hint <= 8 {
		return 1
	}
	// loadFactorNum=13, loadFactorDen=2：hint <= (bucketShift(B)*13)/2
	b := uint64(1) // 2^B
	for uint64(hint) > (b*13)/2 {
		b <<= 1
	}
	return b
}

// TestSetTimerHint_MapPrealloc 实测不同 hint 下 nodeMap 的预分配内存。
// 对应 master 节点 SetTimerHint(1<<21) 的场景。
func TestSetTimerHint_MapPrealloc(t *testing.T) {
	for _, hint := range []int{0, 1 << 10, 1 << 16, 1 << 20, 1 << 21} {
		runtime.GC()
		var before runtime.MemStats
		runtime.ReadMemStats(&before)

		tw := NewTimeWheelWithHint(10*time.Millisecond, hint)
		_ = tw

		var after runtime.MemStats
		runtime.ReadMemStats(&after) // 不 GC，立即读取

		delta := int64(after.HeapAlloc) - int64(before.HeapAlloc)
		t.Logf("hint=%-8d nodeMap HeapAlloc delta=%d KB (~%d MB), Buckets=%d",
			hint, delta>>10, delta>>20, bucketNum(hint))
	}
}

// TestRecurringTimer_MemStable 模拟 timer_actor 负载：
// 1000 个 recurring 定时器（延迟 50ms~1049ms），长时间运行，采样 HeapAlloc 与 ActiveCount。
// 若有节点级泄漏，ActiveCount 会持续增长、HeapAlloc 线性爬升。
func TestRecurringTimer_MemStable(t *testing.T) {
	if testing.Short() {
		t.Skip("skip long-run mem probe in -short mode")
	}

	tw := NewTimeWheelWithHint(10*time.Millisecond, 1<<21)
	tw.Start()
	defer tw.Stop()

	for i := range 1000 {
		delay := time.Duration(50+i%1000) * time.Millisecond
		tw.AddTimer(delay, func() {}, false)
	}

	samples := 6
	interval := 30 * time.Second
	var prev runtime.MemStats

	for i := range samples {
		time.Sleep(interval)
		runtime.GC()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		t.Logf("elapsed=%dm active=%d HeapAlloc=%dMB Sys=%dMB",
			(i+1)*int(interval/time.Minute), tw.ActiveCount(),
			m.HeapAlloc>>20, m.Sys>>20)

		if i > 0 {
			d := int64(m.HeapAlloc) - int64(prev.HeapAlloc)
			if d > 0 {
				t.Logf("  -> HeapAlloc grew +%d KB in %v", d>>10, interval)
			}
		}
		if got := tw.ActiveCount(); got != 1000 {
			t.Fatalf("ActiveCount = %d, want 1000 (leak!)", got)
		}
		prev = m
	}
}

// TestOnceTimer_MemReleased 验证大量一次性定时器全部触发完成后内存能释放：
// 触发完成后 ActiveCount 归零（无残留 nodeMap 条目）；释放句柄后 GC，
// node 空壳应被回收，HeapAlloc 不增长。
// 回归：reset 曾保留链表指针，被句柄持有的已完成 node 会通过残留 next/prev
// 拖住同批其他 node 不 GC。
func TestOnceTimer_MemReleased(t *testing.T) {
	if testing.Short() {
		t.Skip("skip long-run mem probe in -short mode")
	}

	tw := NewTimeWheelWithHint(10*time.Millisecond, 1<<21)
	tw.Start()
	defer tw.Stop()

	const n = 20000
	handles := make([]*Timer, 0, n)
	for i := range n {
		delay := time.Duration(50+i%1000) * time.Millisecond
		handles = append(handles, tw.AddTimer(delay, func() {}, true))
	}

	// 等待全部触发完成（触发即 reset + ActiveCount--，无需等 removeCmd）。
	deadline := time.Now().Add(60 * time.Second)
	for tw.ActiveCount() != 0 {
		if time.Now().After(deadline) {
			t.Fatalf("ActiveCount = %d after expiry, want 0", tw.ActiveCount())
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 句柄仍持有 node 空壳时的基线。
	runtime.GC()
	var held runtime.MemStats
	runtime.ReadMemStats(&held)

	// 释放句柄，node 空壳失去引用，再次 GC。
	for i := range handles {
		handles[i] = nil
	}
	runtime.GC()
	var released runtime.MemStats
	runtime.ReadMemStats(&released)

	t.Logf("HeapAlloc: held=%dKB released=%dKB delta=%dKB",
		held.HeapAlloc>>10, released.HeapAlloc>>10,
		int64(released.HeapAlloc-held.HeapAlloc)>>10)

	if released.HeapAlloc > held.HeapAlloc+1<<20 {
		t.Fatalf("HeapAlloc grew after releasing handles: held=%dMB released=%dMB",
			held.HeapAlloc>>20, released.HeapAlloc>>20)
	}
}
