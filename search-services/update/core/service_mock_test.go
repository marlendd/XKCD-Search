package core_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"yadro.com/course/update/core"
	"yadro.com/course/update/mocks"
)

func TestStats(t *testing.T) {
	cases := []struct{
		name string
		dbStats core.DBStats
		dbErr error
		lastID int
		lastIDErr error
		wantErr bool
		want core.ServiceStats
	}{
		{
			name: "ok",
			dbStats: core.DBStats{WordsTotal: 10, WordsUnique: 5, ComicsFetched: 3},
			lastID: 100,
			want: core.ServiceStats{DBStats: core.DBStats{
				WordsTotal: 10, WordsUnique: 5, ComicsFetched: 3}, ComicsTotal: 100},
		},
		{
			name: "db error",
			dbErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "xkcd error",
			dbStats: core.DBStats{WordsTotal: 10},
			lastIDErr: errors.New("xkcd error"),
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			mockDB := mocks.NewMockDB(c)
			mockXKCD := mocks.NewMockXKCD(c)

			mockDB.EXPECT().Stats(gomock.Any()).Return(tc.dbStats, tc.dbErr)

			if tc.dbErr == nil {
				mockXKCD.EXPECT().LastID(gomock.Any()).Return(tc.lastID, tc.lastIDErr)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			service, err := core.NewService(log, mockDB, mockXKCD, nil, 1, nil)
			require.NoError(t, err)

			result, err := service.Stats(context.Background())
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, result)
			}
		})
	}
}

