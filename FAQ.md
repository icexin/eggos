# Frequently Asked Questions

## General Questions

### What version of go does this project require?

This project requires anything within the `go 1.16.x` range of versions. Anything newer simply wont work, see: [issue 92](https://github.com/icexin/eggos/issues/92)


### Does this mean I have to downgrade go?

Nope. You can install [multiple versions](https://go.dev/doc/manage-install#installing-multiple) of go, which can be used with `egg` as per:

```bash
$ GOROOT=$(go1.16.13 env GOROOT) egg build
```
