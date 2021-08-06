![eggos](assets/files/eggos.png)

![CI](https://github.com/icexin/eggos/workflows/CI/badge.svg)

使用Go语言编写的运行在x86裸机上的unikernel

将Go程序运行在x86裸机上，全部使用Go语言编写(bootloader里面有少量汇编和c)，支持大部分Go的核心功能(GC，goroutine等)，包括大部分Go的标准库，同时也附带一个网络协议栈，从而能运行大部分基于`net`库的第三方库。

不同于传统的操作系统内核，eggos没有对用户态和内核态的代码进行隔离，整个unikernel运行在一个地址空间上。没有进程的概念以及进程间通信，但是完整支持了Go的goroutine和channel。另外也没有传统的ELF加载器，但是附带了一个Javascript的解释器。


# 背景

Go的runtime提供了一些基本的操作系统抽象，goroutine对应进程，channel对应进程间通信，另外Go有自己的虚拟内存管理，因此就产生了将Go程序运行在裸机上的想法。

事实证明，Go有操作硬件的能力：对内存布局的掌控、直接翻译硬件指令、以及类C的语法，这些能力都让Go程序运行在裸机上成为了可能。

然而挑战还是比较有的，Go在很多指令里面打桩来进行协程调度以及内存GC，在一些不能重入的地方，如中断处理和系统调用，都带来了一些麻烦。 总的来说，使用Go来操作硬件是一种乐趣，一方面也让我深入了解了Go的runtime，另一方面也是提供了一种除了c语言之外在裸机上编写操作系统内核的尝试。


# 架构

<img src="https://i.imgur.com/gnq4m9h.png" width="700" />

# 应用截图

![js](https://i.imgur.com/Canhd8D.gif)
![nes](https://i.imgur.com/WugXcTk.gif)
![gui](https://i.imgur.com/jILuMMk.png)



# 功能列表

- Go的内置功能，如GC，goroutine，channel等
- 一个支持行编辑的终端
- 支持TCP/IP的协议栈
- Go风格的VFS抽象，使用[afero](https://github.com/spf13/afero)
- NES模拟器，使用[nes](https://github.com/fogleman/nes)
- Javascript解释器，使用[otto](https://github.com/robertkrimen/otto)
- GUI支持，使用[nucular](https://github.com/aarzilli/nucular)
- 一些简单的应用，如(httpd, sshd)


# 依赖

- Go 1.16.x (高版本可能运行不了)
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

# 快速开始

``` bash
$ mage qemu
```

# 构建自己的unikernel

`eggos` 有将普通Go程序转换成运行于裸机上的 `ELF unikernel` 的能力。

首先获取egg二进制，可以通过 https://github.com/icexin/eggos/releases 下载。也可以直接运行`go install github.com/icexin/eggos/cmd/egg`获取。

在你的项目目录运行`egg build -o kernel.elf`，接着运行`egg run kernel.elf`启动qemu虚拟机。


`egg pack -o eggos.iso -k kernel.elf` 可以将内核打包成一个iso文件，通过 https://github.com/ventoy/Ventoy 即可运行在真实的机器上。

Happy hacking!

# Debug

eggos支持直接使用GDB debug，或者使用vscode这样带图形界面的IDE来debug。

mac用户使用如下命令安装GDB

``` bash
brew install x86_64-elf-gdb
```

vscode用户通过安装`Native Debug` 扩展来支持GDB。

首先执行`mage qemudebug`来让qemu运行于debug模式，之后就可以使用vscode自带的debug功能debug了。项目自带vscode的debug配置文件。

另外Go语言也自带了对GDB的支持，见[Debugging Go Code with GDB](https://golang.org/doc/gdb)

![vscode-gdb](https://i.imgur.com/KIg6l5A.png)


# 文档

[docs/README.md](docs/README.md)

# Roadmap

- [ ] WASM runner
- [x] GUI support
- [x] 3D graphic
- [x] x86_64 support
- [ ] SMP support

# 关于贡献

eggos在活跃开发中，你将会遇到很多bug，包括不限于panic或者死机。如果你想贡献eggos，欢迎提交PR，谢谢！


# 特别感谢

我的小闺女的出生给小家庭带来了很多欢乐，这个工程就是用她的小名`蛋蛋`命名的。另外我的妻子和丈母娘也给了我很大的支持来更新这个项目，特别感谢她们的默默付出. :heart: :heart: :heart:
