package datasource

import (
	"context"
	"time"

	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
	"github.com/go-ping/ping"
)

// PingDetail ping探测器详情
type PingDetail struct {
	// 总包数
	TotalPackets float64 `json:"totalPackets,omitempty"`
	// 成功包数
	SuccessPackets float64 `json:"successPackets,omitempty"`
	// 丢包率
	LossRate float64 `json:"lossRate,omitempty"`
	// 最小延迟
	MinDelay float64 `json:"minDelay,omitempty"`
	// 最大延迟
	MaxDelay float64 `json:"maxDelay,omitempty"`
	// 平均延迟
	AvgDelay float64 `json:"avgDelay,omitempty"`
	// 标准差
	StdDevDelay float64 `json:"stdDevDelay,omitempty"`
}

// EndpointPing 是ping探测器的实现
func EndpointPing(_ context.Context, endpoint string, timeout time.Duration) (map[watch.Indexer]*Point, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	pinger, err := ping.NewPinger(endpoint)
	if err != nil {
		return nil, err
	}

	go func() {
		select {
		case <-ctx.Done():
			pinger.Stop()
		}
	}()
	now := time.Now()
	points := make(map[watch.Indexer]*Point)

	pinger.OnRecv = func(pkt *ping.Packet) {
		//fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
		//	pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	}

	pinger.OnDuplicateRecv = func(pkt *ping.Packet) {
		//fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v (DUP!)\n",
		//	pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
	}

	var detail PingDetail

	pinger.OnFinish = func(stats *ping.Statistics) {
		//fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
		//fmt.Printf("%d packets transmitted, %d packets received, %v%% packet loss\n",
		//	stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		//fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
		//	stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
		detail.TotalPackets = float64(stats.PacketsSent)
		detail.SuccessPackets = float64(stats.PacketsRecv)
		detail.LossRate = float64(stats.PacketLoss)
		detail.MinDelay = float64(stats.MinRtt.Milliseconds())
		detail.MaxDelay = float64(stats.MaxRtt.Milliseconds())
		detail.AvgDelay = float64(stats.AvgRtt.Milliseconds())
		detail.StdDevDelay = float64(stats.StdDevRtt.Milliseconds())
	}

	//fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
	if err := pinger.Run(); err != nil {
		return nil, err
	}
	labels := vobj.NewLabels(map[string]string{
		vobj.Domain: endpoint,
	})
	unix := now.Unix()
	points[labels] = &Point{
		Labels: labels.Map(),
		Values: []*Value{
			// 总包数
			{
				Value:     detail.TotalPackets,
				Timestamp: unix,
			},
			// 成功包数
			{
				Value:     detail.SuccessPackets,
				Timestamp: unix,
			},
			// 丢包率
			{
				Value:     detail.LossRate,
				Timestamp: unix,
			},
			// 最小延迟
			{
				Value:     detail.MinDelay,
				Timestamp: unix,
			},
			// 最大延迟
			{
				Value:     detail.MaxDelay,
				Timestamp: unix,
			},
			// 平均延迟
			{
				Value:     detail.AvgDelay,
				Timestamp: unix,
			},
			// 标准差
			{
				Value:     detail.StdDevDelay,
				Timestamp: unix,
			},
		},
	}
	return points, nil
}
