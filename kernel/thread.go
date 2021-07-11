package kernel

import (
	"unsafe"

	"github.com/icexin/eggos/mm"
	"github.com/icexin/eggos/sys"
)

const (
	_NTHREDS = 20

	_KCODE_IDX  = 1
	_KDATA_IDX  = 2
	_UCODE_IDX  = 3
	_UDATA_IDX  = 4
	_TSS_IDX    = 5
	_GO_TLS_IDX = 6
	_KTLS_IDX   = 7

	_RPL_USER = 3

	_FLAGS_IF        = 0x200
	_FLAGS_IOPL_USER = 0x3000

	_THREAD_STACK_SIZE = 32 << 10
)

const (
	UNUSED = iota
	INITING
	SLEEPING
	RUNNABLE
	RUNNING
	EXIT
)

const (
	_TSS_ESP0 = 1
	_TSS_SS0  = 2
)

var (
	threads     [_NTHREDS]Thread
	ktls        [16]unsafe.Pointer
	scheduler   *context
	taskstate   [27]uint32
	idle_thread threadptr
)

type context struct {
	di uintptr
	si uintptr
	bx uintptr
	bp uintptr
	ip uintptr
}

type TrapFrame struct {
	GS, FS, ES, DS                  uint16
	DI, SI, BP, _SP, BX, DX, CX, AX uintptr
	Trapno                          uintptr

	// pushed by hardware
	Err           uintptr
	IP, CS, FLAGS uintptr
	SP, SS        uintptr
}

type Thread struct {
	// position of tf and fpstate must be synced with trap.s
	stack  uintptr
	tf     *TrapFrame
	kstack uintptr

	// the state of fpu
	fpstate uintptr

	sigstack stackt
	sigset   sigset

	context *context
	id      int
	state   int
	counter int64

	// sysmon 会调用usleep，进而调用sleepon，如果sleepKey是个指针会触发gcWriteBarrier
	// 而sysmon没有P，会导致空指针
	sleepKey uintptr
	tls      userDesc
}

//go:nosplit
func allocThread() *Thread {
	var t *Thread
	for i := 0; i < _NTHREDS; i++ {
		tt := &threads[i]
		if tt.state == UNUSED {
			t = tt
			t.id = i
			break
		}
	}
	if t == nil {
		panic("no thread slot available")
	}
	// t.sigstack.ss_flags = _SS_DISABLE
	t.sigstack.ss_sp = mm.Alloc()
	t.sigstack.ss_size = mm.PGSIZE
	t.state = INITING
	t.kstack = mm.Mmap(0, _THREAD_STACK_SIZE) + _THREAD_STACK_SIZE
	t.fpstate = mm.Alloc()
	return t
}

type threadptr uintptr

func (t threadptr) ptr() *Thread {
	return (*Thread)(unsafe.Pointer(t))
}

//go:nosplit
func set_fs(idx int)

//go:nosplit
func Mythread() *Thread

//go:nosplit
func set_mythread(t *Thread)

//go:nosplit
func set_gs(idx int)

//go:nosplit
func switchThreadContext(t *Thread) {
	// set go tls base address
	settls(_GO_TLS_IDX, uint32(t.tls.baseAddr), uint32(t.tls.limit))
	// flush cache for invisible gs register
	set_gs(_GO_TLS_IDX)
	// use current thread esp0 in tss
	taskstate[_TSS_SS0] = _KDATA_IDX << 3
	taskstate[_TSS_ESP0] = uint32(t.kstack)
}

//go:nosplit
func ktls_init() {
	addr := uintptr(unsafe.Pointer(&ktls[0]))
	settls(_KTLS_IDX, uint32(addr), uint32(unsafe.Sizeof(ktls)))
	set_fs(_KTLS_IDX)
}

//go:nosplit
func go_entry()

//go:nosplit
func thread0_init() {
	t := allocThread()
	t.stack = mm.Mmap(0, _THREAD_STACK_SIZE)
	t.stack += _THREAD_STACK_SIZE

	sp := t.kstack

	// for trap frame
	sp -= unsafe.Sizeof(TrapFrame{})
	tf := (*TrapFrame)(unsafe.Pointer(sp))

	// Because trapret restore fpstate
	// we need a valid fpstate here
	sys.Fxsave(t.fpstate)
	tf.DS = _UDATA_IDX<<3 | _RPL_USER
	tf.ES = _UDATA_IDX<<3 | _RPL_USER
	tf.FS = _KTLS_IDX<<3 | _RPL_USER
	tf.GS = _GO_TLS_IDX<<3 | _RPL_USER
	tf.SS = _UDATA_IDX<<3 | _RPL_USER
	tf.SP = t.stack
	// enable interrupt and io port
	tf.FLAGS = _FLAGS_IF | _FLAGS_IOPL_USER
	tf.CS = _UCODE_IDX<<3 | _RPL_USER
	tf.IP = sys.FuncPC(thread0)
	t.tf = tf

	// for context
	sp -= unsafe.Sizeof(*t.context)
	ctx := (*context)(unsafe.Pointer(sp))
	ctx.ip = sys.FuncPC(trapret)
	t.context = ctx

	t.state = RUNNABLE
}

