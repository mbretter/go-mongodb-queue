package queue

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

func TestQueue_Publish(t *testing.T) {
	setNowFunc(func() time.Time {
		t, _ := time.Parse(time.DateTime, "2023-11-12 15:04:05")
		return t
	})

	tests := []struct {
		name     string
		topic    string
		payload  any
		maxTries uint
		error    error
	}{
		{
			name:     "Success",
			topic:    "topic1",
			payload:  "payload1",
			maxTries: 3,
		},
		{
			name:     "Error",
			topic:    "topic2",
			payload:  "payload2",
			maxTries: 0,
			error:    errors.New("db insert failed"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock := NewDbInterfaceMock(t)
			q := NewQueue(dbMock)

			oId := primitive.NewObjectID()

			taskExpected := Task{
				Topic:    tt.topic,
				Payload:  tt.payload,
				Tries:    0,
				MaxTries: 3,
				Meta: Meta{
					Created: nowFunc(),
				},
				State: StatePending,
			}
			dbMock.EXPECT().InsertOne(taskExpected).Return(oId, tt.error)

			task, err := q.Publish(tt.topic, tt.payload, tt.maxTries)

			if tt.error == nil {
				taskExpected.Id = oId
				assert.Equal(t, taskExpected, *task)
			} else {
				assert.Nil(t, task)
				assert.Equal(t, tt.error, err)
			}

		})
	}
}
