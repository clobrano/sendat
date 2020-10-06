package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

func main() {
	portName := flag.String("port", "/dev/ttyUSB0", "serial port to test (/dev/ttyUSB0, etc)")
	portName = flag.String("p", "/dev/ttyUSB0", "serial port to test (/dev/ttyUSB0, etc) (shorthand)")
	baud := flag.Uint("baud", 115200, "Baud rate")
	stopbits := flag.Uint("stopbits", 1, "Stop bits")
	databits := flag.Uint("databits", 8, "Data bits")
	command := flag.String("command", "", "AT command to be sent (e.g. at+cfun?)")
	command = flag.String("c", "", "AT command to be sent (e.g. at+cfun?) (shorthand)")
	expect := flag.String("expect", "OK", "AT reply to look for before exiting")
	expect = flag.String("e", "OK", "AT reply to look for before exiting (shorthand)")
	timeout := flag.Uint("timeout", 3, "Reply timeout")
	timeout = flag.Uint("t", 3, "Reply timeout (shorthand)")

	flag.Parse()

	options := serial.OpenOptions{
		PortName:        *portName,
		BaudRate:        *baud,
		DataBits:        *databits,
		StopBits:        *stopbits,
		MinimumReadSize: 4,
	}

	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	defer port.Close()

	if !strings.HasPrefix(*command, "at") && !strings.HasPrefix(*command, "AT") {
		*command = "AT" + *command
	}

	// Write 4 bytes to the port.
	b := []byte(*command + "\r\n")
	_, err = port.Write(b)
	if err != nil {
		log.Fatalf("port.Write: %v", err)
	}

	fmt.Println(">>> ", *command)

	r := make(chan string)
	go func() {
		for {
			buf := make([]byte, 32)
			n, err := port.Read(buf)
			if err != nil {
				fmt.Println(err)
				if err != io.EOF {
					r <- fmt.Sprintf("Error reading from serial port: %s", err)
				}
				r <- fmt.Sprintf("end of stream")
			}

			fmt.Println(string(buf[:n]))
			if strings.Contains(string(buf[:n]), *expect) {
				r <- ""
			}
		}
	}()

	select {
	case res := <-r:
		if res != "" {
			fmt.Println(r)
		}
	case <-time.After(time.Duration(*timeout) * time.Second):
		fmt.Println("Timeout expired")
	}
}
