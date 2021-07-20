#include "textflag.h"

// Outb(port uint16, data byte)
TEXT ·Outb(SB), NOSPLIT, $0-3
	MOVW port+0(FP), DX
	MOVB data+2(FP), AX
	OUTB
	RET

// byte Inb(reg uint16)
TEXT ·Inb(SB), NOSPLIT, $0-9
	MOVW port+0(FP), DX
	XORW AX, AX
	INB
	MOVB AX, ret+8(FP)
	RET

// Outl(port uint16, data uint32)
TEXT ·Outl(SB), NOSPLIT, $0-8
	MOVW port+0(FP), DX
	MOVL data+4(FP), AX
	OUTL
	RET

TEXT ·Inl(SB), NOSPLIT, $0-12
	MOVW port+0(FP), DX
	INL
	MOVL AX, ret+8(FP)
	RET

TEXT ·SetAX(SB), NOSPLIT, $0-8
	MOVQ val+0(FP), AX
	RET

TEXT ·Flags(SB), NOSPLIT, $0-8
	PUSHFQ
	POPQ AX
	MOVQ AX, ret+0(FP)
	RET

TEXT ·Cr2(SB), NOSPLIT, $0-8
	MOVQ CR2, AX
	MOVQ AX, ret+0(FP)
	RET

TEXT ·Fxsave(SB), NOSPLIT, $0-8
	MOVQ addr+0(FP), AX
	FXSAVE (AX)
	RET
