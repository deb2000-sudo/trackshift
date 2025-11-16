package transport

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/deb2000-sudo/trackshift/pkg/protocol"
)

// UDPReceiver receives TrackShift UDP packets and forwards payloads to a handler.
type UDPReceiver struct {
	addr   *net.UDPAddr
	conn   *net.UDPConn
	closed chan struct{}
	wg     sync.WaitGroup

	// Handler is invoked for each successfully decoded packet.
	Handler func(p *protocol.Packet, from *net.UDPAddr)
}

// NewUDPReceiver creates a new UDPReceiver bound to the given port.
func NewUDPReceiver(port int) (*UDPReceiver, error) {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort("", fmt.Sprintf("%d", port)))
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	return &UDPReceiver{
		addr:   addr,
		conn:   conn,
		closed: make(chan struct{}),
	}, nil
}

// Start begins the main receive loop in a background goroutine.
func (r *UDPReceiver) Start() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		buf := make([]byte, 64*1024+256)
		for {
			n, from, err := r.conn.ReadFromUDP(buf)
			if err != nil {
				select {
				case <-r.closed:
					return
				default:
					log.Printf("udp receive error: %v", err)
					continue
				}
			}
			raw := make([]byte, n)
			copy(raw, buf[:n])
			p, err := protocol.DeserializePacket(raw)
			if err != nil {
				log.Printf("udp packet decode error: %v", err)
				continue
			}
			if r.Handler != nil {
				r.Handler(p, from)
			}
		}
	}()
}

// Close stops the receiver and closes the socket.
func (r *UDPReceiver) Close() error {
	close(r.closed)
	err := r.conn.Close()
	r.wg.Wait()
	return err
}


