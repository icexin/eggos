package kernel64

import (
	"unsafe"

	"gvisor.dev/gvisor/pkg/abi/linux"
)

const (
	_FUTEX_WAIT         = 0
	_FUTEX_WAKE         = 1
	_FUTEX_PRIVATE_FLAG = 128
	_FUTEX_WAIT_PRIVATE = _FUTEX_WAIT | _FUTEX_PRIVATE_FLAG
	_FUTEX_WAKE_PRIVATE = _FUTEX_WAKE | _FUTEX_PRIVATE_FLAG
)

//go:nosplit
func futex(addr *uintptr, op, val uintptr, ts *linux.Timespec) {
	switch op {
	case _FUTEX_WAIT, _FUTEX_WAIT_PRIVATE:
		if ts != nil {
			sleeptimeout(addr, val, ts)
			return
		}
		for *addr == val {
			sleepon(addr)
		}
		return
	case _FUTEX_WAKE, _FUTEX_WAKE_PRIVATE:
		wakeup(addr, int(val))
	default:
		panic("futex: invalid op")
	}
}

//go:nosplit
func sleeptimeout(addr *uintptr, val uintptr, ts *linux.Timespec) {
	if ts == nil {
		panic("sleeptimeout: nil ts")
	}
	deadline := nanosecond() + int64(ts.Nsec) + int64(ts.Sec)*second
	// check on every timer intr
	for nanosecond() < deadline && *addr == val {
		sleepon(&sleeplock)
	}
}

//go:nosplit
func sleepon(lock *uintptr) {
	t := Mythread()
	t.sleepKey = uintptr(unsafe.Pointer(lock))
	t.state = SLEEPING
	Sched()
	t.sleepKey = 0
}

// wakeup thread sleep on lock, n == -1 means all threads
//go:nosplit
func wakeup(lock *uintptr, n int) {
	limit := uint(n)
	cnt := uint(0)
	for i := 0; i < _NTHREDS; i++ {
		t := &threads[i]
		if t.sleepKey == uintptr(unsafe.Pointer(lock)) && cnt < limit {
			cnt++
			t.state = RUNNABLE
		}
	}
}

type note uintptr

//go:nosplit
func (n *note) sleep(ts *linux.Timespec) {
	futex((*uintptr)(unsafe.Pointer(n)), _FUTEX_WAIT, 0, ts)
}

//go:nosplit
func (n *note) wakeup() {
	*n = 1
	futex((*uintptr)(unsafe.Pointer(n)), _FUTEX_WAKE, 1, nil)
}

//go:nosplit
func (n *note) clear() {
	*n = 0
}
