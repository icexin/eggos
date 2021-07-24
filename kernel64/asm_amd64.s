#include "textflag.h"

TEXT ·rt0(SB), NOSPLIT, $0-0
	// switch to new stack
	MOVQ $0x80000, SP

	CALL ·preinit(SB)
	INT  $3

	// never return

TEXT ·go_entry(SB), NOSPLIT, $0
	SUBQ  $256, SP
	PUSHQ SP
	CALL  ·prepareArgs(SB)
	ADDQ  $8, SP
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

TEXT ·rdmsr(SB),NOSPLIT,$0-16
	MOVL reg+0(FP), CX
	RDMSR
	MOVL	AX, lo+8(FP)
	MOVL	DX, hi+12(FP)
	RET

TEXT ·wrmsr(SB),NOSPLIT,$0-16
	MOVL	reg+0(FP), CX
	MOVL	lo+8(FP), AX
	MOVL	hi+12(FP), DX
	WRMSR
	RET
