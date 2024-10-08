//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var Accounts = newAccountsTable("public", "accounts", "")

type accountsTable struct {
	postgres.Table

	// Columns
	ID         postgres.ColumnInteger
	Name       postgres.ColumnString
	Email      postgres.ColumnString
	Active     postgres.ColumnBool
	FavColor   postgres.ColumnString
	FavNumbers postgres.ColumnString
	Properties postgres.ColumnString
	CreatedAt  postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type AccountsTable struct {
	accountsTable

	EXCLUDED accountsTable
}

// AS creates new AccountsTable with assigned alias
func (a AccountsTable) AS(alias string) *AccountsTable {
	return newAccountsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new AccountsTable with assigned schema name
func (a AccountsTable) FromSchema(schemaName string) *AccountsTable {
	return newAccountsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new AccountsTable with assigned table prefix
func (a AccountsTable) WithPrefix(prefix string) *AccountsTable {
	return newAccountsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new AccountsTable with assigned table suffix
func (a AccountsTable) WithSuffix(suffix string) *AccountsTable {
	return newAccountsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newAccountsTable(schemaName, tableName, alias string) *AccountsTable {
	return &AccountsTable{
		accountsTable: newAccountsTableImpl(schemaName, tableName, alias),
		EXCLUDED:      newAccountsTableImpl("", "excluded", ""),
	}
}

func newAccountsTableImpl(schemaName, tableName, alias string) accountsTable {
	var (
		IDColumn         = postgres.IntegerColumn("id")
		NameColumn       = postgres.StringColumn("name")
		EmailColumn      = postgres.StringColumn("email")
		ActiveColumn     = postgres.BoolColumn("active")
		FavColorColumn   = postgres.StringColumn("fav_color")
		FavNumbersColumn = postgres.StringColumn("fav_numbers")
		PropertiesColumn = postgres.StringColumn("properties")
		CreatedAtColumn  = postgres.TimestampzColumn("created_at")
		allColumns       = postgres.ColumnList{IDColumn, NameColumn, EmailColumn, ActiveColumn, FavColorColumn, FavNumbersColumn, PropertiesColumn, CreatedAtColumn}
		mutableColumns   = postgres.ColumnList{NameColumn, EmailColumn, ActiveColumn, FavColorColumn, FavNumbersColumn, PropertiesColumn, CreatedAtColumn}
	)

	return accountsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:         IDColumn,
		Name:       NameColumn,
		Email:      EmailColumn,
		Active:     ActiveColumn,
		FavColor:   FavColorColumn,
		FavNumbers: FavNumbersColumn,
		Properties: PropertiesColumn,
		CreatedAt:  CreatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
