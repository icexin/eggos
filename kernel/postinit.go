package kernel

// called when go runtime init done
func Init() {
	go traploop()
	go handleForward()
	bootstrapDone = true
}