func TestDrop(t *testing.T) {
	cases := []struct{
		name string
		errDB error
		errPub error
		wantErr bool
	}{
		{
			name: "ok",
			wantErr: false,
		},
		{
			name: "db error",
			errDB: errors.New("db error"),
			wantErr: true,
		},
		{
			name: "publisher error",
			errPub: errors.New("publisher error"),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			mockDB := mocks.NewMockDB(c)
			mockPub := mocks.NewMockPublisher(c)

			mockDB.EXPECT().Drop(gomock.Any()).Return(tc.errDB)
			if tc.errDB == nil {
				mockPub.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(tc.errPub)
			}

			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			service, err := core.NewService(log, mockDB, nil, nil, 1, mockPub)
			require.NoError(t, err)

			err = service.Drop(context.Background())
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdate_EmptyDb(t *testing.T) {
	c := gomock.NewController(t)

	mockDB := mocks.NewMockDB(c)
	mockXKCD := mocks.NewMockXKCD(c)
	mockWords := mocks.NewMockWords(c)
	mockPub := mocks.NewMockPublisher(c)

	mockDB.EXPECT().IDs(gomock.Any()).Return([]int{}, nil)

	mockXKCD.EXPECT().LastID(gomock.Any()).
	Return(3, nil)

	mockXKCD.EXPECT().Get(gomock.Any(), gomock.Any()).
	Return(core.XKCDInfo{URL: "http://testing.com"}, nil).Times(3)

	mockWords.EXPECT().Norm(gomock.Any(), gomock.Any()).
	Return([]string{"test"}, nil).Times(3)

	mockPub.EXPECT().Publish("xkcd.db.updated", gomock.Any()).Return(nil)

	mockDB.EXPECT().Add(gomock.Any(), gomock.Any()).
	Return(nil).Times(3)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := core.NewService(log, mockDB, mockXKCD, mockWords, 1, mockPub)
	require.NoError(t, err)

	err = service.Update(context.Background())
	require.NoError(t, err)
}

func TestUpdate_NotEmptyDb(t *testing.T) {
	c := gomock.NewController(t)

	mockDB := mocks.NewMockDB(c)
	mockXKCD := mocks.NewMockXKCD(c)
	mockWords := mocks.NewMockWords(c)
	mockPub := mocks.NewMockPublisher(c)

	mockDB.EXPECT().IDs(gomock.Any()).Return([]int{1,2}, nil)

	mockXKCD.EXPECT().LastID(gomock.Any()).
	Return(3, nil)

	mockXKCD.EXPECT().Get(gomock.Any(), gomock.Any()).
	Return(core.XKCDInfo{URL: "http://testing.com"}, nil)

	mockWords.EXPECT().Norm(gomock.Any(), gomock.Any()).
	Return([]string{"test"}, nil)

	mockPub.EXPECT().Publish("xkcd.db.updated", gomock.Any()).Return(nil)

	mockDB.EXPECT().Add(gomock.Any(), gomock.Any()).
	Return(nil)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := core.NewService(log, mockDB, mockXKCD, mockWords, 1, mockPub)
	require.NoError(t, err)

	err = service.Update(context.Background())
	require.NoError(t, err)
}

func TestUpdate_NotFound(t *testing.T) {
	c := gomock.NewController(t)

	mockDB := mocks.NewMockDB(c)
	mockXKCD := mocks.NewMockXKCD(c)
	mockWords := mocks.NewMockWords(c)
	mockPub := mocks.NewMockPublisher(c)

	mockDB.EXPECT().IDs(gomock.Any()).Return([]int{1,2}, nil)

	mockXKCD.EXPECT().LastID(gomock.Any()).
	Return(3, nil)

	mockXKCD.EXPECT().Get(gomock.Any(), gomock.Any()).
	Return(core.XKCDInfo{}, core.ErrNotFound)

	mockPub.EXPECT().Publish("xkcd.db.updated", gomock.Any()).Return(nil)

	mockDB.EXPECT().Add(gomock.Any(), gomock.Any()).
	Return(nil)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := core.NewService(log, mockDB, mockXKCD, mockWords, 1, mockPub)
	require.NoError(t, err)

	err = service.Update(context.Background())
	require.NoError(t, err)
}

func TestUpdate_GetError(t *testing.T) {
	c := gomock.NewController(t)

	mockDB := mocks.NewMockDB(c)
	mockXKCD := mocks.NewMockXKCD(c)

	mockDB.EXPECT().IDs(gomock.Any()).Return([]int{1,2}, nil)

	mockXKCD.EXPECT().LastID(gomock.Any()).
	Return(3, nil)

	mockXKCD.EXPECT().Get(gomock.Any(), gomock.Any()).
	Return(core.XKCDInfo{}, errors.New("xkcd error"))

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := core.NewService(log, mockDB, mockXKCD, nil, 1, nil)
	require.NoError(t, err)

	err = service.Update(context.Background())
	require.Error(t, err)
}

func TestUpdate_IdsError(t *testing.T) {
	c := gomock.NewController(t)

	mockDB := mocks.NewMockDB(c)

	mockDB.EXPECT().IDs(gomock.Any()).Return(nil, errors.New("db error"))

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := core.NewService(log, mockDB, nil, nil, 1, nil)
	require.NoError(t, err)

	err = service.Update(context.Background())
	require.Error(t, err)
}

func TestUpdate_WordsError(t *testing.T) {
	c := gomock.NewController(t)

	mockDB := mocks.NewMockDB(c)
	mockXKCD := mocks.NewMockXKCD(c)
	mockWords := mocks.NewMockWords(c)
	mockPub := mocks.NewMockPublisher(c)

	mockDB.EXPECT().IDs(gomock.Any()).Return([]int{1,2}, nil)

	mockXKCD.EXPECT().LastID(gomock.Any()).
	Return(3, nil)

	mockXKCD.EXPECT().Get(gomock.Any(), gomock.Any()).
	Return(core.XKCDInfo{URL: "http://testing.com"}, nil)

	mockWords.EXPECT().Norm(gomock.Any(), gomock.Any()).
	Return([]string{}, core.ErrBadArguments)

	mockPub.EXPECT().Publish("xkcd.db.updated", gomock.Any()).Return(nil)

	mockDB.EXPECT().Add(gomock.Any(), gomock.Any()).
	Return(nil)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := core.NewService(log, mockDB, mockXKCD, mockWords, 1, mockPub)
	require.NoError(t, err)

	err = service.Update(context.Background())
	require.NoError(t, err)
}

func TestUpdate_DoubleUpdate(t *testing.T) {
    c := gomock.NewController(t)
    mockDB := mocks.NewMockDB(c)
    mockXKCD := mocks.NewMockXKCD(c)
    mockWords := mocks.NewMockWords(c)
    mockPub := mocks.NewMockPublisher(c)

    mockDB.EXPECT().IDs(gomock.Any()).Return([]int{}, nil)
    mockXKCD.EXPECT().LastID(gomock.Any()).Return(0, nil)
    mockPub.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

    log := slog.New(slog.NewTextHandler(io.Discard, nil))
    service, err := core.NewService(log, mockDB, mockXKCD, mockWords, 1, mockPub)
    require.NoError(t, err)

    var err1, err2 error
    var wg sync.WaitGroup

    wg.Go(func() {err1 = service.Update(context.Background()) })
    wg.Go(func() {err2 = service.Update(context.Background()) })
    wg.Wait()

    results := []error{err1, err2}
    require.Contains(t, results, core.ErrAlreadyExists)
    require.Contains(t, results, nil)
}
