package trap

var trapHandlers = [256]TrapHandler{}

type TrapHandler func()

//go:nosplit
func Handler(no int) TrapHandler {
	return trapHandlers[no]
}

//go:nosplit
func Register(idx int, handler func()) {
	trapHandlers[idx] = handler
}
