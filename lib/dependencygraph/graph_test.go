package dependencygraph

import "testing"

func TestWalk(t *testing.T) {
	expected_tree := map[string][]string{
		"test_tree/a.js":                 {"foo"},
		"test_tree/b.ts":                 {"foo"},
		"test_tree/util/c.js":            {"lodash", "express", "test_tree/b", "test_tree/util/fake_url/printFunc"},
		"test_tree/src/components/d.jsx": {"react", "@remix-run/react"},
		"test_tree/src/components/e.tsx": {},
		"test_tree/src/components/f.tsx": {"react", "dynamic_data"},
		"test_tree/src/hooks/g.ts":       {},
		"test_tree/src/hooks/h.ts":       {"test_tree/src/hooks/g", "test_tree/a", "test_tree/src/hooks/spurious_imports.txt"},
	}

	dgraph := New()

	tree, err := dgraph.Walk()

	if err != nil {
		t.Fatalf("Expected no error. Got: %s\n", err)
	}

	for k, v := range tree {
		expected, ok := expected_tree[k]
		if !ok {
			t.Errorf("Unexpected path in dependency tree. Got %s\n", k)
		}

		if len(expected) != len(v) {
			t.Fatalf("Wrong number of items in adjaceny list for %q. Got %d. Expected %d\n", k, len(v), len(expected))
		}

		for i, p := range v {
			if expected[i] != p {
				t.Errorf("Expected import path at index %d to be %q. Got %q\n", i, expected[i], p)
			}
		}
	}
}
