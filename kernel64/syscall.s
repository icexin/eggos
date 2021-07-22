#include "textflag.h"


TEXT ·syscallEntry1(SB), NOSPLIT, $0
    CALL ·dotrap(SB)

