package main

import (
	"encoding/binary"
)

type header struct {
	id      uint16
	qr      bool
	opcode  uint8
	aa      bool
	tc      bool
	rd      bool
	ra      bool
	z       uint8
	rcode   uint8
	qdcount uint16
	ancount uint16
	nscount uint16
	arcount uint16
}

func (h header) encodeFlags() uint16 {
	var flags uint16 = 0
	if h.qr {
		flags |= 1 << 15
	}
	if h.opcode != 0 {
		flags |= uint16(h.opcode) << 11
	}
	if h.aa {
		flags |= 1 << 10
	}
	if h.tc {
		flags |= 1 << 9
	}
	if h.rd {
		flags |= 1 << 8
	}
	if h.ra {
		flags |= 1 << 7
	}
	if h.z != 0 {
		flags |= uint16(h.z) << 4
	}
	if h.opcode == 0 {
		flags |= uint16(0)
	} else {
		flags |= uint16(4)
	}

	return flags
}

type label struct {
	length  uint8
	content []byte
}

func (l *label) encode() []byte {
	return append([]byte{byte(l.length)}, l.content...)
}

type question struct {
	qname  []label
	qtype  uint16
	qclass uint16
}

func (q *question) encode() []byte {
	out := []byte{}
	for _, dn := range q.qname {
		out = append(out, dn.encode()...)
	}
	out = append(out, 0)
	out = binary.BigEndian.AppendUint16(out, q.qtype)
	out = binary.BigEndian.AppendUint16(out, q.qclass)

	return out
}

type answer struct {
	aname  []label
	atype  uint16
	aclass uint16
	ttl    uint32
	length uint16
	data   []byte
}

func (a *answer) encode() []byte {
	out := []byte{}

	for _, dn := range a.aname {
		out = append(out, dn.encode()...)
	}
	out = append(out, 0)
	out = binary.BigEndian.AppendUint16(out, a.atype)
	out = binary.BigEndian.AppendUint16(out, a.aclass)
	out = binary.BigEndian.AppendUint32(out, a.ttl)
	out = binary.BigEndian.AppendUint16(out, a.length)
	out = append(out, a.data...)

	return out
}

type Message struct {
	header    header
	questions []question
	answers   []answer
}

func (msg *Message) ToggleQR() {
	msg.header.qr = !msg.header.qr
}

func (msg *Message) AddAnswer(i uint8, ttl uint32, data []byte) {
	msg.answers = append(msg.answers, answer{
		msg.questions[i].qname,
		msg.questions[i].qtype,
		msg.questions[i].qclass,
		ttl,
		uint16(len(data)),
		data,
	})
	msg.header.ancount++
}

func (msg *Message) MarshalMessage() []byte {
	out := []byte{}
	out = binary.BigEndian.AppendUint16(out, msg.header.id)
	out = binary.BigEndian.AppendUint16(out, msg.header.encodeFlags())
	out = binary.BigEndian.AppendUint16(out, uint16(len(msg.questions)))
	out = binary.BigEndian.AppendUint16(out, uint16(len(msg.answers)))
	out = binary.BigEndian.AppendUint16(out, msg.header.nscount)
	out = binary.BigEndian.AppendUint16(out, msg.header.arcount)

	for _, q := range msg.questions {
		out = append(out, q.encode()...)
	}

	for _, a := range msg.answers {
		out = append(out, a.encode()...)
	}

	return out
}

func decodeDomainName(b []byte, i uint16) ([]label, uint16) {
	domainName := []label{}
	for b[i] != 0 {
		domainName = append(domainName, label{
			uint8(b[i]),
			b[i+1 : i+1+uint16(uint8(b[i]))],
		})
		i += 1 + uint16(uint8(b[i]))
		if b[i]&0xc0 == 0xc0 {
			p := binary.BigEndian.Uint16(b[i:i+2]) - 0xc000
			pointedDomainName, _ := decodeDomainName(b, p)
			domainName = append(domainName, pointedDomainName...)
			i += 1
			break
		}
	}
	return domainName, i
}

func UnmarshalMessage(b []byte) Message {
	m := Message{
		header: header{
			binary.BigEndian.Uint16(b[:2]),
			b[2]>>7&1 == 1,
			b[2] >> 3 & 0xF,
			b[2]>>2&1 == 1,
			b[2]>>1&1 == 1,
			b[2]&1 == 1,
			b[3]>>7&1 == 1,
			b[3] >> 4 & 7,
			b[3] & 0xF,
			binary.BigEndian.Uint16(b[4:6]),
			binary.BigEndian.Uint16(b[6:8]),
			binary.BigEndian.Uint16(b[8:10]),
			binary.BigEndian.Uint16(b[10:12]),
		},
		questions: []question{},
	}

	var i uint16 = 12
	for len(m.questions) < int(m.header.qdcount) {
		domainName, d := decodeDomainName(b, i)
		i = d
		m.questions = append(m.questions, question{
			domainName,
			binary.BigEndian.Uint16(b[i+1 : i+3]),
			binary.BigEndian.Uint16(b[i+3 : i+5]),
		})
		i += 5
	}

	for len(m.answers) < int(m.header.ancount) {
		domainName, d := decodeDomainName(b, i)
		i = d
		rdLength := binary.BigEndian.Uint16(b[i+9 : i+11])
		m.answers = append(m.answers, answer{
			domainName,
			binary.BigEndian.Uint16(b[i+1 : i+3]),
			binary.BigEndian.Uint16(b[i+3 : i+5]),
			binary.BigEndian.Uint32(b[i+5 : i+9]),
			rdLength,
			b[i+11 : i+11+rdLength],
		})
		i += 11 + rdLength
	}

	return m
}
