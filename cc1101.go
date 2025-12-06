package cc1101

import (
		"machine"
)

const (
	CC1101_READSINGLE = 0x80
	CC1101_READBURST  = 0xC0
	CC1101_WRITEBURST = 0x40
)

var (
	StateCCMode bool
	m4RxBw      byte

	// Config for 0x12: MDMCFG2
	// | DCOFF   | MODFM    | MANCH   | SYNCM    |
	// | 7th bit | 6-4 bits | 3rd bit | 2-0 bits |
	// Read page 77 https://www.ti.com/lit/ds/symlink/cc1101.pdf
	//
	m2DCOFF, m2MANCH, m2MODFM byte
	frend0                    byte
	m2SYNCM                   byte = 0x02
	// Set Frequency vars
	freq0, freq1, freq2 byte
	mhz                 float32
	marcstate           byte
)

type SPI interface {
	Tx(writeBuffer, readBuffer []byte) error
}

type PinOutput func(state bool)

type Device struct {
	bus  SPI
	cs   PinOutput
	miso machine.Pin
}
func New(bus SPI, cs PinOutput, miso machine.Pin) *Device {
	device := Device{bus: bus, cs: cs, miso: miso}
	return &device
}

func (d *Device) IsConnected() bool {
	d.EnableCS()
	state, _ := d.ReadSingleRegister(0x31)
	if state > 0 {
		return true
	}
	return false
}


// Marcstate register addr : 0xF5
//	Value for each states in marcstate :
//	0x00	SLEEP
//	0x01	IDLE
//	0x02	XOFF
//	0x03	VCOON_MC
//	0x04	REGON_MC
//	0x05	MANCAL
//	0x06	VCOON
//	0x07	REGON
//	0x08	STARTCAL
//	0x09	BWBOOST
//	0x0A	FS_LOCK
//	0x0B	IFADCON
//	0x0C	ENDCAL
//	0x0D	RX
//	0x0E	RX_END
//	0x0F	RX_RST
//	0x10	TXRX_SWITCH
//	0x11	RXFIFO_OVERFLOW
//	0x12	FSTXON
//	0x13	TX
//	0x14	TX_END
//	0x15	RXTX_SWITCH
//	0x16	TXFIFO_UNDERFLOW
//
//	Strobe commands :
//  SRES    = 0x30 // Reset chip
//  SFSTXON = 0x31 // Enable/calibrate freq synthesizer
//  SXOFF   = 0x32 // Turn off crystal oscillator.
//  SCAL    = 0x33 // Calibrate freq synthesizer & disable
//  SRX     = 0x34 // Enable RX.
//  STX     = 0x35 // Enable TX.
//  SIDLE   = 0x36 // Exit RX / TX
//  SAFC    = 0x37 // AFC adjustment of freq synthesizer
//  SWOR    = 0x38 // Start automatic RX polling sequence
//  SPWD    = 0x39 // Enter pwr down mode when CSn goes hi
//  SFRX    = 0x3A // Flush the RX FIFO buffer.
//  SFTX    = 0x3B // Flush the TX FIFO buffer.
//  SWORRST = 0x3C // Reset real time clock.
//  SNOP    = 0x3D // No operation.
