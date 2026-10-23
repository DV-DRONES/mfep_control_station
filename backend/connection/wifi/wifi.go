package wifi

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	DefaultPort           = 50000
	MaxFileBytes          = 100 * 1024 * 1024
	maxFileNameBytes      = 1024
	defaultSendTimeout    = 10 * time.Second
	defaultReceiveTimeout = 2 * time.Minute
)

// send file via tcp
func Send(address, filePath string) error {
	source, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("resolve source file: %w", err)
	}
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("inspect source file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source %q is not a regular file", source)
	}
	if info.Size() > MaxFileBytes {
		return fmt.Errorf("source exceeds the %d-byte transfer limit", MaxFileBytes)
	}

	name := filepath.Base(source)
	if err := validateFileName(name); err != nil {
		return err
	}
	if !utf8.ValidString(name) || len([]byte(name)) > maxFileNameBytes {
		return fmt.Errorf("file name must be valid UTF-8 and at most %d bytes", maxFileNameBytes)
	}

	connection, err := net.DialTimeout("tcp", address, defaultSendTimeout)
	if err != nil {
		return fmt.Errorf("connect to Wi-Fi receiver %q: %w", address, err)
	}
	defer connection.Close()
	if err := connection.SetDeadline(time.Now().Add(defaultSendTimeout)); err != nil {
		return fmt.Errorf("set transfer deadline: %w", err)
	}

	if err := binary.Write(connection, binary.BigEndian, int32(len([]byte(name)))); err != nil {
		return fmt.Errorf("write file name length: %w", err)
	}
	if _, err := io.WriteString(connection, name); err != nil {
		return fmt.Errorf("write file name: %w", err)
	}
	if err := binary.Write(connection, binary.BigEndian, int64(info.Size())); err != nil {
		return fmt.Errorf("write file size: %w", err)
	}

	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer file.Close()
	if _, err := io.CopyN(connection, file, info.Size()); err != nil {
		return fmt.Errorf("send file contents: %w", err)
	}
	return nil
}

// socket
func ReceiveOne(listener net.Listener, outputDirectory string, timeout time.Duration) (string, error) {
	connection, err := listener.Accept()
	if err != nil {
		return "", fmt.Errorf("accept Wi-Fi transfer: %w", err)
	}
	return receiveConnection(connection, outputDirectory, timeout)
}

func receiveConnection(connection net.Conn, outputDirectory string, timeout time.Duration) (string, error) {
	defer connection.Close()
	if timeout > 0 {
		if err := connection.SetDeadline(time.Now().Add(timeout)); err != nil {
			return "", fmt.Errorf("set receive deadline: %w", err)
		}
	}

	var nameLength int32
	if err := binary.Read(connection, binary.BigEndian, &nameLength); err != nil {
		return "", fmt.Errorf("read file name length: %w", err)
	}
	if nameLength <= 0 || nameLength > maxFileNameBytes {
		return "", fmt.Errorf("invalid file name length %d", nameLength)
	}
	nameBytes := make([]byte, nameLength)
	if _, err := io.ReadFull(connection, nameBytes); err != nil {
		return "", fmt.Errorf("read file name: %w", err)
	}
	name := string(nameBytes)
	if !utf8.Valid(nameBytes) {
		return "", fmt.Errorf("file name is not valid UTF-8")
	}
	if err := validateFileName(name); err != nil {
		return "", err
	}

	var size int64
	if err := binary.Read(connection, binary.BigEndian, &size); err != nil {
		return "", fmt.Errorf("read file size: %w", err)
	}
	if size < 0 || size > MaxFileBytes {
		return "", fmt.Errorf("file size %d is invalid or exceeds the %d-byte limit", size, MaxFileBytes)
	}

	if err := os.MkdirAll(outputDirectory, 0o750); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}
	destination := filepath.Join(outputDirectory, name)
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("create received file %q: %w", destination, err)
	}
	if _, err := io.CopyN(file, connection, size); err != nil {
		closeErr := file.Close()
		removeErr := os.Remove(destination)
		return "", errors.Join(
			fmt.Errorf("receive file contents: %w", err),
			closeErr,
			removeErr,
		)
	}
	if err := file.Close(); err != nil {
		removeErr := os.Remove(destination)
		return "", errors.Join(fmt.Errorf("close received file: %w", err), removeErr)
	}
	return destination, nil
}

func Serve(address, outputDirectory string) error {
	if err := os.MkdirAll(outputDirectory, 0o750); err != nil {
		return fmt.Errorf("create receive directory %q: %w", outputDirectory, err)
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen for Wi-Fi transfers on %q: %w", address, err)
	}
	defer listener.Close()
	log.Printf("Wi-Fi receiver listening on %s; saving files to %s", listener.Addr(), outputDirectory)
	return serveListener(listener, outputDirectory)
}

func serveListener(listener net.Listener, outputDirectory string) error {
	for {
		connection, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("transfer%w", err)
		}
		path, err := receiveConnection(connection, outputDirectory, defaultReceiveTimeout)
		if err != nil {
			log.Printf("Wi-Fi transfer failed: %v", err)
			continue
		}
		log.Printf("Received %s", path)
	}
}

func validateFileName(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\:`) {
		return fmt.Errorf("no path")
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".json", ".yaml", ".yml":
		return nil
	default:
		return fmt.Errorf("file name must be .json, .yaml, or .yml")
	}
}
