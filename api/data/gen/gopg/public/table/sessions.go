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

var Sessions = newSessionsTable("public", "sessions", "")

type sessionsTable struct {
	postgres.Table

	// Columns
	Token     postgres.ColumnString
	UserID    postgres.ColumnString
	Expiry    postgres.ColumnTimestampz
	Data      postgres.ColumnString
	CreatedAt postgres.ColumnTimestampz

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type SessionsTable struct {
	sessionsTable

	EXCLUDED sessionsTable
}

// AS creates new SessionsTable with assigned alias
func (a SessionsTable) AS(alias string) *SessionsTable {
	return newSessionsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new SessionsTable with assigned schema name
func (a SessionsTable) FromSchema(schemaName string) *SessionsTable {
	return newSessionsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new SessionsTable with assigned table prefix
func (a SessionsTable) WithPrefix(prefix string) *SessionsTable {
	return newSessionsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new SessionsTable with assigned table suffix
func (a SessionsTable) WithSuffix(suffix string) *SessionsTable {
	return newSessionsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newSessionsTable(schemaName, tableName, alias string) *SessionsTable {
	return &SessionsTable{
		sessionsTable: newSessionsTableImpl(schemaName, tableName, alias),
		EXCLUDED:      newSessionsTableImpl("", "excluded", ""),
	}
}

func newSessionsTableImpl(schemaName, tableName, alias string) sessionsTable {
	var (
		TokenColumn     = postgres.StringColumn("token")
		UserIDColumn    = postgres.StringColumn("user_id")
		ExpiryColumn    = postgres.TimestampzColumn("expiry")
		DataColumn      = postgres.StringColumn("data")
		CreatedAtColumn = postgres.TimestampzColumn("created_at")
		allColumns      = postgres.ColumnList{TokenColumn, UserIDColumn, ExpiryColumn, DataColumn, CreatedAtColumn}
		mutableColumns  = postgres.ColumnList{UserIDColumn, ExpiryColumn, DataColumn, CreatedAtColumn}
	)

	return sessionsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		Token:     TokenColumn,
		UserID:    UserIDColumn,
		Expiry:    ExpiryColumn,
		Data:      DataColumn,
		CreatedAt: CreatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}
