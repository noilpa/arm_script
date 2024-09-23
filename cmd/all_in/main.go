package main

import (
	"context"
	"log"

	"arm_migrator/internal/ci"
	"arm_migrator/internal/docker"
	"arm_migrator/internal/values"
)

func main() {
	ctx := context.Background()
	projectPath := "/Users/ipnovikov/GolandProjects/cas-backend"

	if _, err := docker.Update(ctx, projectPath); err != nil {
		log.Printf("Error: %v\n", err)
	}

	if _, err := values.Update(ctx, projectPath); err != nil {
		log.Printf("Error: %v\n", err)
	}

	if _, err := ci.Update(ctx, projectPath); err != nil {
		log.Printf("Error: %v\n", err)
	}
}
