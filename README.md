Testing this driver on an esp32WROOM32D

PINOUT :
| ESP32   | CC1101 | CC1101 PIN NUMBER |
|---------|--------|--------------------|
| GPIO5   | CS     | 4                |
| GPIO18  | SCK    | 5                |
| GPIO19  | MISO   | 7                |
| GPIO23  | MOSI   | 6                |
| 3V3     | VCC    | 2                |
| GND     | GND    | 1                |


compiling : tinygo flash -target=esp32-coreboard-v2 -monitor main.go