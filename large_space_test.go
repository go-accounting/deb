package deb

import "testing"

func TestLargeSpaceTransactions(t *testing.T) {
	SpaceTester(0).TestTransactions(t, LargeSpaceBuilder(0))
}

func TestLargeSpaceAppend(t *testing.T) {
	SpaceTester(0).TestAppend(t, LargeSpaceBuilder(0))
}

func TestLargeSpaceSlice(t *testing.T) {
	SpaceTester(0).TestSlice(t, LargeSpaceBuilder(0))
}

func TestLargeSpaceProjection(t *testing.T) {
	SpaceTester(0).TestProjection(t, LargeSpaceBuilder(0))
}
