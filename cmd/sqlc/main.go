/*
Build and run some queries using the SQLC generator library.
SQLC is very different from the other libraries, in that it expects you to
provide it both your schema (or database connection) and a file containing a
list of queries you want to convert into code.
It then generates both the struct models and the golang functions to match each
query.
It works wonderfully for any non-dynamic queries, but unfortunately it does not
really cover dynamic queries.
For dynamic queries, it can actually work if you use CASE statement in the WHERE,
but this isn't something you'd want to do when there are more than 2-3 optional
conditions, as it will create a very ugly query that probably won't get
optimized well.
You could choose to make use of the generated models with other libraries for
dynamic SQL generation.
It also a bug around arrays/slices of enum types.
SQLC works with both database/sql and PGX.
*/
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5" // DB Driver
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/veqryn/awesome-go-sql/cmd/sqlc/internal/model"
	"github.com/veqryn/awesome-go-sql/models"
)

// brew install sqlc
// Run with go generate -x ./...
// This will create subdirectories with the generated queries
//go:generate sqlc generate

func main() {
	ctx := context.Background()

	// This is the normal version of pgx
	db, err := pgxpool.New(ctx, "postgresql://postgres:password@localhost:5432/awesome")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dao := DAO{
		Queries: model.New(db),
		db:      db,
	}

	_, err = dao.SelectAccountByID(ctx, 0)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		fmt.Printf("ERROR: %#+v\n", err)
		panic(err)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		panic("ERROR: Account should not be found")
	}
	// fmt.Printf("--------\nQuery by ID\n%s\n", AccountToStr(account))

	accounts, err := dao.SelectAllAccounts(ctx)
	if err != nil {
		fmt.Printf("ERROR: %#+v\n", err)
		panic(err)
	}
	fmt.Println("--------\nQuery All")
	for _, account := range accounts {
		fmt.Printf("%s\n\n", AccountToStr(account))
	}

	accounts, err = dao.SelectAllAccountsByFilter(ctx, model.SelectAllAccountsByFilterParams{
		AnyNames: true,
		Names:    []string{"Jane", "John"},
		IsActive: true,
		Active:   true,
		//AnyFavColor: true, // TODO: currently doesn't work, made a bug ticket
		//FavColors:   []model.Colors{model.ColorsRed, model.ColorsBlue, model.ColorsGreen},
	})
	if err != nil {
		fmt.Printf("ERROR: %#+v\n", err)
		panic(err)
	}
	fmt.Println("--------\nQuery Filter")
	for _, account := range accounts {
		fmt.Printf("%s\n\n", AccountToStr(account))
	}
}

type DAO struct {
	*model.Queries
	db *pgxpool.Pool
}

func AccountToStr(a model.Account) string {
	return fmt.Sprintf("Account:\nID: %d\nName: %s\nEmail: %s\nActive: %t\nFavColor: %s\nFavNumbers: %v\nProperties: %s\nCreatedAt: %s",
		a.ID,
		a.Name,
		a.Email,
		a.Active,
		NullColorsToStr(a.FavColor),
		models.SliceToStr(a.FavNumbers),
		models.SliceToStr(a.Properties),
		a.CreatedAt.Time)
}

func NullColorsToStr(nc model.NullColors) string {
	if !nc.Valid {
		return "<nil>"
	}
	return string(nc.Colors)
}
