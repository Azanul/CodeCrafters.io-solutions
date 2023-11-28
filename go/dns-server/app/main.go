package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

func main() {
	forwardAddress := flag.String("resolver", "", "address to forward questions")
	flag.Parse()

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:2053")
	if err != nil {
		fmt.Println("Failed to resolve UDP address:", err)
		return
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Failed to bind to address:", err)
		return
	}
	defer udpConn.Close()

	var forwardAddr *net.UDPAddr
	var forwardBuf []byte
	if forwardAddress != nil {
		forwardAddr, err = net.ResolveUDPAddr("udp", *forwardAddress)
		if err != nil {
			fmt.Println("Failed to resolve forward UDP address:", err)
			return
		}
		forwardBuf = make([]byte, 512)
	}

	buf := make([]byte, 512)

	for {
		size, source, err := udpConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			break
		}

		receivedData := buf[:size]
		fmt.Printf("Received %d bytes from %s\n", size, source)
		query := UnmarshalMessage(receivedData)
		if query.header.opcode == 0 {
			if forwardAddress == nil {
				for i := range query.questions {
					query.AddAnswer(
						uint8(i),
						60,
						[]byte{8, 8, 8, 8},
					)
				}
			} else {
				allQuestions := query.questions
				allAnswers := []answer{}
				for _, q := range allQuestions {
					query.questions = []question{q}
					dataWithSingleQuestion := query.MarshalMessage()
					_, err = udpConn.WriteToUDP(dataWithSingleQuestion, forwardAddr)
					if err != nil {
						log.Panic("Failed to forward request:", err)
					}

					size, _, err := udpConn.ReadFromUDP(forwardBuf)
					if err != nil {
						fmt.Println("Error receiving data from foward:", err)
						return
					}
					forwardResp := UnmarshalMessage(forwardBuf[:size])
					allAnswers = append(allAnswers, forwardResp.answers[0])
				}
				query.questions = allQuestions
				for i, a := range allAnswers {
					query.AddAnswer(
						uint8(i),
						a.ttl,
						a.data,
					)
				}
			}
		}
		query.ToggleQR()
		reply := query.MarshalMessage()
		_, err = udpConn.WriteToUDP(reply, source)
		if err != nil {
			fmt.Println("Failed to send response:", err)
		}
	}
}
