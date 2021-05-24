package model

import (
	"fmt"
	"sync"
	"time"

	memdb "github.com/hashicorp/go-memdb"

	"github.com/joyrex2001/kubedock/internal/model/types"
	"github.com/joyrex2001/kubedock/internal/util/stringid"
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
		if err == nil {
			instance.loadDefaults()
		}
	})
	return instance, err
}

// createSchema will create the database with schema.
func (in *Database) createSchema() (*memdb.MemDB, error) {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"container": {
				Name: "container",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"shortid": {
						Name:    "shortid",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ShortID"},
					},
				},
			},
			"exec": {
				Name: "exec",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
				},
			},
			"network": {
				Name: "network",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"shortid": {
						Name:    "shortid",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ShortID"},
					},
					"name": {
						Name:    "name",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
				},
			},
			"image": {
				Name: "image",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"shortid": {
						Name:    "shortid",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ShortID"},
					},
					"name": {
						Name:    "name",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Name"},
					},
				},
			},
		},
	}
	return memdb.NewMemDB(schema)
}

// loadDefaults will insert default records into the database.
func (in *Database) loadDefaults() {
	in.SaveNetwork(&types.Network{Name: "null"})
	in.SaveNetwork(&types.Network{Name: "host"})
	in.SaveNetwork(&types.Network{Name: "bridge"})
}

// GetContainer will return a container with given id, or an error if
// the instance does not exist.
func (in *Database) GetContainer(id string) (*types.Container, error) {
	txn := in.db.Txn(false)
	defer txn.Abort()
	idx := "id"
	if stringid.IsShortID(id) {
		idx = "shortid"
	}
	raw, err := txn.First("container", idx, id)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("container %s not found", id)
	}
	return raw.(*types.Container), nil
}

// GetContainers will return all stored containers.
func (in *Database) GetContainers() ([]*types.Container, error) {
	rec := []*types.Container{}
	txn := in.db.Txn(false)
	defer txn.Abort()
	it, err := txn.Get("container", "id")
	if err != nil {
		return rec, err
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		rec = append(rec, obj.(*types.Container))
	}
	return rec, nil
}

// SaveContainer will either update the given container, or create a new
// record. If ID is not provided, it will generate an ID and adds the
// current time in Created.
func (in *Database) SaveContainer(con *types.Container) error {
	if con.ID == "" {
		id := stringid.GenerateRandomID()
		con.ID = id
		con.ShortID = stringid.TruncateID(id)
		con.Created = time.Now()
	}
	return in.save("container", con)
}

// DeleteContainer will delete provided container.
func (in *Database) DeleteContainer(con *types.Container) error {
	return in.delete("container", con)
}

// GetExec will return a exec with given id, or an error if the
// instance does not exist.
func (in *Database) GetExec(id string) (*types.Exec, error) {
	txn := in.db.Txn(false)
	defer txn.Abort()
	raw, err := txn.First("exec", "id", id)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("exec %s not found", id)
	}
	return raw.(*types.Exec), nil
}

// GetExecs will return all stored execs.
func (in *Database) GetExecs() ([]*types.Exec, error) {
	rec := []*types.Exec{}
	txn := in.db.Txn(false)
	defer txn.Abort()
	it, err := txn.Get("exec", "id")
	if err != nil {
		return rec, err
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		rec = append(rec, obj.(*types.Exec))
	}
	return rec, nil
}

// SaveExec will either update the given exec, or create a new
// record. If ID is not provided, it will generate an ID and adds the
// current time in Created.
func (in *Database) SaveExec(exc *types.Exec) error {
	if exc.ID == "" {
		id := stringid.GenerateRandomID()
		exc.ID = id
		exc.Created = time.Now()
	}
	return in.save("exec", exc)
}

// DeleteExec will delete provided exec.
func (in *Database) DeleteExec(exc *types.Exec) error {
	return in.delete("exec", exc)
}

// GetNetwork will return a network with given id, or an error if the
// instance does not exist.
func (in *Database) GetNetwork(id string) (*types.Network, error) {
	txn := in.db.Txn(false)
	defer txn.Abort()
	idx := "id"
	if stringid.IsShortID(id) {
		idx = "shortid"
	}
	raw, err := txn.First("network", idx, id)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("network %s not found", id)
	}
	return raw.(*types.Network), nil
}

