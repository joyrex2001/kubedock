package model

import (
	"testing"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestDatabase(t *testing.T) {
	db, _ := New()
	for i := 0; i < 2; i++ {
		_db, _ := New()
		if _db != db && db != nil {
			t.Errorf("New failed %d - got different instance", i)
		}
	}

	// container tests
	if _, err := db.LoadContainer("someid"); err == nil {
		t.Errorf("Expected error when loading a container that doesn't exist")
	}
	con := &types.Container{}
	if err := db.SaveContainer(con); err != nil {
		t.Errorf("Unexpected error when creating a new container")

	}
	if con.ID == "" {
		t.Errorf("Expected ID when saving a new container")
	}
	con.Image = "busybox"
	if err := db.SaveContainer(con); err != nil {
		t.Errorf("Unexpected error when updating a container")

	}
	if conl, err := db.LoadContainer(con.ID); err != nil {
		t.Errorf("Unexpected error when loading an existing container")
	} else {
		if conl.ID != con.ID || conl.Image != con.Image {
			t.Errorf("Loaded container differs to saved container")
		}
	}
	if cons, err := db.GetContainers(); err != nil {
		t.Errorf("Unexpected error when loading all existing containers")
	} else {
		if len(cons) != 1 {
			t.Errorf("Expected 1 container records, but got %d", len(cons))
		}
	}
	conid := con.ID
	if err := db.DeleteContainer(con); err != nil {
		t.Errorf("Unexpected error when deleting a container")

	}
	if _, err := db.LoadContainer(conid); err == nil {
		t.Errorf("Expected error when loading a container that doesn't exist")
	}

	// exec tests
	if _, err := db.LoadExec("someid"); err == nil {
		t.Errorf("Expected error when loading an exec that doesn't exist")
	}
	exc := &types.Exec{}
	if err := db.SaveExec(exc); err != nil {
		t.Errorf("Unexpected error when creating a new exec")
	}
	if exc.ID == "" {
		t.Errorf("Expected ID when saving a new exec")
	}
	exc.ContainerID = "1234"
	if err := db.SaveExec(exc); err != nil {
		t.Errorf("Unexpected error when updating an exec")
	}

	if excl, err := db.LoadExec(exc.ID); err != nil {
		t.Errorf("Unexpected error when loading an existing exec")
	} else {
		if excl.ID != exc.ID || excl.ContainerID != exc.ContainerID {
			t.Errorf("Loaded container differs to saved exec")
		}
	}
	if excs, err := db.GetExecs(); err != nil {
		t.Errorf("Unexpected error when loading all existing execs")
	} else {
		if len(excs) != 1 {
			t.Errorf("Expected 1 exec records, but got %d", len(excs))
		}
	}
	excid := exc.ID
	if err := db.DeleteExec(exc); err != nil {
		t.Errorf("Unexpected error when deleting an exec")

	}
	if _, err := db.LoadExec(excid); err == nil {
		t.Errorf("Expected error when loading an exec that doesn't exist")
	}
}
