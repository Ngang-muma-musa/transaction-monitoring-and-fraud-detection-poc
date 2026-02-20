package router

import (
	"frauddetection/internal/presentation/restapi/handler"
	"net/http"

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
}
