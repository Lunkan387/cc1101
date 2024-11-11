package cc1101

import (
	"errors"
	"machine"
	"time"
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
	m2DCOFF, m2SYNCM, m2MANCH, m2MODFM byte
	frend0                             byte
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

// EnableCS sets the CS pin to 0V. While sending data through the bus, the CS pin has to be enabled.
// DisableCS sets the CS pin to a high state, indicating the end of communication.

func (d *Device) EnableCS() {
	d.cs(false)
}
func (d *Device) DisableCS() {
	d.cs(true)
}

func New(bus SPI, cs PinOutput, miso machine.Pin) *Device {
	device := Device{bus: bus, cs: cs, miso: miso}
	return &device
}

func (d *Device) Reset() error {
	d.DisableCS()
	time.Sleep(1 * time.Microsecond)
	d.EnableCS()
	time.Sleep(1 * time.Microsecond)
	for d.miso.Get() != false {
		time.Sleep(1 * time.Microsecond)
	}
	err := d.bus.Tx([]byte{CC1101_SRES}, nil)
	if err != nil {
		d.DisableCS()
		return err
	}
	for d.miso.Get() != false {
		time.Sleep(1 * time.Microsecond)
	}
	d.DisableCS()
	return nil
}

func (d *Device) ReadSingleRegister(addr byte) (byte, error) {
	var temp = addr | CC1101_READSINGLE
	var readBuffer = []byte{0x00}
	var writeBuffer = []byte{temp}

	d.EnableCS()
	for d.miso.Get() != false {
		time.Sleep(1 * time.Microsecond)
	}
	if err := d.bus.Tx(writeBuffer, nil); err != nil {
		d.DisableCS()
		return 0, err
	}
	if err := d.bus.Tx([]byte{0x00}, readBuffer); err != nil {
		d.DisableCS()
		return 0, err
	}
	d.DisableCS()
	return readBuffer[0], nil
}

func (d *Device) ReadBurstRegister(addr byte, length int) ([]byte, error) {
	var temp = addr | CC1101_READBURST
	data := make([]byte, length)
	d.EnableCS()
	for d.miso.Get() != false {
		time.Sleep(1 * time.Microsecond)
	}
	if err := d.bus.Tx([]byte{temp}, nil); err != nil {
		d.DisableCS()
		return nil, err
	}
	if err := d.bus.Tx(make([]byte, length), data); err != nil {
		d.DisableCS()
		return nil, err
	}
	d.DisableCS()

	return data, nil
}

func (d *Device) WriteSingleRegister(addr, value byte) error {
	d.EnableCS()
	for d.miso.Get() != false {
		time.Sleep(1 * time.Microsecond)
	}
	if err := d.bus.Tx([]byte{addr}, nil); err != nil {
		d.DisableCS()
		return err
	}
	if err := d.bus.Tx([]byte{value}, nil); err != nil {
		d.DisableCS()
		return err
	}
	d.DisableCS()
	return nil
}

func (d *Device) WriteBurstRegister(addr byte, data []byte) error {
	temp := addr | CC1101_WRITEBURST
	d.EnableCS()
	for d.miso.Get() != false {
		time.Sleep(1 * time.Microsecond)
	}
	if err := d.bus.Tx([]byte{temp}, nil); err != nil {
		d.DisableCS()
		return err
	}
	for _, byteData := range data {
		if err := d.bus.Tx([]byte{byteData}, nil); err != nil {
			d.DisableCS()
			return err
		}
	}
	d.DisableCS()
	return nil
}

func (d *Device) IsConnected() bool {
	d.EnableCS()
	state, _ := d.ReadSingleRegister(0x31)
	if state > 0 {
		return true
	}
	return false
}

func (d *Device) setCCMode(state bool) {
	StateCCMode = state
	if StateCCMode == true {
		d.WriteSingleRegister(CC1101_IOCFG2, 0x0b)
		d.WriteSingleRegister(CC1101_IOCFG0, 0x06)
		d.WriteSingleRegister(CC1101_PKTCTRL0, 0x05)
		d.WriteSingleRegister(CC1101_MDMCFG3, 0xF8)
		d.WriteSingleRegister(CC1101_MDMCFG4, 11+m4RxBw)
	} else {
		d.WriteSingleRegister(CC1101_IOCFG2, 0x0D)
		d.WriteSingleRegister(CC1101_IOCFG0, 0x0D)
		d.WriteSingleRegister(CC1101_PKTCTRL0, 0x32)
		d.WriteSingleRegister(CC1101_MDMCFG3, 0x93)
		d.WriteSingleRegister(CC1101_MDMCFG4, 7+m4RxBw)
	}
}

func (d *Device) SetModulation(modulation string) error {
	switch modulation {
	case "2FSK":
		m2MODFM = 0x00
		frend0 = 0x10
	case "GFSK":
		m2MODFM = 0x10
		frend0 = 0x10
	case "OOK":
		m2MODFM = 0x30
		frend0 = 0x11
	case "4FSK":
		m2MODFM = 0x40
		frend0 = 0x10
	case "MSK":
		m2MODFM = 0x70
		frend0 = 0x10
	default:
		return errors.New("Unsupported modulation type, please use 2FSK,GFSK,OOK,4FSK,MSK ")
	}
	d.WriteSingleRegister(CC1101_MDMCFG2, m2DCOFF+m2MODFM+m2MANCH+m2SYNCM)
	d.WriteSingleRegister(CC1101_FREND0, frend0)

	return nil
}
