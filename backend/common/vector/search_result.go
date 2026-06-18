package vector

import "sort"

type SearchResult struct {
	ID       int64   `json:"id"`
	Score    float32 `json:"score"`
	LawID    int64   `json:"lawId"`
	Content  string  `json:"content"`
	Keywords string  `json:"keywords"`
	VectorID string  `json:"vectorId"`
	Distance float32 `json:"distance"`
}

type SearchResultList []*SearchResult

func (s SearchResultList) Len() int           { return len(s) }

func (s SearchResultList) Less(i, j int) bool {
	return s[i].Score > s[j].Score
}

func (s SearchResultList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func SortByScore(results []*SearchResult) []*SearchResult {
	sort.Sort(SearchResultList(results))
	return results
}

func FilterByThreshold(results []*SearchResult, threshold float32) []*SearchResult {
	filtered := make([]*SearchResult, 0, len(results))
	for _, r := range results {
		if r.Score >= threshold {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func TopKResults(results []*SearchResult, k int) []*SearchResult {
	sorted := SortByScore(results)
	if k <= 0 || k >= len(sorted) {
		return sorted
	}
	return sorted[:k]
}