//go:nosplit
func sys_clone(pc, stack uintptr) uintptr

//go:nosplit
func sys_yield()

//go:nosplit
func sys_hlt()

// thread0 is the first thread
//go:nosplit
func thread0() {
	// thread0 clone idle thread
	stack := mm.Mmap(0, _THREAD_STACK_SIZE) + _THREAD_STACK_SIZE
	tid := sys_clone(sys.FuncPC(idle), stack)
	idle_thread = (threadptr)(unsafe.Pointer(&threads[tid]))

	// make idle thread running at ring0, so that it can call HLT instruction.
	tf := idle_thread.ptr().tf
	tf.CS = _KCODE_IDX << 3

	// jump to go rt0
	go_entry()
	panic("main return")
}

//go:nosplit
func idle() {
	for {
		sys.Hlt()
		sys_yield()
	}
}

//go:nosplit
func clone(pc, usp uintptr) int {
	my := Mythread()
	chld := allocThread()

	sp := chld.kstack
	// for trap frame
	sp -= unsafe.Sizeof(TrapFrame{})
	tf := (*TrapFrame)(unsafe.Pointer(sp))
	*tf = *my.tf

	// copy fpstate
	fpsrc := (*[512]byte)(unsafe.Pointer(my.fpstate))
	fpdst := (*[512]byte)(unsafe.Pointer(chld.fpstate))
	*fpdst = *fpsrc

	tf.SP = usp
	tf.IP = pc
	tf.AX = 0

	// for context
	sp -= unsafe.Sizeof(context{})
	ctx := (*context)(unsafe.Pointer(sp))
	ctx.ip = sys.FuncPC(trapret)

	chld.context = ctx
	// *(*uintptr)(unsafe.Pointer(&chld.context)) = sp
	chld.tf = tf
	chld.stack = usp
	chld.state = RUNNABLE
	return chld.id
}

//go:nosplit
func exit() {
	t := Mythread()
	t.state = EXIT
	Yield()
	// TODO: handle thread exit in scheduler
}

//go:nosplit
func thread_init() {
	ktls_init()
	thread0_init()
}

//go:nosplit
func swtch(old **context, _new *context)

//go:nosplit
func schedule() {
	var t *Thread
	var idx int
	for {
		t = pickup(&idx)
		switchto(t)
	}
}

// pickup selects the next runnable thread
//go:nosplit
func pickup(pidx *int) *Thread {
	curr := *pidx
	if traptask != 0 && traptask.ptr().state == RUNNABLE {
		return traptask.ptr()
	}
	if syscalltask != 0 && syscalltask.ptr().state == RUNNABLE {
		return syscalltask.ptr()
	}

	var t *Thread
	for i := 0; i < _NTHREDS; i++ {
		idx := (curr + i + 1) % _NTHREDS
		*pidx = idx
		tt := &threads[idx]
		if tt.state == RUNNABLE && tt != idle_thread.ptr() {
			t = tt
			break
		}
	}
	if t == nil {
		t = idle_thread.ptr()
	}
	return t
}

// switchto switch thread context from scheduler to t
//go:nosplit
func switchto(t *Thread) {
	begin := nanosecond()
	// assert interrupt is enableds
	if t.tf != nil && t.tf.FLAGS&0x200 == 0 {
		panic("bad eflags")
	}
	set_mythread(t)
	t.state = RUNNING

	swtch(&scheduler, t.context)

	used := nanosecond() - begin
	t.counter += used
}

func ThreadStat(stat *[_NTHREDS]int64) {
	for i := 0; i < _NTHREDS; i++ {
		stat[i] = threads[i].counter
	}
}

//go:nosplit
func Sched() {
	my := Mythread()
	swtch(&my.context, scheduler)
}

//go:nosplit
func Yield() {
	my := Mythread()
	my.state = RUNNABLE
	swtch(&my.context, scheduler)
}
