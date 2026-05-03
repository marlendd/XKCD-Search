package core

import "github.com/golang-jwt/jwt/v5"

type UpdateStatus string

const (
	StatusUpdateUnknown UpdateStatus = "unknown"
	StatusUpdateIdle    UpdateStatus = "idle"
	StatusUpdateRunning UpdateStatus = "running"
)

var jwtMethod jwt.SigningMethod = jwt.SigningMethodHS256

func GetJWTMethod() jwt.SigningMethod {
    return jwtMethod
}


type UpdateStats struct {
	WordsTotal    int
	WordsUnique   int
	ComicsFetched int
	ComicsTotal   int
}

type Comic struct {
	ID  int
	URL string
}

type SearchResult struct {
	Comics []Comic
}
