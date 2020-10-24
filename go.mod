module github.com/icexin/eggos

go 1.13

require (
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/fogleman/gg v1.3.0
	github.com/fogleman/nes v0.0.0-20200820131603-8c4b9cf54c35
	github.com/gin-gonic/gin v1.6.3
	github.com/gliderlabs/ssh v0.3.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/netstack v0.0.0-20191123085552-55fcc16cd0eb
	github.com/hirochachacha/go-smb2 v1.0.3
	github.com/jakecoffman/cp v1.0.0
	github.com/klauspost/cpuid v1.3.1
	github.com/mattn/go-shellwords v1.0.10
	github.com/peterh/liner v1.2.0
	github.com/rakyll/statik v0.1.7
	github.com/robertkrimen/otto v0.0.0-20191219234010-c382bd3c16ff
	github.com/spf13/afero v1.4.0
	github.com/stretchr/testify v1.6.1 // indirect
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/image v0.0.0-20200801110659-972c09e46d76
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
)

replace (
	github.com/fogleman/nes v0.0.0-20200820131603-8c4b9cf54c35 => github.com/icexin/nes v0.0.0-20200906065456-8ff789fac016
	github.com/google/netstack v0.0.0-20191123085552-55fcc16cd0eb => github.com/icexin/netstack v0.0.0-20201005132454-bd9d0399feb1
)
