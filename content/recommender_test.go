package content

import (
	"testing"
)

func TestTagBasedRecommender(t *testing.T) {
	index := createIndexWithID("test")
	index.Add([]*Content{{ID: "0", Tags: []string{"t1", "t2"}}, {ID: "1", Tags: []string{"t3"}}})

	recommender := TagBasedRecommender{}

	recs, err := recommender.Recommend(index, map[string]interface{}{"tags": "t1 t3", "lang": ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 0 {
		t.Error("Should not have found a recommendation")
	}

	recs, err = recommender.Recommend(index, map[string]interface{}{"tags": "t1 t2", "lang": ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("Expected exactly one recommendation, but got %v", len(recs))
	}
	if recs[0].ID != "0" {
		t.Error("Expected different recommendation")
	}

	recs, err = recommender.Recommend(index, map[string]interface{}{"tags": "t4", "lang": ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 0 {
		t.Error("Should not have found a recommendation")
	}

	recs, err = recommender.Recommend(index, map[string]interface{}{"tags": "t1,tX", "lang": ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("Expected exactly one recommendation, but got %v", len(recs))
	}
	if recs[0].ID != "0" {
		t.Error("Expected different recommendation")
	}

	recs, err = recommender.Recommend(index, map[string]interface{}{"tags": "t1,t3", "lang": ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Errorf("Expected exactly two recommendations, but got %v", len(recs))
	}
}

func BenchmarkTagBasedRecommender(b *testing.B) {
	for i := 0; i < b.N; i++ {
		index := createIndexWithID("test")
		index.Add([]*Content{{ID: "0", Tags: []string{"t1", "t2"}}, {ID: "1", Tags: []string{"t3"}}})

		recommender := TagBasedRecommender{}

		recs, err := recommender.Recommend(index, map[string]interface{}{"tags": "t1 t3", "lang": ""})
		if err != nil {
			b.Fatal(err)
		}
		if len(recs) != 0 {
			b.Error("Should not have found a recommendation")
		}
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

func TestLocaleBasedRecommender(t *testing.T) {
	index := createIndexWithID("test")
	index.Add([]*Content{{ID: "1", Language: "de"}, {ID: "2", Language: "de", Regions: []string{"AT"}}})

	recommender := &LocaleBasedRecommender{}

	content, err := recommender.Recommend(index, map[string]interface{}{"locale": "en"})
	if err != nil {
		t.Fatal("Failed to compute recommendations for provider", err)
	}
	if len(content) != 0 {
		t.Errorf("Expected content of length 0, but got %v", len(content))
	}

	content, err = recommender.Recommend(index, map[string]interface{}{"locale": "de"})
	if err != nil {
		t.Fatal("Failed to compute recommendations for provider", err)
	}
	if len(content) != 1 {
		// Won't get de-AT content as it's restricted to AT
		t.Errorf("Expected content of length 1, but got %v", len(content))
	}
	if content[0].ID != "1" {
		t.Errorf("Expected content relevant to de, but got: %v", content[0])
	}

	content, err = recommender.Recommend(index, map[string]interface{}{"locale": "de-AT"})
	if err != nil {
		t.Fatal("Failed to compute recommendations for provider", err)
	}
	if len(content) != 2 {
		t.Errorf("Expected content of length 2, but got %v", len(content))
	}
	if content[0].ID != "1" {
		t.Errorf("Expected content relevant to de-AT, but got: %v", content[0])
	}
	if content[1].ID != "2" {
		// "de" content does not provide a region and is therefore relevant to all "de" regions incl. AT
		t.Errorf("Expected content relevant to de-AT, but got: %v", content[1])
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
