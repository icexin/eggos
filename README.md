# EggOS

A Go unikernel running on x86 bare metal

Run a single Go applications on x86 bare metal, written entirely in Go (a small amount of c and some assembly), support most features of Go (like GC, goroutine) and standard libraries, also come with a network stack that can run most `net` based libraries.

# Snapshot

![js](https://i.imgur.com/Canhd8D.gif)
![nes](https://i.imgur.com/WugXcTk.gif)


# Feature list

- Basic Go features, such as GC, goroutine, channel.
- A simple console support basic line editting.
- Network stack support tcp/udp.
- Go style vfs abstraction using [afero](https://github.com/spf13/afero)
- A nes game simulator using [nes](https://github.com/fogleman/nes)
- A Javascript interpreter using [otto](https://github.com/robertkrimen/otto)
- VBE based frame buffer.
- Some simple network apps(httpd, sshd).


# Dependencies

- Go 1.13.x (only tested on Go1.13.x)
- i386-elf-gcc
- qemu

## MacOS

``` bash
$ brew tap nativeos/i386-elf-toolchain
$ brew install brew install i386-elf-binutils i386-elf-gcc i386-elf-gdb
$ brew install qemu
```

## Ubuntu

``` bash
$ sudo apt-get install build-essential qemu
```

# Quickstart

``` bash
$ make
$ make qemu
```

# Debug

Debug can directly use gdb for debugging, or use vscode for graphical debugging

First you need to install gdb, if you are under macos, execute the following command

``` bash
brew install i386-elf-gdb
```

Use the extension `Native Debug` in vscode to support debugging with gdb

If you need to debug, first execute the `make qemu-debug` command to let qemu start the gdb server, and then use gdb's `target remote: 1234` to attach to qemu's gdb server, or use vscode debug directly, the project has built-in vscode debug configuration file.

![vscode-gdb](https://i.imgur.com/KIg6l5A.png)

# Grub boot

TODO

# Documentations

TODO