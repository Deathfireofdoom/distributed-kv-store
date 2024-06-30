package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Deathfireofdoom/distributed-kv-store/internal/api"
	"github.com/Deathfireofdoom/distributed-kv-store/internal/config"
	"github.com/Deathfireofdoom/distributed-kv-store/internal/kvstore"
	"github.com/Deathfireofdoom/distributed-kv-store/internal/raft"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	nodeID := cfg.NodeID
	httpPort := cfg.HTTPPort
	grpcPort := cfg.GRPCPort
	peers := cfg.Peers

	kvstore := kvstore.NewStore()

	raftNode := raft.NewRaftNode(nodeID, peers, kvstore)
	go func() {
		grpcAddress := fmt.Sprintf(":%d", grpcPort)
		if err := raftNode.StartGRPCServer(grpcAddress); err != nil {
			log.Fatalf("failed to start grpc server: %v", err)
		}
	}()

	router := api.NewRouter()
	log.Println("Starting server")
	httpAddress := fmt.Sprintf(":%d", httpPort)
	if err := http.ListenAndServe(httpAddress, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
