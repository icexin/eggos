# Frequently Asked Questions

## General Questions

### What version of go does this project require?

This project requires anything within the `go 1.16.x` range of versions. Anything newer simply wont work, see: [issue 92](https://github.com/icexin/eggos/issues/92)


### Does this mean I have to downgrade go?

Nope. You can install [multiple versions](https://go.dev/doc/manage-install#installing-multiple) of go, which can be used with `egg` as per:

```bash
$ GOROOT=$(go1.16.13 env GOROOT) egg build
```

### How do I disable mouse/ network/ filesystem support?

By default, running `egg build` will transparently load a set of well tested, useful drivers and kernel components into your application by generating an init function which configures things.

This works wonderfully for turning any application into a useful, and usable, unikernel but is not without drawbacks- especially where things like mice, or filesystem support is not needed.

You can run `egg generate` at the root of your project and then edit the file `zz_load_eggos.go` accordingly. The presence of this file will stop `egg build` generating anything its self.
