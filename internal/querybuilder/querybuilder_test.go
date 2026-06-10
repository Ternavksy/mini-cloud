package querybuilder

import (
	"reflect"
	"testing"
)

func TestQueryBuilders(t *testing.T) {
	selectSQL, selectArgs := Select("id", "name").
		From("collections").
		Where("id = ?", "collection-1").
		OrderBy("created_at").
		SQL()
	if selectSQL != "SELECT id, name FROM collections WHERE id = ? ORDER BY created_at" {
		t.Fatalf("select SQL = %q", selectSQL)
	}
	if !reflect.DeepEqual(selectArgs, []any{"collection-1"}) {
		t.Fatalf("select args = %v", selectArgs)
	}

	insertSQL, insertArgs := InsertInto("collections").
		Columns("id", "name").
		Values("collection-1", "album").
		SQL()
	if insertSQL != "INSERT INTO collections (id, name) VALUES (?, ?)" {
		t.Fatalf("insert SQL = %q", insertSQL)
	}
	if !reflect.DeepEqual(insertArgs, []any{"collection-1", "album"}) {
		t.Fatalf("insert args = %v", insertArgs)
	}

	deleteSQL, deleteArgs := DeleteFrom("collections").
		Where("id = ?", "collection-1").
		SQL()
	if deleteSQL != "DELETE FROM collections WHERE id = ?" {
		t.Fatalf("delete SQL = %q", deleteSQL)
	}
	if !reflect.DeepEqual(deleteArgs, []any{"collection-1"}) {
		t.Fatalf("delete args = %v", deleteArgs)
	}
}
