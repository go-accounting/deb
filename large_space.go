package deb

import "fmt"

type LargeSpace struct {
	blockSize uint
	in        func() chan *DataBlock
	out       chan []*DataBlock
	errc      chan error
}

type DataBlock struct {
	Key interface{}
	M   []int64 // Moments
	D   []int32 // Dates
	A   []int32 // Accounts
	V   []int64 // Values
	B   []int16 // Accounts bound
	MD  []byte  // Metadata
	BMD []int32 // Metadata bounds
}

func NewLargeSpace(
	blockSize uint,
	in func() chan *DataBlock,
	out chan []*DataBlock,
	errc chan error,
) *LargeSpace {
	return &LargeSpace{blockSize: blockSize, in: in, out: out, errc: errc}
}

func (ls *LargeSpace) Append(s Space) error {
	var (
		lastBlock *DataBlock
		blocks    []*DataBlock
	)
	c, errc := s.Transactions()
	count := 0
	moments := map[int64]int{}
	for t := range c {
		if block, err := ls.freeBlock(lastBlock, t); err != nil {
			return err
		} else {
			if block == nil {
				block = ls.NewDataBlock()
			}
			if block != lastBlock {
				lastBlock = block
				blocks = append(blocks, block)
			}
			block.append(t)
		}
		count++
		moments[int64(t.Moment)]++
	}
	if logger != nil {
		logger(fmt.Sprintf("largeSpace.Append: %v transactions, %v moments, %v blocks\n",
			count, len(moments), len(blocks)))
	}
	if err := <-errc; err != nil {
		return err
	}
	ls.out <- blocks
	if err := <-ls.errc; err != nil {
		return err
	}
	return nil
}

func (ls *LargeSpace) Slice(a []Account, d []DateRange, m []MomentRange) (Space, error) {
	out := make(chan *Transaction)
	var err error
	go func() {
		defer close(out)
		err = ls.iterateWithFilter(a, d, m, func(block *DataBlock, i int) {
			out <- block.newTransaction(i)
		})
	}()
	return ChannelSpace(out), err
}

func (ls *LargeSpace) Projection(a []Account, d []DateRange, m []MomentRange) (Space, error) {
	type key struct {
		moment Moment
		date   Date
	}
	transactions := map[key]*Transaction{}
	err := ls.iterateWithFilter(a, d, m, func(block *DataBlock, i int) {
		k := key{startMoment(m, Moment(block.M[i])), startDate(d, Date(block.D[i]))}
		nt := block.newTransaction(i)
		if t, ok := transactions[k]; !ok {
			transactions[k] = nt
			transactions[k].Metadata = []byte{}
		} else {
			for ek, ev := range nt.Entries {
				if ov, ok := t.Entries[ek]; ok {
					t.Entries[ek] = ov + ev
				} else {
					t.Entries[ek] = ev
				}
			}
		}
	})
	if err != nil {
		return nil, err
	}
	out := make(chan *Transaction)
	go func() {
		defer close(out)
		for _, t := range transactions {
			out <- t
		}
	}()
	return ChannelSpace(out), nil
}

func (ls *LargeSpace) Transactions() (chan *Transaction, chan error) {
	out := make(chan *Transaction)
	go func() {
		defer close(out)
		for block := range ls.in() {
			for i := 0; i < len(block.M); i++ {
				out <- block.newTransaction(i)
			}
		}
	}()
	return out, ls.errc
}

func (ls *LargeSpace) String() string {
	blocksAsString := []string{}
	count := 0
	for block := range ls.in() {
		blocksAsString = append(blocksAsString, fmt.Sprintf("%v", *block))
		count += 1
	}
	err := <-ls.errc
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("{%v %v %v %v}", ls.blockSize, count, ls.capacity(), blocksAsString)
}

func (ls *LargeSpace) Pop() (int32, []int64, error) {
	var block *DataBlock
	for b := range ls.in() {
		if len(b.M) > 0 && len(b.A) > 0 && len(b.MD) > 0 {
			block = b
		}
	}
	if err := <-ls.errc; err != nil {
		return 0, nil, err
	}
	if block == nil {
		return 0, nil, fmt.Errorf("No nonempty block found")
	}
	mLen := len(block.M)
	aLen := len(block.A)
	mdLen := len(block.MD)
	if mLen == 0 || aLen == 0 || mdLen == 0 {
		return 0, nil, fmt.Errorf("At least one empty array M:%v A:%v MD:%v", mLen, aLen, mdLen)
	}
	entriesCount := block.B[mLen*2-1 : mLen*2][0] - block.B[mLen*2-2 : mLen*2-1][0]
	metadataSize := block.BMD[mLen*2-1 : mLen*2][0] - block.BMD[mLen*2-2 : mLen*2-1][0]
	block.M = block.M[0 : mLen-1]
	d := block.D[mLen-1]
	v := block.V[int16(aLen)-entriesCount:]
	block.D = block.D[0 : mLen-1]
	block.A = block.A[0 : int16(aLen)-entriesCount]
	block.V = block.V[0 : int16(aLen)-entriesCount]
	block.B = block.B[0 : mLen*2-2]
	block.MD = block.MD[0 : int32(mdLen)-metadataSize]
	block.BMD = block.BMD[0 : mLen*2-2]
	ls.out <- []*DataBlock{block}
	return d, v, <-ls.errc
}

