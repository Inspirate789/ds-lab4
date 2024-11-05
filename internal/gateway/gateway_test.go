package gateway_test

import (
	"context"
	"github.com/Inspirate789/ds-lab2/internal/gateway"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"log/slog"
	"os"
	"testing"
)

type GatewaySuite struct {
	suite.Suite
}

func (s *GatewaySuite) TestHealthCheck(t provider.T) {
	t.Epic("MVP")
	t.Severity(allure.NORMAL)

	// arrange
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	carsAPI := new(carsApiMock)
	rentalAPI := new(rentalApiMock)
	paymentAPI := new(paymentApiMock)

	carsAPI.On("HealthCheck", ctx).Return(nil)
	rentalAPI.On("HealthCheck", ctx).Return(nil)
	paymentAPI.On("HealthCheck", ctx).Return(nil)

	g := gateway.New(carsAPI, rentalAPI, paymentAPI, logger)
	// act
	err := g.HealthCheck(ctx)
	// assert
	t.Require().NoError(err)
	carsAPI.AssertExpectations(t)
	rentalAPI.AssertExpectations(t)
	paymentAPI.AssertExpectations(t)
	carsAPI.AssertNumberOfCalls(t, "HealthCheck", 1)
	rentalAPI.AssertNumberOfCalls(t, "HealthCheck", 1)
	paymentAPI.AssertNumberOfCalls(t, "HealthCheck", 1)
}

func TestUseCase(t *testing.T) {
	t.Parallel()

	suite.RunSuite(t, new(GatewaySuite))
}
