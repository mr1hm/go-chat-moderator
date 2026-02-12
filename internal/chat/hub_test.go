package chat

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// waitFor polls a condition until it returns true or timeout expires
func waitFor(t *testing.T, timeout time.Duration, condition func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal(msg)
}

// mockClient creates a test client without WebSocket connection
func mockClient(hub *Hub, userID, roomID string) *Client {
	return &Client{
		Hub:    hub,
		Send:   make(chan []byte, 256),
		UserID: userID,
		RoomID: roomID,
	}
}

func TestHub_ConcurrentRegisterUnregister(t *testing.T) {
	hub := NewHub()

	// Start hub in background (without Redis subscription for testing)
	go func() {
		for {
			select {
			case client := <-hub.register:
				hub.addClient(client)
			case client := <-hub.unregister:
				hub.removeClient(client)
			case item := <-hub.broadcast:
				if item.msg != nil {
					hub.broadcastToRoom(item.msg)
				} else {
					hub.broadcastRaw(item.roomID, item.raw)
				}
			}
		}
	}()

	var wg sync.WaitGroup
	clientCount := 100

	// Spawn many clients connecting/disconnecting rapidly
	for i := 0; i < clientCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			client := mockClient(hub, fmt.Sprintf("user-%d", id), "room1")
			hub.register <- client

			// Small random delay
			time.Sleep(time.Microsecond * time.Duration(id%10))

			hub.unregister <- client
		}(i)
	}

	wg.Wait()

	// Verify room is cleaned up
	waitFor(t, time.Second, func() bool {
		hub.mtx.RLock()
		defer hub.mtx.RUnlock()
		return len(hub.rooms) == 0
	}, "expected 0 rooms after all clients left")
}

func TestHub_ConcurrentBroadcastDuringJoinLeave(t *testing.T) {
	hub := NewHub()

	// Start hub
	go func() {
		for {
			select {
			case client := <-hub.register:
				hub.addClient(client)
			case client := <-hub.unregister:
				hub.removeClient(client)
			case item := <-hub.broadcast:
				if item.msg != nil {
					hub.broadcastToRoom(item.msg)
				} else {
					hub.broadcastRaw(item.roomID, item.raw)
				}
			}
		}
	}()

	var wg sync.WaitGroup
	roomID := "test-room"

	// Register some persistent clients
	persistentClients := make([]*Client, 10)
	for i := 0; i < 10; i++ {
		persistentClients[i] = mockClient(hub, fmt.Sprintf("persistent-%d", i), roomID)
		hub.register <- persistentClients[i]
	}

	// Wait for all clients to be registered
	waitFor(t, time.Second, func() bool {
		hub.mtx.RLock()
		defer hub.mtx.RUnlock()
		return len(hub.rooms[roomID]) == 10
	}, "expected 10 persistent clients to be registered")

	// Concurrent broadcasts
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			hub.broadcast <- broadcastItem{
				roomID: roomID,
				raw:    []byte(fmt.Sprintf(`{"type":"message","payload":"msg-%d"}`, i)),
			}
			time.Sleep(time.Microsecond * 100)
		}
	}()

	// Concurrent join/leave
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			client := mockClient(hub, fmt.Sprintf("transient-%d", i), roomID)
			hub.register <- client
			time.Sleep(time.Microsecond * 50)
			hub.unregister <- client
		}
	}()

	wg.Wait()

	expectedMessages := 50

	// Verify ALL messages arrived at each persistent client
	for _, client := range persistentClients {
		received := 0
		waitFor(t, 2*time.Second, func() bool {
			received = len(client.Send)
			return received >= expectedMessages
		}, fmt.Sprintf("client %s: expected %d messages, got %d", client.UserID, expectedMessages, received))
	}

	// Cleanup
	for _, client := range persistentClients {
		hub.unregister <- client
	}
}

