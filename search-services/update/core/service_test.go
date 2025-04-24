package core_test

import (
	"context"
	"errors"
	"testing"
	"yadro.com/course/update/core"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	mocks "yadro.com/course/update/core/mock"
)

func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		concurrency int
		expectedErr string
	}{
		{
			name:        "successful creation",
			concurrency: 5,
		},
		{
			name:        "zero concurrency",
			concurrency: 0,
			expectedErr: "wrong concurrency specified: 0",
		},
		{
			name:        "negative concurrency",
			concurrency: -1,
			expectedErr: "wrong concurrency specified: -1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			mockXKCD := mocks.NewMockXKCD(ctrl)
			mockWords := mocks.NewMockWords(ctrl)

			service, err := core.NewService(nil, mockDB, mockXKCD, mockWords, tt.concurrency)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockDB, *mocks.MockXKCD, *mocks.MockWords)
		expectedErr string
	}{
		{
			name: "successful update with new comics",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords) {
				// Setup mocks
				xkcd.EXPECT().LastID(gomock.Any()).Return(3, nil)
				db.EXPECT().IDs(gomock.Any()).Return([]int{1}, nil)

				// Comics 2
				xkcd.EXPECT().Get(gomock.Any(), 2).Return(core.XKCDInfo{
					NUM:         2,
					URL:         "http://example.com/2",
					Title:       "Test 2",
					Description: "Description 2",
				}, nil)
				words.EXPECT().Norm(gomock.Any(), "Test 2 Description 2").Return([]string{"test", "two"}, nil)
				db.EXPECT().Add(gomock.Any(), core.Comics{
					ID:    2,
					URL:   "http://example.com/2",
					Words: []string{"test", "two"},
				}).Return(nil)

				// Comics 3
				xkcd.EXPECT().Get(gomock.Any(), 3).Return(core.XKCDInfo{
					NUM:         3,
					URL:         "http://example.com/3",
					Title:       "Test 3",
					Description: "Description 3",
				}, nil)
				words.EXPECT().Norm(gomock.Any(), "Test 3 Description 3").Return([]string{"test", "three"}, nil)
				db.EXPECT().Add(gomock.Any(), core.Comics{
					ID:    3,
					URL:   "http://example.com/3",
					Words: []string{"test", "three"},
				}).Return(nil)
			},
		},
		{
			name: "skip existing comics",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords) {
				xkcd.EXPECT().LastID(gomock.Any()).Return(2, nil)
				db.EXPECT().IDs(gomock.Any()).Return([]int{1, 2}, nil)
				// No calls to Get or Add expected
			},
		},
		{
			name: "handle not found comics",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords) {
				xkcd.EXPECT().LastID(gomock.Any()).Return(2, nil)
				db.EXPECT().IDs(gomock.Any()).Return([]int{1}, nil)
				xkcd.EXPECT().Get(gomock.Any(), 2).Return(core.XKCDInfo{}, core.ErrNotFound)
			},
		},
		{
			name: "error getting last ID",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords) {
				xkcd.EXPECT().LastID(gomock.Any()).Return(0, errors.New("last id error"))
				// No other calls expected
			},
			expectedErr: "failed to get last id: last id error",
		},
		{
			name: "error getting existing IDs",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords) {
				xkcd.EXPECT().LastID(gomock.Any()).Return(2, nil)
				db.EXPECT().IDs(gomock.Any()).Return(nil, errors.New("db error"))
				// No other calls expected
			},
			expectedErr: "failed to get existing IDs: db error",
		},
		{
			name: "error getting comic info",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords) {
				xkcd.EXPECT().LastID(gomock.Any()).Return(2, nil)
				db.EXPECT().IDs(gomock.Any()).Return([]int{1}, nil)
				xkcd.EXPECT().Get(gomock.Any(), 2).Return(core.XKCDInfo{}, errors.New("get error"))
			},
			expectedErr: "failed to get comics 2: get error",
		},
		{
			name: "error normalizing words",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords) {
				xkcd.EXPECT().LastID(gomock.Any()).Return(2, nil)
				db.EXPECT().IDs(gomock.Any()).Return([]int{1}, nil)
				xkcd.EXPECT().Get(gomock.Any(), 2).Return(core.XKCDInfo{
					NUM:         2,
					Title:       "Test",
					Description: "Desc",
				}, nil)
				words.EXPECT().Norm(gomock.Any(), "Test Desc").Return(nil, errors.New("norm error"))
			},
			expectedErr: "failed to normalize words for comics 2: norm error",
		},
		{
			name: "error adding to db",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD, words *mocks.MockWords) {
				xkcd.EXPECT().LastID(gomock.Any()).Return(2, nil)
				db.EXPECT().IDs(gomock.Any()).Return([]int{1}, nil)
				xkcd.EXPECT().Get(gomock.Any(), 2).Return(core.XKCDInfo{
					NUM:         2,
					URL:         "http://example.com/2",
					Title:       "Test",
					Description: "Desc",
				}, nil)
				words.EXPECT().Norm(gomock.Any(), "Test Desc").Return([]string{"test"}, nil)
				db.EXPECT().Add(gomock.Any(), core.Comics{
					ID:    2,
					URL:   "http://example.com/2",
					Words: []string{"test"},
				}).Return(errors.New("add error"))
			},
			expectedErr: "failed to add comics 2 to db: add error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			mockXKCD := mocks.NewMockXKCD(ctrl)
			mockWords := mocks.NewMockWords(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB, mockXKCD, mockWords)
			}

			service, err := core.NewService(nil, mockDB, mockXKCD, mockWords, 2)
			assert.NoError(t, err)

			err = service.Update(context.Background())

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Stats(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockDB, *mocks.MockXKCD)
		expected    core.ServiceStats
		expectedErr string
	}{
		{
			name: "successful stats",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD) {
				db.EXPECT().Stats(gomock.Any()).Return(core.DBStats{
					WordsTotal:    100,
					WordsUnique:   80,
					ComicsFetched: 10,
				}, nil)
				xkcd.EXPECT().LastID(gomock.Any()).Return(15, nil)
				xkcd.EXPECT().MissingIds(gomock.Any()).Return([]int{404, 405})
			},
			expected: core.ServiceStats{
				DBStats: core.DBStats{
					WordsTotal:    100,
					WordsUnique:   80,
					ComicsFetched: 10,
				},
				ComicsTotal: 13,
			},
		},
		{
			name: "error getting db stats",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD) {
				db.EXPECT().Stats(gomock.Any()).Return(core.DBStats{}, errors.New("db stats error"))
			},
			expectedErr: "failed to get db stats: db stats error",
		},
		{
			name: "error counting comics",
			mockSetup: func(db *mocks.MockDB, xkcd *mocks.MockXKCD) {
				db.EXPECT().Stats(gomock.Any()).Return(core.DBStats{
					WordsTotal:    100,
					WordsUnique:   80,
					ComicsFetched: 10,
				}, nil)
				xkcd.EXPECT().LastID(gomock.Any()).Return(0, errors.New("last id error"))
			},
			expectedErr: "failed to count comics: failed to get last ID: last id error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			mockXKCD := mocks.NewMockXKCD(ctrl)
			mockWords := mocks.NewMockWords(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB, mockXKCD)
			}

			service, err := core.NewService(nil, mockDB, mockXKCD, mockWords, 1)
			assert.NoError(t, err)

			stats, err := service.Stats(context.Background())

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, stats)
			}
		})
	}
}

