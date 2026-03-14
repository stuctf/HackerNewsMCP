package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
)

func main() {
	// MCP servers must not write anything to stdout except JSON-RPC responses.
	// Log to stderr for debugging.
	log.SetOutput(os.Stderr)

	scanner := bufio.NewScanner(os.Stdin)
	// Increase buffer for large responses.
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("invalid JSON-RPC: %v", err)
			continue
		}

		resp := HandleRequest(req)
		if resp == nil {
			continue // Notification, no response needed.
		}

		if err := encoder.Encode(resp); err != nil {
			log.Printf("encode response: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("stdin read error: %v", err)
	}
}
