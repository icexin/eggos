#include "textflag.h"

// lgdt(gdtptr uint64) - Load Global Descriptor Table Register.
TEXT ·lgdt(SB), NOSPLIT, $0-8
	MOVQ gdtptr+0(FP), AX
	LGDT (AX)
	RET

// lidt(idtptr uint64) - Load Interrupt Descriptor Table Register.
TEXT ·lidt(SB), NOSPLIT, $0-8
	MOVQ idtptr+0(FP), AX
	LIDT (AX)
	RET

// ltr(sel uint64) - Load Task Register.
TEXT ·ltr(SB), NOSPLIT, $0-8
	MOVQ sel+0(FP), AX
	LTR  AX
	RET

// reloadCS returns from the current interrupt handler.
TEXT ·reloadCS(SB), NOSPLIT, $0
	// save ip
	MOVQ 0(SP), AX

	// save sp
	MOVQ SP, BX
	ADDQ $8, BX

	// rerange the stack, as in an interrupt stack
	PUSHQ $0x10 // SS
	PUSHQ BX
	PUSHFQ
	PUSHQ $8
	PUSHQ AX

	// IRET
	IRETQ

