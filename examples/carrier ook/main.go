package main

import (
	"cc1101"
	"fmt"
	"machine"
	"time"
)

const (
	SCK  = machine.GPIO18
	MISO = machine.GPIO19
	MOSI = machine.GPIO23
	CS   = machine.GPIO5
)

func main() {
	// Configuration des pins
	csPin := machine.Pin(CS)
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Configuration SPI - AUGMENTE LA VITESSE !
	spi := machine.SPI2
	err := spi.Configure(machine.SPIConfig{
		Frequency: 4_000_000, 
		SCK:       machine.Pin(SCK),
		SDO:       machine.Pin(MOSI),
		SDI:       machine.Pin(MISO),
		Mode:      0,
	})
	if err != nil {
		panic(err)
	}

	// Création du device
	cc := cc1101.New(spi, csPin.Set, machine.Pin(MISO))

	fmt.Println("Reset du CC1101...")
	if err := cc.Reset(); err != nil {
		panic(err)
	}
	time.Sleep(100 * time.Millisecond)

	// Vérification de la connexion
	version, _ := cc.ReadSingleRegister(cc1101.VERSION)
	partnum, _ := cc.ReadSingleRegister(cc1101.PARTNUM)
	fmt.Printf("CC1101 détecté - Part: 0x%02X, Version: 0x%02X\n", partnum, version)

	// Configuration complète OOK
	fmt.Println("Configuration OOK...")
	if err := cc.ConfigureOOKCarrierWave(); err != nil {
		panic(err)
	}

	// Réglage de la fréquence
	fmt.Println("Réglage fréquence 433.92 MHz...")
	if err := cc.SetFrequency(433.92); err != nil {
		panic(err)
	}

	// Vérification fréquence
	freq, _ := cc.GetFrequency()
	fmt.Printf("Fréquence configurée: %.2f MHz\n", freq)

	// Puissance max pour test
	fmt.Println("Réglage puissance TX à +10 dBm...")
	if err := cc.SetTxPower(cc1101.Power_10dBm); err != nil {
		panic(err)
	}

	// Calibration
	fmt.Println("Calibration...")
	cc.SpiStrobe(cc1101.SCAL)
	time.Sleep(100 * time.Millisecond)

	// Passage en TX
	fmt.Println("=== ÉMISSION CONTINUE (Carrier Wave) ===")
	fmt.Println("Surveille 433.92 MHz sur ton HackRF...")
	cc.SetTx()

	// Surveillance de l'état
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			state, _ := cc.ReadSingleRegister(cc1101.MARCSTATE)
			marcState := state & cc1101.MARCSTATE_MASK
			
			stateStr := getStateString(marcState)
			fmt.Printf("État: %s (0x%02X) | ", stateStr, marcState)
			
			txBytes, _ := cc.ReadSingleRegister(cc1101.TXBYTES)
			fmt.Printf("TXBYTES: %d\n", txBytes&0x7F)

			if marcState != cc1101.MARCSTATE_TX {
				fmt.Println("⚠️  Pas en TX, relance...")
				cc.SpiStrobe(cc1101.SFTX) // Flush TX FIFO
				time.Sleep(10 * time.Millisecond)
				cc.SetTx()
			}
		}
	}
}

func getStateString(state byte) string {
	switch state {
	case cc1101.MARCSTATE_SLEEP:
		return "SLEEP"
	case cc1101.MARCSTATE_IDLE:
		return "IDLE"
	case cc1101.MARCSTATE_FSTXON:
		return "FSTXON"
	case cc1101.MARCSTATE_TX:
		return "TX ✓"
	case cc1101.MARCSTATE_TX_END:
		return "TX_END"
	case cc1101.MARCSTATE_RX:
		return "RX"
	case cc1101.MARCSTATE_TX_UNDERFLOW:
		return "TX_UNDERFLOW ⚠"
	default:
		return fmt.Sprintf("UNKNOWN_%d", state)
	}
}