package kernel

import (
	"unsafe"

	"github.com/icexin/eggos/kernel/mm"
	"github.com/icexin/eggos/sys"
)

const (
	_NTHREDS = 20

	_FLAGS_IF        = 0x200
	_FLAGS_IOPL_USER = 0x3000

	_RPL_USER = 3

	_THREAD_STACK_SIZE         = 32 << 10
	_THREAD_STACK_GUARD_OFFSET = 1 << 10
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
	threads    [_NTHREDS]Thread
	scheduler  *context
	idleThread threadptr
)

type context struct {
	r15 uintptr
	r14 uintptr
	r13 uintptr
	r12 uintptr
	r11 uintptr
	bx  uintptr
	bp  uintptr
	ip  uintptr
}

// position of threadTLS and fpstate must be synced with trap.s and syscall.s
type Thread struct {
	// store thread tls, the pointer to Thread
	threadTLS [4]uintptr

	// the state of fpu
	fpstate uintptr

	kstack uintptr
	stack  uintptr
	tf     *trapFrame

	context *context
	id      int
	state   int
	counter int64

	// sysmon 会调用usleep，进而调用sleepon，如果sleepKey是个指针会触发gcWriteBarrier
	// 而sysmon没有P，会导致空指针
	sleepKey uintptr

	// store goroutine tls
	fsBase uintptr

	// 用于保存需要转发的系统调用栈帧
	systf trapFrame
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
		throw("no thread slot available")
	}

	t.state = INITING
	t.kstack = allocThreadStack()
	t.fpstate = mm.Alloc()
	t.threadTLS[0] = uintptr(unsafe.Pointer(t))
	return t
}

//go:nosplit
func allocThreadStack() uintptr {
	stack := mm.Mmap(0, _THREAD_STACK_SIZE)
	stack += _THREAD_STACK_SIZE - _THREAD_STACK_GUARD_OFFSET
	return stack
}

type threadptr uintptr

//go:nosplit
func (t threadptr) ptr() *Thread {
	return (*Thread)(unsafe.Pointer(t))
}

//go:nosplit
func setFS(addr uintptr) {
	wrmsr(_MSR_FS_BASE, addr)
}

//go:nosplit
func setGS(addr uintptr) {
	wrmsr(_MSR_GS_BASE, addr)
}

//go:nosplit
func Mythread() *Thread

//go:nosplit
func setMythread(t *Thread) {
	switchThreadContext(t)
}

//go:nosplit
func switchThreadContext(t *Thread) {
	// set go tls base address
	if t.fsBase != 0 {
		setFS(t.fsBase)
	}
	// set current thread base address
	setGS(uintptr(unsafe.Pointer(&t.threadTLS)))

	// use current thread esp0 in tss
	setTssSP0(t.kstack)
}

//go:nosplit
func thread0Init() {
	t := allocThread()
	t.stack = allocThreadStack()

	sp := t.kstack

	// for trap frame
	sp -= unsafe.Sizeof(trapFrame{})
	tf := (*trapFrame)(unsafe.Pointer(sp))

	// Because trapret restore fpstate
	// we need a valid fpstate here
	sys.Fxsave(t.fpstate)
	tf.SS = _UDATA_IDX<<3 | _RPL_USER
	tf.SP = t.stack
	// enable interrupt and io port
	// TODO: enable interrupt
	tf.FLAGS = _FLAGS_IF | _FLAGS_IOPL_USER
	// tf.FLAGS = _FLAGS_IF
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
func ksysClone(pc, stack uintptr) uintptr

//go:nosplit
func ksysYield()

// thread0 is the first thread
//go:nosplit
func thread0() {
	// jump to go rt0
	go_entry()
	panic("main return")
}

// run when after main init
func idleInit() {
	// thread0 clone idle thread
	stack := mm.SysMmap(0, _THREAD_STACK_SIZE) +
		_THREAD_STACK_SIZE - _THREAD_STACK_GUARD_OFFSET

	tid := ksysClone(sys.FuncPC(idle), stack)
	idleThread = (threadptr)(unsafe.Pointer(&threads[tid]))

	// make idle thread running at ring0, so that it can call HLT instruction.
	tf := idleThread.ptr().tf
	tf.CS = _KCODE_IDX << 3
	tf.SS = _KDATA_IDX << 3
}

//go:nosplit
func idle() {
	for {
		if sys.CS() != 8 {
			throw("bad cs")
		}
		sys.Hlt()
		ksysYield()
	}
}

//go:nosplit
func clone(pc, usp, tls uintptr) int {
	my := Mythread()
	chld := allocThread()

	sp := chld.kstack
	// for trap frame
	sp -= unsafe.Sizeof(trapFrame{})
	tf := (*trapFrame)(unsafe.Pointer(sp))
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
	chld.fsBase = tls
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
func threadInit() {
	thread0Init()
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
		if tt.state == RUNNABLE && tt != idleThread.ptr() {
			t = tt
			break
		}
	}
	if t == nil {
		t = idleThread.ptr()
	}
	if t == nil {
		throw("no runnable thread")
	}
	return t
}

// switchto switch thread context from scheduler to t
//go:nosplit
func switchto(t *Thread) {
	begin := nanosecond()
	// assert interrupt is enableds
	// TODO: enable check
	if t.tf != nil && t.tf.FLAGS&0x200 == 0 {
		throw("bad eflags")
	}
	setMythread(t)
	t.state = RUNNING

	if t == idleThread.ptr() && t.tf.CS != 8 {
		throw("bad idle cs")

	}
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
