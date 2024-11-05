package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Inspirate789/ds-lab2/internal/models"
	"github.com/Inspirate789/ds-lab2/internal/pkg/app"
	"github.com/Inspirate789/ds-lab2/internal/rental/delivery"
	"github.com/pkg/errors"
	"github.com/sony/gobreaker/v2"
	"go.uber.org/multierr"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const ErrServiceUnavailable = "Rental Service unavailable"

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type RequestBacklog interface {
	app.HealthChecker
	Push(ctx context.Context, req *http.Request) error
}

type rentals struct {
	items      []models.Rental
	totalCount uint64
}

type rental struct {
	item      models.Rental
	found     bool
	permitted bool
}

type RentalsAPI struct {
	baseURL   string
	client    *http.Client
	backlog   RequestBacklog
	rentalsCB *gobreaker.CircuitBreaker[rentals]
	rentalCB  *gobreaker.CircuitBreaker[rental]
	logger    *slog.Logger
}

func New(baseURL string, client *http.Client, backlog RequestBacklog, maxFails uint, logger *slog.Logger) *RentalsAPI {
	logCB := func(name string, from gobreaker.State, to gobreaker.State) {
		logger.Debug("change circuit breaker state",
			slog.String("name", name),
			slog.String("from", from.String()),
			slog.String("to", to.String()),
		)
	}

	rentalsCB := gobreaker.NewCircuitBreaker[rentals](gobreaker.Settings{
		Name:          "get_rentals",
		MaxRequests:   uint32(maxFails),
		Timeout:       time.Second,
		OnStateChange: logCB,
	})

	rentalCB := gobreaker.NewCircuitBreaker[rental](gobreaker.Settings{
		Name:          "get_rental",
		MaxRequests:   uint32(maxFails),
		Timeout:       time.Second,
		OnStateChange: logCB,
	})

	return &RentalsAPI{
		baseURL:   baseURL,
		client:    client,
		backlog:   backlog,
		rentalsCB: rentalsCB,
		rentalCB:  rentalCB,
		logger:    logger,
	}
}

func (api *RentalsAPI) HealthCheck(ctx context.Context) (err error) {
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

func (api *RentalsAPI) getUserRentals(ctx context.Context, username string, offset, limit uint64) ([]models.Rental, uint64, error) {
	endpoint := api.baseURL + fmt.Sprintf("/api/v1/rentals?offset=%d&limit=%d", offset, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("X-User-Name", username)

	resp, err := api.client.Do(req)
	if err != nil {
		var DNSError *net.DNSError
		if errors.As(err, &DNSError) {
			err = errors.Wrap(err, ErrServiceUnavailable)
		}

		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, 0, errors.New(string(body))
	}

	var rentals delivery.RentalsDTO

	err = json.Unmarshal(body, &rentals)
	if err != nil {
		return nil, 0, err
	}

	model, err := rentals.ToModel()
	if err != nil {
		return nil, 0, err
	}

	return model, rentals.Count, nil
}

func (api *RentalsAPI) GetUserRentals(ctx context.Context, username string, offset, limit uint64) ([]models.Rental, uint64, error) {
	res, err := api.rentalsCB.Execute(func() (rentals, error) {
		items, totalCount, err := api.getUserRentals(ctx, username, offset, limit)
		return rentals{
			items:      items,
			totalCount: totalCount,
		}, err
	})
	if err != nil {
		api.logger.Warn(err.Error())
		return make([]models.Rental, 0), 0, nil
	}

	return res.items, res.totalCount, nil
}

func (api *RentalsAPI) getUserRental(ctx context.Context, rentalUID, username string) (res models.Rental, found, permitted bool, err error) {
	endpoint := api.baseURL + "/api/v1/rentals/" + rentalUID

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return models.Rental{}, false, false, err
	}

	req.Header.Set("X-User-Name", username)

	resp, err := api.client.Do(req)
	if err != nil {
		var DNSError *net.DNSError
		if errors.As(err, &DNSError) {
			err = errors.Wrap(err, ErrServiceUnavailable)
		}

		return models.Rental{}, false, false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.Rental{}, false, false, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return models.Rental{}, false, false, nil
	} else if resp.StatusCode == http.StatusForbidden {
		return models.Rental{}, true, false, nil
	} else if resp.StatusCode != http.StatusOK {
		return models.Rental{}, false, false, errors.New(string(body))
	}

	var rental delivery.RentalDTO

	err = json.Unmarshal(body, &rental)
	if err != nil {
		return models.Rental{}, true, true, err
	}

	model, err := rental.ToModel()
	if err != nil {
		return models.Rental{}, true, true, err
	}

	return model, true, true, nil
}

func (api *RentalsAPI) GetUserRental(ctx context.Context, rentalUID, username string) (models.Rental, bool, bool, error) {
	res, err := api.rentalCB.Execute(func() (rental, error) {
		item, found, permitted, err := api.getUserRental(ctx, rentalUID, username)
		return rental{
			item:      item,
			found:     found,
			permitted: permitted,
		}, err
	})
	if err != nil {
		api.logger.Warn(err.Error())
		return models.Rental{
			RentalUID:        rentalUID,
			RentalProperties: models.RentalProperties{Username: username},
		}, true, false, nil
	}

	return res.item, res.found, res.permitted, nil
}

func (api *RentalsAPI) CreateRental(ctx context.Context, properties models.RentalProperties) (models.Rental, error) {
	endpoint := api.baseURL + "/api/v1/rentals"
	dto := delivery.NewRentalPropertiesDTO(properties)

	body, err := json.Marshal(dto)
	if err != nil {
		return models.Rental{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return models.Rental{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		var DNSError *net.DNSError
		if errors.As(err, &DNSError) {
			err = errors.Wrap(err, ErrServiceUnavailable)
		}

		return models.Rental{}, multierr.Combine(err, api.backlog.Push(ctx, req))
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return models.Rental{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return models.Rental{}, errors.New(string(body))
	}

	var rental delivery.RentalDTO

	err = json.Unmarshal(body, &rental)
	if err != nil {
		return models.Rental{}, err
	}

	model, err := rental.ToModel()
	if err != nil {
		return models.Rental{}, err
	}

	return model, nil
}

func (api *RentalsAPI) SetRentalStatus(ctx context.Context, rentalUID string, status models.RentalStatus) (found bool, err error) {
	endpoint := api.baseURL + "/api/v1/rentals/" + rentalUID + "/status"

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewBufferString(fmt.Sprint(status)))
	if err != nil {
		return false, err
	}

	resp, err := api.client.Do(req)
	if err != nil {
		var DNSError *net.DNSError
		if errors.As(err, &DNSError) {
			err = errors.Wrap(err, ErrServiceUnavailable)
		}

		return false, multierr.Combine(err, api.backlog.Push(ctx, req))
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
