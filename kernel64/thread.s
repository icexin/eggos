#include "textflag.h"

#define SYS_clone	     56
#define SYS_sched_yield	 158

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

TEXT ·ksysClone(SB), NOSPLIT, $0-24
	MOVQ $SYS_clone, AX
	MOVL pc+0(FP), DX
	MOVL stack+4(FP), CX
	INT  $0x80

	// In parent, return.
	CMPL AX, $0
	JEQ  3(PC)
	MOVL AX, ret+8(FP)
	RET

	NOP SP // tell vet SP changed - stop checking offsets
	JMP DX

TEXT ·ksysYield(SB), NOSPLIT, $0
	MOVL $SYS_sched_yield, AX
	INT  $0x80
	RET
