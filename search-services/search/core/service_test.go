package core_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"yadro.com/course/search/core"
	"yadro.com/course/search/mocks"
)

func TestSearch_Search(t *testing.T) {
	c := gomock.NewController(t)

	mockWords := mocks.NewMockWords(c)
	mockDB := mocks.NewMockDB(c)
	
	mockWords.EXPECT().Norm(gomock.Any(), gomock.Any()).
	Return([]string{}, nil)

	mockDB.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any()).
	Return([]core.Comic{}, nil)
	
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := core.NewService(log, mockDB, mockWords)
	_, err := service.Search(context.Background(), "", 1)
	require.NoError(t, err)
}

func TestSearch_SearchWordsErr(t *testing.T) {
	c := gomock.NewController(t)

	mockWords := mocks.NewMockWords(c)
	
	mockWords.EXPECT().Norm(gomock.Any(), gomock.Any()).
	Return([]string{}, errors.New("words error"))
	
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := core.NewService(log, nil, mockWords)
	
	result, err := service.Search(context.Background(), "", 1)
	require.Equal(t, core.SearchResult{}, result)
	require.Error(t, err)
}

func TestSearch_SearchDbErr(t *testing.T) {
	c := gomock.NewController(t)

	mockWords := mocks.NewMockWords(c)
	mockDB := mocks.NewMockDB(c)
	
	mockWords.EXPECT().Norm(gomock.Any(), gomock.Any()).
	Return([]string{}, nil)

	mockDB.EXPECT().Search(gomock.Any(), gomock.Any(), gomock.Any()).
	Return([]core.Comic{}, errors.New("db error"))
	
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := core.NewService(log, mockDB, mockWords)
	_, err := service.Search(context.Background(), "", 1)
	require.Error(t, err)
}

func TestSearch_ISearch(t *testing.T) {
	c := gomock.NewController(t)

	mockWords := mocks.NewMockWords(c)
	
	mockWords.EXPECT().Norm(gomock.Any(), gomock.Any()).
	Return([]string{}, nil)
	
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := core.NewService(log, nil, mockWords)
	_, err := service.ISearch(context.Background(), "", 1)
	require.NoError(t, err)
}

func TestSearch_ISearchWithIndex(t *testing.T) {
    c := gomock.NewController(t)
    mockDB := mocks.NewMockDB(c)
    mockWords := mocks.NewMockWords(c)

    mockDB.EXPECT().AllComics(gomock.Any()).Return([]core.StoredComic{
        {ID: 1, URL: "http://a.com", Words: []string{"linux"}},
        {ID: 2, URL: "http://b.com", Words: []string{"linux", "cpu"}},
    }, nil)
    mockWords.EXPECT().Norm(gomock.Any(), gomock.Any()).Return([]string{"linux"}, nil)

    log := slog.New(slog.NewTextHandler(io.Discard, nil))
    service := core.NewService(log, mockDB, mockWords)

    err := service.BuildIndex(context.Background())
    require.NoError(t, err)

    result, err := service.ISearch(context.Background(), "linux", 10)
    require.NoError(t, err)
    require.Len(t, result.Comics, 2)
}


func TestSearch_ISearchWordsErr(t *testing.T) {
	c := gomock.NewController(t)

	mockWords := mocks.NewMockWords(c)
	
	mockWords.EXPECT().Norm(gomock.Any(), gomock.Any()).
	Return([]string{}, errors.New("words error"))
	
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := core.NewService(log, nil, mockWords)
	
	result, err := service.ISearch(context.Background(), "", 1)
	require.Equal(t, core.SearchResult{}, result)
	require.Error(t, err)
}

func TestSearch_BuildIndex(t *testing.T) {
	c := gomock.NewController(t)

	mockDB := mocks.NewMockDB(c)

	mockDB.EXPECT().AllComics(gomock.Any()).
	Return([]core.StoredComic{}, nil)
	
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := core.NewService(log, mockDB, nil)
	err := service.BuildIndex(context.Background())
	require.NoError(t, err)
}

func TestSearch_BuildIndexDbErr(t *testing.T) {
	c := gomock.NewController(t)

	mockDB := mocks.NewMockDB(c)

	mockDB.EXPECT().AllComics(gomock.Any()).
	Return([]core.StoredComic{}, errors.New("db error"))
	
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := core.NewService(log, mockDB, nil)
	err := service.BuildIndex(context.Background())
	require.Error(t, err)
}