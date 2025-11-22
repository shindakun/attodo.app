package main

import (
	"fmt"
	"log"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	// Generate VAPID keys (returns base64-encoded strings)
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		log.Fatalf("Failed to generate VAPID keys: %v", err)
	}

	fmt.Println("VAPID Keys Generated Successfully!")
	fmt.Println()
	fmt.Println("Add these to your .env file:")
	fmt.Println()
	fmt.Printf("VAPID_PUBLIC_KEY=%s\n", publicKey)
	fmt.Printf("VAPID_PRIVATE_KEY=%s\n", privateKey)
	fmt.Println()
	fmt.Println("The public key should also be included in your client-side JavaScript")
	fmt.Println("to subscribe users to push notifications.")
}
