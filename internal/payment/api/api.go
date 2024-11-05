package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"github.com/Inspirate789/ds-lab2/internal/payment/delivery"
	"github.com/Inspirate789/ds-lab2/internal/pkg/app"
	"github.com/pkg/errors"
	"github.com/sony/gobreaker/v2"
	"go.uber.org/multierr"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

const ErrServiceUnavailable = "Payment Service unavailable"

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type RequestBacklog interface {
	app.HealthChecker
	Push(ctx context.Context, req *http.Request) error
}

type payment struct {
	item  models.Payment
	found bool
}

type PaymentsAPI struct {
	baseURL   string
	client    *http.Client
	backlog   RequestBacklog
	paymentCB *gobreaker.CircuitBreaker[payment]
	logger    *slog.Logger
}

func New(baseURL string, client *http.Client, backlog RequestBacklog, maxFails uint, logger *slog.Logger) *PaymentsAPI {
	paymentCB := gobreaker.NewCircuitBreaker[payment](gobreaker.Settings{
		Name:        "get_payment",
		MaxRequests: uint32(maxFails),
		Timeout:     time.Second,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Debug("change circuit breaker state",
				slog.String("name", name),
				slog.String("from", from.String()),
				slog.String("to", to.String()),
			)
		},
	})

	return &PaymentsAPI{
		baseURL:   baseURL,
		client:    client,
		backlog:   backlog,
		paymentCB: paymentCB,
		logger:    logger,
	}
}

func (api *PaymentsAPI) HealthCheck(ctx context.Context) (err error) {
	defer func() {
		err = multierr.Append(err, api.backlog.HealthCheck(ctx))
	}()

	endpoint := api.baseURL + "/manage/health"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := api.client.Do(req)
	if err != nil {
		var DNSError *net.DNSError
		if errors.As(err, &DNSError) {
			err = errors.Wrap(err, ErrServiceUnavailable)
		}

		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	return nil

}

func (api *PaymentsAPI) CreatePayment(ctx context.Context, price uint64) (res models.Payment, err error) {
	endpoint := api.baseURL + "/api/v1/payments?price=" + strconv.FormatUint(price, 10)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return models.Payment{}, err
	}

	resp, err := api.client.Do(req)
	if err != nil {
		var DNSError *net.DNSError
		if errors.As(err, &DNSError) {
			err = errors.Wrap(err, ErrServiceUnavailable)
		}

		return models.Payment{}, multierr.Combine(err, api.backlog.Push(ctx, req))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Payment{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return models.Payment{}, errors.New(string(body))
	}

	var payment delivery.PaymentDTO

	err = json.Unmarshal(body, &payment)
	if err != nil {
		return models.Payment{}, err
	}

	return payment.ToModel(), nil
}

func (api *PaymentsAPI) SetPaymentStatus(ctx context.Context, paymentUID string, status models.PaymentStatus) (found bool, err error) {
	endpoint := api.baseURL + "/api/v1/payments/" + paymentUID + "/status"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewBufferString(fmt.Sprint(status)))
	if err != nil {
		return false, err
	}

	resp, err := api.client.Do(req)
	if err != nil {
		var DNSError *net.DNSError
		if errors.As(err, &DNSError) {
			err = nil
		}

		return true, multierr.Combine(err, api.backlog.Push(ctx, req))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if resp.StatusCode != http.StatusOK {
		return false, errors.New(string(body))
	}

	return true, nil
}

func (api *PaymentsAPI) getPayment(ctx context.Context, paymentUID string) (res models.Payment, found bool, err error) {
	endpoint := api.baseURL + "/api/v1/payments/" + paymentUID

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return models.Payment{}, false, err
	}

	resp, err := api.client.Do(req)
	if err != nil {
		var DNSError *net.DNSError
		if errors.As(err, &DNSError) {
			err = errors.Wrap(err, ErrServiceUnavailable)
		}

		return models.Payment{}, false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Payment{}, false, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return models.Payment{}, false, nil
	} else if resp.StatusCode != http.StatusOK {
		return models.Payment{}, false, errors.New(string(body))
	}

	var payment delivery.PaymentDTO

	err = json.Unmarshal(body, &payment)
	if err != nil {
		return models.Payment{}, false, err
	}

	return payment.ToModel(), true, nil
}

func (api *PaymentsAPI) GetPayment(ctx context.Context, paymentUID string) (models.Payment, bool, error) {
	res, err := api.paymentCB.Execute(func() (payment, error) {
		item, found, err := api.getPayment(ctx, paymentUID)
		return payment{
			item:  item,
			found: found,
		}, err
	})
	if err != nil {
		api.logger.Warn(err.Error())
		return models.Payment{}, true, nil
	}

	return res.item, res.found, nil
}
