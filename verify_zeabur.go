package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/GoPolymarket/polymarket-go-sdk/pkg/auth"
)

const ProductionSignerURL = "https://polymarket.zeabur.app/v1/sign-builder"

func main() {
	fmt.Println("🚀 Starting Zeabur Production Verification...")
	fmt.Printf("TARGET: %s\n", ProductionSignerURL)

	config := &auth.BuilderConfig{
		Remote: &auth.BuilderRemoteConfig{
			Host: ProductionSignerURL,
		},
	}

	ctx := context.Background()
	method := "POST"
	path := "/order"
	body := `{"test":"true"}`
	timestamp := time.Now().Unix()

	fmt.Println("📡 Sending signature request to Zeabur...")
	start := time.Now()

	headers, err := config.Headers(ctx, method, path, &body, timestamp)
	if err != nil {
		log.Fatalf("❌ FAILED: %v", err)
	}

	duration := time.Since(start)
	apiKey := headers.Get(auth.HeaderPolyBuilderAPIKey)
	signature := headers.Get(auth.HeaderPolyBuilderSignature)

	fmt.Println("\n✅ ZEABUR VERIFICATION SUCCESSFUL!")
	fmt.Printf("⏱️  Latency: %v\n", duration)
	fmt.Printf("🔑 Builder Key ID: %s\n", apiKey[:8]+"...")
	fmt.Printf("✍️  Signature:      %s...\n", signature[:10])
}
