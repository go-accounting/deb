package deb

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
	ls := newLargeSpace(1014*1024, in, out, errc)
	if arr != nil {
		ls.Append(NewSmallSpaceWithOffset(arr, uint64(do), uint64(mo), metadata))
	}
	return ls
}