// GetNetworkByName will return a network with given name, or an error if the
// instance does not exist.
func (in *Database) GetNetworkByName(name string) (*types.Network, error) {
	txn := in.db.Txn(false)
	defer txn.Abort()
	raw, err := txn.First("network", "name", name)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("network %s not found", name)
	}
	return raw.(*types.Network), nil
}

// GetNetworkByNameOrID will return a network with id/name, or an error if the
// instance does not exist.
func (in *Database) GetNetworkByNameOrID(id string) (*types.Network, error) {
	netw, err := in.GetNetwork(id)
	if err == nil {
		return netw, nil
	}
	return in.GetNetworkByName(id)
}

// GetNetworks will return all stored networks.
func (in *Database) GetNetworks() ([]*types.Network, error) {
	rec := []*types.Network{}
	txn := in.db.Txn(false)
	defer txn.Abort()
	it, err := txn.Get("network", "id")
	if err != nil {
		return rec, err
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		rec = append(rec, obj.(*types.Network))
	}
	return rec, nil
}

// GetNetworksByIDs will return all networks that are in the
// given set of network ids.
func (in *Database) GetNetworksByIDs(ids map[string]interface{}) ([]*types.Network, error) {
	rec := []*types.Network{}
	txn := in.db.Txn(false)
	defer txn.Abort()
	it, err := txn.Get("network", "id")
	if err != nil {
		return rec, err
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		netw := obj.(*types.Network)
		if _, ok := ids[netw.ID]; ok {
			rec = append(rec, netw)
		}
	}
	return rec, nil
}

// SaveNetwork will either update the given network, or create a new
// record. If ID is not provided, it will generate an ID and adds the
// current time in Created.
func (in *Database) SaveNetwork(netw *types.Network) error {
	if netw.ID == "" {
		id := stringid.GenerateRandomID()
		netw.ID = id
		netw.ShortID = stringid.TruncateID(id)
		netw.Created = time.Now()
	}
	return in.save("network", netw)
}

// DeleteNetwork will delete provided network.
func (in *Database) DeleteNetwork(netw *types.Network) error {
	return in.delete("network", netw)
}

// GetImage will return an image with given id, or an error if the
// instance does not exist.
func (in *Database) GetImage(id string) (*types.Image, error) {
	txn := in.db.Txn(false)
	defer txn.Abort()
	idx := "id"
	if stringid.IsShortID(id) {
		idx = "shortid"
	}
	raw, err := txn.First("image", idx, id)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("image %s not found", id)
	}
	return raw.(*types.Image), nil
}

// GetImageByName will return an image with given name, or an error if the
// instance does not exist.
func (in *Database) GetImageByName(name string) (*types.Image, error) {
	txn := in.db.Txn(false)
	defer txn.Abort()
	raw, err := txn.First("image", "name", name)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, fmt.Errorf("image %s not found", name)
	}
	return raw.(*types.Image), nil
}

// GetImageByNameOrID will return an image with id/name, or an error if the
// instance does not exist.
func (in *Database) GetImageByNameOrID(id string) (*types.Image, error) {
	netw, err := in.GetImage(id)
	if err == nil {
		return netw, nil
	}
	return in.GetImageByName(id)
}

// GetImages will return all stored execs.
func (in *Database) GetImages() ([]*types.Image, error) {
	rec := []*types.Image{}
	txn := in.db.Txn(false)
	defer txn.Abort()
	it, err := txn.Get("image", "id")
	if err != nil {
		return rec, err
	}
	for obj := it.Next(); obj != nil; obj = it.Next() {
		rec = append(rec, obj.(*types.Image))
	}
	return rec, nil
}

// SaveImage will either update the given image, or create a new
// record. If ID is not provided, it will generate an ID and adds the
// current time in Created.
func (in *Database) SaveImage(img *types.Image) error {
	if img.ID == "" {
		id := stringid.GenerateRandomID()
		img.ID = id
		img.ShortID = stringid.TruncateID(id)
		img.Created = time.Now()
	}
	return in.save("image", img)
}

// DeleteImage will delete provided image.
func (in *Database) DeleteImage(img *types.Image) error {
	return in.delete("image", img)
}

// save is a generic save method to store or update a record in the
// database.
func (in *Database) save(table string, rec interface{}) error {
	txn := in.db.Txn(true)
	if err := txn.Insert(table, rec); err != nil {
		txn.Abort()
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
		txn.Abort()
		return err
	}
	txn.Commit()
	return nil
}
