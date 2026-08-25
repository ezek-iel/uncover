package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
)

func (app *application) setupRoutes() {
	app.echo.GET("/", hello)
	app.echo.POST("/generate", app.generateAIResponseHandler)
}

func hello(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"Hello": "World"})
}

func (app *application) generateAIResponseHandler(c *echo.Context) error {
	var req AIRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request"})
	}
	if req.Question == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "question is required"})
	}

	rsp := c.Response()
	rsp.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	rsp.Header().Set("Cache-Control", "no-cache")
	rsp.Header().Set("Connection", "keep-alive")
	rsp.Header().Set("X-Accel-Buffering", "no")
	rc := http.NewResponseController(rsp)

	// Cancel the flow when the client disconnects so we stop hitting the model.
	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	for v, err := range app.generateFlow.Stream(ctx, req) {
		if err != nil {
			c.Logger().Error("generate flow failed", "err", err)
			fmt.Fprint(rsp, (&EventMessage{eventType: "error", data: err.Error()}).String())
			rc.Flush()
			return nil
		}
		if v.Done {
			break
		}
		fmt.Fprint(rsp, v.Stream.String())
		if err := rc.Flush(); err != nil {
			return nil // client went away; ctx cancel stops the flow
		}
	}
	return nil
}
