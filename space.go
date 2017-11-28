package deb

import "time"

type Account uint32
type Date uint32
type Moment uint64
type DateRange struct{ Start, End Date }
type MomentRange struct{ Start, End Moment }
type Entries map[Account]int64

type Transaction struct {
	Moment   Moment
	Date     Date
	Entries  Entries
	Metadata []byte
}

type Space interface {
	Append(Space) error
	Slice([]Account, []DateRange, []MomentRange) (Space, error)
	Projection([]Account, []DateRange, []MomentRange) (Space, error)
	Transactions() (chan *Transaction, chan error)
}

var logger func(string)

func RegisterLogger(l func(string)) {
	logger = l
}

func (d Date) ToTime() time.Time {
	return time.Date(int(d%100000000/10000), time.Month(d%10000/100), int(d%100), 0, 0, 0, 0, time.UTC)
}

func DateFromTime(t time.Time) Date {
	return Date(t.Year()*10000 + int(t.Month())*100 + t.Day())
}

func (m Moment) ToTime() time.Time {
	return time.Unix(0, int64(m))
}

func MomentFromTime(t time.Time) Moment {
	return Moment(t.UnixNano())
}
