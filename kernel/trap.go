package kernel

import (
	"unsafe"

	"github.com/jspc/eggos/drivers/pic"
	"github.com/jspc/eggos/kernel/isyscall"
	"github.com/jspc/eggos/kernel/sys"
	"github.com/jspc/eggos/kernel/trap"
	"github.com/jspc/eggos/log"
)

var (
	trapnum = [...]string{
		"#DE(0) Divide by zero",
		"#DB(1) Debug exception",
		"#NMI(2) Non maskable interrupt exception",
		"#BP(3) Breakpoint exception",
		"#OF(4) Overflow exception",
		"#BR(5) Bound range exception",
		"#UD(6) Invalid opcode exception",
		"#NM(7) Device not avaiable exception",
		"#DF(8) Double fault exception",
		"(9)Coprocessor segment overrun exception",
		"#TS(10) Invalid TSS exception",
		"#NP(11) Segment not present exception",
		"#SS(12) Stack exception",
		"#GP(13) General protection exception",
		"#PF(14) Page fault exception",
		"#MF(15) x87 floating point exception",
		"#AC(16) Alignment check exception",
		"#MC(17) Machine check exception",
		"#XF(18) SIMD floating point exception",
		"#HV(19) Hypervisor injection exception",
		"#VC(20) VMM communication exception",
		"#SX(21) Security Exception",
	}
)

//go:notinheap
type trapFrame struct {
	AX, BX, CX, DX    uintptr
	BP, SI, DI, R8    uintptr
	R9, R10, R11, R12 uintptr
	R13, R14, R15     uintptr

	Trapno, Err uintptr

	// pushed by hardware
	IP, CS, FLAGS, SP, SS uintptr
}

func (t *trapFrame) SyscallRequest() isyscall.Request {
	return isyscall.NewRequest(uintptr(unsafe.Pointer(t)))
}

//go:nosplit
func trapret()

//go:nosplit
func trapPanic() {
	tf := Mythread().tf
	log.PrintStr("trap panic: ")
	if tf.Trapno < uintptr(len(trapnum)) {
		log.PrintStr(trapnum[tf.Trapno])
	}
	log.PrintStr("\n")
	throwtf(tf, "stack trace:")
}

//go:nosplit
func pageFaultPanic() {
	panic("nil pointer or invalid memory access")
}

//go:nosplit
func ignoreHandler() {
}

//go:nosplit
func pageFaultHandler() {
	t := Mythread()
	checkKernelPanic(t)
	changeReturnPC(t.tf, sys.FuncPC(pageFaultPanic))
}

//go:nosplit
func faultHandler() {
	trapPanic()
}

//go:nosplit
func printReg(name string, reg uintptr) {
	log.PrintStr(name)
	log.PrintStr("=")
	log.PrintHex(reg)
	log.PrintStr("\n")
}

//go:nosplit
func checkKernelPanic(t *Thread) {
	tf := t.tf
	if tf.CS != _KCODE_IDX<<3 {
		return
	}
	printReg("tid", uintptr(t.id))
	printReg("no", tf.Trapno)
	printReg("err", tf.Err)
	printReg("cr2", sys.Cr2())
	printReg("ip", tf.IP)
	printReg("sp", tf.SP)
	printReg("ax", tf.AX)
	printReg("bx", tf.BX)
	printReg("cx", tf.CX)
	printReg("dx", tf.CX)
	printReg("cs", tf.CS)
	throw("trap fault in kernel\n")
}

//go:nosplit
func preparePanic(tf *trapFrame) {
	changeReturnPC(tf, sys.FuncPC(trapPanic))
}

// ChangeReturnPC change the return pc of a trap
// must be called in trap handler
//go:nosplit
func changeReturnPC(tf *trapFrame, pc uintptr) {
	// tf.Err, tf.IP, tf.CS, tf.FLAGS = pc, tf.CS, tf.FLAGS, tf.IP
	sp := tf.SP
	sp -= sys.PtrSize
	*(*uintptr)(unsafe.Pointer(sp)) = tf.IP
	tf.SP = sp
	tf.IP = pc
}

//go:nosplit
func dotrap(tf *trapFrame) {
	if sys.Flags()&_FLAGS_IF != 0 {
		throw("IF should clear")
	}
	my := Mythread()
	// ugly as it is, avoid writeBarrier
	// my.tf = tf
	*(*uintptr)(unsafe.Pointer(&my.tf)) = uintptr(unsafe.Pointer(tf))

	handler := trap.Handler(int(tf.Trapno))
	if handler == nil {
		faultHandler()
		return
	}
	// timer and syscall interrupts are processed synchronously
	if tf.Trapno > 32 && tf.Trapno != 0x80 {
		// pci using level trigger irq, cause dead lock on trap handler
		// FIXME: hard code network irq line
		if tf.Trapno == 43 {
			pic.DisableIRQ(43 - pic.IRQ_BASE)
		}
		wakeIRQ(tf.Trapno)
		return
	}
	handler()
}

//go:nosplit
func trapInit() {
	trap.Register(14, pageFaultHandler)
	trap.Register(39, ignoreHandler)
	trap.Register(47, ignoreHandler)
}
