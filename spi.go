package cc1101

import (
	"time"
)

func (d *Device) EnableCS() {
	d.cs(false)
}
func (d *Device) DisableCS() {
	d.cs(true)
}


func (d *Device) Reset() error {
	d.EnableCS()
	time.Sleep(10 * time.Microsecond)
	d.DisableCS()
	time.Sleep(40 * time.Microsecond)

	err := d.SpiStrobe(SRES)
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Millisecond)

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

func (d *Device) SpiStrobe(strobe byte) error {
	d.EnableCS()
	for d.miso.Get() != false {
		time.Sleep(1 * time.Microsecond)
	}
	if err := d.bus.Tx([]byte{strobe}, nil); err != nil {
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