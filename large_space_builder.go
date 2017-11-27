package deb

type LargeSpaceBuilder int

func (lsb LargeSpaceBuilder) NewSpace(arr Array, metadata [][][]byte) Space {
	return lsb.NewSpaceWithOffset(arr, 0, 0, metadata)
}

func (LargeSpaceBuilder) NewSpaceWithOffset(arr Array, do, mo int, metadata [][][]byte) Space {
	blocks := []*DataBlock{}
	errc := make(chan error, 1)
	in := func() chan *DataBlock {
		c := make(chan *DataBlock)
		go func() {
			for i, block := range blocks {
				block.Key = i
				c <- block
			}
			close(c)
			errc <- nil
		}()
		return c
	}
	out := make(chan []*DataBlock)
	go func() {
		for blocks_ := range out {
			for _, block := range blocks_ {
				if block.Key == nil {
					blocks = append(blocks, block)
				} else {
					index := block.Key.(int)
					blocks[index] = block
				}
			}
			errc <- nil
		}
	}()
	ls := NewLargeSpace(1014*1024, in, out, errc)
	if arr != nil {
		ls.Append(NewSmallSpaceWithOffset(arr, uint64(do), uint64(mo), metadata))
	}
	return ls
}
