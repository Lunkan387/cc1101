package cc1101

import (
	"errors"
	"fmt"
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
		d.WriteSingleRegister(IOCFG2, 0x0b)
		d.WriteSingleRegister(IOCFG0, 0x06)
		d.WriteSingleRegister(PKTCTRL0, 0x05)
		d.WriteSingleRegister(MDMCFG3, 0xF8)
		d.WriteSingleRegister(MDMCFG4, 11+m4RxBw)
	} else {
		d.WriteSingleRegister(IOCFG2, 0x0D)
		d.WriteSingleRegister(IOCFG0, 0x0D)
		d.WriteSingleRegister(PKTCTRL0, 0x32)
		d.WriteSingleRegister(MDMCFG3, 0x93)
		d.WriteSingleRegister(MDMCFG4, 7+m4RxBw)
	}
}

// The mask ensures that only the relevant bits are modified in the register
// while leaving the others unchanged.
// The following mask operations isolate specific bits for each field:
// - (m2DCOFF & 0x80) keeps only bit 7 for DCOFF.
// - (m2MODFM & 0x70) keeps bits 6-4 for MODFM (modulation type).
// - (m2MANCH & 0x08) keeps bit 3 for Manchester encoding.
// - (m2SYNCM & 0x07) keeps bits 2-0 for SYNC mode.
//
// These are combined using the OR operator (|) to form the final register value.

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
	registerValue := (m2DCOFF & 0x80) | (m2MODFM & 0x70) | (m2MANCH & 0x08) | (m2SYNCM & 0x07)
	err := d.WriteSingleRegister(MDMCFG2, registerValue)
	if err != nil {
		return fmt.Errorf("Error writing in the register : %v", err)
	}
	d.WriteSingleRegister(FREND0, frend0)

	return nil
}

func (d *Device) SetSYNC_MODE(choice int) error {
	switch choice {
	case 0:
		m2SYNCM = 0x00 // Aucune synchronisation (pas de préambule/sync)
	case 1:
		m2SYNCM = 0x01 // 15/16 bits du mot de synchronisation détectés
	case 2:
		m2SYNCM = 0x02 // 16/16 bits du mot de synchronisation détectés
	case 3:
		m2SYNCM = 0x03 // 30/32 bits du mot de synchronisation détectés
	case 4:
		m2SYNCM = 0x04 // Aucune synchronisation avec détection du seuil de porteuse
	case 5:
		m2SYNCM = 0x05 // 15/16 bits du mot de synchronisation + détection du seuil de porteuse
	case 6:
		m2SYNCM = 0x06 // 16/16 bits du mot de synchronisation + détection du seuil de porteuse
	case 7:
		m2SYNCM = 0x07 // 30/32 bits du mot de synchronisation + détection du seuil de porteuse
	default:
		return fmt.Errorf("invalid SYNC_MODE choice: %d", choice)
	}
	registerValue := (m2DCOFF & 0x80) | (m2MODFM & 0x70) | (m2MANCH & 0x08) | (m2SYNCM & 0x07)
	err := d.WriteSingleRegister(MDMCFG2, registerValue)
	if err != nil {
		return fmt.Errorf("Error writing in the register : %v", err)
	}
	return nil
}

func (d *Device) EnableManchester() error {
	m2MANCH = 0x08
	registerValue := (m2DCOFF & 0x80) | (m2MODFM & 0x70) | (m2MANCH & 0x08) | (m2SYNCM & 0x07)
	err := d.WriteSingleRegister(MDMCFG2, registerValue)
	if err != nil {
		return fmt.Errorf("Error writing in the register : %v", err)
	}
	return nil
}

func (d *Device) DisableManchester() error {
	m2MANCH = 0x00
	registerValue := (m2DCOFF & 0x80) | (m2MODFM & 0x70) | (m2MANCH & 0x08) | (m2SYNCM & 0x07)
	err := d.WriteSingleRegister(MDMCFG2, registerValue)
	if err != nil {
		return fmt.Errorf("Error writing in the register : %v", err)
	}
	return nil
}

func (d *Device) EnableDCFilter() error {
	m2DCOFF = 0x80
	registerValue := (m2DCOFF & 0x80) | (m2MODFM & 0x70) | (m2MANCH & 0x08) | (m2SYNCM & 0x07)
	err := d.WriteSingleRegister(MDMCFG2, registerValue)
	if err != nil {
		return fmt.Errorf("Error writing in the register : %v", err)
	}
	return nil
}

func (d *Device) DisableDCFilter() error {
	m2DCOFF = 0x00
	registerValue := (m2DCOFF & 0x80) | (m2MODFM & 0x70) | (m2MANCH & 0x08) | (m2SYNCM & 0x07)
	err := d.WriteSingleRegister(MDMCFG2, registerValue)
	if err != nil {
		return fmt.Errorf("Error writing in the register : %v", err)
	}
	return nil
}

// Example : 433.92 mhz
// [16 176 113]
// freq2 = 16  | 26
// freq1 = 176 | 0.1015625
// freq0 = 113 | 0.00039675
// Calcul = 16 * 26 + 176 * 0.1015625  + 113 * 0.00039675 = 433.91983275

func (d *Device) SetFrequency(frequency float32) error {
	// Convertir la fréquence en valeurs pour les registres FREQ2, FREQ1 et FREQ0
	freq := uint32((frequency * 1_000_000) / 26_000_000 * (1 << 16))

	freq2 := byte((freq >> 16) & 0xFF)
	freq1 := byte((freq >> 8) & 0xFF)
	freq0 := byte(freq & 0xFF)

	// Écrire dans les registres du CC1101
	err := d.WriteSingleRegister(FREQ2, freq2)
	if err != nil {
		return err
	}
	err = d.WriteSingleRegister(FREQ1, freq1)
	if err != nil {
		return err
	}
	err = d.WriteSingleRegister(FREQ0, freq0)
	if err != nil {
		return err
	}

	return nil
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

func (d *Device) SetRx() {
	d.SpiStrobe(SRX)
	for marcstate != MARCSTATE_RX {
		marcstate, _ = d.ReadSingleRegister(MARCSTATE)
	}
}

func (d *Device) SetTx() {
	d.SpiStrobe(STX)
	for marcstate != MARCSTATE_TX {
		marcstate, _ = d.ReadSingleRegister(MARCSTATE)
	}
}
