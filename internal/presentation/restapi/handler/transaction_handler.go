package handler

import (
	"errors"
	"fmt"
	"frauddetection/internal/application"
	"frauddetection/internal/domain"
	orm "frauddetection/internal/infrastructure/postgresql"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Payment interface {
	ProcessPayment(c echo.Context) error
	GetPaymentByID(c echo.Context) error
	GetAllPayments(c echo.Context) error
}

type PaymentServiceHandler struct {
	paymentServiceApp application.PaymentServiceApp
}

func NewPaymentServiceHandler(
	paymentServiceApp application.PaymentServiceApp,
) Payment {
	return &PaymentServiceHandler{
		paymentServiceApp: paymentServiceApp,
	}
}

func (p *PaymentServiceHandler) ProcessPayment(c echo.Context) error {
	r := new(domain.Transaction)
	if err := c.Bind(r); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Invalid payload: %v", err),
		})
	}

	if r.UserID == "" || r.Amount <= 0 || r.Currency == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "userId, amount, and currency are required",
		})
	}
	payment, err := p.paymentServiceApp.CreateAndQueuePayment(
		c.Request().Context(),
		r.UserID,
		r.Amount,
		r.Currency,
	)

	if err != nil {
		if errors.Is(err, application.ErrRateLimitExceeded) {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":       "Error occurred creating payment",
			"err_message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, payment)
}

func (p *PaymentServiceHandler) GetPaymentByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "A valid payment Id is required",
		})
	}

	payment, err := p.paymentServiceApp.GetPaymentByID(
		c.Request().Context(),
		id,
	)

	if err != nil {
		if errors.Is(err, application.ErrRateLimitExceeded) {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": err.Error(),
			})
		}

		if errors.Is(err, orm.ErrPaymentNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":       "Error occurred getting payment",
			"err_message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, payment)
}

func (p *PaymentServiceHandler) GetAllPayments(c echo.Context) error {
	payments, err := p.paymentServiceApp.GetAllPayments(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":       "Error occurred getting payments",
			"err_message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, payments)
}
