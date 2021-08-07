This directory contains examples of how `eggos` can be used to write common go applications.

# helloworld

``` sh
$ cd helloworld
$ egg run
```

# concurrent prime sieve

``` sh
$ cd prime-sieve
$ egg run
```

# http server

``` sh
$ cd httpd
$ egg run -p 8000:8000
```

Access this address from browser http://127.0.0.1:8000/hello

# simple repl program

``` sh
$ cd repl
$ egg run
```

# simple animation

Code from `The Go Programming Language`

``` sh
$ cd graphic
$ egg pack -o graphic.iso
$ egg run graphic.iso
```

# handle syscall

This example shows how to add or modify syscall

``` sh
$ cd syscall
$ egg run
```