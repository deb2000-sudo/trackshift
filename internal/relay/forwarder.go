package relay

import (
	"log"
	"net"
	"sync"
	"time"
)

// Forwarder is a minimal UDP packet forwarder used by edge relays.
type Forwarder struct {
	ListenAddr      *net.UDPAddr
	ForwardAddr     *net.UDPAddr
	RelayID         string
	OrchestratorURL string

	conn   *net.UDPConn
	closed chan struct{}
	wg     sync.WaitGroup
}

// NewForwarder creates a new Forwarder.
func NewForwarder(listen, forward, relayID, orchestratorURL string) (*Forwarder, error) {
	laddr, err := net.ResolveUDPAddr("udp", listen)
	if err != nil {
		return nil, err
	}
	faddr, err := net.ResolveUDPAddr("udp", forward)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}
	return &Forwarder{
		ListenAddr:      laddr,
		ForwardAddr:     faddr,
		RelayID:         relayID,
		OrchestratorURL: orchestratorURL,
		conn:            conn,
		closed:          make(chan struct{}),
	}, nil
}

// Start begins forwarding packets until Close is called.
func (f *Forwarder) Start() {
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		buf := make([]byte, 64*1024+256)
		for {
			n, addr, err := f.conn.ReadFromUDP(buf)
			if err != nil {
				select {
				case <-f.closed:
					return
				default:
					log.Printf("[relay %s] read error from %v: %v", f.RelayID, addr, err)
					continue
				}
			}
			// best-effort forward
			if _, err := f.conn.WriteToUDP(buf[:n], f.ForwardAddr); err != nil {
				log.Printf("[relay %s] forward error to %v: %v", f.RelayID, f.ForwardAddr, err)
			}
		}
	}()

	// heartbeat/metrics ticker placeholder
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Printf("[relay %s] heartbeat (forwarding to %s)", f.RelayID, f.ForwardAddr.String())
			case <-f.closed:
				return
			}
		}
	}()
}

// Close stops forwarding and closes the socket.
func (f *Forwarder) Close() error {
	close(f.closed)
	err := f.conn.Close()
	f.wg.Wait()
	return err
}


