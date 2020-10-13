# The process of starting the program

Take GOOS=linux GOARCH=386 as an example, the normal go program uses `_rt0_386_linux` in `runtime/rt0_linux_386.s` as the entrance of the whole program

Since go assumes that it is running on Linux, and we are running on bare metal, a simple hardware initialization is required before starting the runtime of go. We need to modify the entry address of the go program and redirect it to our own entry function `kernel.rt0`.

Basic initialization actions are performed in our own entry function, such as `gdt`, `thread`, `memory`, etc. At the end of initialization, the thread scheduler will be started, and the thread scheduler will then schedule `thread0`, the real go The runtime entry is called in thread0, and then the real go runtime initialization operation starts.

# Memory layout

```
.------------------------------------------.
|  Virtual memory(managed by go runtime)   |
:------------------------------------------: 1GB
| ........                                 |
:------------------------------------------: memtop(256MB or more)
| Physical memory(managed by eggos kernel) |
:------------------------------------------: 20MB
| Kernel image                             |
:------------------------------------------: 1MB
| Unused                                   |
'------------------------------------------' 0 
```

The first 1KB of memory is not page-mapped, and the subsequent memory until `memtop` is a direct mapping from virtual memory to physical memory.

The virtual address space available for go runtime starts from 1GB, and it manages the virtual address space itself, so what the kernel does is allocate physical pages according to the mmap system call of go runtime and map them to virtual memory.


# Trap

# Syscall