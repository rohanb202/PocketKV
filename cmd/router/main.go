package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	
	"dist-cache/cluster"
	"dist-cache/node"
	"dist-cache/router"
	"github.com/joho/godotenv"
)

func getEnv(key, defaultValue string) string {

	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}

func main() {

	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	routerPort := getEnv(
		"ROUTER_PORT",
		"8080",
	)

	nodesEnv := getEnv(
		"NODES",
		"localhost:8081,localhost:8082,localhost:8083",
	)

	addresses := strings.Split(
		nodesEnv,
		",",
	)

	cl := cluster.NewCluster()

	for i, address := range addresses {

		id := fmt.Sprintf(
			"node%d",
			i+1,
		)

		n := node.NewNode(
			ctx,
			id,
			address,
		)

		cl.AddNode(n)
	}

	cl.Start(ctx)

	rt := router.NewRouter(cl)

	http.HandleFunc(
		"/cache",
		rt.CacheHandler,
	)

	slog.Info(
		"router started",
		slog.String("port", routerPort),
	)

	if err := http.ListenAndServe(
		":"+routerPort,
		nil,
	); err != nil {

		slog.Error(
			"router stopped",
			slog.Any("error", err),
		)

		os.Exit(1)
	}
}