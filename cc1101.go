package cc1101

import (
	"machine"
	"time"
)

const (
	CC1101_READSINGLE = 0x80
	CC1101_READBURST  = 0xC0
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

func (d *Device) Reset() {
	d.DisableCS()
	time.Sleep(1 * time.Millisecond)
	d.EnableCS()
	time.Sleep(1 * time.Millisecond)
	for d.miso.Get() != false {
		time.Sleep(1 * time.Microsecond)
	}
	d.bus.Tx([]byte{CC1101_SRES}, nil)
	for d.miso.Get() != false {
		time.Sleep(1 * time.Microsecond)
	}
	d.DisableCS()
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
