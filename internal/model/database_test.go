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
	con.Name = "testymctestface"
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
	if conl, err := db.GetContainerByNameOrID(con.Name); err != nil {
		t.Errorf("Unexpected error when loading an existing container name; %s", err)
	} else {
		if conl.ID != con.ID || conl.Image != con.Image {
			t.Errorf("Loaded shortid container differs to saved container")
		}
	}
	if conl, err := db.GetContainerByNameOrID("somepodname-" + con.ShortID); err != nil {
		t.Errorf("Unexpected error when loading an existing container with a podname; %s", err)
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

func TestDeadlock(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Errorf("Unexpected error creating database: %s", err)
	}
	for i := 0; i < 1000; i++ {
		con := &types.Container{}
		if err := db.SaveContainer(con); err != nil {
			t.Errorf("Unexpected error when creating a new container: %s", err)
		}
		if err := db.DeleteContainer(con); err != nil {
			t.Errorf("Unexpected error when deleting container: %s", err)
		}
		netw := &types.Network{Name: "tb303"}
		if err := db.SaveNetwork(netw); err != nil {
			t.Errorf("Unexpected error when creating a new network: %s", err)
		}
		if err := db.DeleteNetwork(netw); err != nil {
			t.Errorf("Unexpected error when deleting network: %s", err)
		}
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
		netw := &types.Network{Name: n, ID: fmt.Sprintf("%d", i+1), ShortID: fmt.Sprintf("%d", i+1)}
		if err := db.SaveNetwork(netw); err != nil {
			t.Errorf("Unexpected error when creating network %s: %s", n, err)
		}
	}

	if netws, err := db.GetNetworks(); err != nil {
		t.Errorf("Unexpected error when loading all existing networks")
	} else {
		if len(netws) != 7 {
			t.Errorf("Expected 7 network records, but got %d", len(netws))
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
		t.Errorf("Unexpected error when loading network net1: %s", err)
	}
	if err := db.DeleteNetwork(net1); err != nil {
		t.Errorf("Unexpected error when deleting network net1: %s", err)
	}
	net1, err = db.GetNetworkByNameOrID("net1")
	if err == nil {
		t.Errorf("Expected error when loading deleted network net1: %s", err)
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

func TestImage(t *testing.T) {
	db, _ := New()

	if _, err := db.GetImageByNameOrID("roland/tb303:latest"); err == nil {
		t.Errorf("Expected an error when loading an non existing image")
	}

	img := &types.Image{Name: "roland/tr606:8.0.8"}
	if err := db.SaveImage(img); err != nil {
		t.Errorf("Unexpected error when creating image %s", err)
	}

	for i, n := range []string{"roland/tr606:9.0.9", "roland/tr808:9.0.9", "roland/tr606:3.0.3"} {
		img := &types.Image{Name: n, ID: fmt.Sprintf("%d", i+1), ShortID: fmt.Sprintf("%d", i+1)}
		if err := db.SaveImage(img); err != nil {
			t.Errorf("Unexpected error when creating image %s: %s", n, err)
		}
	}

	if imgs, err := db.GetImages(); err != nil {
		t.Errorf("Unexpected error when loading all existing images")
	} else {
		if len(imgs) != 4 {
			t.Errorf("Expected 4 network images, but got %d", len(imgs))
		}
	}

	img1, err := db.GetImageByNameOrID("roland/tr606:9.0.9")
	if err != nil {
		t.Errorf("Unexpected error when loading network img1")
	}
	if img1.ID != "1" {
		t.Errorf("Invalid id for image img1")
	}
	img1, err = db.GetImageByNameOrID("1")
	if err != nil {
		t.Errorf("Unexpected error when loading image img1: %s", err)
	}
	if err := db.DeleteImage(img1); err != nil {
		t.Errorf("Unexpected error when deleting image img1: %s", err)
	}
	img1, err = db.GetImageByNameOrID("roland/tr606:9.0.9")
	if err == nil {
		t.Errorf("Expected error when loading deleted image img1: %s", err)
	}
}
