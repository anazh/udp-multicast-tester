package main

import (
	"fmt"
	"flag"
	"net"
	"encoding/binary"
)

const (
	UINT64MAX = 18446744073709551615
)

func sendUdpFlood(multicastAddr net.UDPAddr) {
	// we want to mimic UDP MPEG traffic
	buf := make([]byte, 188)
	conn, err := net.ListenUDP("udp4", &multicastAddr)

	if nil != err {
		panic(err)
	}

	var counter uint64;
	for counter = 0; counter < UINT64MAX; counter++ {
		binary.LittleEndian.PutUint64(buf, counter)
		conn.WriteToUDP(buf, &multicastAddr)
	}
}

func receiveStream(chanReceive chan uint64, multicastAddr net.UDPAddr) {
	buf := make([]byte, 188)
	conn, err := net.ListenMulticastUDP("udp4", nil, &multicastAddr)
	if nil != err {
		panic(err)
	}

	var counter uint64

	for {
		conn.ReadFromUDP(buf)
		counter = binary.LittleEndian.Uint64(buf[0:8])
		chanReceive <- counter
	}
}

func processChannel(chanReceive chan uint64) {
	var diff, receiveCounter uint64
	var displayCounter int
	var result string
	for counter := range chanReceive {
		receiveCounter++
		displayCounter++

		if (displayCounter == 10000) {
			if (counter > receiveCounter) {
				diff = counter - receiveCounter
				result = fmt.Sprintf("More sent than received. diff=%d", diff)

			} else if counter < receiveCounter {
				diff = receiveCounter - counter
				result = fmt.Sprintf("More received than sent?!. diff=%d", diff)

			} else {
				diff = 0
				result = "We are in sync"
			}
			displayCounter = 0
			fmt.Printf("%s\n", result)
		}
	}
}

func main() {

	fmt.Println("UDP Broadcast Tester")

	var server, client bool
	var multicastIp string
	var multicastPort int
	flag.BoolVar(&server, "server", false, "Set this instance up as a server")
	flag.BoolVar(&client, "client", false, "Set this instance up as a client")
	flag.StringVar(&multicastIp, "multicast-ip", "224.0.1.50", "The Multicast IP to broadcast to")
	flag.IntVar(&multicastPort, "multicast-port", 12345, "The Multicast Port to broadcast on")

	flag.Parse()

	var addr = net.UDPAddr{IP: net.ParseIP(multicastIp), Port: multicastPort}
	if (server) {
		fmt.Printf("Sending UDP Multicast on: %s:%d\n", multicastIp, multicastPort);
		sendUdpFlood(addr)
	} else if (client) {
		fmt.Printf("Listening for UDP Multicast on: %s:%d\n", multicastIp, multicastPort);

		chanReceive := make(chan uint64, 1000)
		go receiveStream(chanReceive, addr)
		go processChannel(chanReceive)
		select {}

		// setup a received message channel.  make it buffer a decent number of messages (1000?).
		// Thread 1: put all received messages in receive channel.
		// Thread 2: mark all elements for deletion and increment the expecting num when we find a packet
		//
	} else {
		// no options - show usage
		flag.Usage()
	}
}
