package cc1101
import (
	"errors"
	"fmt"
)


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


func (d *Device) SetModulation(modulation string) error {
    switch modulation {
    case "2FSK":
        m2MODFM = 0x00
    case "GFSK":
        m2MODFM = 0x10
    case "OOK":
        m2MODFM = 0x30
    case "4FSK":
        m2MODFM = 0x40
    case "MSK":
        m2MODFM = 0x70
    default:
        return errors.New("Unsupported modulation type")
    }
    
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



// Fichier : config.go

func (d *Device) SetFrequency(frequency float32) error {
    // La formule correcte est : FREQ = (f_target / f_ref) * 2^16
    // f_ref = 26_000_000 Hz
    freq := uint32((frequency * 1_000_000 * (1 << 16)) / 26_000_000)

    freqBytes := []byte{
        byte((freq >> 16) & 0xFF), // FREQ2
        byte((freq >> 8) & 0xFF),  // FREQ1
        byte(freq & 0xFF),         // FREQ0
    }
    err := d.WriteBurstRegister(FREQ2, freqBytes)
    if err != nil {
        return err
    }

    return nil
}


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

func (d *Device) SetTxPower(powerSetting byte) error {
	// CRITIQUE: initialiser TOUS les 8 bytes de la PATABLE
	// Sinon le CC1101 peut utiliser des valeurs indéfinies
	paTable := []byte{
		powerSetting, // Index 0 (utilisé si FREND0 = 0x10)
		0x00,         // Index 1
		0x00,         // Index 2
		0x00,         // Index 3
		0x00,         // Index 4
		0x00,         // Index 5
		0x00,         // Index 6
		0x00,         // Index 7
	}

	err := d.WriteBurstRegister(PATABLE, paTable)
	if err != nil {
		return err
	}
	
	// FREND0 = 0x10 signifie utiliser PATABLE[0]
	return d.WriteSingleRegister(FREND0, 0x10)
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


func (d *Device )GetFrequency() (float32, error){
	freqs := make([]byte,8)
	freqs, err := d.ReadBurstRegister(FREQ2, 3)
	if err != nil {
		return 0.0, err
	}
	freqWord := uint32(freqs[0])<<16 | uint32(freqs[1])<<8 | uint32(freqs[2])
	mhz := float32(freqWord) * (26.0 / 65536.0)

	return mhz, nil

}


