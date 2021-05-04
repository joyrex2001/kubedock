package model

import (
	"sync"
	"time"

	memdb "github.com/hashicorp/go-memdb"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/uuid"
)

// Database is the object contains the in-memory database.
type Database struct {
	db *memdb.MemDB
}

var instance *Database
var once sync.Once

// New will create return the singleton Database instance.
func New() (*Database, error) {
	var err error
	var db *memdb.MemDB
	once.Do(func() {
		instance = &Database{}
		db, err = instance.createSchema()
		instance.db = db
	})
	return instance, err
}

// createSchema will create the database with schema.
func (in *Database) createSchema() (*memdb.MemDB, error) {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"container": &memdb.TableSchema{
				Name: "container",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
				},
			},
			"exec": &memdb.TableSchema{
				Name: "exec",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
				},
			},
		},
	}
	return memdb.NewMemDB(schema)
}

// LoadContainer will return a container with given id, or an error if
// the instance does not exist.
func (in *Database) LoadContainer(id string) (*types.Container, error) {
	txn := in.db.Txn(false)
	raw, err := txn.First("container", "id", id)
	if err != nil {
		return nil, err
	}
	return raw.(*types.Container), nil
}

// SaveContainer will either update the given container, or create a new
// record. If ID is not provided, it will generate an ID and adds the
// current time in Created.
func (in *Database) SaveContainer(con *types.Container) error {
	if con.ID == "" {
		id, err := uuid.New()
		if err != nil {
			return err
		}
		con.ID = id
		con.Created = time.Now()
	}
	return in.save("container", con)
}

// DeleteContainer will delete provided container.
func (in *Database) DeleteContainer(con *types.Container) error {
	return in.delete("container", con)

}

// LoadExec will return a exec with given id, or an error if the
// instance does not exist.
func (in *Database) LoadExec(id string) (*types.Exec, error) {
	txn := in.db.Txn(false)
	raw, err := txn.First("exec", "id", id)
	if err != nil {
		return nil, err
	}
	return raw.(*types.Exec), nil
}

// SaveExec will either update the given exec, or create a new
// record. If ID is not provided, it will generate an ID and adds the
// current time in Created.
func (in *Database) SaveExec(exc *types.Exec) error {
	if exc.ID == "" {
		id, err := uuid.New()
		if err != nil {
			return err
		}
		exc.ID = id
		exc.Created = time.Now()
	}
	return in.save("exec", exc)
}

// DeleteContainer will delete provided exec.
func (in *Database) DeleteExec(exc *types.Exec) error {
	return in.delete("exec", exc)
}

// save is a generic save method to store or update a record in the
// database.
func (in *Database) save(table string, rec interface{}) error {
	txn := in.db.Txn(true)
	if err := txn.Insert(table, rec); err != nil {
		return err
	}
	txn.Commit()
	return nil
}

// delete is a generic delete method to remove a record from the
// database.
func (in *Database) delete(table string, rec interface{}) error {
	txn := in.db.Txn(true)
	if err := txn.Delete(table, rec); err != nil {
		return err
	}
	txn.Commit()
	return nil
}
