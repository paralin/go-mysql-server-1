// Copyright 2020-2021 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sql

import (
	"fmt"
	"strings"
	"sync"

	"github.com/dolthub/go-mysql-server/internal/similartext"

	gerrors "errors"

	"gopkg.in/src-d/go-errors.v1"
)

// ErrDatabaseNotFound is thrown when a database is not found
var ErrDatabaseNotFound = errors.NewKind("database not found: %s")

// ErrNoDatabaseSelected is thrown when a database is not selected and the query requires one
var ErrNoDatabaseSelected = errors.NewKind("no database selected")

// ErrAsOfNotSupported is thrown when an AS OF query is run on a database that can't support it
var ErrAsOfNotSupported = errors.NewKind("AS OF not supported for database %s")

// ErrIncompatibleAsOf is thrown when an AS OF clause is used in an incompatible manner, such as when using an AS OF
// expression with a view when the view definition has its own AS OF expressions.
var ErrIncompatibleAsOf = errors.NewKind("incompatible use of AS OF: %s")

// Catalog holds databases, tables and functions.
type Catalog struct {
	FunctionRegistry
	*ProcessList
	*MemoryManager
	DatabaseCatalog

	mu    sync.RWMutex
	locks sessionLocks
}

type tableLocks map[string]struct{}

type dbLocks map[string]tableLocks

type sessionLocks map[uint32]dbLocks

// NewCatalog returns a new empty Catalog.
func NewCatalog() *Catalog {
	return &Catalog{
		FunctionRegistry: NewFunctionRegistry(),
		MemoryManager:    NewMemoryManager(ProcessMemory),
		ProcessList:      NewProcessList(),
		DatabaseCatalog:  NewDatabases(nil),
		locks:            make(sessionLocks),
	}
}

// DatabaseCatalog contains a list of databases.
type DatabaseCatalog interface {
	// AllDatabases returns all databases in the catalog.
	AllDatabases() (DatabasesSlice, error)
	// Database returns the Database with the given name if it exists.
	Database(name string) (Database, error)
	// CreateDatabase creates a new database and returns an error if it exists.
	CreateDatabase(name string) (Database, error)
	// RemoveDatabase removes a database from the catalog.
	// Does not return an error if does not exist.
	RemoveDatabase(dbName string) error
	// HasDB returns if the database exists.
	HasDB(db string) (bool, error)
	// Table returns the Table with the given name if it exists.
	Table(ctx *Context, dbName string, tableName string) (Table, Database, error)
	// TableAsOf returns the table with the name given at the time given, if it
	// existed. The database named must implement sql.VersionedDatabase or an
	// error is returned.
	TableAsOf(ctx *Context, dbName string, tableName string, asOf interface{}) (Table, Database, error)
}

// CreateDbFn is a function to create a database.
type CreateDbFn func(name string) (Database, error)

// Databases contains a set of databases.
type Databases struct {
	mu         sync.RWMutex
	dbs        DatabasesSlice
	createDbFn CreateDbFn
}

// NewDatabases constructs a new empty databases catalog.
func NewDatabases(createDbFn CreateDbFn) *Databases {
	return &Databases{createDbFn: createDbFn}
}

// AllDatabases returns all databases in the catalog.
func (c *Databases) AllDatabases() (DatabasesSlice, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result = make(DatabasesSlice, len(c.dbs))
	copy(result, c.dbs)
	return result, nil
}

// AddDatabase adds a new database to the catalog.
func (c *Databases) AddDatabase(db Database) {
	c.mu.Lock()
	c.dbs.Add(db)
	c.mu.Unlock()
}

