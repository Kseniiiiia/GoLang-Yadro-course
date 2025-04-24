package core

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDB(ctrl)
	mockWords := NewMockWords(ctrl)
	logger := slog.Default()

	t.Run("successful creation", func(t *testing.T) {
		service, err := NewService(logger, mockDB, mockWords)
		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, mockDB, service.db)
		assert.Equal(t, mockWords, service.words)
	})
}

func TestService_Search(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDB(ctrl)
	mockWords := NewMockWords(ctrl)
	logger := slog.Default()
	service, _ := NewService(logger, mockDB, mockWords)

	t.Run("successful search", func(t *testing.T) {
		expectedWords := []string{"test", "phrase"}
		expectedComics := []Comics{
			{ID: 1, URL: "http://example.com/1"},
			{ID: 2, URL: "http://example.com/2"},
		}

		mockWords.EXPECT().
			Norm(gomock.Any(), "test phrase").
			Return(expectedWords, nil)

		mockDB.EXPECT().
			SearchComics(gomock.Any(), expectedWords, 10).
			Return(expectedComics, nil)

		result, err := service.Search(context.Background(), "test phrase", 10)
		assert.NoError(t, err)
		assert.Equal(t, expectedComics, result.Comics)
		assert.Equal(t, 2, result.Total)
	})

	t.Run("normalization error", func(t *testing.T) {
		mockWords.EXPECT().
			Norm(gomock.Any(), "error phrase").
			Return(nil, errors.New("normalization error"))

		_, err := service.Search(context.Background(), "error phrase", 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "normalization failed")
	})

	t.Run("db search error", func(t *testing.T) {
		mockWords.EXPECT().
			Norm(gomock.Any(), "db error").
			Return([]string{"test"}, nil)

		mockDB.EXPECT().
			SearchComics(gomock.Any(), []string{"test"}, 10).
			Return(nil, errors.New("db error"))

		_, err := service.Search(context.Background(), "db error", 10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db search failed")
	})
}

func TestService_IndexSearch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDB(ctrl)
	mockWords := NewMockWords(ctrl)
	logger := slog.Default()
	service, _ := NewService(logger, mockDB, mockWords)

	service.index = Index{
		"test":  []int{1, 2},
		"word":  []int{2, 3},
		"other": []int{4},
	}

	t.Run("successful index search", func(t *testing.T) {
		expectedWords := []string{"test", "word"}
		expectedComics := []Comics{
			{ID: 2, URL: "http://example.com/2", Words: []string{"test", "word"}},
			{ID: 1, URL: "http://example.com/1", Words: []string{"test"}},
			{ID: 3, URL: "http://example.com/3", Words: []string{"word"}},
		}

		mockWords.EXPECT().
			Norm(gomock.Any(), "test word").
			Return(expectedWords, nil)

		mockDB.EXPECT().
			GetComicsByIDs(gomock.Any(), []int{1, 2, 3}).
			Return(expectedComics, nil)

		result, err := service.IndexSearch(context.Background(), "test word", 10)
		assert.NoError(t, err)
		assert.Len(t, result.Comics, 3)
		assert.Equal(t, 3, result.Total)
	})

	t.Run("normalization error", func(t *testing.T) {
		mockWords.EXPECT().
			Norm(gomock.Any(), "error phrase").
			Return(nil, errors.New("normalization error"))

		_, err := service.IndexSearch(context.Background(), "error phrase", 10)
		assert.Error(t, err)
	})
}

func TestService_BuildIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDB(ctrl)
	mockWords := NewMockWords(ctrl)
	logger := slog.Default()
	service, _ := NewService(logger, mockDB, mockWords)

	t.Run("successful build", func(t *testing.T) {
		comics := []Comics{
			{ID: 1, Words: []string{"test", "one"}},
			{ID: 2, Words: []string{"test", "two"}},
			{ID: 3, Words: []string{"three"}},
		}

		mockDB.EXPECT().
			AllComics(gomock.Any()).
			Return(comics, nil)

		err := service.BuildIndex(context.Background())
		assert.NoError(t, err)

		index := service.GetIndex(context.Background())
		assert.Len(t, index["test"], 2)
		assert.Len(t, index["one"], 1)
		assert.Len(t, index["two"], 1)
		assert.Len(t, index["three"], 1)
	})

	t.Run("db error", func(t *testing.T) {
		mockDB.EXPECT().
			AllComics(gomock.Any()).
			Return(nil, errors.New("db error"))

		err := service.BuildIndex(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get comics")
	})
}

func TestService_Stats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockDB(ctrl)
	mockWords := NewMockWords(ctrl)
	logger := slog.Default()
	service, _ := NewService(logger, mockDB, mockWords)

	t.Run("successful stats", func(t *testing.T) {
		expectedStats := DBStats{
			WordsTotal:    100,
			WordsUnique:   50,
			ComicsFetched: 10,
		}

		mockDB.EXPECT().
			Stats(gomock.Any()).
			Return(expectedStats, nil)

		stats, err := service.Stats(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedStats, stats)
	})

	t.Run("db error", func(t *testing.T) {
		mockDB.EXPECT().
			Stats(gomock.Any()).
			Return(DBStats{}, errors.New("db error"))

		_, err := service.Stats(context.Background())
		assert.Error(t, err)
	})
}
