#include "textflag.h"

// Outb(port uint16, data byte)
TEXT ·Outb(SB), NOSPLIT, $0-3
	MOVW port+0(FP), DX
	MOVB data+2(FP), AX
	OUTB
	RET

// byte Inb(reg uint16)
TEXT ·Inb(SB), NOSPLIT, $0-5
	MOVW port+0(FP), DX
	XORW AX, AX
	INB
	MOVB AX, ret+4(FP)
	RET

// Outl(port uint16, data uint32)
TEXT ·Outl(SB), NOSPLIT, $0-8
	MOVW port+0(FP), DX
	MOVL data+4(FP), AX
	OUTL
	RET

TEXT ·Inl(SB), NOSPLIT, $0-8
	MOVW port+0(FP), DX
	INL
	MOVL AX, ret+4(FP)
	RET

TEXT ·Cli(SB), NOSPLIT, $0
	CLI
	RET

TEXT ·Sti(SB), NOSPLIT, $0
	STI
	RET

TEXT ·Hlt(SB), NOSPLIT, $0
	HLT
	RET

TEXT ·Cr2(SB), NOSPLIT, $0-4
	MOVL CR2, AX
	MOVL AX, ret+0(FP)
	RET

TEXT ·Flags(SB), NOSPLIT, $0-4
	PUSHFL
	POPL AX
	MOVL AX, ret+0(FP)
	RET

TEXT ·Mfence(SB), NOSPLIT, $0
	MFENCE
	RET

TEXT ·Fxsave(SB), NOSPLIT, $0-4
	MOVL addr+0(FP), AX
	FXSAVE (AX)
	RET

TEXT ·SetAX(SB), NOSPLIT, $0-4
	MOVL val+0(FP), AX
	RET
