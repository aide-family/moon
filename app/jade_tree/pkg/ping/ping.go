// Package ping implements ICMP echo (ping) over IPv4 using raw sockets.
package ping

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

// ICMP echo types (IPv4).
const (
	TypeEchoRequest = 8
	TypeEchoReply   = 0
)

var (
	// ErrTruncated indicates the buffer is too short for the claimed IPv4/ICMP layout.
	ErrTruncated = errors.New("ping: packet truncated")
	// ErrNotIPv4 indicates the buffer does not start with an IPv4 header.
	ErrNotIPv4 = errors.New("ping: not IPv4")
)

// Reply is one successful echo reply.
type Reply struct {
	Seq        uint16
	RTT        time.Duration
	TTL        uint8
	PayloadLen int
}

// Pinger sends ICMP echo requests to a host.
type Pinger struct {
	// Count is the number of echo requests. If zero, defaults to 4.
	Count int
	// Timeout is the read deadline for each echo exchange. If zero, defaults to 5s.
	Timeout time.Duration
	// ID is the ICMP identifier in host byte order. If zero, defaults to low 16 bits of PID.
	ID uint16
	// Payload is echoed by the peer. If nil, defaults to "HelloPing".
	Payload []byte
}

// Ping resolves host, opens an IPv4 ICMP socket, sends Count echo requests, and collects replies.
// It requires elevated privileges (e.g. root) on Unix-like systems.
func (p *Pinger) Ping(ctx context.Context, host string) ([]Reply, error) {
	count := p.Count
	if count <= 0 {
		count = 4
	}
	timeout := p.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	id := p.ID
	if id == 0 {
		id = uint16(os.Getpid() & 0xffff)
	}
	payload := p.Payload
	if payload == nil {
		payload = []byte("HelloPing")
	}

	dst, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve host: %w", err)
	}
	if dst.IP.To4() == nil {
		return nil, fmt.Errorf("resolve host: need IPv4 address")
	}

	conn, err := net.DialIP("ip4:icmp", nil, dst)
	if err != nil {
		return nil, fmt.Errorf("dial icmp: %w", err)
	}
	defer conn.Close()

	var out []Reply
	for seq := uint16(1); seq <= uint16(count); seq++ {
		if err := ctx.Err(); err != nil {
			return out, err
		}

		packet := EncodeEchoRequest(id, seq, payload)
		start := time.Now()
		if _, err := conn.Write(packet); err != nil {
			continue
		}

		deadline := time.Now().Add(timeout)
		for time.Now().Before(deadline) {
			if err := ctx.Err(); err != nil {
				return out, err
			}
			_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			buf := make([]byte, 1500)
			n, rerr := conn.Read(buf)
			if rerr != nil {
				if ne, ok := rerr.(net.Error); ok && ne.Timeout() {
					continue
				}
				break
			}
			elapsed := time.Since(start)
			matched, plen, ttl, perr := ParseIPv4EchoReply(buf[:n], id, seq)
			if perr != nil {
				continue
			}
			if matched {
				out = append(out, Reply{
					Seq:        seq,
					RTT:        elapsed,
					TTL:        ttl,
					PayloadLen: plen,
				})
				break
			}
		}
	}
	return out, nil
}

// EncodeEchoRequest builds an ICMP echo request (type 8, code 0) with checksum set.
func EncodeEchoRequest(id, seq uint16, data []byte) []byte {
	b := make([]byte, 8+len(data))
	b[0] = TypeEchoRequest
	b[1] = 0
	b[2] = 0
	b[3] = 0
	binary.BigEndian.PutUint16(b[4:6], id)
	binary.BigEndian.PutUint16(b[6:8], seq)
	copy(b[8:], data)
	cs := ChecksumICMP(b)
	binary.BigEndian.PutUint16(b[2:4], cs)
	return b
}

// ChecksumICMP computes the ICMP checksum for b (RFC 1071). The checksum field in b should be zero.
func ChecksumICMP(b []byte) uint16 {
	var sum uint32
	length := len(b)
	for i := 0; i < length-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(b[i : i+2]))
	}
	if length%2 == 1 {
		sum += uint32(b[length-1]) << 8
	}
	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return uint16(^sum)
}

// ParseIPv4EchoReply inspects an IPv4 datagram (header + ICMP) and reports whether it is an echo
// reply for the given id and seq. When matched is false and err is nil, the packet was not the
// expected echo reply (e.g. different ICMP type or identifier).
func ParseIPv4EchoReply(packet []byte, wantID, wantSeq uint16) (matched bool, payloadLen int, ttl uint8, err error) {
	ihl, ttl, err := ipv4HeaderLen(packet)
	if err != nil {
		return false, 0, 0, err
	}
	icmp := packet[ihl:]
	if len(icmp) < 8 {
		return false, 0, ttl, ErrTruncated
	}
	typ := icmp[0]
	code := icmp[1]
	rid := binary.BigEndian.Uint16(icmp[4:6])
	rseq := binary.BigEndian.Uint16(icmp[6:8])
	if typ != TypeEchoReply || code != 0 || rid != wantID || rseq != wantSeq {
		return false, 0, ttl, nil
	}
	return true, len(icmp) - 8, ttl, nil
}

func ipv4HeaderLen(b []byte) (ihl int, ttl uint8, err error) {
	if len(b) < 20 {
		return 0, 0, ErrTruncated
	}
	if b[0]>>4 != 4 {
		return 0, 0, ErrNotIPv4
	}
	ihl = int(b[0]&0x0f) * 4
	if ihl < 20 || len(b) < ihl {
		return 0, 0, ErrTruncated
	}
	return ihl, b[8], nil
}