// CreateDatabase creates a new database and returns an error if it exists.
func (c *Databases) CreateDatabase(name string) (Database, error) {
	if c.createDbFn == nil {
		return nil, gerrors.New("create database not implemented")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	edb, err := c.dbs.Database(name)
	if edb != nil {
		return nil, ErrCannotCreateDatabaseExists.New(name)
	}
	if err != nil && !ErrDatabaseNotFound.Is(err) {
		return nil, err
	}

	ndb, err := c.createDbFn(name)
	if err != nil {
		return nil, err
	}
	c.dbs.Add(ndb)
	return ndb, nil
}

// RemoveDatabase removes a database from the catalog.
// Does not return an error if does not exist.
func (c *Databases) RemoveDatabase(dbName string) error {
	c.mu.Lock()
	c.dbs.Delete(dbName)
	c.mu.Unlock()
	return nil
}

func (c *Databases) HasDB(db string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, err := c.dbs.Database(db)

	return err == nil, nil
}

// Database returns the database with the given name.
func (c *Databases) Database(db string) (Database, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dbs.Database(db)
}

// Table returns the table in the given database with the given name.
func (c *Databases) Table(ctx *Context, db, table string) (Table, Database, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dbs.Table(ctx, db, table)
}

// TableAsOf returns the table in the given database with the given name, as it existed at the time given. The database
// named must support timed queries.
func (c *Databases) TableAsOf(ctx *Context, db, table string, time interface{}) (Table, Database, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.dbs.TableAsOf(ctx, db, table, time)
}

// DatabasesSlice is a collection of Database.
type DatabasesSlice []Database

// Database returns the Database with the given name if it exists.
func (d DatabasesSlice) Database(name string) (Database, error) {
	if len(d) == 0 {
		return nil, ErrDatabaseNotFound.New(name)
	}

	name = strings.ToLower(name)
	var dbNames []string
	for _, db := range d {
		if strings.ToLower(db.Name()) == name {
			return db, nil
		}
		dbNames = append(dbNames, db.Name())
	}
	similar := similartext.Find(dbNames, name)
	return nil, ErrDatabaseNotFound.New(name + similar)
}

// Add adds a new database.
func (d *DatabasesSlice) Add(db Database) {
	*d = append(*d, db)
}

// Delete removes a database.
func (d *DatabasesSlice) Delete(dbName string) {
	idx := -1
	for i, db := range *d {
		if db.Name() == dbName {
			idx = i
			break
		}
	}

	if idx != -1 {
		*d = append((*d)[:idx], (*d)[idx+1:]...)
	}
}

// Table returns the Table with the given name if it exists.
func (d DatabasesSlice) Table(ctx *Context, dbName string, tableName string) (Table, Database, error) {
	db, err := d.Database(dbName)
	if err != nil {
		return nil, nil, err
	}

	tbl, ok, err := db.GetTableInsensitive(ctx, tableName)

	if err != nil {
		return nil, nil, err
	} else if !ok {
		return nil, nil, suggestSimilarTables(db, ctx, tableName)
	}

	return tbl, db, nil
}

// _ is a type assertion
var _ DatabaseCatalog = ((*Databases)(nil))

func suggestSimilarTables(db Database, ctx *Context, tableName string) error {
	tableNames, err := db.GetTableNames(ctx)
	if err != nil {
		return err
	}

	similar := similartext.Find(tableNames, tableName)
	return ErrTableNotFound.New(tableName + similar)
}

// TableAsOf returns the table with the name given at the time given, if it existed. The database named must implement
// sql.VersionedDatabase or an error is returned.
func (d DatabasesSlice) TableAsOf(ctx *Context, dbName string, tableName string, asOf interface{}) (Table, Database, error) {
	db, err := d.Database(dbName)
	if err != nil {
		return nil, nil, err
	}

	versionedDb, ok := db.(VersionedDatabase)
	if !ok {
		return nil, nil, ErrAsOfNotSupported.New(tableName)
	}

	tbl, ok, err := versionedDb.GetTableInsensitiveAsOf(ctx, tableName, asOf)

	if err != nil {
		return nil, nil, err
	} else if !ok {
		return nil, nil, suggestSimilarTablesAsOf(versionedDb, ctx, tableName, asOf)
	}

	return tbl, versionedDb, nil
}

func suggestSimilarTablesAsOf(db VersionedDatabase, ctx *Context, tableName string, time interface{}) error {
	tableNames, err := db.GetTableNamesAsOf(ctx, time)
	if err != nil {
		return err
	}

	similar := similartext.Find(tableNames, tableName)
	return ErrTableNotFound.New(tableName + similar)
}

// LockTable adds a lock for the given table and session client. It is assumed
// the database is the current database in use.
func (c *Catalog) LockTable(ctx *Context, table string) {
	id := ctx.ID()
	db := ctx.GetCurrentDatabase()

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.locks[id]; !ok {
		c.locks[id] = make(dbLocks)
	}

	if _, ok := c.locks[id][db]; !ok {
		c.locks[id][db] = make(tableLocks)
	}

	c.locks[id][db][table] = struct{}{}
}

// UnlockTables unlocks all tables for which the given session client has a
// lock.
func (c *Catalog) UnlockTables(ctx *Context, id uint32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errors []string
	for db, tables := range c.locks[id] {
		for t := range tables {
			table, _, err := c.DatabaseCatalog.Table(ctx, db, t)
			if err == nil {
				if lockable, ok := table.(Lockable); ok {
					if e := lockable.Unlock(ctx, id); e != nil {
						errors = append(errors, e.Error())
					}
				}
			} else {
				errors = append(errors, err.Error())
			}
		}
	}

	delete(c.locks, id)
	if len(errors) > 0 {
		return fmt.Errorf("error unlocking tables for %d: %s", id, strings.Join(errors, ", "))
	}

	return nil
}
