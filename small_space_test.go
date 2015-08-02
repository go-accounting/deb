package deb

import "testing"

type SmallSpaceBuilder int

func (SmallSpaceBuilder) NewSpace(arr Array, metadata [][][]byte) Space {
	return NewSmallSpace(arr, metadata)
}

func (SmallSpaceBuilder) NewSpaceWithOffset(arr Array, do, mo int, metadata [][][]byte) Space {
	return NewSmallSpaceWithOffset(arr, uint64(do), uint64(mo), metadata)
}

func TestSmallSpaceTransactions(t *testing.T) {
	SpaceTester(0).TestTransactions(t, SmallSpaceBuilder(0))
}
func TestSmallSpaceAppend(t *testing.T) {
	SpaceTester(0).TestAppend(t, SmallSpaceBuilder(0))
}
func TestSmallSpaceSlice(t *testing.T) {
	SpaceTester(0).TestSlice(t, SmallSpaceBuilder(0))
}
func TestSmallSpaceProjection(t *testing.T) {
	SpaceTester(0).TestProjection(t, SmallSpaceBuilder(0))
}
