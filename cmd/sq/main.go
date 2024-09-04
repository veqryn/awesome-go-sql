/*
Build and run some queries using the sq and sqddl libraries.
SQ is supposed to be a type-safe SQL builder, that also generates table
definitions of your database schema.
In practice, it Errors out on relatively simple queries.
The way it scans structs is also very cumbersome and verbose.
It does not do dynamic queries.
It does not work with PGX.
*/
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bokwoon95/sq"
	_ "github.com/jackc/pgx/v5/stdlib" // DB Driver
	"github.com/veqryn/awesome-go-sql/cmd/sq/internal/table"
	"github.com/veqryn/awesome-go-sql/models"
)

// go install -tags=fts5 github.com/bokwoon95/sqddl@latest
// Run with go generate -x ./...
// This will create subdirectories with the model
//go:generate mkdir -p ./internal/table
//go:generate sqddl tables -db 'postgres://postgres:password@localhost:5432/awesome?sslmode=disable' -schemas public -file ./internal/table/gen_tables.go -pkg table

func (d DAO) SelectAccountByID(ctx context.Context, id uint64) (models.AccountIdeal, bool, error) {
	// Manually set the scan column names
	account, err := sq.FetchOneContext(ctx, sq.VerboseLog(d.db), // d.db,
		sq.Queryf(
			`SELECT {*}
		FROM accounts
		WHERE id = {}`,
			id).SetDialect(sq.DialectPostgres),
		func(row *sq.Row) models.AccountIdeal {
			rval := models.AccountIdeal{
				ID:       uint64(row.Int64("id")),
				Name:     row.String("name"),
				Email:    row.String("email"),
				Active:   row.Bool("active"),
				FavColor: NullStringToPtr(row.NullString("fav_color")),
				// FavNumbers has to be done separately for some reason
				Properties: BytesToJsonRawPtr(row.Bytes("properties")),
				CreatedAt:  row.Time("created_at"),
			}
			row.Array(&rval.FavNumbers, "fav_numbers")
			return rval
		},
	)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return account, false, nil
	case err != nil:
		return account, false, err
	default:
		return account, true, nil
	}
}

func (d DAO) SelectAllAccounts(ctx context.Context) ([]models.AccountIdeal, error) {
	// Use the generated table definition to set the column names
	a := sq.New[table.ACCOUNTS]("accounts")
	return sq.FetchAllContext(ctx, sq.VerboseLog(d.db), // d.db,
		sq.From(a).SetDialect(sq.DialectPostgres),
		func(row *sq.Row) models.AccountIdeal {
			rval := models.AccountIdeal{
				ID:       uint64(row.Int64Field(a.ID)),
				Name:     row.StringField(a.NAME),
				Email:    row.StringField(a.EMAIL),
				Active:   row.BoolField(a.ACTIVE),
				FavColor: NullStringToPtr(row.NullString(a.FAV_COLOR.GetAlias())), // No NullEnumField method exists yet
				// FavNumbers has to be done separately for some reason
				Properties: BytesToJsonRawPtr(row.BytesField(a.PROPERTIES)),
				CreatedAt:  row.TimeField(a.CREATED_AT),
			}
			row.ArrayField(&rval.FavNumbers, a.FAV_NUMBERS)
			return rval
		})
}

func (d DAO) SelectAllAccountsByFilter(ctx context.Context, filters models.Filters) ([]models.AccountIdeal, error) {
	query := `
		SELECT {*}
		FROM accounts`

	// Sadly, we have to manually build dynamic queries
	var wheres []string
	var args []any
	if len(filters.Names) > 0 {
		wheres = append(wheres, "name = ANY({})")
		args = append(args, filters.Names)
	}
	if filters.Active != nil {
		wheres = append(wheres, "active = {}")
		args = append(args, *filters.Active)
	}
	if len(filters.FavColors) > 0 {
		wheres = append(wheres, "fav_color = ANY({})")
		args = append(args, filters.FavColors)
	}

	if len(wheres) > 0 {
		query += " WHERE " + strings.Join(wheres, " AND ")
	}

	fmt.Printf("--------\nDynamic Query SQL:\n%s\n\nDynamic Query Args:\n%#+v\n", query, args)

	// Use the generated table definition to set the column names
	a := sq.New[table.ACCOUNTS]("accounts")
	return sq.FetchAllContext(ctx, sq.VerboseLog(d.db), // d.db,
		sq.Queryf(query, args...).SetDialect(sq.DialectPostgres),
		func(row *sq.Row) models.AccountIdeal {
			rval := models.AccountIdeal{
				ID:       uint64(row.Int64Field(a.ID)),
				Name:     row.StringField(a.NAME),
				Email:    row.StringField(a.EMAIL),
				Active:   row.BoolField(a.ACTIVE),
				FavColor: NullStringToPtr(row.NullString(a.FAV_COLOR.GetAlias())), // No NullEnumField method exists yet
				// FavNumbers has to be done separately for some reason
				Properties: BytesToJsonRawPtr(row.BytesField(a.PROPERTIES)),
				CreatedAt:  row.TimeField(a.CREATED_AT),
			}
			row.ArrayField(&rval.FavNumbers, a.FAV_NUMBERS)
			return rval
		})
}

func main() {
	ctx := context.Background()

	// This is the database/sql version of pgx
	db, err := sql.Open("pgx", "postgresql://postgres:password@localhost:5432/awesome")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dao := DAO{db: db}

	// Query 1
	_, ok, err := dao.SelectAccountByID(ctx, 2)
	if err != nil {
		fmt.Printf("ERROR: %#+v\n", err)
		panic(err)
	}
	// There are errors when the row doesn't exist
	if ok {
		panic("ERROR: Account should not be found")
	}
	// fmt.Printf("--------\nQuery by ID\n%s\n", account)

	// These also error out

	// Query multiple
	accounts, err := dao.SelectAllAccounts(ctx)
	if err != nil {
		fmt.Printf("ERROR: %#+v\n", err)
		panic(err)
	}
	fmt.Println("--------\nQuery All")
	for _, account := range accounts {
		fmt.Printf("%s\n\n", account)
	}

	// Dynamic Query of multiple
	active := true
	accounts, err = dao.SelectAllAccountsByFilter(ctx, models.Filters{
		Names:     []string{"Jane", "John"},
		Active:    &active,
		FavColors: []string{"red", "blue", "green"},
	})
	if err != nil {
		fmt.Printf("ERROR: %#+v\n", err)
		panic(err)
	}
	fmt.Println("--------\nQuery Filter")
	for _, account := range accounts {
		fmt.Printf("%s\n\n", account)
	}
}

type DAO struct {
	db *sql.DB
}

func NullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func BytesToJsonRawPtr(b []byte) *json.RawMessage {
	if b == nil {
		return nil
	}
	raw := json.RawMessage(b)
	return &raw
}
