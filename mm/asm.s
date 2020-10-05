#include "textflag.h"

#define CR0_WP          0x00010000      // Write Protect
#define CR0_PG          0x80000000      // Paging

TEXT ·page_enable(SB),NOSPLIT,$0-0
	// enable page
	MOVL	CR0, AX
	ORL		$(CR0_WP|CR0_PG), AX
	MOVL	AX, CR0
	RET

TEXT ·lcr3(SB),NOSPLIT,$0-4
	// setup page dir
	MOVL	pgdir+0(FP), AX
	MOVL	AX, CR3
	RET


