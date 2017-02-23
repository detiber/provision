package backend

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"

	"github.com/digitalrebar/digitalrebar/go/common/store"
)

var (
	backingStore store.SimpleStore
	tmpDir       string
)

func loadExample(dt *DataTracker, kind, p string) (bool, error) {
	buf, err := os.Open(p)
	if err != nil {
		return false, err
	}
	defer buf.Close()
	var res store.KeySaver
	switch kind {
	case "users":
		res = dt.NewUser()
	case "machines":
		res = dt.NewMachine()
	case "templates":
		res = dt.NewTemplate()
	case "bootenvs":
		res = dt.NewBootEnv()
	case "leases":
		res = dt.NewLease()
	case "reservations":
		res = dt.NewReservation()
	case "subnets":
		res = dt.NewSubnet()
	}

	dec := json.NewDecoder(buf)
	if err := dec.Decode(&res); err != nil {
		return false, err
	}
	return dt.create(res)
}

func mkDT(bs store.SimpleStore) *DataTracker {
	dt := NewDataTracker(bs, true, true, tmpDir, "CURL", "local", "local", "FURL", "127.0.0.1", log.New(os.Stdout, "dt", 0))
	if err := dt.ExtractAssets(); err != nil {
		log.Printf("Unable to extract assets: %v", err)
		os.Exit(1)
	}
	return dt
}

func TestBackingStorePersistence(t *testing.T) {
	bs, err := store.NewSimpleLocalStore(tmpDir)
	if err != nil {
		t.Errorf("Could not create boltdb: %v", err)
		return
	}
	dt := mkDT(bs)
	explDirs := []string{"users",
		"templates",
		"bootenvs",
		"machines",
		"leases",
		"reservations",
		"subnets",
	}

	for _, d := range explDirs {
		p := path.Join("test-data", d, "default.json")
		created, err := loadExample(dt, d, p)
		if !created {
			t.Errorf("Error loading test data: %v", err)
			return
		}
	}
	t.Logf("Example data loaded into the data tracker")
	t.Logf("Creating new DataTracker using the same backing store")
	dt = nil
	dt = mkDT(bs)
	// There should be one of everything in the cache now.
	for _, ot := range explDirs {
		var items []store.KeySaver
		switch ot {
		case "users":
			items = dt.fetchAll(dt.NewUser())
		case "templates":
			items = dt.fetchAll(dt.NewTemplate())
		case "bootenvs":
			items = dt.fetchAll(dt.NewBootEnv())
		case "machines":
			items = dt.fetchAll(dt.NewMachine())
		case "leases":
			items = dt.fetchAll(dt.NewLease())
		case "reservations":
			items = dt.fetchAll(dt.NewReservation())
		case "subnets":
			items = dt.fetchAll(dt.NewSubnet())
		}
		if len(items) != 1 {
			t.Errorf("Expected to find 1 %s, instead found %d", ot, len(items))
		} else {
			t.Logf("Found 1 %s, as expected", ot)
		}
	}
}

func TestMain(m *testing.M) {
	var err error
	tmpDir, err = ioutil.TempDir("", "datatracker-")
	if err != nil {
		log.Printf("Creating temp dir for file root failed: %v", err)
		os.Exit(1)
	}
	ret := m.Run()
	os.RemoveAll(tmpDir)
	os.Exit(ret)
}