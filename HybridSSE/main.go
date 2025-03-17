package main

import (
	"bufio"
	"fmt"
	"log"
	"strings"
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

	// Fragment data store
	// Simulated employee database
	type Employee struct {
		ID         string
		Name       string
		Position   string
		Department string
	}

	var employeeDatabase = map[string]Employee{
		"123": {ID: "123", Name: "John Smith", Position: "Software Engineer", Department: "Engineering"},
		"535": {ID: "535", Name: "Sarah Johnson", Position: "Project Manager", Department: "Management"},
		"121": {ID: "121", Name: "Mike Davis", Position: "Systems Analyst", Department: "IT"},
		"553": {ID: "553", Name: "Lisa Chen", Position: "UX Designer", Department: "Design"},
		"802": {ID: "802", Name: "Alex Morgan", Position: "DevOps Engineer", Department: "Operations"},
	}

	var fragmentData = map[string]fiber.Map{
		"fragment1": {
			"title": "Welcome to Fragment 1",
			"items": []string{
				"Fragment 1 - Item 1: Server Updates",
				"Fragment 1 - Item 2: System Status",
				"Fragment 1 - Item 3: Performance Metrics",
			},
			"description": "Real-time server monitoring information",
		},
		"fragment2": {
			"title":       "Employee Directory",
			"items":       []string{},
			"description": "Current employee records",
			"employees":   []Employee{},
		},
		"fragment3": {
			"title": "System Analytics",
			"items": []string{
				"Fragment 3 - Item 1: CPU Usage",
				"Fragment 3 - Item 2: Memory Stats",
				"Fragment 3 - Item 3: Network Traffic",
			},
			"description": "System performance analytics",
		},
	}

	app.Get("/static/fragments/:fragment", func(c *fiber.Ctx) error {
		fragment := c.Params("fragment")
		fragmentName := strings.TrimSuffix(fragment, ".html")

		data, exists := fragmentData[fragmentName]
		if !exists {
			return c.Status(fiber.StatusNotFound).SendString("Fragment not found")
		}

		// Add timestamp to the data
		data["timestamp"] = time.Now().Format(time.RFC3339)

		return c.Render("fragments/"+fragmentName, data)
	})

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

	app.Post("/scans", func(c *fiber.Ctx) error {
		ids := strings.Split(c.FormValue("ids"), ",")
		var foundEmployees []Employee

		for _, id := range ids {
			if emp, exists := employeeDatabase[strings.TrimSpace(id)]; exists {
				foundEmployees = append(foundEmployees, emp)
			}
		}

		var fragmentToRender string
		if len(foundEmployees) > 0 {
			// Update fragment2 data with found employees
			fragmentData["fragment2"] = fiber.Map{
				"title":       "Employee Directory",
				"description": fmt.Sprintf("Found %d employees", len(foundEmployees)),
				"employees":   foundEmployees,
			}
			fragmentToRender = "fragment2"
		} else {
			// Use error fragment when no employees found
			fragmentData["error"] = fiber.Map{
				"title":   "No Results Found",
				"message": fmt.Sprintf("No employees found matching IDs: %s", c.FormValue("ids")),
			}
			fragmentToRender = "error"
		}

		// Notify all clients to update with appropriate fragment
		htmxInstruction := fmt.Sprintf("<div hx-get=\"/static/fragments/%s.html\" hx-trigger=\"load\" hx-swap=\"innerHTML\" hx-target=\"#content\"></div>", fragmentToRender)

		clientsMux.Lock()
		for client := range clients {
			select {
			case client.channel <- htmxInstruction:
			default:
			}
		}
		clientsMux.Unlock()

		return c.SendString(fmt.Sprintf("Found %d employees", len(foundEmployees)))
	})

	app.Post("/send", func(c *fiber.Ctx) error {
		fragment := c.FormValue("fragment")
		if fragment == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Fragment parameter required")
		}

		htmxInstruction := fmt.Sprintf("<div hx-get=\"/static/fragments/%s.html\" hx-trigger=\"load\" hx-swap=\"innerHTML\" hx-target=\"#content\"></div>", fragment)

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
