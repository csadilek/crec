package content

import (
	"testing"
)

func TestQueryBasedRecommender(t *testing.T) {
	content := []*Content{}
	var searchQuery = ""
	recommender := &QueryBasedRecommender{Search: func(q string) ([]*Content, error) {
		searchQuery = q
		return []*Content{}, nil
	}}

	recommender.Recommend(content, map[string]string{"query": "keyword"})
	if searchQuery != "keyword" {
		t.Errorf("Query function not invoked with provided query string")
	}
}

func TestTagBasedRecommender(t *testing.T) {
	content := []*Content{{ID: "0", Tags: []string{"t1", "t2"}}, {ID: "1", Tags: []string{"t3"}}}
	recommender := TagBasedRecommender{}

	recs, err := recommender.Recommend(content, map[string]string{"tags": "t1 t3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 0 {
		t.Error("Should not have found a recommendation")
	}

	recs, err = recommender.Recommend(content, map[string]string{"tags": "t1 t2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Errorf("Expected exactly one recommendation, but got %v", len(recs))
	}
	if recs[0].ID != "0" {
		t.Error("Expected different recommendation")
	}

	recs, err = recommender.Recommend(content, map[string]string{"tags": "t4"})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 0 {
		t.Error("Should not have found a recommendation")
	}

	recs, err = recommender.Recommend(content, map[string]string{"tags": "t1,t2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Errorf("Expected exactly one recommendation, but got %v", len(recs))
	}
	if recs[0].ID != "0" {
		t.Error("Expected different recommendation")
	}

	recs, err = recommender.Recommend(content, map[string]string{"tags": "t1,t3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Errorf("Expected exactly two recommendations, but got %v", len(recs))
	}
}
