package main

import (
	"context"
	"log"

	"arm_migrator/internal/values"
)

func main() {
	ctx := context.Background()
	projectPath := "/Users/ipnovikov/GolandProjects/cas-backend"

	_, err := values.Update(ctx, projectPath)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
}