func TestService_Status(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)
	mockXKCD := mocks.NewMockXKCD(ctrl)
	mockWords := mocks.NewMockWords(ctrl)

	service, err := core.NewService(nil, mockDB, mockXKCD, mockWords, 1)
	assert.NoError(t, err)

	assert.Equal(t, core.StatusIdle, service.Status(context.Background()))
}

func TestService_Drop(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockDB)
		expectedErr string
	}{
		{
			name: "successful drop",
			mockSetup: func(db *mocks.MockDB) {
				db.EXPECT().Drop(gomock.Any()).Return(nil)
			},
		},
		{
			name: "error dropping db",
			mockSetup: func(db *mocks.MockDB) {
				db.EXPECT().Drop(gomock.Any()).Return(errors.New("drop error"))
			},
			expectedErr: "failed to drop database: drop error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			mockXKCD := mocks.NewMockXKCD(ctrl)
			mockWords := mocks.NewMockWords(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockDB)
			}

			service, err := core.NewService(nil, mockDB, mockXKCD, mockWords, 1)
			assert.NoError(t, err)

			err = service.Drop(context.Background())

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Count(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockXKCD)
		expected    int
		expectedErr string
	}{
		{
			name: "successful count",
			mockSetup: func(xkcd *mocks.MockXKCD) {
				xkcd.EXPECT().LastID(gomock.Any()).Return(10, nil)
				xkcd.EXPECT().MissingIds(gomock.Any()).Return([]int{404, 405})
			},
			expected: 8,
		},
		{
			name: "error getting last ID",
			mockSetup: func(xkcd *mocks.MockXKCD) {
				xkcd.EXPECT().LastID(gomock.Any()).Return(0, errors.New("last id error"))
			},
			expectedErr: "failed to get last ID: last id error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockDB(ctrl)
			mockXKCD := mocks.NewMockXKCD(ctrl)
			mockWords := mocks.NewMockWords(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockXKCD)
			}

			service, err := core.NewService(nil, mockDB, mockXKCD, mockWords, 1)
			assert.NoError(t, err)

			count, err := service.Count(context.Background())

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, count)
			}
		})
	}
}
