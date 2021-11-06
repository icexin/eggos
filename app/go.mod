module github.com/icexin/eggos/app

go 1.16

require (
	github.com/aarzilli/nucular v0.0.0-20210408133902-d3dd7b05a80a
	github.com/fogleman/fauxgl v0.0.0-20200818143847-27cddc103802
	github.com/fogleman/gg v1.3.0
	github.com/fogleman/nes v0.0.0-20210605215016-0aace4b1814a
	github.com/gin-gonic/gin v1.7.2
	github.com/gliderlabs/ssh v0.3.3
	github.com/icexin/eggos v0.0.0-00010101000000-000000000000
	github.com/icexin/nk v0.1.0
	github.com/jakecoffman/cp v1.1.0
	github.com/klauspost/cpuid v1.3.1
	github.com/mattn/go-shellwords v1.0.12
	github.com/peterh/liner v1.2.1
	github.com/prometheus/client_golang v1.7.1
	github.com/robertkrimen/otto v0.0.0-20210614181706-373ff5438452
	github.com/spf13/afero v1.4.0
	golang.org/x/exp v0.0.0-20210729172720-737cce5152fc
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d
	golang.org/x/mobile v0.0.0-20210716004757-34ab1303b554
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
)

replace (
	github.com/aarzilli/nucular => github.com/icexin/nucular v0.0.0-20210713192454-c3f236ca56cb
	github.com/fogleman/nes => github.com/icexin/nes v0.0.0-20200906065456-8ff789fac016
	github.com/icexin/eggos => ../
)
