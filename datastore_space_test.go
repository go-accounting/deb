// +build appengine

package deb

import (
	"fmt"
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
		fmt.Println(err)
		return
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
		fmt.Println(err)
		return nil
	}
	if arr != nil {
		if err := ds.Append(NewSmallSpaceWithOffset(arr, uint64(do), uint64(mo),
			metadata)); err != nil {
			fmt.Println(err)
			return nil
		}
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

/*
func TestDatastoreSpaceBigSpace(t *testing.T) {
	size := 50
	space := DatastoreSpaceBuilder(0).NewSpace(nil, nil)
	arr := make([][][]int64, size)
	md := make([][][]byte, 1)
	md[0] = make([][]byte, size)
	tx := [][]int64{{1, -1}}
	md_ := []byte{1}
	for i := 0; i < size; i++ {
		arr[i] = tx
		md[0][i] = md_
	}
	ss := NewSmallSpace(Array(arr).Transposed(), md)
	if err := space.Append(ss); err != nil {
		fmt.Println(err)
	}
}
*/
