package gqlgen

import (
	"fmt"

	"github.com/vektah/gqlparser/ast"
	"testing"
)

func checkStringSlicesEqual(expected, results []string) string {
	if len(expected) != len(results) {
		return fmt.Sprintf("Slice lengths are not equal. Expected: %v, but got %v", len(expected), len(results))
	}

	for i := range expected {
		if expected[i] != results[i] {
			return fmt.Sprintf("Slices are not equal at index %v. Expected: %v, but got %v", i, expected[i], results[i])
		}
	}

	return ""
}

func TestGetRequestFieldsSingle(t *testing.T) {

	ctx := testContext(ast.SelectionSet{
		&ast.Field{
			Name:  "TF1",
			Alias: "TF1A",
		},
	})

	results := GetRequestFields(ctx)
	expected := []string{"TF1A"}

	msg := checkStringSlicesEqual(expected, results)

	if msg != "" {
		t.Errorf(msg)
		return
	}
}

func TestGetRequestFieldsTopLevelMultiple(t *testing.T) {

	ctx := testContext(ast.SelectionSet{
		&ast.Field{
			Name:  "TF1",
			Alias: "TF1A",
		},
		&ast.Field{
			Name:  "TF2",
			Alias: "TF2A",
		},
		&ast.Field{
			Name:  "TF3",
			Alias: "TF3A",
		},
	})

	results := GetRequestFields(ctx)
	expected := []string{"TF1A", "TF2A", "TF3A"}

	msg := checkStringSlicesEqual(expected, results)

	if msg != "" {
		t.Errorf(msg)
		return
	}
}

func TestGetRequestFieldsMultiLevel(t *testing.T) {

	ctx := testContext(ast.SelectionSet{
		&ast.Field{
			Name:  "TF1",
			Alias: "TF1A",
			SelectionSet: ast.SelectionSet{
				&ast.Field{
					Name:  "TF11",
					Alias: "TF11A",
				},
				&ast.Field{
					Name:  "TF12",
					Alias: "TF12A",
				},
				&ast.Field{
					Name:  "TF13",
					Alias: "TF13A",
					SelectionSet: ast.SelectionSet{
						&ast.Field{
							Name:  "TF131",
							Alias: "TF131A",
							SelectionSet: ast.SelectionSet{
								&ast.Field{
									Name:  "TF1311",
									Alias: "TF1311A",
								},
								&ast.Field{
									Name:  "TF1312",
									Alias: "TF1312A",
								},
								&ast.Field{
									Name:  "TF1312",
									Alias: "TF1312A",
								},
							},
						},
					},
				},
			},
		},
		&ast.Field{
			Name:  "TF2",
			Alias: "TF2A",
		},
	})

	results := GetRequestFields(ctx)
	expected := []string{"TF1A", "TF11A", "TF12A", "TF13A", "TF131A", "TF1311A", "TF1312A", "TF2A"}

	msg := checkStringSlicesEqual(expected, results)

	if msg != "" {
		t.Errorf(msg)
		return
	}
}
