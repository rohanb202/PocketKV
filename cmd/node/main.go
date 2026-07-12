package main

import (
	"context"
	"log/slog"
	"os"
	
	"github.com/joho/godotenv"

	"dist-cache/node"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	_ = godotenv.Load()

	ctx := context.Background()

	id := getEnv("NODE_ID", "node1")
	address := getEnv("NODE_ADDRESS", ":8081")

	n := node.NewNode(
		ctx,
		id,
		address,
	)

	slog.Info(
		"starting node",
		slog.String("id", id),
		slog.String("address", address),
	)

	if err := n.Start(); err != nil {
		slog.Error(
			"node stopped",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
}