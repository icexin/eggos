package kernel

import (
	"unsafe"

	"github.com/icexin/eggos/sys"
	"github.com/icexin/eggos/uart"
)

const (
	_NSIG = 128

	_SS_DISABLE = 2
)

var (
	sigs   [_NSIG]signal
	sinfo  siginfo
	sigctx ucontext
)

type sighandler func(sig uint32, info *siginfo, ctx *ucontext)

type signal struct {
	sigactiont sigactiont
}

type siginfo struct {
	si_signo int32
	si_errno int32
	si_code  int32
	// below here is a union; si_addr is the only field we use
	si_addr uint32
}

type ucontext struct {
	uc_flags    uint32
	uc_link     *ucontext
	uc_stack    stackt
	uc_mcontext sigcontext
	uc_sigmask  uint32
}

type sigcontext struct {
	gs            uint16
	__gsh         uint16
	fs            uint16
	__fsh         uint16
	es            uint16
	__esh         uint16
	ds            uint16
	__dsh         uint16
	edi           uint32
	esi           uint32
	ebp           uint32
	esp           uint32
	ebx           uint32
	edx           uint32
	ecx           uint32
	eax           uint32
	trapno        uint32
	err           uint32
	eip           uint32
	cs            uint16
	__csh         uint16
	eflags        uint32
	esp_at_signal uint32
	ss            uint16
	__ssh         uint16
	fpstate       *fpstate
	oldmask       uint32
	cr2           uint32
}

type fpstate struct {
	cw        uint32
	sw        uint32
	tag       uint32
	ipoff     uint32
	cssel     uint32
	dataoff   uint32
	datasel   uint32
	_st       [8]fpreg
	status    uint16
	magic     uint16
	_fxsr_env [6]uint32
	mxcsr     uint32
	reserved  uint32
	_fxsr_st  [8]fpxreg
	_xmm      [8]xmmreg
	padding1  [44]uint32
	anon0     [48]byte
}

type fpreg struct {
	significand [4]uint16
	exponent    uint16
}

type fpxreg struct {
	significand [4]uint16
	exponent    uint16
	padding     [3]uint16
}

type xmmreg struct {
	element [4]uint32
}

type sigactiont struct {
	sa_handler  uintptr
	sa_flags    uint32
	sa_restorer uintptr
	sa_mask     uint64
}

type stackt struct {
	ss_sp    uintptr
	ss_flags int32
	ss_size  uintptr
}

type sigset [2]uint32

func (s *sigset) block(s1 *sigset) {
	s[0] |= s1[0]
	s[1] |= s1[1]
}

func (s *sigset) unblock(s1 *sigset) {
	s[0] &^= s1[0]
	s[1] &^= s1[1]
}

//go:nosplit
func sigaltstack(new, old *stackt) {
	my := Mythread()
	if old != nil {
		*old = my.sigstack
	}
	if new == nil {
		return
	}
	uart.WriteString("signal stack ")
	uart.WriteByte('0' + byte(my.id))
	uart.WriteByte('\n')
	my.sigstack = *new
}

const (
	_SIG_BLOCK   = 0
	_SIG_UNBLOCK = 1
	_SIG_SETMASK = 2
)

//go:nosplit
func rtsigprocmask(how int32, new, old *sigset, size int32) {
	my := Mythread()
	if old != nil {
		*old = my.sigset
	}
	if new == nil {
		return
	}
	switch how {
	case _SIG_SETMASK:
		my.sigset = *new
	case _SIG_UNBLOCK:
		my.sigset.unblock(new)
	case _SIG_BLOCK:
		my.sigset.block(new)
	}
}

//go:nosplit
func rt_sigaction(sig uintptr, new, old *sigactiont, size uintptr) int32 {
	if sig >= _NSIG {
		return -1
	}
	if old != nil {
		*old = sigs[int(sig)].sigactiont
	}
	if new == nil {
		return 0
	}
	sigs[int(sig)].sigactiont = *new
	return 0
}

//go:nosplit
func callSigHandler(pc, sp uintptr, no uintptr, info *siginfo, ctx *ucontext)

//go:nosplit
func Signal(signo, sigcode, sigaddr uintptr) {
	my := Mythread()

	sig := sigs[signo]
	if sig.sigactiont.sa_handler == 0 {
		return
	}

	sinfo = siginfo{
		si_signo: int32(signo),
		si_addr:  uint32(sigaddr),
		si_code:  int32(sigcode),
	}
	var dummy int
	sigctx.uc_stack = my.sigstack
	sigctx.uc_mcontext.eip = uint32(my.tf.IP)
	sigctx.uc_mcontext.esp = uint32(uintptr(unsafe.Pointer(&dummy)))
	if my.sigstack.ss_sp == 0 {
		uart.WriteString("signal stack nil ")
		uart.WriteByte('0' + byte(my.id))
		uart.WriteByte('\n')
		panic("signal stack nil")
	}

	sp := my.sigstack.ss_sp
	sp += my.sigstack.ss_size

	sp -= sys.PtrSize
	*(*uintptr)(unsafe.Pointer(sp)) = uintptr(unsafe.Pointer(&sigctx))
	sp -= sys.PtrSize
	*(*uintptr)(unsafe.Pointer(sp)) = uintptr(unsafe.Pointer(&sinfo))
	sp -= sys.PtrSize
	*(*uintptr)(unsafe.Pointer(sp)) = signo

	callSigHandler(sig.sigactiont.sa_handler, sp, signo, &sinfo, &sigctx)
	if my.tf.IP != uintptr(sigctx.uc_mcontext.eip) {
		// panic happend
		ChangeReturnPC(my.tf, uintptr(sigctx.uc_mcontext.eip))
	}
}
