package e1000

const (
	/// Controls the device features and states.
	REG_CTRL = 0x0000
	/// Auto-Speed Detetion Enable.
	CTRL_ASDE = (1 << 5)
	/// Set link up.
	CTRL_SLU = (1 << 6)
	/// Device Reset.
	CTRL_RST = (1 << 26)

	/// Interrupt Mask Set.
	REG_IMS  = 0x00d0
	IMS_RXT0 = (1 << 7)

	/// Interrupt Mask Clear.
	REG_IMC = 0x00d8

	/// Interrupt Cause Read.
	REG_ICR = 0x00c0
	/// Receiver Timer Interrupt.
	ICR_RXT0 = (1 << 7)

	/// Multicast Table Array.
	REG_MTA_BASE = 0x5200
	/// The lower bits of the Ethernet address.
	REG_RECEIVE_ADDR_LOW = 0x5400
	/// The higher bits of the Ethernet address and some extra bits.
	REG_RECEIVE_ADDR_HIGH = 0x5404

	/// Receive Control.
	REG_RCTL = 0x0100
	/// Receiver Enable.
	RCTL_EN = (1 << 1)
	/// Strip Ethernet CRC from receiving packet.
	RCTL_SECRC = (1 << 26)
	/// Receive Buffer Size: 2048 bytes (assuming RCTL.BSEX == 0).
	RCTL_BSIZE = 0 << 16
	/// Broadcast Accept Mode.
	RCTL_BAM = (1 << 15)

	/// Receive Descriptor Base Low.
	REG_RDBAL = 0x2800
	/// Receive Descriptor Base High.
	REG_RDBAH = 0x2804
	/// Length of Receive Descriptors.
	REG_RDLEN = 0x2808
	/// Receive Descriptor Head.
	REG_RDH = 0x2810
	/// Receive Descriptor Tail.
	REG_RDT = 0x2818

	/// Transmit Control.
	REG_TCTL = 0x0400
	// Transmit Inter Packet Gap
	REG_TIPG = 0x0410
	/// Receiver Enable.
	TCTL_EN = (1 << 1)
	/// Pad Short Packets.
	TCTL_PSP = (1 << 3)

	/// Transmit Descriptor Base Low.
	REG_TDBAL = 0x3800
	/// Transmit Descriptor Base High.
	REG_TDBAH = 0x3804
	/// Length of Transmit Descriptors.
	REG_TDLEN = 0x3808
	/// Transmit Descriptor Head.
	REG_TDH = 0x3810
	/// Transmit Descriptor Tail.
	REG_TDT = 0x3818

	/// Insert FCS.
	TX_DESC_IFCS = (1 << 1)
	/// End Of Packet.
	TX_DESC_EOP = (1 << 0)
	// Report Status
	TX_DESC_RS = (1 << 3)

	// eeprom register
	REG_EEPROM = 0x0014
	REG_RXADDR = 0x5400

	/// Descriptor Done.
	RX_DESC_DD = (1 << 0)
	/// End Of Packet.
	RX_DESC_EOP = (1 << 1)
)

const (
	/// The size of buffer to store received/transmtting packets.
	BUFFER_SIZE = 2048
	/// Number of receive descriptors.
	NUM_RX_DESCS = 32
	/// Number of receive descriptors.
	NUM_TX_DESCS = 32

	TX_BUF_SZ = 1024
)

type rxdesc struct {
	paddr    uint64
	len      uint16
	checksum uint16
	status   uint8
	errors   uint8
	special  uint16
}

type txdesc struct {
	paddr   uint64
	len     uint16
	cso     uint8
	cmd     uint8
	status  uint8
	css     uint8
	special uint16
}
