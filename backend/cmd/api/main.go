package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai/deepseek"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type application struct {
	echo         *echo.Echo
	genkit       *genkit.Genkit
	searchTool   *ai.ToolAction[SearchInput, string]
	generateFlow *core.Flow[AIRequest, string, EventMessage]
}

func NewApplication(e *echo.Echo, g *genkit.Genkit) *application {
	toolDescription := `Searches Parallel AI for the latest information. 
The input must be a JSON object with two fields:
- "objective": a string describing the research goal.
- "search_queries": an array of strings, e.g., ["query1", "query2"].
Example: {"objective": "Find new AI models", "search_queries": ["latest AI models", "AI announcements 2025"]}`

	search := genkit.DefineTool(g, "search", toolDescription, func(ctx *ai.ToolContext, input SearchInput) (string, error) {
		return search(input, ctx)
	})
	app := &application{echo: e, genkit: g, searchTool: search}
	app.generateFlow = newGenerateFlow(app)
	return app
}

func main() {
	godotenv.Load(".env")
	port := flag.Int("port", 4000, "Port on which to run the server")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)

	ctx := context.Background()
	e := echo.New()
	e.Use(middleware.CORS("http://localhost:5173"))
	g := genkit.Init(ctx, genkit.WithPlugins(&deepseek.DeepSeek{}))

	app := NewApplication(e, g)

	app.setupRoutes()

	app.echo.Logger.Info("Starting Server", "port", *port)
	if err := e.Start(addr); err != nil {
		e.Logger.Error("Failed to start server: ", "error", err)
	}
}
