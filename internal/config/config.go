package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	NodeID   string
	HTTPPort int
	GRPCPort int
	Peers    []string
}

func LoadConfig() (*Config, error) {
	nodeID := os.Getenv("NODE_ID")
	httpPortStr := os.Getenv("HTTP_PORT")
	grpcPortStr := os.Getenv("GRPC_PORT")
	peersStr := os.Getenv("PEERS")

	if nodeID == "" || httpPortStr == "" || grpcPortStr == "" || peersStr == "" {
		log.Fatal("NODE_ID, HTTP_PORT, GRPC_PORT, and PEERS environment variables must be set")
	}

	httpPort, err := strconv.Atoi(httpPortStr)
	if err != nil {
		return nil, err
	}

	grpcPort, err := strconv.Atoi(grpcPortStr)
	if err != nil {
		return nil, err
	}

	peers := strings.Split(peersStr, ",")

	config := &Config{
		NodeID:   nodeID,
		HTTPPort: httpPort,
		GRPCPort: grpcPort,
		Peers:    peers,
	}

	return config, nil
}
