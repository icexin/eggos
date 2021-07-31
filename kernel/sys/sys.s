#include "textflag.h"

TEXT ·Cli(SB), NOSPLIT, $0
	CLI
	RET

TEXT ·Sti(SB), NOSPLIT, $0
	STI
	RET

TEXT ·Hlt(SB), NOSPLIT, $0
	HLT
	RET