func (block *DataBlock) newTransaction(i int) *Transaction {
	t := Transaction{Moment(block.M[i]), Date(block.D[i]), make(Entries), nil}
	t.Metadata = block.MD[block.BMD[i*2]:block.BMD[i*2+1]]
	for j := block.B[i*2]; j < block.B[i*2+1]; j++ {
		t.Entries[Account(block.A[j])] = block.V[j]
	}
	return &t
}

func (ls *LargeSpace) capacity() uint {
	return (ls.blockSize / 2) / (64 + 32 + 32*2 + 64*2 + 16*2 + 32*2)
}

func (ls *LargeSpace) freeBlock(block *DataBlock, t *Transaction) (*DataBlock, error) {
	if block != nil && block.hasRoomFor(t, ls) {
		return block, nil
	}
	var result *DataBlock
	for block := range ls.in() {
		if block.hasRoomFor(t, ls) {
			result = block
		}
	}
	return result, <-ls.errc
}

func (ls *LargeSpace) NewDataBlock() *DataBlock {
	block := new(DataBlock)
	block.M = make([]int64, 0, ls.capacity())
	block.D = make([]int32, 0, ls.capacity())
	block.A = make([]int32, 0, ls.capacity()*2)
	block.V = make([]int64, 0, ls.capacity()*2)
	block.B = make([]int16, 0, ls.capacity()*2)
	block.MD = make([]byte, 0, ls.blockSize/2)
	block.BMD = make([]int32, 0, ls.capacity()*2)
	return block
}

func (ls *LargeSpace) iterateWithFilter(a []Account, d []DateRange, m []MomentRange,
	f func(*DataBlock, int)) error {
	for block := range ls.in() {
		for i := 0; i < len(block.M); i++ {
			if containsMoment(m, Moment(block.M[i])) && containsDate(d, Date(block.D[i])) {
				for j := block.B[i*2]; j < block.B[i*2+1]; j++ {
					if containsAccount(a, Account(block.A[j])) {
						f(block, i)
						break
					}
				}
			}
		}
	}
	return <-ls.errc
}

func (block *DataBlock) append(t *Transaction) {
	mLen := len(block.M)
	aLen := len(block.A)
	mdLen := len(block.MD)
	block.M = block.M[0 : mLen+1]
	block.D = block.D[0 : mLen+1]
	block.A = block.A[0 : aLen+len(t.Entries)]
	block.V = block.V[0 : aLen+len(t.Entries)]
	block.B = block.B[0 : mLen*2+2]
	block.BMD = block.BMD[0 : mLen*2+2]
	block.M[mLen] = int64(t.Moment)
	block.D[mLen] = int32(t.Date)
	i := 0
	for a, v := range t.Entries {
		block.A[aLen+i] = int32(a)
		block.V[aLen+i] = v
		i++
	}
	block.B[mLen*2] = int16(aLen)
	block.B[mLen*2+1] = int16(aLen + len(t.Entries))
	if t.Metadata != nil {
		block.BMD[mLen*2] = int32(mdLen)
		block.BMD[mLen*2+1] = int32(mdLen + len(t.Metadata))
		block.MD = block.MD[0 : mdLen+len(t.Metadata)]
		copy(block.MD[mdLen:mdLen+len(t.Metadata)], t.Metadata)
	} else {
		block.BMD[mLen*2] = 0
		block.BMD[mLen*2+1] = 0
	}
}

func (block *DataBlock) hasRoomFor(t *Transaction, ls *LargeSpace) bool {
	return uint(len(block.A)+len(t.Entries)) <= ls.capacity()*2 &&
		(t.Metadata == nil || uint(len(block.MD)+len(t.Metadata)) <= ls.blockSize/2)
}

func startDate(d []DateRange, elem Date) Date {
	for _, each := range d {
		if each.Start <= elem && each.End >= elem {
			return each.Start
		}
	}
	return 0
}

func startMoment(m []MomentRange, elem Moment) Moment {
	for _, each := range m {
		if each.Start <= elem && each.End >= elem {
			return each.Start
		}
	}
	return 0
}
