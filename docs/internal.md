# The process of starting the program

Take GOOS=linux GOARCH=386 as an example, the normal go program uses `_rt0_386_linux` in `runtime/rt0_linux_386.s` as the entrance of the whole program

Since go assumes that it is running on Linux, and we are running on bare metal, a simple hardware initialization is required before starting the runtime of go. We need to modify the entry address of the go program and redirect it to our own entry function `kernel.rt0`.

Basic initialization actions are performed in our own entry function, such as `gdt`, `thread`, `memory`, etc. At the end of initialization, the thread scheduler will be started, and the thread scheduler will then schedule `thread0`, the real go The runtime entry is called in thread0, and then the real go runtime initialization operation starts.

# Memory layout

# Trap

# Syscall