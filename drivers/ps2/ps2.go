package ps2

import "github.com/icexin/eggos/kernel/sys"

const (
	_CMD_PORT  = 0x64
	_DATA_PORT = 0x60
)

func waitCanWrite() {
	timeout := 1000
	for timeout > 0 {
		timeout--
		x := sys.Inb(_CMD_PORT)
		// input buffer empty means we can write to controller
		if x&0x02 == 0 {
			return
		}
	}
}

func waitCanRead() {
	timeout := 1000
	for timeout > 0 {
		timeout--
		x := sys.Inb(_CMD_PORT)
		// output buffer full means we can read from controller
		if x&0x01 != 0 {
			return
		}
	}
}

func ReadDataNoWait() byte {
	return sys.Inb(_DATA_PORT)
}

func ReadData() byte {
	waitCanRead()
	return sys.Inb(_DATA_PORT)
}

func WriteData(x byte, needAck bool) {
	waitCanWrite()
	sys.Outb(_DATA_PORT, x)
	if needAck {
		ReadAck()
	}
}

func ReadAck() {
	x := ReadData()
	if x != 0xFA {
		panic("not a ps2 ack packet")
	}
}

func ReadCmd() byte {
	return sys.Inb(_CMD_PORT)
}

func WriteCmd(x byte) {
	waitCanWrite()
	sys.Outb(_CMD_PORT, x)
}

func WriteMouseData(x byte) {
	WriteCmd(0xD4)
	WriteData(x, true)
}

func ReadMouseData(x byte) byte {
	WriteCmd(0xD4)
	WriteData(x, true)
	return ReadData()
}
