#include "textflag.h"

#define SYS_clone	     56
#define SYS_sched_yield	 24

#define tls_my 0

// func swtch(old **context, _new *context)
TEXT ·swtch(SB), NOSPLIT, $0-16
	MOVQ old+0(FP), AX
	MOVQ _new+8(FP), DX

	// Save old callee-saved registers
	PUSHQ BP
	PUSHQ BX
	PUSHQ R11
	PUSHQ R12
	PUSHQ R13
	PUSHQ R14
	PUSHQ R15

	// Switch stacks
	MOVQ SP, (AX)
	MOVQ DX, SP

	POPQ R15
	POPQ R14
	POPQ R13
	POPQ R12
	POPQ R11
	POPQ BX
	POPQ BP
	RET

TEXT ·Mythread(SB), NOSPLIT, $0-8
	MOVQ tls_my(GS), AX
	MOVQ AX, ret+0(FP)
	RET

TEXT ·ksysClone(SB), NOSPLIT, $0-32
	MOVQ $SYS_clone, AX
	MOVQ pc+0(FP), R12
	MOVQ stack+8(FP), SI
	MOVQ flags+16(FP), DI

	// clear tls
	XORQ R8, R8

	INT $0x80

	// In parent, return.
	CMPQ AX, $0
	JEQ  3(PC)
	MOVQ AX, ret+24(FP)
	RET

	NOP SP  // tell vet SP changed - stop checking offsets
	JMP R12

TEXT ·ksysYield(SB), NOSPLIT, $0
	MOVQ $SYS_sched_yield, AX
	INT  $0x80
	RET
