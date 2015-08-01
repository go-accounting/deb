package deb

type channelSpace chan *Transaction

func (c channelSpace) Append(s Space) error {
	panic("Not implemented")
}

func (c channelSpace) Slice([]Account, []DateRange, []MomentRange) (Space, error) {
	panic("Not implemented")
}

func (c channelSpace) Projection([]Account, []DateRange, []MomentRange) (Space, error) {
	panic("Not implemented")
}

func (c channelSpace) Transactions() (chan *Transaction, chan error) {
	errc := make(chan error, 1)
	errc <- nil
	defer close(errc)
	return c, errc
}
