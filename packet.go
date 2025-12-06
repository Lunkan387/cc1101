// Fichier : packet.go
package cc1101

import (
    "fmt"
    "time"
)


func (d *Device) SendData(packet []byte) error {

    if len(packet) > FIFOBUFFER {
        return fmt.Errorf("packet too long: %d bytes (max %d)", len(packet), FIFOBUFFER)
    }

    d.SpiStrobe(SIDLE)
    d.SpiStrobe(SFTX)

    fifoPayload := make([]byte, 1+len(packet))
    fifoPayload[0] = byte(len(packet))
    copy(fifoPayload[1:], packet)

    err := d.WriteBurstRegister(TXFIFO_SINGLE_BYTE, fifoPayload)
    if err != nil {
        return fmt.Errorf("failed to write to TX FIFO: %w", err)
    }


    d.SpiStrobe(STX)

    for {
        state, err := d.ReadSingleRegister(MARCSTATE)
        if err != nil {
            return fmt.Errorf("failed to read MARCSTATE: %w", err)
        }

        currentState := state & MARCSTATE_MASK

        if currentState != MARCSTATE_TX && currentState != MARCSTATE_TX_END {
            break
        }

        time.Sleep(1 * time.Millisecond)
    }

    return nil
}