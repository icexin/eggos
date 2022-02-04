#include "textflag.h"

// pageEnable enables translation from virtual address (linear address) to
// physical address, based on the page directory set in the CR3 register.
TEXT ·pageEnable(SB), NOSPLIT, $0-0
	// enable PAE
	MOVQ CR4, AX
	BTSQ $5, AX
	MOVQ AX, CR4

	// enable page
	MOVQ CR0, AX
	BTSQ $31, AX
	MOVQ AX, CR0
	RET

// lcr3(topPage uint64) sets the CR3 register.
TEXT ·lcr3(SB), NOSPLIT, $0-8
	// setup page dir
	MOVQ topPage+0(FP), AX
	MOVQ AX, CR3
	RET

