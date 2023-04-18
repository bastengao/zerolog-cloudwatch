package writer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type event struct {
	Time int64
}

type Writer struct {
	client            *cloudwatchlogs.Client
	groupName         string
	streamName        string
	nextSequenceToken *string

	bufferSize    int
	flushInterval time.Duration

	ctx     context.Context
	cancel  func()
	msgChan chan []byte
	evt     event // reuse event
}

func New(cfg aws.Config, groupName, streamName string) *Writer {
	client := cloudwatchlogs.NewFromConfig(cfg)
	ctx, cancel := context.WithCancel(context.Background())

	writer := &Writer{
		client:     client,
		groupName:  groupName,
		streamName: streamName,

		bufferSize:    100,
		flushInterval: time.Second * 5,
		ctx:           ctx,
		cancel:        cancel,
		msgChan:       make(chan []byte, 10),
	}

	go writer.start()

	return writer
}

func (w *Writer) Write(p []byte) (int, error) {
	select {
	case <-w.ctx.Done():
		return 0, nil
	default:
	}

	// because zerolog will reuse the buffer, we need to copy it
	dup := make([]byte, len(p))
	copy(dup, p)
	w.msgChan <- dup
	return len(p), nil
}

func (w *Writer) Close() error {
	w.cancel()
	close(w.msgChan)
	return nil
}

func (w *Writer) send(buffer [][]byte) error {
	if len(buffer) == 0 {
		return nil
	}

	events := make([]types.InputLogEvent, 0, len(buffer))
	for _, b := range buffer {
		err := json.Unmarshal(b, &w.evt)
		if err != nil {
			fmt.Println(err)
		}
		var timestamp *int64
		if w.evt.Time != 0 {
			timestamp = &w.evt.Time
		} else {
			timestamp = aws.Int64(time.Now().UnixMilli())
		}

		events = append(events, types.InputLogEvent{
			Message:   aws.String(string(b)),
			Timestamp: timestamp,
		})
	}

	out, err := w.client.PutLogEvents(context.TODO(), &cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  &w.groupName,
		LogStreamName: &w.streamName,
		LogEvents:     events,
		SequenceToken: w.nextSequenceToken,
	})
	if err != nil {
		return err
	}

	w.nextSequenceToken = out.NextSequenceToken
	return nil
}

func (w *Writer) start() {
	ticker := time.NewTicker(w.flushInterval)

	var buffer [][]byte

	for {
		select {
		case <-ticker.C:
			err := w.send(buffer)
			if err != nil {
				fmt.Printf("send: %v\n", err)
			}
			buffer = buffer[:0]
		case m, ok := <-w.msgChan:
			if ok {
				buffer = append(buffer, m)
			} else {
				ticker.Stop()
				err := w.send(buffer)
				if err != nil {
					fmt.Printf("send: %v\n", err)
				}
				return
			}

			if len(buffer) >= w.bufferSize {
				err := w.send(buffer)
				if err != nil {
					fmt.Printf("send: %v\n", err)
				}
				buffer = buffer[:0]
			}
		}
	}
}
