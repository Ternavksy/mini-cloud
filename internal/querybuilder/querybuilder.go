package querybuilder

import (
	"strings"
)

type SelectBuilder struct {
	columns []string
	table   string
	where   []string
	orderBy string
	args    []any
}

func Select(columns ...string) SelectBuilder {
	return SelectBuilder{columns: columns}
}

func (b SelectBuilder) From(table string) SelectBuilder {
	b.table = table
	return b
}

func (b SelectBuilder) Where(condition string, args ...any) SelectBuilder {
	b.where = append(b.where, condition)
	b.args = append(b.args, args...)
	return b
}

func (b SelectBuilder) OrderBy(orderBy string) SelectBuilder {
	b.orderBy = orderBy
	return b
}

func (b SelectBuilder) SQL() (string, []any) {
	query := "SELECT " + strings.Join(b.columns, ", ") + " FROM " + b.table
	if len(b.where) > 0 {
		query += " WHERE " + strings.Join(b.where, " AND ")
	}
	if b.orderBy != "" {
		query += " ORDER BY " + b.orderBy
	}
	return query, b.args
}

type InsertBuilder struct {
	table   string
	columns []string
	values  []any
}

func InsertInto(table string) InsertBuilder {
	return InsertBuilder{table: table}
}

func (b InsertBuilder) Columns(columns ...string) InsertBuilder {
	b.columns = columns
	return b
}

func (b InsertBuilder) Values(values ...any) InsertBuilder {
	b.values = values
	return b
}

func (b InsertBuilder) SQL() (string, []any) {
	placeholders := make([]string, 0, len(b.values))
	for range b.values {
		placeholders = append(placeholders, "?")
	}

	query := "INSERT INTO " + b.table + " (" + strings.Join(b.columns, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
	return query, b.values
}

type DeleteBuilder struct {
	table string
	where []string
	args  []any
}

func DeleteFrom(table string) DeleteBuilder {
	return DeleteBuilder{table: table}
}

func (b DeleteBuilder) Where(condition string, args ...any) DeleteBuilder {
	b.where = append(b.where, condition)
	b.args = append(b.args, args...)
	return b
}

func (b DeleteBuilder) SQL() (string, []any) {
	query := "DELETE FROM " + b.table
	if len(b.where) > 0 {
		query += " WHERE " + strings.Join(b.where, " AND ")
	}
	return query, b.args
}
