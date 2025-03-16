package main

import (
	"bufio"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

type Client struct {
	channel chan string
}

var (
	clients    = make(map[*Client]bool)
	clientsMux sync.RWMutex
)

func main() {
	engine := html.New("./templates/", ".html")

	app := fiber.New(fiber.Config{
		Views:        engine,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 0,
	})

	app.Use(logger.New())

	app.Static("/static", "./templates")

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})

	// SSE endpoint
	app.Get("/events", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Access-Control-Allow-Origin", "*")

		client := &Client{
			channel: make(chan string, 10),
		}

		clientsMux.Lock()
		clients[client] = true
		clientsMux.Unlock()

		c.Context().SetUserValue("client", client)

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			defer func() {
				clientsMux.Lock()
				delete(clients, client)
				close(client.channel)
				clientsMux.Unlock()
			}()

			fmt.Fprintf(w, "data: {\"message\": \"Connected to SSE\"}\n\n")
			w.Flush()

			for {
				select {
				case msg, ok := <-client.channel:
					if !ok {
						return
					}
					_, err := fmt.Fprintf(w, "data: %s\n\n", msg)
					if err != nil {
						return
					}
					if err := w.Flush(); err != nil {
						return
					}
				}
			}
		})

		return nil
	})

	app.Post("/send", func(c *fiber.Ctx) error {
		fragment := c.FormValue("fragment")
		if fragment == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Fragment parameter required")
		}

		htmxInstruction := fmt.Sprintf("<div hx-get=\"/%s.html\" hx-trigger=\"load\" hx-swap=\"innerHTML\" hx-target=\"#content\"></div>", fragment)

		clientsMux.Lock()
		for client := range clients {
			select {
			case client.channel <- htmxInstruction:
			default:
			}
		}
		clientsMux.Unlock()

		return c.SendString(fmt.Sprintf("Fragment %s sent to all connected clients", fragment))
	})

	log.Fatal(app.Listen(":9080"))
}
