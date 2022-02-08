//go:build sshd
// +build sshd

package sshd

import (
	"net"

	"github.com/gliderlabs/ssh"
	"github.com/icexin/eggos/app"
)

var rsaContent = `
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEAvuGE6drEKbdWh7jncCgN9y1xbmd8x9FRRf2GcEE5WlKM4u4ewKm2
fblRy9Ens9fr805EOmdjVGGN9bN+6XxdsDRxuNTHVU2C08r7Ei5CG77fdTtePAil1CV8u8
JzodWS+ngQy/6v0qmObCrBldwpNhBZeBYgEWwc2pCByj4UffxRpK7Oxij30K9FZwWsHdSg
f+ktvqaatdURqQWRbS/2VAcNczZgNtzj3kpAnx8yZxk5HwrFuRNPOrxnd3i7mmwux5ZVy8
q7338NL15/Hx/mqZBSs/0HGkP5HRSAkpohX9ZEm2YWKs08lQ1r/zoMM3AxD78uJhqc6ZT0
u276z4WsfQAAA9BATdEVQE3RFQAAAAdzc2gtcnNhAAABAQC+4YTp2sQpt1aHuOdwKA33LX
FuZ3zH0VFF/YZwQTlaUozi7h7AqbZ9uVHL0Sez1+vzTkQ6Z2NUYY31s37pfF2wNHG41MdV
TYLTyvsSLkIbvt91O148CKXUJXy7wnOh1ZL6eBDL/q/SqY5sKsGV3Ck2EFl4FiARbBzakI
HKPhR9/FGkrs7GKPfQr0VnBawd1KB/6S2+ppq11RGpBZFtL/ZUBw1zNmA23OPeSkCfHzJn
GTkfCsW5E086vGd3eLuabC7HllXLyrvffw0vXn8fH+apkFKz/QcaQ/kdFICSmiFf1kSbZh
YqzTyVDWv/OgwzcDEPvy4mGpzplPS7bvrPhax9AAAAAwEAAQAAAQAx5epE57dX4GFyYVe+
7fmYn/yDC/KGmaVRUpEOTz6a6fGCcRUA8FyQSR2k1iw2yz8W/2K+kcBZkpb1n9KRXr1vDo
ab9qOVHQoSK4GuowENF7x6fOaJcwlGh/YvbwmjSJ1/dFuPuChmPYTJqfOpJUBwrZ110vLX
Gxf/2r7TC593v0YvaMfaTK1JLUofjTVmei/rl/6jP9AvoMSpFx+uVlCXaEdeULt9/EUgDS
aYe5jMcr9dTDIJ2GXCyJEN+n5aSOyJlHqcMW5MqJsTHllq7MlXF9GRCfq2RXD7PzXYVHv0
9bDIggFyh9NHSFCGOo8+yzyv0zgJkUc5GOfFzfDdSZtRAAAAgBpJQdv0NwX8yGLpHmx2mt
E2A+vi36QsHxlqGGF8ulDSqhppSQfNi5d+TZp7tcWCczUvyziiyaXSRlLK5OAaoUwSdpOM
rSkMr2KTM056EbHjlT6HM7vvk6uzMtnSbekx/1S16+PMxye3+sGsJDLjAwE2hBMiuMVOBL
i0T7Z9qk3ZAAAAgQDokbmEPJxJhFkNbc8IeCpn2xZalkI7vnTroCuPFjOwxYaySiV3yhCM
yvFOqslXpH0JIkBYFGKVFzUGum68tNDzDrevTPxqi6cxXGd2auR99VbChC2DqUfQQdCDfq
2bCzuMgCoNmSB7c1X4u6Y1y1tDUxfOlWru3fazJhpXNTISEwAAAIEA0hyYxsFtfIp12Rqi
sPsViNSFr8jXUcPl+NHdhixWuLXCEBva6FMZozgh6oS1tVe++iII5ivbOvU7LAq/k1tdzl
z8eS3BbfyIqnULDKFAbQJjZKkX1x4FMjr01dClgyQa9kubhMP7r2SbTQg9BxmQqavrVc+L
vt2IXe8kH/A9mS8AAAAbZmFuYmluZ3hpbkBmYW5iaW5neGluLmxvY2Fs
-----END OPENSSH PRIVATE KEY-----
`

func main(ctx *app.Context) error {

	ssh.Handle(func(s ssh.Session) {
		shell := app.Get("sh")
		ctx := &app.Context{
			Stdin:  s,
			Stdout: s,
			Stderr: s,
		}
		ctx.Init()
		shell(ctx)
	})
	l, err := net.Listen("tcp", "0.0.0.0:22")
	if err != nil {
		return err
	}
	srv := &ssh.Server{
		Handler: ssh.DefaultHandler,
	}
	srv.SetOption(ssh.NoPty())
	srv.SetOption(ssh.HostKeyPEM([]byte(rsaContent)))
	ctx.Printf("ssh server start at :22\n")
	err = srv.Serve(l)
	return err
}

func init() {
	app.Register("sshd", main)
}
