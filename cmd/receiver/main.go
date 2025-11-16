package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/deb2000-sudo/trackshift/internal/crypto"
	"github.com/deb2000-sudo/trackshift/internal/session"
	"github.com/deb2000-sudo/trackshift/internal/transport"
	"github.com/deb2000-sudo/trackshift/pkg/models"
	"github.com/deb2000-sudo/trackshift/pkg/utils"
)

func main() {
	port := flag.Int("port", 8080, "listening port")
	outputDir := flag.String("output-dir", "received", "output directory for completed files")
	tempDir := flag.String("temp-dir", "", "temporary directory for chunk storage")
	sessionDir := flag.String("sessions-dir", "sessions", "session state directory")
	protocolFlag := flag.String("protocol", "tcp", "transport protocol: tcp or udp")
	logFile := flag.String("log-file", "", "path to log file (optional)")
	flag.Parse()

	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			log.Fatalf("open log file: %v", err)
		}
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	sessMgr, err := session.NewSessionManager(*sessionDir)
	if err != nil {
		log.Fatalf("create session manager: %v", err)
	}
	switch *protocolFlag {
	case "tcp":
		runTCPReceiver(*port, *outputDir, *tempDir, sessMgr)
	case "udp":
		log.Println("UDP receiver mode not yet implemented; starting TCP receiver instead")
		runTCPReceiver(*port, *outputDir, *tempDir, sessMgr)
	default:
		log.Fatalf("unknown protocol %q", *protocolFlag)
	}
}

func runTCPReceiver(port int, outputDir, tempDir string, sessMgr *session.SessionManager) {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen on %s: %v", addr, err)
	}
	defer ln.Close()

	recv, err := transport.NewTCPReceiver(outputDir, tempDir)
	if err != nil {
		log.Fatalf("create receiver: %v", err)
	}

	log.Printf("Receiver listening on %s (tcp)", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go handleConnection(conn, recv, sessMgr)
	}
}

func handleConnection(conn net.Conn, recv *transport.TCPReceiver, sessMgr *session.SessionManager) {
	defer conn.Close()

	// For MVP, we assume a single session per connection. We'll create it lazily
	// on receiving the first chunk.
	var sess *models.TransferSession

	for {
		data, meta, err := recv.Receive(conn)
		if err != nil {
			if err == io.EOF {
				break
			}
			if opErr, ok := err.(net.Error); ok && !opErr.Temporary() {
				log.Printf("connection closed: %v", err)
				break
			}
			log.Printf("receive error: %v", err)
			break
		}

		// Handle file metadata control frame
		if meta.ID == "__filemeta__" {
			var fileMeta models.FileMetadata
			if err := json.Unmarshal(data, &fileMeta); err != nil {
				log.Printf("invalid file metadata frame: %v", err)
				return
			}
			var err error
			sess, err = sessMgr.CreateSession(fileMeta)
			if err != nil {
				log.Printf("create session: %v", err)
				return
			}
			continue
		}

		if sess == nil {
			log.Printf("received data chunk before file metadata; dropping")
			continue
		}

		// Verify hash on decompressed data
		expectedHashBytes, err := hex.DecodeString(meta.SHA256)
		if err != nil {
			log.Printf("invalid hash encoding for chunk %s: %v", meta.ID, err)
			continue
		}
		var expectedHash [32]byte
		copy(expectedHash[:], expectedHashBytes)
		if !crypto.VerifyChunk(data, expectedHash) {
			log.Printf("hash mismatch for chunk %s", meta.ID)
			continue
		}

		meta.SessionID = sess.ID
		if sess.Chunks == nil {
			sess.Chunks = make(map[string]*models.ChunkMetadata)
		}
		sess.Chunks[meta.ID] = meta

		if _, err := recv.StoreChunk(sess.ID, meta, data); err != nil {
			log.Printf("store chunk %s: %v", meta.ID, err)
			continue
		}

		if err := sessMgr.UpdateChunkStatus(sess.ID, meta.ID, models.ChunkStatusCompleted); err != nil {
			log.Printf("update chunk status: %v", err)
		}
	}

	if sess != nil {
		outPath, err := recv.AssembleFile(sess)
		if err != nil {
			log.Printf("assemble file: %v", err)
			return
		}
		log.Printf("Assembled file at %s (%s)", outPath, utils.HumanBytes(sess.File.Size))
	}
}


