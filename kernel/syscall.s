#include "textflag.h"

#define tls_my 0
#define tls_ax 8
#define m_kstack 40
#define ucode_idx  3
#define	udata_idx  4
#define rpl_user   3

#define SYS_clockgettime 228

TEXT ·syscallEntry(SB), NOSPLIT, $0
    // save AX
    MOVQ AX, tls_ax(GS)
    // AX == pointer of current thread
    MOVQ tls_my(GS), AX
    // AX == kernel stack
    MOVQ m_kstack(AX), AX

    // push regs like INT 0x80
    SUBQ $40, AX
    // CX store IP
    MOVQ CX, 0(AX)
    // save CS
    MOVQ $ucode_idx<<3|rpl_user, 8(AX)
    // R11 store FLAGS
    MOVQ R11, 16(AX)
    // save SP
    MOVQ SP, 24(AX)
    // save SS
    MOVQ $udata_idx<<3|rpl_user, 32(AX)

    // change SP
    MOVQ AX, SP

    // restore AX
    MOVQ tls_ax(GS), AX

    // jmp INT 0x80
    JMP ·trap128(SB)
    
TEXT ·vdsoGettimeofday(SB), NOSPLIT, $0
    MOVQ $SYS_clockgettime, AX
    // DI store *TimeSpec, but clockgettime need SI
    MOVQ DI, SI
    MOVQ $0, DI
    INT $0x80
    RET

