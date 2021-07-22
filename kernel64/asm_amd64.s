#include "textflag.h"

TEXT ·rt0(SB), NOSPLIT, $0-0
	// switch to new stack
	MOVQ $0x80000, SP

	CALL ·preinit(SB)
	INT  $3

	// never return

TEXT ·go_entry(SB), NOSPLIT, $0
	JMP _rt0_amd64_linux(SB)

TEXT ·sseInit(SB), NOSPLIT, $0
	MOVL CR0, AX
	ANDW $0xFFFB, AX
	ORW  $0x2, AX
	MOVL AX, CR0
	MOVL CR4, AX
	ORW  $3<<9, AX
	MOVL AX, CR4
	RET

TEXT ·lgdt(SB), NOSPLIT, $0-8
	MOVQ gdtptr+0(FP), AX
	LGDT (AX)
	RET

TEXT ·lidt(SB), NOSPLIT, $0-8
	MOVQ idtptr+0(FP), AX
	LIDT (AX)
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
