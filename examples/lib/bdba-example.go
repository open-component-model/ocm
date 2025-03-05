package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin" // Vulnerable dependency, older version
	"github.com/unknwon/com"   // GPL 2.0 license
)

func main() {
	// Simple code to trigger imports. No functional impact.
	fmt.Println("Testing Blackduck Rapid Scan")

	// Triggering a potentially vulnerable function from gin (older versions had CVEs)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	//Triggering a function from com which has a GPL 2.0 license.
	result, err := com.IsDir("some/non/existent/path") // Using com package.
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Is Dir:", result)

	// Start the server (just for demonstration, not needed for scan)
	go func() {
		if err := r.Run(":8081"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Keep the main function running, otherwise server will stop immediately.
	select {}
}
