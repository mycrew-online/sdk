package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/mycrew-online/sdk/pkg/client"
)

func main() {
	fmt.Println("🚀 Testing async safety improvements...")

	// Create a new SimConnect client
	sdk := client.New("AsyncSafetyDemo")
	defer func() {
		fmt.Println("🔄 Closing connection...")
		if err := sdk.Close(); err != nil {
			fmt.Printf("❌ Error closing: %v\n", err)
		} else {
			fmt.Println("✅ Connection closed gracefully")
		}
	}()

	// Test concurrent operations
	var wg sync.WaitGroup

	// Try to connect
	fmt.Println("📡 Attempting to connect to SimConnect...")
	if err := sdk.Open(); err != nil {
		fmt.Printf("⚠️  Connection failed (MSFS not running?): %v\n", err)
		fmt.Println("✅ Testing concurrent operations without connection...")
		testConcurrentOpsWithoutConnection(sdk)
		return
	}

	fmt.Println("✅ Connected successfully!")

	// Test 1: Multiple concurrent Listen() calls (should be safe with sync.Once)
	fmt.Println("\n🧪 Test 1: Multiple concurrent Listen() calls...")
	wg.Add(3)

	var channels []<-chan any
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		go func(id int) {
			defer wg.Done()
			fmt.Printf("   Goroutine %d: Starting Listen()...\n", id)
			ch := sdk.Listen()

			mu.Lock()
			channels = append(channels, ch)
			mu.Unlock()

			if ch != nil {
				fmt.Printf("   Goroutine %d: ✅ Got channel\n", id)
			} else {
				fmt.Printf("   Goroutine %d: ❌ Got nil channel\n", id)
			}
		}(i)
	}

	wg.Wait()

	// Verify all goroutines got the same channel (should be the case with sync.Once)
	mu.Lock()
	if len(channels) > 0 {
		firstChan := channels[0]
		allSame := true
		for _, ch := range channels[1:] {
			if ch != firstChan {
				allSame = false
				break
			}
		}
		if allSame {
			fmt.Println("   ✅ All Listen() calls returned the same channel (sync.Once working)")
		} else {
			fmt.Println("   ❌ Listen() calls returned different channels (sync.Once failed)")
		}
	}
	mu.Unlock()

	// Test 2: Concurrent message processing
	fmt.Println("\n🧪 Test 2: Concurrent message processing...")
	if len(channels) > 0 && channels[0] != nil {
		messageChannel := channels[0]

		// Start multiple message readers
		wg.Add(2)
		messageCount := make([]int, 2)

		for i := 0; i < 2; i++ {
			go func(readerID int) {
				defer wg.Done()
				count := 0
				timeout := time.After(3 * time.Second)

				for {
					select {
					case msg := <-messageChannel:
						if msg != nil {
							count++
							if msgMap, ok := msg.(map[string]any); ok {
								fmt.Printf("   Reader %d: Message %d - Type=%v\n",
									readerID, count, msgMap["type"])
							}
						}
					case <-timeout:
						messageCount[readerID] = count
						fmt.Printf("   Reader %d: Finished with %d messages\n", readerID, count)
						return
					}
				}
			}(i)
		}

		wg.Wait()

		totalMessages := messageCount[0] + messageCount[1]
		fmt.Printf("   ✅ Total messages processed: %d\n", totalMessages)
	}

	// Test 3: Concurrent Close operations
	fmt.Println("\n🧪 Test 3: Concurrent Close() operations...")
	wg.Add(3)

	closeResults := make([]error, 3)

	for i := 0; i < 3; i++ {
		go func(id int) {
			defer wg.Done()
			fmt.Printf("   Goroutine %d: Calling Close()...\n", id)
			closeResults[id] = sdk.Close()
			if closeResults[id] == nil {
				fmt.Printf("   Goroutine %d: ✅ Close() succeeded\n", id)
			} else {
				fmt.Printf("   Goroutine %d: ❌ Close() failed: %v\n", id, closeResults[id])
			}
		}(i)
	}

	wg.Wait()

	// Check that all close operations were handled gracefully
	successCount := 0
	for _, err := range closeResults {
		if err == nil {
			successCount++
		}
	}
	fmt.Printf("   ✅ %d/3 Close() operations succeeded\n", successCount)

	fmt.Println("\n🎉 Async safety tests completed!")
}

func testConcurrentOpsWithoutConnection(sdk client.Connection) {
	var wg sync.WaitGroup

	fmt.Println("🧪 Testing concurrent operations without connection...")

	// Test concurrent Listen() calls without connection
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func(id int) {
			defer wg.Done()
			ch := sdk.Listen()
			if ch == nil {
				fmt.Printf("   Goroutine %d: ✅ Correctly got nil channel (not connected)\n", id)
			} else {
				fmt.Printf("   Goroutine %d: ❌ Got non-nil channel when not connected\n", id)
			}
		}(i)
	}
	wg.Wait()

	// Test concurrent Close() calls without connection
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func(id int) {
			defer wg.Done()
			err := sdk.Close()
			if err == nil {
				fmt.Printf("   Goroutine %d: ✅ Close() handled gracefully\n", id)
			} else {
				fmt.Printf("   Goroutine %d: ❌ Close() error: %v\n", id, err)
			}
		}(i)
	}
	wg.Wait()

	fmt.Println("✅ Concurrent operations without connection handled correctly")
}
