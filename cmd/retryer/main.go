package main

import (
	"context"
	"github.com/Inspirate789/ds-lab2/internal/pkg/app"
	"github.com/Inspirate789/ds-lab2/pkg/retryer"
	"github.com/lmittmann/tint"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/pflag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var configPath string
	pflag.StringVarP(&configPath, "config", "c", "configs/retryer.yaml", "Config file path")
	pflag.Parse()

	config, err := app.ReadLocalConfig(configPath)
	if err != nil {
		panic(err)
	}

	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.Level(config.Logging.Level)}))

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Kafka.Addresses,
		Topic:   config.Kafka.Topic,
	})

	requestBacklog := retryer.NewKafkaRequestBacklog(kafkaReader, nil, logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	for {
		select {
		case <-quit:
			cancel()
			kafkaReader.Close()
			return
		default:
			err := requestBacklog.HandleRequest(ctx, func(req *http.Request) error {
				_, err := http.DefaultClient.Do(req)
				return err
			})
			if err != nil {
				logger.Error(err.Error())
			}
		}
	}
}
