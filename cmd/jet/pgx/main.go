/*
Build and run some queries using the Jet generator library.
Jet is a type-safe SQL builder, that also generates models of your existing db
schema, and can scan directly to those models.
Jet works by connecting to your database and inspecting the schema, in order to
generate models.
This allows its SQL builder to be both type safe and use the actual column names
defined in your database schema.
Some disadvantages of it are:
  - the default schema assumes all postgres arrays and json/jsonb types become
    golang strings. This apparently can be customized though.
  - the sql builder can be a bit cumbersome and verbose, owing to its type safety.
  - the sql builder requires wrappers to support slices.
  - the sql scanner does not work with PGX.
    Jet's SQL Builder works with both database/sql and PGX, but the scanner only
    works with database/sql.
  - if using just the sql builder with PGX, you can not use the generated models
    if they contain any arrays.
  - A very minor disadvantage is that it doesn't use postgres' =ANY($1) format
    for IN queries using a slice, and instead enumerates all values in the slice.
*/
package main

import (
	"context"
	"errors"
	"fmt"

	. "github.com/go-jet/jet/v2/postgres" // Dot import for fluent sql writing, but optional
	"github.com/jackc/pgx/v5"             // DB Driver
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/veqryn/awesome-go-sql/cmd/jet/internal/awesome/public/table" // Dot import for fluent sql writing, but optional
	"github.com/veqryn/awesome-go-sql/models"
)

func (d DAO) SelectAccountByID(ctx context.Context, id uint64) (models.AccountIdeal, bool, error) {
	query := SELECT(
		// This would also work: Accounts.AllColumns
		Accounts.ID,
		Accounts.Name,
		Accounts.Email,
		Accounts.Active,
		Accounts.FavColor,
		Accounts.FavNumbers,
		Accounts.Properties,
		Accounts.CreatedAt,
	).FROM(
		Accounts,
	).WHERE(
		Accounts.ID.EQ(Uint64(id)),
	)

	sqlStr, args := query.Sql()

	var account models.AccountIdeal
	err := d.db.QueryRow(ctx, sqlStr, args...).Scan(
		&account.ID,
		&account.Name,
		&account.Email,
		&account.Active,
		&account.FavColor,
		&account.FavNumbers,
		&account.Properties,
		&account.CreatedAt,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return account, false, nil
	case err != nil:
		return account, false, err
	default:
		return account, true, nil
	}
}

func (d DAO) SelectAllAccounts(ctx context.Context) ([]models.AccountIdeal, error) {
	query := SELECT(
		Accounts.AllColumns,
	).FROM(
		Accounts,
	).ORDER_BY(Accounts.ID)

	sqlStr, args := query.Sql()

	rows, err := d.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.AccountIdeal
	for rows.Next() {
		var account models.AccountIdeal
		scanErr := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Email,
			&account.Active,
			&account.FavColor,
			&account.FavNumbers,
			&account.Properties,
			&account.CreatedAt)
		if scanErr != nil {
			// Check for a scan error. Query rows will be closed with defer.
			return nil, scanErr
		}
		accounts = append(accounts, account)
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (d DAO) SelectAllAccountsByFilter(ctx context.Context, filters models.Filters) ([]models.AccountIdeal, error) {
	// Create a slice of conditions (where expressions) dynamically,
	// then build the SQL statement.
	var wheres []BoolExpression
	if len(filters.Names) > 0 {
		wheres = append(wheres, Accounts.Name.IN(Strings(filters.Names)...))
	}
	if filters.Active != nil {
		wheres = append(wheres, Accounts.Active.EQ(Bool(*filters.Active)))
	}
	if len(filters.FavColors) > 0 {
		wheres = append(wheres, Accounts.FavColor.IN(Enums(filters.FavColors)...))
	}

	query := SELECT(
		Accounts.AllColumns,
	).FROM(
		Accounts,
	).WHERE(WhereAnd(wheres))

	sqlStr, args := query.Sql()
	fmt.Printf("--------\nDynamic Query SQL:\n%s\n\nDynamic Query Args:\n%#+v\n", sqlStr, args)

	rows, err := d.db.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.AccountIdeal
	for rows.Next() {
		var account models.AccountIdeal
		scanErr := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Email,
			&account.Active,
			&account.FavColor,
			&account.FavNumbers,
			&account.Properties,
			&account.CreatedAt)
		if scanErr != nil {
			// Check for a scan error. Query rows will be closed with defer.
			return nil, scanErr
		}
		accounts = append(accounts, account)
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func main() {
	ctx := context.Background()

	// This is the normal version of pgx
	db, err := pgxpool.New(ctx, "postgresql://postgres:password@localhost:5432/awesome")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dao := DAO{db: db}

	// Query 1
	_, ok, err := dao.SelectAccountByID(ctx, 0)
	if err != nil {
		fmt.Printf("ERROR: %#+v\n", err)
		panic(err)
	}
	if ok {
		panic("ERROR: Account should not be found")
	}
	// fmt.Printf("--------\nQuery by ID\n%s\n", account)

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
	db *pgxpool.Pool
}

// Integers converts slice of integers into slice of jet.Expression, useful for IN queries.
func Integers[T ~int | ~int8 | ~int16 | ~int32 | ~int64](s []T) []Expression {
	expressions := make([]Expression, 0, len(s))
	for _, v := range s {
		expressions = append(expressions, Expression(Int(int64(v))))
	}
	return expressions
}

// Strings converts slice of strings into slice of jet.Expression's, useful for IN queries.
func Strings[T ~string](s []T) []Expression {
	expressions := make([]Expression, 0, len(s))
	for _, v := range s {
		expressions = append(expressions, Expression(String(string(v))))
	}
	return expressions
}

// Enums converts slice of strings into slice of jet.Expression's, useful for IN queries.
func Enums[T ~string](s []T) []Expression {
	expressions := make([]Expression, 0, len(s))
	for _, v := range s {
		expressions = append(expressions, Expression(NewEnumValue(string(v))))
	}
	return expressions
}

// WhereAnd joins multiple jet.BoolExpression together with AND
func WhereAnd(wheres []BoolExpression) BoolExpression {
	var where BoolExpression
	for _, w := range wheres {
		if where == nil {
			where = w
		} else {
			where = where.AND(w)
		}
	}
	return where
}
