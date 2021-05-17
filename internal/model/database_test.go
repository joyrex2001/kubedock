package model

import (
	"fmt"
	"testing"

	"github.com/joyrex2001/kubedock/internal/model/types"
)

func TestDatabase(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Errorf("Unexpected error creating database: %s", err)
	}
	for i := 0; i < 2; i++ {
		_db, _ := New()
		if _db != db && db != nil {
			t.Errorf("New failed %d - got different instance", i)
		}
	}

	// container tests
	if _, err := db.GetContainer("someid"); err == nil {
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
	if conl, err := db.GetContainer(con.ID); err != nil {
		t.Errorf("Unexpected error when loading an existing container")
	} else {
		if conl.ID != con.ID || conl.Image != con.Image {
			t.Errorf("Loaded container differs to saved container")
		}
	}
	if conl, err := db.GetContainer(con.ShortID); err != nil {
		t.Errorf("Unexpected error when loading an existing container shortid")
	} else {
		if conl.ID != con.ID || conl.Image != con.Image {
			t.Errorf("Loaded shortid container differs to saved container")
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
	if _, err := db.GetContainer(conid); err == nil {
		t.Errorf("Expected error when loading a container that doesn't exist")
	}

	// exec tests
	if _, err := db.GetExec("someid"); err == nil {
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

	if excl, err := db.GetExec(exc.ID); err != nil {
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
	if _, err := db.GetExec(excid); err == nil {
		t.Errorf("Expected error when loading an exec that doesn't exist")
	}
}

func TestIDWorkaround(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Errorf("Unexpected error creating database: %s", err)
	}
	for i := 0; i < 1000; i++ {
		con := &types.Container{}
		if err := db.SaveContainer(con); err != nil {
			t.Errorf("Unexpected error when creating a new container")
		}
		if con.ID[:1] == "c" {
			t.Errorf("Container ID that start with a c cause problems in the server router setup...")
			return
		}
		// netw := &types.Network{}
		// if err := db.SaveNetwork(netw); err != nil {
		// 	t.Errorf("Unexpected error when creating a new network")
		// }
		// if netw.ID[:1] == "c" {
		// 	t.Errorf("Network ID that start with a c cause problems in the server router setup...")
		// 	return
		// }
	}
}

func TestNetwork(t *testing.T) {
	db, _ := New()

	if _, err := db.GetNetworkByNameOrID("bridge"); err != nil {
		t.Errorf("Unexpected error when loading the bridge network")
	}

	netw := &types.Network{Name: "net0"}
	if err := db.SaveNetwork(netw); err != nil {
		t.Errorf("Unexpected error when creating network net0")
	}

	for i, n := range []string{"net1", "net2", "net3"} {
		netw := &types.Network{Name: n, ID: fmt.Sprintf("%d", i+1)}
		if err := db.SaveNetwork(netw); err != nil {
			t.Errorf("Unexpected error when creating network %s", n)
		}
	}

	if netws, err := db.GetNetworks(); err != nil {
		t.Errorf("Unexpected error when loading all existing networks")
	} else {
		if len(netws) != 6 {
			t.Errorf("Expected 5 network records, but got %d", len(netws))
		}
	}

	net1, err := db.GetNetworkByNameOrID("net1")
	if err != nil {
		t.Errorf("Unexpected error when loading network net1")
	}
	if net1.ID != "1" {
		t.Errorf("Invalid id for network net1")
	}
	net1, err = db.GetNetworkByNameOrID("1")
	if err != nil {
		t.Errorf("Unexpected error when loading network net1")
	}
	if err := db.DeleteNetwork(net1); err != nil {
		t.Errorf("Unexpected error when deleting network net1")
	}
	net1, err = db.GetNetworkByNameOrID("net1")
	if err == nil {
		t.Errorf("Expected error when loading deleted network net1")
	}

	netws, err := db.GetNetworksByIDs(map[string]interface{}{})
	if err != nil {
		t.Errorf("Unexpected error when loading networks by empty ids mapping")
	}
	if len(netws) != 0 {
		t.Errorf("Expected 0 networks for empty ids mapping")
	}
	netws, err = db.GetNetworksByIDs(map[string]interface{}{"2": 1, "3": 1})
	if err != nil {
		t.Errorf("Unexpected error when loading networks by ids mapping")
	}
	if len(netws) != 2 {
		t.Errorf("Expected 2 networks for empty ids mapping, but got %#v", netws)
	}
}
