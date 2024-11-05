package retryer

import (
	"bytes"
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type BacklogError string

func (e BacklogError) Error() string {
	return string(e)
}

const (
	ErrNoWriter BacklogError = "backlog has no writer"
	ErrNoReader BacklogError = "backlog has no reader"
)

type KafkaRequestBacklog struct {
	reader *kafka.Reader
	writer *kafka.Writer
	logger *slog.Logger
}

func NewKafkaRequestBacklog(reader *kafka.Reader, writer *kafka.Writer, logger *slog.Logger) *KafkaRequestBacklog {
	return &KafkaRequestBacklog{
		reader: reader,
		writer: writer,
		logger: logger,
	}
}

type header struct {
	Key    string
	Values []string
}

type request struct {
	Method  string
	URL     string
	Body    []byte
	Headers []header
}

func (backlog *KafkaRequestBacklog) HealthCheck(_ context.Context) error {
	return nil // TODO ?
}

func (backlog *KafkaRequestBacklog) Push(ctx context.Context, req *http.Request) error {
	if backlog.writer == nil {
		return ErrNoWriter
	}

	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return err
		}
	}

	headers := make([]header, 0, len(req.Header))
	for key, values := range req.Header {
		headers = append(headers, header{Key: key, Values: values})
	}

	payload, err := kafka.Marshal(request{
		Method:  req.Method,
		URL:     req.URL.String(),
		Body:    body,
		Headers: headers,
	})
	if err != nil {
		return err
	}

	uid := uuid.New().String()
	msg := kafka.Message{
		Key:   []byte(uid),
		Value: payload,
	}
	backlog.logger.Debug("write message to kafka...",
		slog.String("topic", backlog.writer.Topic),
		slog.String("key", uid),
	)

	err = backlog.writer.WriteMessages(ctx, msg)
	if errors.Is(err, kafka.UnknownTopicOrPartition) {
		time.Sleep(5 * time.Second) // Wait for auto creating topic
		err = backlog.writer.WriteMessages(ctx, msg)
	}

	return err
}

func (backlog *KafkaRequestBacklog) HandleRequest(ctx context.Context, do func(*http.Request) error) (err error) {
	if backlog.reader == nil {
		return ErrNoReader
	}

	msg, err := backlog.reader.ReadMessage(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = multierror.Append(err, backlog.reader.SetOffset(msg.Offset))
		}
	}()

	backlog.logger.Debug("read message from kafka",
		slog.String("topic", msg.Topic),
		slog.Int("partition", msg.Partition),
		slog.Int64("offset", msg.Offset),
		slog.String("key", string(msg.Key)),
	)

	var rawRequest request
	err = kafka.Unmarshal(msg.Value, &rawRequest)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, rawRequest.Method, rawRequest.URL, bytes.NewBuffer(rawRequest.Body))
	if err != nil {
		return err
	}

	req.Header = make(http.Header)
	for _, h := range rawRequest.Headers {
		req.Header.Set(h.Key, h.Values[0])
	}

	return do(req)
}
