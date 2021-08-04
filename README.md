# EggOS

![CI](https://github.com/icexin/eggos/workflows/CI/badge.svg)

[中文说明](README_CN.md)

A Go unikernel running on x86 bare metal

Run a single Go applications on x86 bare metal, written entirely in Go (only a small amount of C and some assembly), support most features of Go (like GC, goroutine) and standard libraries, also come with a network stack that can run most `net` based libraries.

The entire kernel is a go application running on ring0. There are no processes and process synchronization primitives, only goroutines and channels. There is no elf loader, but there is a Javascript interpreter that can run js script files, and a WASM interpreter will be added to run WASM files later.

# Background

Go's runtime provides some basic operating system abstractions. Goroutine corresponds to processes and channel corresponds to inter-process communication. In addition, go has its own virtual memory management, so the idea of running Go programs on bare metal was born.

It turns out that Go has the ability to manipulate hardware resources, thanks to Go's controllable memory layout, the ability to directly translate hardware instructions without a virtual machine, and C-like syntax. These all make it possible for Go to write programs that run on bare metal.
However, there are also some challenges. Go piling in many instructions to perform coroutine scheduling and memory GC, which brings some troubles in some places that cannot be reentrant, such as interrupt handling and system calls.

In general, writing kernel in Go is a very interesting experience. On the one hand, it gave me a deep understanding of Go's runtime. On the other hand, it also provided an attempt to write the operating system kernel on bare metal in addition to the C language.

# Architecture

<img src="https://i.imgur.com/gnq4m9h.png" width="700" />

# Snapshot

![js](https://i.imgur.com/Canhd8D.gif)
![nes](https://i.imgur.com/WugXcTk.gif)
![gui](https://i.imgur.com/jILuMMk.png)



# Feature list

- Basic Go features, such as GC, goroutine, channel.
- A simple console support basic line editting.
- Network stack support tcp/udp.
- Go style vfs abstraction using [afero](https://github.com/spf13/afero)
- A nes game emulator using [nes](https://github.com/fogleman/nes)
- A Javascript interpreter using [otto](https://github.com/robertkrimen/otto)
- VBE based frame buffer.
- Some simple network apps(httpd, sshd).
- GUI support by [nucular](https://github.com/aarzilli/nucular).


# Dependencies

- Go 1.16.x (higher versions may not work)
- gcc
- qemu
- mage

## MacOS

``` bash
$ go get github.com/magefile/mage
$ brew install x86_64-elf-binutils x86_64-elf-gcc x86_64-elf-gdb
$ brew install qemu
```

## Ubuntu

``` bash
$ go get github.com/magefile/mage
$ sudo apt-get install build-essential qemu
```

# Quickstart

``` bash
$ mage qemu
```

# Build you own unikernel

eggos has the ability to convert normal go program into an ELF unikernel which an be running on bare metal. 

First, get the `egg` binary, which can be accessed through https://github.com/icexin/eggos/releases, or directly through `go install github.com/icexin/eggos/cmd/egg`

Run `egg build -o kernel.elf` in your project directory to get the kernel file, and then run `egg run -k kernel.elf` to start the qemu virtual machine to run the kernel.

`egg pack -o eggos.iso` can pack the kernel as an iso file, and use https://github.com/ventoy/Ventoy to run the iso file on a bare metal.

Happy hacking!

# Debug

You can directly use the gdb command to debug, or use vscode for graphical debugging.

First you need to install gdb, if you are under macos, execute the following command

``` bash
brew install x86_64-elf-gdb
```

Use the extension `Native Debug` in vscode to support debugging with gdb

First execute the `mage qemudebug` command to let qemu start the gdb server, and then use the debug function of vscode to start a debug session. The debug configuration file of vscode is built into the project.

Go provides simple support for gdb, see [Debugging Go Code with GDB](https://golang.org/doc/gdb) for details

![vscode-gdb](https://i.imgur.com/KIg6l5A.png)

# Running on bare metal

If you want eggos to run on bare metal, it is recommended to use grub as the bootloader.

The multiboot.elf generated after executing the make command is a kernel image conforming to the multiboot specification, which can be directly recognized by grub and booted on a bare metal. The sample configuration file refer to `boot/grub.cfg`

# Documentation

[docs/README.md](docs/README.md)

# Roadmap

- [ ] WASM runner
- [x] GUI support
- [x] 3D graphic
- [x] x86_64 support
- [ ] SMP support

# Bugs

The program still has a lot of bugs, and often loses response or panic. If you are willing to contribute, please submit a PR, thank you!

# Special thanks

The birth of my little daughter brought a lot of joy to the family. This project was named after her name `xiao dan dan`. My wife and mother also gave me a lot of support and let me update this project in my spare time. :heart: :heart: :heart:
