#include "textflag.h"

#define SYS_clone	     120
#define SYS_sched_yield	 158

#define _KCODE_IDX 1
#define _KDATA_IDX 2
#define _TSS_IDX 5

// copy from go_tls.h
#define get_tls(r)	MOVL TLS, r
#define g(r)	0(r)(TLS*1)

TEXT ·rt0(SB), NOSPLIT, $0-8
	// save multiboot from old stack
	MOVL magic+0(FP), AX
	MOVL mbi+4(FP), BX

	// switch to new stack
	MOVL $0x80000, SP

	SUBL $8, SP
	MOVL AX, 0(SP)
	MOVL BX, 4(SP)
	CALL ·preinit(SB)
	INT  $3

	// never return

TEXT ·go_entry(SB), NOSPLIT, $0
	SUBL  $256, SP
	PUSHL SP
	CALL  ·prepareArgs(SB)
	ADDL  $4, SP

	JMP _rt0_386_linux(SB)

TEXT ·sse_init(SB), NOSPLIT, $0
	MOVL CR0, AX
	ANDW $0xFFFB, AX
	ORW  $0x2, AX
	MOVL AX, CR0
	MOVL CR4, AX
	ORW  $3<<9, AX
	MOVL AX, CR4
	RET

TEXT load_cs(SB), NOSPLIT, $0
	// save ip
	MOVL 0(SP), AX

	// rerange the stack, as in an interrupt stack
	ADDL  $4, SP // skip old IP register in stack
	PUSHFL
	PUSHL $_KCODE_IDX<<3
	PUSHL AX

	// IRET
	BYTE $0xCF

TEXT ·gdt_init(SB), NOSPLIT, $0
	CALL ·fillgdt(SB)

	LGDT ·gdtptr(SB)
	MOVL $_KDATA_IDX<<3, AX
	MOVW AX, DS
	MOVW AX, ES
	MOVW AX, SS

	MOVL $0x00, AX
	MOVW AX, FS
	MOVW AX, GS

	MOVL $_TSS_IDX<<3, AX
	LTR  AX

	CALL load_cs(SB)

	RET

TEXT ·set_fs(SB), NOSPLIT, $0-4
	MOVW idx+0(FP), AX
	SHLL $3, AX
	ADDL $3, AX
	MOVW AX, FS
	RET

TEXT ·set_gs(SB), NOSPLIT, $0-4
	MOVW idx+0(FP), AX
	SHLL $3, AX
	ADDL $3, AX
	MOVW AX, GS
	RET

// func swtch(old **context, _new *context)
TEXT ·swtch(SB), NOSPLIT, $0-8
	MOVL old+0(FP), AX
	MOVL _new+4(FP), DX

	// Save old callee-saved registers
	PUSHL BP
	PUSHL BX
	PUSHL SI
	PUSHL DI

	// Switch stacks
	MOVL SP, (AX)
	MOVL DX, SP

	POPL DI
	POPL SI
	POPL BX
	POPL BP
	RET

// func callSigHandler(pc, sp uintptr, no uintptr, info *siginfo, ctx *ucontext)
TEXT ·callSigHandler(SB), NOSPLIT, $0-20
	PUSHL BP
	MOVL  SP, BP

	MOVL sp+4(FP), CX
	SUBL $12, CX

	MOVL no+8(FP), DX
	MOVL DX, 0(CX)

	MOVL info+12(FP), DX
	MOVL DX, 4(CX)

	MOVL ctx+16(FP), DX
	MOVL DX, 8(CX)

	MOVL pc+0(FP), AX
	MOVL CX, SP
	CALL AX

	MOVL BP, SP
	POPL BP
	RET

// func call(pc,a0,a1,a2 uintptr)
TEXT ·call(SB), NOSPLIT, $12-16
	MOVL pc+0(FP), CX

	MOVL a0+4(FP), AX
	MOVL AX, 0(SP)

	MOVL a1+8(FP), AX
	MOVL AX, 4(SP)

	MOVL a2+12(FP), AX
	MOVL AX, 8(SP)
	CALL CX
	RET

TEXT ·set_mythread(SB), NOSPLIT, $4-4
	MOVL tid+0(FP), AX
	MOVL AX, 0(FS)
	MOVL AX, 0(SP)
	CALL ·switchThreadContext(SB)
	RET

TEXT ·sys_clone(SB), NOSPLIT, $0-12
	MOVL $SYS_clone, AX
	MOVL pc+0(FP), DX
	MOVL stack+4(FP), CX
	INT  $0x80

	// In parent, return.
	CMPL AX, $0
	JEQ  3(PC)
	MOVL AX, tid+8(FP)
	RET

	NOP SP // tell vet SP changed - stop checking offsets
	JMP DX

TEXT ·sys_yield(SB), NOSPLIT, $0
	MOVL $SYS_sched_yield, AX
	INT  $0x80
	RET

TEXT ·Mythread(SB), NOSPLIT, $0-4
	MOVL 0(FS), AX
	MOVL AX, ret+0(FP)
	RET

TEXT ·getg(SB), NOSPLIT, $0-4
	get_tls(CX)
	MOVL g(CX), BX
	MOVL BX, ret+0(FP)
	RET


