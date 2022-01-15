package main

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"time"

	"golang.org/x/image/colornames"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

func GetMuteSyncPort() (string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		return "", fmt.Errorf("No serial ports found")
	} else {
		for _, port := range ports {
			if port.IsUSB {
				if port.VID == "10C4" && port.PID == "EA60" {
					return port.Name, nil
				}
			}
		}
	}

	return "", fmt.Errorf("No MuteSync found")
}

var msport serial.Port

func main() {
	port, err := GetMuteSyncPort()

	if err != nil {
		log.Fatal(err)
	}

	msport, err = serial.Open(port, &serial.Mode{})
	if err != nil {
		log.Fatal(err)
	}

	flutterLights()

	offLights()

	pressbuff := make([]byte, 100)

	for {
		n, err := msport.Read(pressbuff)
		if err != nil {
			log.Fatal(err)
		}

		if n == 0 {
			fmt.Println("EOF")
		}

		for i := 0; i < n; i++ {
			if pressbuff[i] == 51 {
				buttonPressed()
			} else if pressbuff[i] == 52 {
				buttonReleased()
			}
		}

	}
}

// offLights: turns off the lights
func offLights() {
	colorbytes := makeOff()

	msport.Write(colorbytes)
}

// flutterLights: flips between two patterns leaving the latter pattern lit
func flutterLights() {
	colorbytes := makeColors(colornames.Red,
		colornames.Green,
		colornames.Red,
		colornames.Green)

	msport.Write(colorbytes)

	time.Sleep(time.Millisecond * 100)

	colorbytes = makeColors(colornames.Green,
		colornames.Red,
		colornames.Green,
		colornames.Red)

	msport.Write(colorbytes)

	time.Sleep(time.Millisecond * 100)

	return
}

func buttonPressed() {
	flutterLights()
}

func buttonReleased() {
	offLights()
}

// Function to translate named color RGBA to RGB, preserving some A
func rgba2rgb(incolor color.RGBA) []byte {
	var alpha = incolor.A

	return []byte{byte(255 / alpha * incolor.R),
		byte(255 / alpha * incolor.G),
		byte(255 / alpha * incolor.B)}

}

// makeColors: turns four colours into a byte array for the MuteSync
// sending this to the MuteSync will set the colours
func makeColors(first, second, third, fourth color.RGBA) []byte {
	var buff bytes.Buffer
	buff.WriteByte(65)
	buff.Write(rgba2rgb(first))
	buff.Write(rgba2rgb(second))
	buff.Write(rgba2rgb(third))
	buff.Write(rgba2rgb(fourth))
	return buff.Bytes()
}

// makeOff: creates a byte array with all RGB setting to 0 for the MuteSync
// sending this effectively turns the MuteSync LEDs off
func makeOff() []byte {
	var buff bytes.Buffer
	buff.WriteByte(65)
	buff.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	return buff.Bytes()
}
