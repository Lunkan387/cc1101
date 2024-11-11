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

var (
	SPI2_SCK_PIN  = SCK
	SPI2_MISO_PIN = MISO
	SPI2_MOSI_PIN = MOSI
	SPI2_CS_PIN   = CS
)

func main() {
	SPI2_CS_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	spi := machine.SPI2
	err := spi.Configure(machine.SPIConfig{
		Frequency: 1_000_000,
		SCK:       SCK,
		SDO:       MOSI,
		SDI:       MISO,
		Mode:      0,
	})
	if err != nil {
		panic(err)
	}
	cc := cc1101.New(spi, SPI2_CS_PIN.Set, SPI2_MISO_PIN)
	if err := cc.Reset(); err != nil {
		panic(err)
	}
	time.Sleep(1 * time.Second)

	cc.SetModulation("OOK")
	for {
		data, err := cc.ReadSingleRegister(0x12)
		if err != nil {
			fmt.Println("Error while reading 0x12")
		} else {
			fmt.Println(data)
		}
		time.Sleep(1 * time.Second)
	}

}
