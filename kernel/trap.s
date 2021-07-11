#include "textflag.h"

TEXT ·idt_init(SB), NOSPLIT, $0
	CALL ·fillidt(SB)
	LIDT ·idtptr(SB)
	RET

#define m_stack 0
#define m_tf 4
#define m_fpstate 12

TEXT alltraps(SB), NOSPLIT, $0
	PUSHAL
	PUSHW DS
	PUSHW ES
	PUSHW FS
	PUSHW GS
	

	MOVL 0(FS), CX    // CX store mythread
	MOVL SP, m_tf(CX)

	// save FPU
	MOVL m_fpstate(CX), DX
	FXSAVE (DX)

	// call trap(tf)
	PUSHL SP
	CALL  ·dotrap(SB)
	ADDL  $4, SP
	JMP   ·trapret(SB)

TEXT ·trapret(SB), NOSPLIT, $0
	MOVL 0(FS), CX    // CX store mythread
	MOVL $0, m_tf(CX) // clear tf

	// restore FPU
	MOVL m_fpstate(CX), DX
	FXRSTOR (DX)
	
	POPW GS
	POPW FS
	POPW ES
	POPW DS
	POPAL

	ADDL $8, SP // skip trapno and errcode

	// IRET
	BYTE $0xCF

