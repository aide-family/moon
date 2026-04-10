package ping

import (
	"context"
	"encoding/binary"
	"os"
	"testing"
	"time"
)

func TestPingerPing_loopback127(t *testing.T) {
	if testing.Short() {
		t.Skip("skip raw ICMP integration test when -short is set")
	}
	if os.Getenv("JADE_TREE_TEST_ICMP") != "1" {
		t.Skip(`set JADE_TREE_TEST_ICMP=1 and run with elevated privileges (e.g. sudo) to exercise raw ICMP; skipped to avoid "operation not permitted" when running from an IDE`)
	}
	count := 2
	p := &Pinger{
		Count:   count,
		Timeout: 3 * time.Second,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	replies, err := p.Ping(ctx, "127.0.0.1")
	if err != nil {
		t.Fatalf("ping 127.0.0.1: %v", err)
	}
	if len(replies) != count {
		t.Fatalf("want %d replies, got len=%d replies=%+v", count, len(replies), replies)
	}
	t.Logf("replies: %+v", replies)
	for i, r := range replies {
		if r.Seq != uint16(i+1) {
			t.Fatalf("seq: want %d, got %d", i+1, r.Seq)
		}
		wantPayload := len([]byte("HelloPing"))
		if r.PayloadLen != wantPayload {
			t.Fatalf("payload len: want %d, got %d", wantPayload, r.PayloadLen)
		}
		if r.RTT < 0 || r.RTT > p.Timeout {
			t.Fatalf("unexpected RTT: %v", r.RTT)
		}
	}
}

func TestChecksumICMP_includesZeroField(t *testing.T) {
	// RFC 1071 style: checksum computed with checksum field zeroed, then written.
	raw := []byte{TypeEchoRequest, 0, 0, 0, 0x12, 0x34, 0, 1, 'h', 'i'}
	cs := ChecksumICMP(raw)
	binary.BigEndian.PutUint16(raw[2:4], cs)
	if !onesComplementSumIsFFFF(raw) {
		t.Fatalf("expected valid ICMP checksum, got bytes %x", raw)
	}
}

func TestEncodeEchoRequest_fieldsAndChecksum(t *testing.T) {
	payload := []byte("HelloPing")
	p := EncodeEchoRequest(0xabcd, 7, payload)
	if p[0] != TypeEchoRequest || p[1] != 0 {
		t.Fatalf("type/code: got %d/%d", p[0], p[1])
	}
	if binary.BigEndian.Uint16(p[4:6]) != 0xabcd || binary.BigEndian.Uint16(p[6:8]) != 7 {
		t.Fatalf("id/seq mismatch")
	}
	if string(p[8:]) != string(payload) {
		t.Fatalf("payload mismatch: %q", p[8:])
	}
	if !onesComplementSumIsFFFF(p) {
		t.Fatal("checksum invalid")
	}
}

func TestParseIPv4EchoReply_match(t *testing.T) {
	id, seq := uint16(0x1122), uint16(3)
	payload := []byte("x")
	icmp := buildEchoReply(id, seq, payload)
	ip := ipv4TestHeader(20, 64)
	pkt := append(ip, icmp...)

	matched, plen, ttl, err := ParseIPv4EchoReply(pkt, id, seq)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !matched || plen != len(payload) || ttl != 64 {
		t.Fatalf("got matched=%v plen=%d ttl=%d", matched, plen, ttl)
	}
}

func TestParseIPv4EchoReply_wrongSeq(t *testing.T) {
	id := uint16(1)
	icmp := buildEchoReply(id, 2, nil)
	ip := ipv4TestHeader(20, 1)
	pkt := append(ip, icmp...)

	matched, _, _, err := ParseIPv4EchoReply(pkt, id, 3)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if matched {
		t.Fatal("expected no match")
	}
}

func TestParseIPv4EchoReply_truncatedIP(t *testing.T) {
	_, _, _, err := ParseIPv4EchoReply([]byte{0x45}, 1, 1)
	if err != ErrTruncated {
		t.Fatalf("want ErrTruncated, got %v", err)
	}
}

func TestParseIPv4EchoReply_truncatedICMP(t *testing.T) {
	ip := ipv4TestHeader(20, 5)
	// Only 4 bytes of ICMP — too short for header.
	pkt := append(ip, []byte{0, 0, 0, 0}...)
	_, _, _, err := ParseIPv4EchoReply(pkt, 1, 1)
	if err != ErrTruncated {
		t.Fatalf("want ErrTruncated, got %v", err)
	}
}

func TestParseIPv4EchoReply_notIPv4(t *testing.T) {
	// IPv6 traffic class / version nibble 6.
	pkt := make([]byte, 40)
	pkt[0] = 0x60
	_, _, _, err := ParseIPv4EchoReply(pkt, 1, 1)
	if err != ErrNotIPv4 {
		t.Fatalf("want ErrNotIPv4, got %v", err)
	}
}

func ipv4TestHeader(ihl int, ttl uint8) []byte {
	if ihl < 20 || ihl%4 != 0 {
		panic("invalid IHL")
	}
	b := make([]byte, ihl)
	b[0] = 0x40 | byte(ihl/4)
	b[8] = ttl
	return b
}

func buildEchoReply(id, seq uint16, payload []byte) []byte {
	icmp := make([]byte, 8+len(payload))
	icmp[0] = TypeEchoReply
	icmp[1] = 0
	binary.BigEndian.PutUint16(icmp[4:6], id)
	binary.BigEndian.PutUint16(icmp[6:8], seq)
	copy(icmp[8:], payload)
	cs := ChecksumICMP(icmp)
	binary.BigEndian.PutUint16(icmp[2:4], cs)
	return icmp
}

func onesComplementSumIsFFFF(b []byte) bool {
	var sum uint32
	for i := 0; i < len(b)-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(b[i : i+2]))
	}
	if len(b)%2 == 1 {
		sum += uint32(b[len(b)-1]) << 8
	}
	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return uint16(sum) == 0xffff
}
