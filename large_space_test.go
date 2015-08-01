package deb

import "testing"

type LargeSpaceBuilder int

func (lsb LargeSpaceBuilder) NewSpace(arr Array, metadata [][][]byte) Space {
	return lsb.NewSpaceWithOffset(arr, 0, 0, metadata)
}

func (LargeSpaceBuilder) NewSpaceWithOffset(arr Array, do, mo int, metadata [][][]byte) Space {
	blocks := []*dataBlock{}
	errc := make(chan error, 1)
	in := func() chan *dataBlock {
		c := make(chan *dataBlock)
		go func() {
			for i, block := range blocks {
				block.key = i
				c <- block
			}
			close(c)
			errc <- nil
		}()
		return c
	}
	out := make(chan *dataBlock)
	go func() {
		for block := range out {
			if block.key == nil {
				blocks = append(blocks, block)
			} else {
				index := block.key.(int)
				blocks[index] = block
			}
			errc <- nil
		}
	}()
	ls := newLargeSpace(10*1024, in, out, errc)
	ls.Append(NewSmallSpaceWithOffset(arr, do, mo, metadata))
	return ls
}

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
