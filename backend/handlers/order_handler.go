package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"backend/db"
	"backend/models"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// Hub represents the central hub for managing WebSocket connections.
type Hub struct {
	upgrader           websocket.Upgrader
	cashierClients     map[*websocket.Conn]bool
	cashierMutex       sync.Mutex
	kitchenClients     map[*websocket.Conn]bool
	kitchenMutex       sync.Mutex
	orderBroadcast     chan models.Order
	completedBroadcast chan models.Order
	kitchenResponse    chan models.Order
	cashierResponse    chan models.Order
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all connections from different client (e.g Postman/Web)
			},
		},
		cashierClients:  make(map[*websocket.Conn]bool),
		kitchenClients:  make(map[*websocket.Conn]bool),
		orderBroadcast:  make(chan models.Order),
		kitchenResponse: make(chan models.Order),
		cashierResponse: make(chan models.Order),
	}
}

// HandleCashierWebSocket handles WebSocket connections from cashiers.
func (h *Hub) HandleCashierWebSocket(c echo.Context) error {
	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Register cashier client
	h.cashierMutex.Lock()
	h.cashierClients[conn] = true
	h.cashierMutex.Unlock()

	for {
		var order models.Order
		// Handling cashier client on error
		err := conn.ReadJSON(&order)
		if err != nil {
			h.cashierMutex.Lock()
			delete(h.cashierClients, conn)
			h.cashierMutex.Unlock()
			break
		}

		db.DB.Create(&order)

		// Broadcast order to kitchen
		h.orderBroadcast <- order
		response := <-h.cashierResponse
		messages := fmt.Sprintf("Order %d Created: %s x %d x %s x %s", response.ID, response.Item, response.Quantity, response.Notes, response.Status)

		err = conn.WriteMessage(websocket.TextMessage, []byte(messages))
		if err != nil {
			// Unregister kitchen client on error
			h.kitchenMutex.Lock()
			delete(h.kitchenClients, conn)
			h.kitchenMutex.Unlock()
			break
		}
	}
	return nil
}

// // HandleKitchenWebSocket handles WebSocket connections from the kitchen.
func (h *Hub) HandleKitchenWebSocket(c echo.Context) error {
	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Register kitchen client
	h.kitchenMutex.Lock()
	h.kitchenClients[conn] = true
	h.kitchenMutex.Unlock()

	for {
		// Wait for message from the kitchen
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			h.kitchenMutex.Lock()
			delete(h.kitchenClients, conn)
			h.kitchenMutex.Unlock()
			break
		}

		// Process message from the kitchen
		if messageType == websocket.TextMessage {
			var data models.Order
			err := json.Unmarshal(message, &data)
			if err != nil {
				log.Println("Error decoding order ID:", err)
				continue
			}

			// Retrieve the order from the database by ID
			order, err := getOrderByID(int(data.ID))
			if err != nil {
				log.Println("Error retrieving order:", err)
				continue
			}

			// Update the order status
			order.Status = data.Status
			if err := db.DB.Save(&order).Error; err != nil {
				log.Println("Error updating order status:", err)
				continue
			}

			// Send a response to the cashier
			h.cashierResponse <- order
		}
	}

	return nil
}

// Function to get an order by its ID from the DB
func getOrderByID(orderID int) (models.Order, error) {
	var order models.Order
	if err := db.DB.Where("id = ?", orderID).First(&order).Error; err != nil {
		return order, err
	}
	return order, nil
}

// Processesing orders and sends responses to the kitchen.
func (h *Hub) ProcessOrders() {
	for {
		// Wait for new orders from cashiers
		order := <-h.orderBroadcast

		fmt.Printf("New Order: %v\n", order)

		// Send a response to the kitchen
		h.kitchenResponse <- order
	}
}

// Processing orders and updating orders from kitchen.
func (h *Hub) ProcessComplete() {
	for {
		// Wait for new orders from kitchen
		order := <-h.completedBroadcast

		fmt.Printf("Processed Order: %d\n", order.ID)

		// Send a response from the kitchen
		h.cashierResponse <- order
	}
}

func (h *Hub) HandleListAll(c echo.Context) error {
	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		orders, err := getAllOrders()
		if err != nil {
			log.Println(err)
			return err
		}

		// Marshalling the data to become desired struct
		jsonData, err := json.Marshal(orders)
		if err != nil {
			fmt.Println(err)
			return err
		}

		// Send data to the client
		if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			log.Println(err)
			return err
		}

		// Simulate real-time updates with a delay
		time.Sleep(10 * time.Second)
	}
}

// Function to get all orders from DB
func getAllOrders() (*[]models.Order, error) {
	var orders *[]models.Order
	if err := db.DB.Find(&orders).Error; err != nil {
		return orders, err
	}
	return orders, nil
}
