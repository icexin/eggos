#include "textflag.h"

// CLI - Clear Interrupt Flag
TEXT ·Cli(SB), NOSPLIT, $0
	CLI
	RET

// STI - Set Interrupt Flag
TEXT ·Sti(SB), NOSPLIT, $0
	STI
	RET

// HLT - Halt
TEXT ·Hlt(SB), NOSPLIT, $0
	HLT
	RET

