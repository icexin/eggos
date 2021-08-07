package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/icexin/eggos/kernel/isyscall"
)

var handler isyscall.Handler

func handleUname(req *isyscall.Request) {
	fmt.Println("syscall `uname` called")
	handler(req)
}

func main() {
	handler = isyscall.GetHandler(syscall.SYS_UNAME)
	isyscall.Register(syscall.SYS_UNAME, handleUname)
	_, err := os.Hostname()
	if err != nil {
		panic(err)
	}
}
