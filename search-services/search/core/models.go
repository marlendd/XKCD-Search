package core

type Comic struct {
	ID  int
	URL string
}

type StoredComic struct {
	ID  int
	URL string
	Words []string
}

type SearchResult struct {
	Comics []Comic
}
