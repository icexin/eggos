#include "textflag.h"

TEXT ·rt0(SB), NOSPLIT, $0-0
	// switch to new stack
	MOVQ $0x80000, SP

	CALL ·preinit(SB)
	INT  $3

	// never return

TEXT ·go_entry(SB), NOSPLIT, $0
	JMP _rt0_amd64_linux(SB)

