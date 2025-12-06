package cc1101

import (
	"fmt"
)
func (d *Device) ConfigureOOKCarrierWave() error {
	if err := d.Configure(); err != nil {
		return err
	}

	// Mode asynchrone, transmission infinie
	d.WriteSingleRegister(PKTCTRL0, 0x32)
	
	// Désactiver le sync word pour carrier wave pur
	d.WriteSingleRegister(MDMCFG2, 0x32) // OOK, no sync

	// GDO0 en serial data output
	d.WriteSingleRegister(IOCFG0, 0x0D)

	return nil
}
func (d *Device) Configure() error {
	// Reset complet
	if err := d.Reset(); err != nil {
		return fmt.Errorf("reset failed: %v", err)
	}

	// Configuration des GPIO
	d.WriteSingleRegister(IOCFG2, 0x29) // GDO2 = chip ready
	d.WriteSingleRegister(IOCFG0, 0x06) // GDO0 = sync word sent/received

	// Configuration du packet handler
	d.WriteSingleRegister(PKTCTRL1, 0x04) // No address check, append status
	d.WriteSingleRegister(PKTCTRL0, 0x32) // Async serial mode, infinite packet length
	d.WriteSingleRegister(PKTLEN, 0xFF)   // Max packet length

	// FIFO thresholds
	d.WriteSingleRegister(FIFOTHR, 0x47) // TX: 33 bytes, RX: 32 bytes

	// Sync word (pour mode test, peut être désactivé)
	d.WriteSingleRegister(SYNC1, 0xD3)
	d.WriteSingleRegister(SYNC0, 0x91)

	// Configuration Modem pour OOK
	// MDMCFG4: Data rate config
	// RX filter bandwidth = 58 kHz
	d.WriteSingleRegister(MDMCFG4, 0xC8) // CHANBW_E=2, CHANBW_M=0, DRATE_E=8

	// MDMCFG3: Data rate config (mantissa)
	d.WriteSingleRegister(MDMCFG3, 0x93) // DRATE_M = 147 (~4.8 kBaud)

	// MDMCFG2: Modem configuration
	// OOK modulation (0x30), No Manchester, 16/16 sync word bits
	d.WriteSingleRegister(MDMCFG2, 0x30)

	// MDMCFG1: Channel spacing and preamble
	d.WriteSingleRegister(MDMCFG1, 0x22) // 4 preamble bytes, CHANSPC_E=2

	// MDMCFG0: Channel spacing (mantissa)
	d.WriteSingleRegister(MDMCFG0, 0xF8)

	// Deviation (important même pour OOK)
	d.WriteSingleRegister(DEVIATN, 0x15) // ~5 kHz deviation

	// Main Radio Control State Machine
	d.WriteSingleRegister(MCSM2, 0x07) // RX_TIME = jusqu'à timeout
	d.WriteSingleRegister(MCSM1, 0x30) // CCA disabled, stay in RX after packet
	d.WriteSingleRegister(MCSM0, 0x18) // Auto calibrate when going from IDLE to RX/TX

	// Frequency Offset Compensation
	d.WriteSingleRegister(FOCCFG, 0x16) // FOC settings

	// Bit synchronization
	d.WriteSingleRegister(BSCFG, 0x6C)

	// AGC Control
	d.WriteSingleRegister(AGCCTRL2, 0x03) // Max DVGA gain, target amplitude
	d.WriteSingleRegister(AGCCTRL1, 0x40) // AGC settings
	d.WriteSingleRegister(AGCCTRL0, 0x91) // AGC filter, wait time

	// Wake on Radio (désactivé pour test)
	d.WriteSingleRegister(WORCTRL, 0xFB)

	// Front End RX/TX Configuration
	d.WriteSingleRegister(FREND1, 0x56) // Front end RX configuration
	d.WriteSingleRegister(FREND0, 0x10) // Front end TX configuration (PATABLE index 0)

	// Frequency Synthesizer Calibration
	d.WriteSingleRegister(FSCAL3, 0xE9)
	d.WriteSingleRegister(FSCAL2, 0x2A)
	d.WriteSingleRegister(FSCAL1, 0x00)
	d.WriteSingleRegister(FSCAL0, 0x1F)

	// RC Oscillator
	d.WriteSingleRegister(RCCTRL1, 0x41)
	d.WriteSingleRegister(RCCTRL0, 0x00)

	// Test settings (valeurs recommandées par TI)
	d.WriteSingleRegister(TEST2, 0x81)
	d.WriteSingleRegister(TEST1, 0x35)
	d.WriteSingleRegister(TEST0, 0x09)

	return nil
}

func (d *Device) ConfigureOOKPacket() error {
    if err := d.Reset(); err != nil {
        return fmt.Errorf("reset failed: %v", err)
    }

    // --- Configuration du gestionnaire de paquets ---
    // PKTCTRL0 = 0x45
    // Bits: WHITENING=1, CRC_EN=1, LENGTH_CONFIG=01 (Variable length)
    // NOTE: Le data whitening est activé ! C'est un point clé.
    d.WriteSingleRegister(PKTCTRL0, 0x05)
    d.WriteSingleRegister(PKTCTRL1, 0x04) // Append status, no address check
    d.WriteSingleRegister(PKTLEN, 0xFF)   // Max packet length

    // --- Configuration du Modem ---
    // MDMSFG4 = 0xF7 (DRATE_E=7)
    // MDMSFG3 = 0x83 (DRATE_M=131)
    // Calcul: (256+131)*2^7 * (26e6/2^28) ≈ 10000 Baud
    d.WriteSingleRegister(MDMCFG4, 0xF7)
    d.WriteSingleRegister(MDMCFG3, 0x83)
    
    // MDMCFG2 = 0x32 -> OOK + 16/16 sync (identique à avant)
    d.WriteSingleRegister(MDMCFG2, 0x32)

    // --- Préambule et Sync Word ---
    // Le préambule est de 64 bits, comme dans le code Arduino.
    d.WriteSingleRegister(MDMCFG1, 0x22) // 4 preamble bytes
    d.WriteSingleRegister(SYNC1, 0x12)
    d.WriteSingleRegister(SYNC0, 0x34)

    // Les autres registres sont standards et peuvent rester les mêmes
    d.WriteSingleRegister(MDMCFG0, 0xF8)
    d.WriteSingleRegister(DEVIATN, 0x15)
    d.WriteSingleRegister(MCSM2, 0x07)
    d.WriteSingleRegister(MCSM1, 0x30)
    d.WriteSingleRegister(MCSM0, 0x18)
    d.WriteSingleRegister(FOCCFG, 0x16)
    d.WriteSingleRegister(BSCFG, 0x6C)
    d.WriteSingleRegister(AGCCTRL2, 0x03)
    d.WriteSingleRegister(AGCCTRL1, 0x40)
    d.WriteSingleRegister(AGCCTRL0, 0x91)
    d.WriteSingleRegister(WORCTRL, 0xFB)
    d.WriteSingleRegister(FREND1, 0x56)
    d.WriteSingleRegister(FREND0, 0x10)
    d.WriteSingleRegister(FSCAL3, 0xE9)
    d.WriteSingleRegister(FSCAL2, 0x2A)
    d.WriteSingleRegister(FSCAL1, 0x00)
    d.WriteSingleRegister(FSCAL0, 0x1F)
    d.WriteSingleRegister(RCCTRL1, 0x41)
    d.WriteSingleRegister(RCCTRL0, 0x00)
    d.WriteSingleRegister(TEST2, 0x81)
    d.WriteSingleRegister(TEST1, 0x35)
    d.WriteSingleRegister(TEST0, 0x09)

    return nil
}