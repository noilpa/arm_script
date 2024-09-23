package main

import (
	"context"
	"log"

	"arm_migrator/internal/docker"
)

func main() {
	ctx := context.Background()
	projectPath := "/Users/ipnovikov/GolandProjects/cas-backend"

	_, err := docker.Update(ctx, projectPath)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
}
