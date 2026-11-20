package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/DimaChekashov/fuchs-budget-platform/backend/services/ledger-service/internal/domain"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Projections interface {
	Apply(context.Context, domain.Operation) error
	Reject(context.Context, string, int32, int64, string) error
}

var amountPattern = regexp.MustCompile(`^-?[0-9]{1,16}(\.[0-9]{1,2})?$`)

func decode(payload []byte) (domain.Operation, error) {
	var event domain.Operation
	if err := json.Unmarshal(payload, &event); err != nil {
		return event, errors.New("invalid JSON event")
	}
	if event.Version != 1 || event.ID == uuid.Nil || event.UserID == uuid.Nil ||
		event.WalletID == uuid.Nil || event.RecordedAt.IsZero() || utf8.RuneCountInString(event.Currency) > 10 ||
		!amountPattern.MatchString(event.Amount) ||
		(event.Type != "deposit" && event.Type != "withdrawal") {
		return event, errors.New("invalid wallet operation event")
	}
	return event, nil
}

func RunConsumer(ctx context.Context, brokers []string, topic string, store Projections) {
	for ctx.Err() == nil {
		if err := consume(ctx, brokers, topic, store); err != nil && ctx.Err() == nil {
			log.Printf("ledger consumer retrying: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func consume(ctx context.Context, brokers []string, topic string, store Projections) error {
	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...), kgo.ConsumeTopics(topic), kgo.ConsumerGroup("fuchs-ledger-v1"),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), kgo.DisableAutoCommit(), kgo.BlockRebalanceOnPoll())
	if err != nil {
		return err
	}
	defer client.CloseAllowingRebalance()

	for ctx.Err() == nil {
		fetched := client.PollRecords(ctx, 1)
		if err := fetched.Err(); err != nil {
			return err
		}
		records := fetched.RecordIter()
		if records.Done() {
			continue
		}
		record := records.Next()
		event, err := decode(record.Value)
		work, cancel := context.WithTimeout(ctx, 10*time.Second)
		if err != nil {
			err = store.Reject(work, record.Topic, record.Partition, record.Offset, err.Error())
		} else {
			err = store.Apply(work, event)
		}
		if err == nil {
			err = client.CommitRecords(work, record)
		}
		cancel()
		if err != nil {
			return fmt.Errorf("process offset %d: %w", record.Offset, err)
		}
		client.AllowRebalance()
	}
	return ctx.Err()
}
