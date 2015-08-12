package deb

type ChannelSpace chan *Transaction

func (c ChannelSpace) Append(s Space) error {
	panic("Not implemented")
}

func (c ChannelSpace) Slice([]Account, []DateRange, []MomentRange) (Space, error) {
	panic("Not implemented")
}

func (c ChannelSpace) Projection([]Account, []DateRange, []MomentRange) (Space, error) {
	panic("Not implemented")
}

func (c ChannelSpace) Transactions() (chan *Transaction, chan error) {
	errc := make(chan error, 1)
	errc <- nil
	defer close(errc)
	return c, errc
}
