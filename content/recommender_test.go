package content

import (
	"testing"

	"golang.org/x/text/language"
)

func TestTagBasedRecommender(t *testing.T) {
	index := createIndexWithID("test")
	index.Add([]*Content{{ID: "0", Tags: []string{"t1", "t2"}}, {ID: "1", Tags: []string{"t3"}}})

	recommender := TagBasedRecommender{}

	recs, err := recommender.Recommend(index, map[string]interface{}{"tags": "t1 t3", "lang-tags": []language.Tag{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 0 {
		t.Error("Should not have found a recommendation")
	}

	recs, err = recommender.Recommend(index, map[string]interface{}{"tags": "t1 t2", "lang-tags": []language.Tag{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("Expected exactly one recommendation, but got %v", len(recs))
	}
	if recs[0].ID != "0" {
		t.Error("Expected different recommendation")
	}

	recs, err = recommender.Recommend(index, map[string]interface{}{"tags": "t4", "lang-tags": []language.Tag{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 0 {
		t.Error("Should not have found a recommendation")
	}

	recs, err = recommender.Recommend(index, map[string]interface{}{"tags": "t1,tX", "lang-tags": []language.Tag{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("Expected exactly one recommendation, but got %v", len(recs))
	}
	if recs[0].ID != "0" {
		t.Error("Expected different recommendation")
	}

	recs, err = recommender.Recommend(index, map[string]interface{}{"tags": "t1,t3", "lang-tags": []language.Tag{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Errorf("Expected exactly two recommendations, but got %v", len(recs))
	}
}

func TestProviderBasedRecommender(t *testing.T) {
	index := createIndexWithID("test")
	index.Add([]*Content{{ID: "1", Source: "p1"}, {ID: "2", Source: "p2"}})
	recommender := &ProviderBasedRecommender{}

	content, err := recommender.Recommend(index, map[string]interface{}{"provider": "p1"})
	if err != nil {
		t.Fatal("Failed to compute recommendations for provider", err)
	}
	if len(content) != 1 {
		t.Errorf("Expected content of length 1, but got %v", len(content))
	}
	if content[0].ID != "1" {
		t.Errorf("Expected content from provider p1 but got: %v", content[0])
	}
}

func TestQueryBasedRecommender(t *testing.T) {
	index := createIndexWithID("test")
	recommender := &QueryBasedRecommender{}
	_, err := recommender.Recommend(index, map[string]interface{}{"query": "keyword"})
	if err != nil {
		t.Fatal("Failed to compute recommendation using QueryBasedRecommender: ", err)
	}
}
