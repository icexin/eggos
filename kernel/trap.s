#include "textflag.h"

TEXT alltraps(SB), NOSPLIT, $0
    PUSHQ R15
    PUSHQ R14
    PUSHQ R13
    PUSHQ R12
    PUSHQ R11
    PUSHQ R10
    PUSHQ R9
    PUSHQ R8
    PUSHQ DI
    PUSHQ SI
    PUSHQ BP
    PUSHQ DX
    PUSHQ CX
    PUSHQ BX
    PUSHQ AX

    PUSHQ SP
    CALL  ·dotrap(SB)
	ADDQ  $8, SP
	JMP   ·trapret(SB)

TEXT ·trapret(SB), NOSPLIT, $0
    POPQ AX
    POPQ BX
    POPQ CX
    POPQ DX
    POPQ BP
    POPQ SI
    POPQ DI
    POPQ R8
    POPQ R9
    POPQ R10
    POPQ R11
    POPQ R12
    POPQ R13
    POPQ R14
    POPQ R15

	ADDQ $16, SP // skip trapno and errcode

    IRETQ

