package core

type Comics struct {
	ID    int
	URL   string
	Words []string
}

type SearchResult struct {
	Comics []Comics
	Total  int
}

type DBStats struct {
	WordsTotal    int
	WordsUnique   int
	ComicsFetched int
}

type Index map[string][]int
