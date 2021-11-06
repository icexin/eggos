#include "textflag.h"

TEXT ·lgdt(SB), NOSPLIT, $0-8
	MOVQ gdtptr+0(FP), AX
	LGDT (AX)
	RET

TEXT ·lidt(SB), NOSPLIT, $0-8
	MOVQ idtptr+0(FP), AX
	LIDT (AX)
	RET

TEXT ·ltr(SB), NOSPLIT, $0-8
	MOVQ sel+0(FP), AX
	LTR  AX
	RET

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

