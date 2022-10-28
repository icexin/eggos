package kernel

import "github.com/jspc/eggos/drivers/clock"

// called when go runtime init done
func Init() {
	clockTimeInit()
	idleInit()
	go runTrapThread()
	go runSyscallThread()
	bootstrapDone = true
}

func clockTimeInit() {
	t := clock.ReadCmosTime()
	baseUnixTime = t.Time().Unix()
	clockBaseCounter = counter
}
