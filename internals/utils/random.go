package utils

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

func GetAvailableRandomPort() (int, error) {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {
		port := rand.Intn(65535-1024) + 1024

		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue // port used
		}
		_ = ln.Close()

		return port, nil
	}

	return 0, fmt.Errorf("failed to find available port")
}
