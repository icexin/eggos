package kernel

import "github.com/icexin/eggos/clock"

// called when go runtime init done
func Init() {
	initClockTime()
	idle_init()
	go traploop()
	go handleForward()
	bootstrapDone = true
}

func initClockTime() {
	t := clock.ReadCmosTime()
	baseUnixTime = t.Time().Unix()
	clockBaseCounter = counter
}
