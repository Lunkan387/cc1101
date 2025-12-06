package main

import (
    "cc1101"
    "fmt"
    "machine"
    "time"
)

// Définition des broches pour l'ESP32
const (
    SCK  = machine.GPIO18
    MISO = machine.GPIO19
    MOSI = machine.GPIO23
    CS   = machine.GPIO5
)

func main() {
    // --- 1. Initialisation du matériel ---
    csPin := machine.Pin(CS)
    csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
    csPin.High() // Mettre CS à l'état haut par défaut (désactivé)

    spi := machine.SPI2
    err := spi.Configure(machine.SPIConfig{
        Frequency: 4_000_000,
        SCK:       machine.Pin(SCK),
        SDO:       machine.Pin(MOSI),
        SDI:       machine.Pin(MISO),
        Mode:      0,
    })
    if err != nil {
        panic(fmt.Sprintf("Erreur de configuration SPI: %v", err))
    }

    // Création de l'instance du pilote CC1101
    cc := cc1101.New(spi, csPin.Set, machine.Pin(MISO))

    fmt.Println("Vérification de la connexion avec le CC1101...")
    if !cc.IsConnected() {
        panic("CC1101 non trouvé. Vérifiez le câblage.")
    }
    fmt.Println("✅ CC1101 détecté et connecté !")


    fmt.Println("Configuration du CC1101 pour le mode paquet OOK...")
    if err := cc.ConfigureOOKPacket(); err != nil {
        panic(fmt.Sprintf("Erreur de configuration OOK: %v", err))
    }

 
    fmt.Println("Réglage de la fréquence à 433.8 MHz...")
    if err := cc.SetFrequency(433.8); err != nil {
        panic(fmt.Sprintf("Erreur de réglage de la fréquence: %v", err))
    }

    fmt.Println("Réglage de la puissance d'émission à +10 dBm...")
    if err := cc.SetTxPower(cc1101.Power_10dBm); err != nil {
        panic(fmt.Sprintf("Erreur de réglage de la puissance: %v", err))
    }

    // Calibration
    fmt.Println("Calibration de la radio...")
    cc.SpiStrobe(cc1101.SCAL)
    time.Sleep(100 * time.Millisecond)

    fmt.Println("========================================")
    fmt.Println("Début de l'émission du message 'Hello #X'...")
    fmt.Println("Surveille 433.8 MHz avec un SDR pour voir les paquets !")
    fmt.Println("========================================")

    packetCounter := 0

    for {
        messageToSend := fmt.Sprintf("Hello #%d", packetCounter)

        fmt.Printf("-> Envoi du paquet #%d : '%s'\n", packetCounter, messageToSend)

        dataToSend := []byte(messageToSend)

        err := cc.SendData(dataToSend)
        if err != nil {
            fmt.Printf("❌ Erreur lors de l'envoi : %v\n", err)
        } else {
            fmt.Println("✅ Paquet envoyé avec succès.")
        }

        packetCounter++
        time.Sleep(1 * time.Second) 
    }
}