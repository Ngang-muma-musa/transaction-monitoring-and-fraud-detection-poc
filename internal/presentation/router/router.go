package router

import (
	"frauddetection/internal/presentation/restapi/handler"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/labstack/echo/v4"
)

type Router interface {
	Register(e *echo.Echo) error
}

type router struct {
	paymentServiceHandler handler.Payment
}

func NewRouter(paymentServiceHandler handler.Payment) Router {
	return &router{
		paymentServiceHandler: paymentServiceHandler,
	}
}

func (r *router) Register(e *echo.Echo) error {
	r.registerAPI(e)
	return nil
}

func (r *router) registerAPI(e *echo.Echo) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "OK",
		})
	})

	e.POST("/payments", r.paymentServiceHandler.ProcessPayment)
	e.GET("/payments/:id", r.paymentServiceHandler.GetPaymentByID)
	e.GET("/payments", r.paymentServiceHandler.GetAllPayments)

	// Serve SPA index at root
	indexPath := filepath.Join(projectRoot(), "web", "index.html")
	e.File("/", indexPath)

	// Serve static files from web directory
	e.Static("/", filepath.Join(projectRoot(), "web"))

	// SPA fallback: if file not found, return index.html so client-side routing works
	e.GET("/*", func(c echo.Context) error {
		reqPath := c.Param("*")
		tryPath := filepath.Join(projectRoot(), "web", reqPath)
		if reqPath == "" {
			return c.File(indexPath)
		}
		if _, err := os.Stat(tryPath); err == nil {
			return c.File(tryPath)
		}
		return c.File(indexPath)
	})
}

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..")
}
