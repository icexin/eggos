#include "textflag.h"

TEXT ·idt_init(SB), NOSPLIT, $0
	CALL ·fillidt(SB)
	LIDT ·idtptr(SB)
	RET

#define m_stack 0
#define m_tf 4

TEXT alltraps(SB), NOSPLIT, $0
	// 因为我们没有切换ring，因此需要显式保存一下sp指针，
	// 便于在中断处理程序中修改sp
	PUSHL SP
	PUSHAL
	PUSHW DS
	PUSHW ES
	PUSHW FS
	PUSHW GS

	MOVL 0(FS), CX    // CX store mythread
	MOVL SP, m_tf(CX)

	// call trap(tf)
	PUSHL SP
	CALL  ·dotrap(SB)
	ADDL  $4, SP
	JMP   ·trapret(SB)

TEXT ·trapret(SB), NOSPLIT, $0
	MOVL 0(FS), CX    // CX store mythread
	MOVL $0, m_tf(CX) // clear tf

	POPW GS
	POPW FS
	POPW ES
	POPW DS
	POPAL
	POPL SP

	ADDL $8, SP // skip trapno and errcode

	// IRET
	BYTE $0xCF

