package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/deb2000-sudo/trackshift/internal/chunker"
	"github.com/deb2000-sudo/trackshift/internal/crypto"
	"github.com/deb2000-sudo/trackshift/internal/session"
	"github.com/deb2000-sudo/trackshift/internal/transport"
	"github.com/deb2000-sudo/trackshift/pkg/models"
	"github.com/deb2000-sudo/trackshift/pkg/utils"
)

func main() {
	filePath := flag.String("file", "", "input file path")
	receiverAddr := flag.String("receiver", "", "receiver address (host:port)")
	chunkSizeFlag := flag.Int64("chunk-size", 50*1024*1024, "chunk size in bytes (default 50MB)")
	sessionDir := flag.String("output-dir", "sessions", "session state directory")
	protocolFlag := flag.String("protocol", "tcp", "transport protocol: tcp or udp")
	parallelStreams := flag.Int("parallel-streams", 32, "number of parallel streams for UDP")
	resumeSession := flag.String("resume", "", "resume existing session ID instead of creating a new one")
	flag.Parse()

	if *filePath == "" || *receiverAddr == "" {
		flag.Usage()
		os.Exit(1)
	}

	info, err := os.Stat(*filePath)
	if err != nil {
		log.Fatalf("stat input file: %v", err)
	}

	fileHash, err := utils.HashFileSHA256(*filePath)
	if err != nil {
		log.Fatalf("hash input file: %v", err)
	}

	fileMeta := models.FileMetadata{
		Name: info.Name(),
		Size: info.Size(),
		Hash: fileHash,
	}

	sessMgr, err := session.NewSessionManager(*sessionDir)
	if err != nil {
		log.Fatalf("create session manager: %v", err)
	}

	var sess *models.TransferSession
	if *resumeSession != "" {
		sess, err = sessMgr.GetSession(*resumeSession)
		if err != nil {
			log.Fatalf("load session %s: %v", *resumeSession, err)
		}
		log.Printf("Resuming session %s", sess.ID)
	} else {
		sess, err = sessMgr.CreateSession(fileMeta)
		if err != nil {
			log.Fatalf("create session: %v", err)
		}
	}

	ch := chunker.NewChunker(chunker.ChunkerConfig{})
	chunkMetas, err := ch.ChunkFile(*filePath, *chunkSizeFlag)
	if err != nil {
		log.Fatalf("chunk file: %v", err)
	}
	sess.TotalChunks = len(chunkMetas)

	if err := sessMgr.SaveSession(sess); err != nil {
		log.Fatalf("save session: %v", err)
	}

	log.Printf("Starting transfer: %s (%s) to %s, %d chunks over %s\n",
		fileMeta.Name, utils.HumanBytes(fileMeta.Size), *receiverAddr, len(chunkMetas), *protocolFlag)

	switch *protocolFlag {
	case "tcp":
		runTCPSender(*receiverAddr, *filePath, fileMeta, sess, sessMgr, chunkMetas, info.Size())
	case "udp":
		runUDPSender(*receiverAddr, *filePath, fileMeta, sess, sessMgr, chunkMetas, info.Size(), *parallelStreams)
	default:
		log.Fatalf("unknown protocol %q", *protocolFlag)
	}
}

func runTCPSender(receiver, filePath string, fileMeta models.FileMetadata, sess *models.TransferSession,
	sessMgr *session.SessionManager, chunkMetas []*models.ChunkMetadata, totalSize int64) {

	sender := transport.NewTCPSender()
	conn, err := sender.Connect(receiver)
	if err != nil {
		log.Fatalf("connect to receiver: %v", err)
	}
	defer conn.Close()

	bar := progressbar.NewOptions64(
		totalSize,
		progressbar.OptionSetDescription("transferring"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(),
	)

	// Handle Ctrl+C
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		log.Println("\nInterrupt received, shutting down gracefully...")
		conn.Close()
		os.Exit(1)
	}()

	// open file for reading chunks
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("open input file: %v", err)
	}
	defer f.Close()

	// send file metadata frame first
	metaPayload, err := json.Marshal(fileMeta)
	if err != nil {
		log.Fatalf("marshal file metadata: %v", err)
	}
	metaFrame := &models.ChunkMetadata{
		ID:       "__filemeta__",
		Size:     int64(len(metaPayload)),
		Offset:   0,
		SHA256:   "",
		IsParity: false,
		Status:   models.ChunkStatusPending,
	}
	compMetaPayload, err := crypto.CompressChunk(metaPayload)
	if err != nil {
		log.Fatalf("compress file metadata frame: %v", err)
	}
	if err := sender.Send(conn, compMetaPayload, metaFrame); err != nil {
		log.Fatalf("send file metadata frame: %v", err)
	}

	for _, meta := range chunkMetas {
		buf := make([]byte, meta.Size)
		if _, err := f.ReadAt(buf, meta.Offset); err != nil {
			log.Fatalf("read chunk at offset %d: %v", meta.Offset, err)
		}

		// hash original data
		dataHash := crypto.HashChunk(buf)
		meta.SHA256 = fmt.Sprintf("%x", dataHash[:])
		meta.SessionID = sess.ID

		// compress for transport
		compressed, err := crypto.CompressChunk(buf)
		if err != nil {
			log.Fatalf("compress chunk: %v", err)
		}

		if err := sender.Send(conn, compressed, meta); err != nil {
			log.Fatalf("send chunk %s: %v", meta.ID, err)
		}

		sess.BytesSent += meta.Size
		if err := sessMgr.UpdateChunkStatus(sess.ID, meta.ID, models.ChunkStatusCompleted); err != nil {
			log.Printf("update chunk status: %v", err)
		}

		_ = bar.Add64(meta.Size)
	}

	log.Println("Transfer complete.")
}

func runUDPSender(receiver, filePath string, fileMeta models.FileMetadata, sess *models.TransferSession,
	sessMgr *session.SessionManager, chunkMetas []*models.ChunkMetadata, totalSize int64, parallelStreams int) {
	// UDP implementation will be added in the next iteration; for now fall back to TCP
	log.Println("UDP protocol not yet fully implemented; falling back to TCP for now")
	runTCPSender(receiver, filePath, fileMeta, sess, sessMgr, chunkMetas, totalSize)
}

