package initiator

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mockinit "yadro.com/course/search/adapters/initiator/mock"
)

func TestNewInit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIndexer := mockinit.NewMockIndexer(ctrl)
	log := slog.Default()
	ttl := 1 * time.Minute

	initiator := NewInit(log, mockIndexer, ttl)

	assert.NotNil(t, initiator)
	assert.Equal(t, log, initiator.log)
	assert.Equal(t, mockIndexer, initiator.service)
	assert.Equal(t, ttl, initiator.ttl)
}

func TestInitiator_Start(t *testing.T) {
	t.Run("index building error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockIndexer := mockinit.NewMockIndexer(ctrl)
		mockIndexer.EXPECT().BuildIndex(gomock.Any()).Return(errors.New("build error")).Times(1)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		initiator := NewInit(slog.Default(), mockIndexer, 1*time.Hour)
		initiator.buildIndex(ctx)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockIndexer := mockinit.NewMockIndexer(ctrl)
		mockIndexer.EXPECT().BuildIndex(gomock.Any()).Return(nil).Times(1)

		ctx, cancel := context.WithCancel(context.Background())
		initiator := NewInit(slog.Default(), mockIndexer, 1*time.Hour)

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		initiator.Start(ctx)
	})
}

func TestInitiator_buildIndex(t *testing.T) {
	t.Run("successful build", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockIndexer := mockinit.NewMockIndexer(ctrl)
		mockIndexer.EXPECT().BuildIndex(gomock.Any()).Return(nil)

		initiator := NewInit(slog.Default(), mockIndexer, 1*time.Hour)
		initiator.buildIndex(context.Background())
	})

	t.Run("build error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockIndexer := mockinit.NewMockIndexer(ctrl)
		mockIndexer.EXPECT().BuildIndex(gomock.Any()).Return(errors.New("build error"))

		initiator := NewInit(slog.Default(), mockIndexer, 1*time.Hour)
		initiator.buildIndex(context.Background())
	})

	t.Run("context canceled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockIndexer := mockinit.NewMockIndexer(ctrl)
		mockIndexer.EXPECT().BuildIndex(gomock.Any()).Return(context.Canceled)

		initiator := NewInit(slog.Default(), mockIndexer, 1*time.Hour)
		initiator.buildIndex(context.Background())
	})
}
