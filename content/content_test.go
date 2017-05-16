package content

import "testing"

func TestFilterContent(t *testing.T) {
	content := []*Content{{ID: "0"}, {ID: "1"}, {ID: "2"}}

	content = filter(content, func(c *Content) bool {
		return c.ID == "1"
	})
	if len(content) != 1 {
		t.Error("Expected exactly one item")
	}
	if content[0].ID != "1" {
		t.Error("Filtered out incorrect item")
	}
}

func TestAnyTagFilter(t *testing.T) {
	content := []*Content{{Tags: []string{"t1", "t2", "t3"}}}

	filtered := filter(content, anyTagFilter(map[string]bool{"t4": true}))
	if len(filtered) > 0 {
		t.Error("Should not have found match")
	}

	filtered = filter(content, anyTagFilter(map[string]bool{"t2": true}))
	if len(filtered) != 1 {
		t.Errorf("Should have found exactly one match, but found %v", len(filtered))
	}
}

func TestAllTagFilter(t *testing.T) {
	content := []*Content{{Tags: []string{"t1", "t2", "t3"}}}

	filtered := filter(content, allTagFilter(map[string]bool{"t2": true, "t4": true}))
	if len(filtered) > 0 {
		t.Error("Should not have found match")
	}

	filtered = filter(content, allTagFilter(map[string]bool{"t1": true, "t2": true, "t3": true}))
	if len(filtered) != 1 {
		t.Errorf("Should have found exactly one match, but found %v", len(filtered))
	}
}

func TestTransformContent(t *testing.T) {
	content := []*Content{{ID: "0"}, {ID: "1"}, {ID: "2"}}

	content = transform(content, func(c Content) *Content {
		c.Title = "Transformed"
		return &c
	})
	if len(content) != 3 {
		t.Error("Length of array should not have changed")
	}
	for i := range content {
		if content[i].Title != "Transformed" {
			t.Errorf("Failed to transform item at index: %v", i)
		}
	}
}

func TestTransformContentIsThreadSafe(t *testing.T) {
	content := []*Content{{ID: "0"}, {ID: "1"}, {ID: "2"}}

	transform(content, func(c Content) *Content {
		c.ID = "Transformed"
		return &c
	})
	if len(content) != 3 {
		t.Error("Length of array should not have changed")
	}
	for i := range content {
		if content[i].ID == "Transformed" {
			t.Errorf("Tranform should NOT change original array")
		}
	}
}