package dump

import (
	"github.com/aweris/postgres-data-dump/database"
	"github.com/aweris/postgres-data-dump/internal/log"
)

// navigator is a simple iterator for navigating tables.
type navigator struct {
	logger   log.Logger
	db       database.DB
	manifest *manifest
	todo     map[string]table
	done     map[string]table
	stack    []string
}

// newNavigator returns new Navigator instance.
func newNavigator(logger log.Logger, db database.DB, manifest *manifest) *navigator {
	nav := navigator{
		logger,
		db,
		manifest,
		make(map[string]table),
		make(map[string]table),
		make([]string, 0),
	}

	for _, item := range nav.manifest.Tables {
		nav.todo[item.TableName] = item
		nav.stack = append(nav.stack, item.TableName)
	}

	return &nav
}

// hasNext returns true if navigator has more item to consume.
func (nav *navigator) hasNext() bool {
	return len(nav.stack) > 0
}

// next returns next manifest item from stack.
func (nav *navigator) next() (*table, error) {
	if len(nav.stack) == 0 {
		return nil, nil
	}

	// pop table from stack
	tableName := nav.stack[0]
	nav.stack = nav.stack[1:]

	// table not in the list. Move to next table
	if _, ok := nav.todo[tableName]; !ok {
		return nav.next()
	}

	// get table dependencies
	deps, err := nav.db.GetTableDependencies(tableName)
	if err != nil {
		return nil, err
	}

	// find which tables needs to process
	todoDeps := make([]string, 0)

	for _, dep := range deps {
		_, isTodo := nav.todo[dep]
		_, isDone := nav.done[dep]
		if !isTodo && !isDone {
			// A new dependency table not present in the manifest file was
			// found, create a default entry for it
			nav.todo[dep] = table{TableName: dep}
		}
		if _, ok := nav.todo[dep]; ok && tableName != dep {
			todoDeps = append(todoDeps, dep)
		}
	}

	// update stack with new dependencies and get next table
	if len(todoDeps) > 0 {
		nav.stack = append(todoDeps, append([]string{tableName}, nav.stack...)...)
		return nav.next()
	}

	next := nav.todo[tableName]

	nav.done[tableName] = nav.todo[tableName]
	delete(nav.todo, tableName)

	if cols := next.Columns; len(cols) == 0 {
		cols, err = nav.db.GetTableColumns(next.TableName)
		if err != nil {
			return nil, err
		}
		next.Columns = cols
	}

	return &next, nil
}