func TestHub_DoubleUnregister(t *testing.T) {
	hub := NewHub()

	// Start hub
	go func() {
		for {
			select {
			case client := <-hub.register:
				hub.addClient(client)
			case client := <-hub.unregister:
				hub.removeClient(client)
			case item := <-hub.broadcast:
				if item.msg != nil {
					hub.broadcastToRoom(item.msg)
				} else {
					hub.broadcastRaw(item.roomID, item.raw)
				}
			}
		}
	}()

	client := mockClient(hub, "user1", "room1")
	hub.register <- client

	// Wait for registration
	waitFor(t, time.Second, func() bool {
		hub.mtx.RLock()
		defer hub.mtx.RUnlock()
		_, exists := hub.rooms["room1"][client]
		return exists
	}, "client should be registered")

	// Unregister twice - should not panic
	hub.unregister <- client
	hub.unregister <- client

	// Wait for channel to be closed
	// Verify channel is closed (range should exit)
	closed := make(chan bool)
	go func() {
		for range client.Send {
		}
		closed <- true
	}()

	select {
	case <-closed:
		// success - channel was closed
	case <-time.After(100 * time.Millisecond):
		t.Error("channel was not closed after unregister")
	}
}

func TestHub_BroadcastToEmptyRoom(t *testing.T) {
	hub := NewHub()

	// Start hub
	go func() {
		for {
			select {
			case client := <-hub.register:
				hub.addClient(client)
			case client := <-hub.unregister:
				hub.removeClient(client)
			case item := <-hub.broadcast:
				if item.msg != nil {
					hub.broadcastToRoom(item.msg)
				} else {
					hub.broadcastRaw(item.roomID, item.raw)
				}
			}
		}
	}()

	// Broadcast to non-existent room - should not panic
	done := make(chan struct{})
	go func() {
		hub.broadcast <- broadcastItem{
			roomID: "nonexistent",
			raw:    []byte(`{"test": true}`),
		}
		close(done)
	}()

	select {
	case <-done:
		// success - broadcast completed without panic
	case <-time.After(time.Second):
		t.Fatal("broadcast to empty room timed out")
	}
}

func TestHub_ClientBufferFull(t *testing.T) {
	hub := NewHub()

	// Start hub
	go func() {
		for {
			select {
			case client := <-hub.register:
				hub.addClient(client)
			case client := <-hub.unregister:
				hub.removeClient(client)
			case item := <-hub.broadcast:
				if item.msg != nil {
					hub.broadcastToRoom(item.msg)
				} else {
					hub.broadcastRaw(item.roomID, item.raw)
				}
			}
		}
	}()

	// Create client with small buffer for testing
	client := &Client{
		Hub:    hub,
		Send:   make(chan []byte, 2), // Small buffer
		UserID: "slow-client",
		RoomID: "room1",
	}
	hub.register <- client

	// Wait for registration
	waitFor(t, time.Second, func() bool {
		hub.mtx.RLock()
		defer hub.mtx.RUnlock()
		_, exists := hub.rooms["room1"][client]
		return exists
	}, "client should be registered")

	// Flood with messages - should disconnect client when buffer full
	for i := 0; i < 10; i++ {
		hub.broadcast <- broadcastItem{
			roomID: "room1",
			raw:    []byte(fmt.Sprintf(`{"msg": %d}`, i)),
		}
	}

	// Client should be removed from room
	waitFor(t, time.Second, func() bool {
		hub.mtx.RLock()
		defer hub.mtx.RUnlock()
		clients := hub.rooms["room1"]
		_, exists := clients[client]
		return !exists
	}, "slow client should have been removed when buffer filled")
}

func TestClient_CloseOnce(t *testing.T) {
	client := &Client{
		Send: make(chan []byte, 1),
	}

	// Close multiple times - should not panic
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.close()
		}()
	}
	wg.Wait()

	// Verify channel is closed
	select {
	case _, ok := <-client.Send:
		if ok {
			t.Error("expected channel to be closed")
		}
	default:
		t.Error("channel should be closed and readable")
	}
}
