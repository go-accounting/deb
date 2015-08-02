// +build appengine

package deb

import (
	"os"
	"testing"

	"appengine/aetest"
)

type DatastoreSpaceBuilder int

var ctx aetest.Context

func TestMain(m *testing.M) {
	var err error
	ctx, err = aetest.NewContext(nil)
	if err != nil {
		panic(err)
	}
	code := m.Run()
	ctx.Close()
	os.Exit(code)
}

func (dsb DatastoreSpaceBuilder) NewSpace(arr Array, metadata [][][]byte) Space {
	return dsb.NewSpaceWithOffset(arr, 0, 0, metadata)
}

func (DatastoreSpaceBuilder) NewSpaceWithOffset(arr Array, do, mo int, metadata [][][]byte) Space {
	ds, _, err := NewDatastoreSpace(ctx, nil)
	if err != nil {
		panic(err)
	}
	if err := ds.Append(NewSmallSpaceWithOffset(arr, do, mo, metadata)); err != nil {
		panic(err)
	}
	return ds
}

func TestDatastoreSpaceTransactions(t *testing.T) {
	SpaceTester(0).TestTransactions(t, DatastoreSpaceBuilder(0))
}
func TestDatastoreSpaceAppend(t *testing.T) {
	SpaceTester(0).TestAppend(t, DatastoreSpaceBuilder(0))
}
func TestDatastoreSpaceSlice(t *testing.T) {
	SpaceTester(0).TestSlice(t, DatastoreSpaceBuilder(0))
}
func TestDatastoreSpaceProjection(t *testing.T) {
	SpaceTester(0).TestProjection(t, DatastoreSpaceBuilder(0))
}
