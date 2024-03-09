package main

import (
	"backend/db"
	"backend/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize the database connection
	db.InitDB()

	// Create a new Hub instance
	hub := handlers.NewHub()

	// WebSocket route for Cashier
	e.GET("/cashier", func(c echo.Context) error {
		return hub.HandleCashierWebSocket(c)
	})

	// WebSocket route for Kitchen
	e.GET("/kitchen", func(c echo.Context) error {
		return hub.HandleKitchenWebSocket(c)
	})

	// Websocket route for get all order lists
	e.GET("/list", func(c echo.Context) error {
		return hub.HandleListAll(c)
	})

	// Start WebSocket hub
	go hub.ProcessOrders()
	go hub.ProcessComplete()

	// Start the server
	e.Start(":8080")
}
