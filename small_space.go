package deb

import (
	"math"

	xmath "github.com/mcesarhm/geek-accounting/go-server/extensions/math"
)

type smallSpace struct {
	arr                      Array
	dateOffset, momentOffset uint64
	metadata                 [][][]byte
}

// NewSmallSpace creates a Space based on the given array and metadata.
// Metadata is a []byte and is indexed by date and moment,
// therefore the type [][][]byte.
func NewSmallSpace(arr Array, metadata [][][]byte) Space {
	return NewSmallSpaceWithOffset(arr, 0, 0, metadata)
}

// NewSmallSpaceWithOffset creates a Space based on the given array, offsets and metadata.
// Metadata is a []byte and is indexed by date and moment,
// therefore the type [][][]byte.
func NewSmallSpaceWithOffset(arr Array, dateOffset, momentOffset uint64,
	metadata [][][]byte) Space {
	return &smallSpace{arr.Copy(), dateOffset, momentOffset, metadata}
}

func (ss *smallSpace) Append(s Space) error {
	var (
		y, z int
		md   [][][]byte
	)
	_, y0, z0 := ss.arr.Dimensions()
	makeMd := func() {
		_, y, z = ss.arr.Dimensions()
		md = make([][][]byte, y)
		for i := 0; i < y; i++ {
			md[i] = make([][]byte, z)
			for j := 0; j < z; j++ {
				if ss.metadata != nil && i < y0 && j < z0 {
					md[i][j] = ss.metadata[i][j]
				}
			}
		}
	}
	if other_ss, ok := s.(*smallSpace); ok {
		do := other_ss.dateOffset - ss.dateOffset
		mo := other_ss.momentOffset - ss.momentOffset
		ss.arr.Append(&other_ss.arr, do, mo)
		if ss.metadata != nil || other_ss.metadata != nil {
			makeMd()
			for i := 0; i < y; i++ {
				for j := 0; j < z; j++ {
					if other_ss.metadata != nil && j >= z0 {
						md[i][j] =
							other_ss.metadata[uint64(i)-other_ss.dateOffset][uint64(j)-other_ss.momentOffset]
					}
				}
			}
			ss.metadata = md
		}
		return nil
	} else {
		// Computes the array size
		var minDate, minMoment, maxAccount, maxDate, maxMoment uint64
		minDate, minMoment = math.MaxUint64, math.MaxUint64
		c, errc := s.Transactions()
		hasMetadata := false
		for t := range c {
			if t.Date < Date(minDate) {
				minDate = uint64(t.Date)
			}
			if t.Moment < Moment(minMoment) {
				minMoment = uint64(t.Moment)
			}
			for a, _ := range t.Entries {
				if a > Account(maxAccount) {
					maxAccount = uint64(a)
				}
			}
			if t.Date > Date(maxDate) {
				maxDate = uint64(t.Date)
			}
			if t.Moment > Moment(maxMoment) {
				maxMoment = uint64(t.Moment)
			}
			if t.Metadata != nil {
				hasMetadata = true
			}
		}
		if err := <-errc; err != nil {
			return err
		}
		if maxAccount == 0 || maxDate == 0 || maxMoment == 0 {
			return nil
		}
		// Builds the array
		other_arr := NewArray(int(maxAccount), int(maxDate-minDate), int(maxMoment-minMoment))
		if ss.metadata != nil || hasMetadata {
			makeMd()
		}
		c, errc = s.Transactions()
		for t := range c {
			for a, v := range t.Entries {
				other_arr[a-1][t.Date-1][t.Moment-1] = v
			}
			if hasMetadata {
				d := uint64(t.Date) - uint64(minDate) - ss.dateOffset
				m := uint64(t.Moment) - uint64(minMoment) - ss.momentOffset
				md[d][m] = t.Metadata
			}
		}
		ss.arr.Append(&other_arr, uint64(minDate), uint64(minMoment))
		if ss.metadata != nil || hasMetadata {
			ss.metadata = md
		}
		return <-errc
	}
}

func (ss *smallSpace) Slice(a []Account, d []DateRange, m []MomentRange) (Space, error) {
	result := smallSpace{ss.arr.Copy(), ss.dateOffset, ss.momentOffset, ss.metadata}
	x, y, z := ss.arr.Dimensions()
	for i := 0; i < x; i++ {
		for j := 0; j < y; j++ {
			for k := 0; k < z; k++ {
				if !containsDate(d, Date(uint64(j+1)+ss.dateOffset)) ||
					!containsMoment(m, Moment(uint64(k+1)+ss.momentOffset)) {
					result.arr[i][j][k] = 0
					if ss.metadata != nil {
						ss.metadata[j][k] = nil
					}
				}
			}
		}
	}
	for j := 0; j < y; j++ {
		for k := 0; k < z; k++ {
			found := false
			for i := 0; i < x; i++ {
				if containsAccount(a, Account(i+1)) && result.arr[i][j][k] != 0 {
					found = true
					break
				}
			}
			if !found {
				for i := 0; i < x; i++ {
					result.arr[i][j][k] = 0
				}
				if ss.metadata != nil {
					ss.metadata[j][k] = nil
				}
			}
		}
	}
	return &result, nil
}

func (ss *smallSpace) Projection(a []Account, d []DateRange, m []MomentRange) (Space, error) {
	result := smallSpace{ss.arr.Copy(), ss.dateOffset, ss.momentOffset, nil}
	x, y, z := ss.arr.Dimensions()
	for i := 0; i < x; i++ {
		for _, each_dr := range d {
			for _, each_mr := range m {
				sum := int64(0)
				j_start := xmath.Min(int(each_dr.Start-1-Date(ss.dateOffset)), y)
				j_end := xmath.Min(int(each_dr.End-Date(ss.dateOffset)), y)
				k_start := xmath.Min(int(each_mr.Start-1-Moment(ss.momentOffset)), z)
				k_end := xmath.Min(int(each_mr.End-Moment(ss.momentOffset)), z)
				for j := j_start; j < j_end; j++ {
					for k := k_start; k < k_end; k++ {
						sum += result.arr[i][j][k]
					}
				}
				result.arr[i][j_start][k_start] = sum
			}
		}
	}
	for j := 0; j < y; j++ {
		for k := 0; k < z; k++ {
			found := false
			if containsDateStart(d, Date(uint64(j+1)+ss.dateOffset)) &&
				containsMomentStart(m, Moment(uint64(k+1)+ss.momentOffset)) {
				for i := 0; i < x; i++ {
					if containsAccount(a, Account(i+1)) && result.arr[i][j][k] != 0 {
						found = true
						break
					}
				}
			}
			if !found {
				for i := 0; i < x; i++ {
					result.arr[i][j][k] = 0
				}
			}
		}
	}
	return &result, nil
}

func (ss *smallSpace) Transactions() (chan *Transaction, chan error) {
	out := make(chan *Transaction)
	errc := make(chan error, 1)
	go func() {
		defer close(out)
		if ss.arr.Empty() {
			return
		}
		x, y, z := ss.arr.Dimensions()
		for j := 0; j < y; j++ {
			for k := 0; k < z; k++ {
				var metadata []byte
				if ss.metadata != nil {
					metadata = ss.metadata[j][k]
				}
				t := Transaction{Moment(uint64(k+1) + ss.momentOffset),
					Date(uint64(j+1) + ss.dateOffset), make(Entries), metadata}
				for i := 0; i < x; i++ {
					if ss.arr[i][j][k] != 0 {
						t.Entries[Account(i+1)] = ss.arr[i][j][k]
					}
				}
				if len(t.Entries) > 0 {
					out <- &t
				}
			}
		}
		errc <- nil
	}()
	return out, errc
}

func containsAccount(a []Account, elem Account) bool {
	if a == nil || len(a) == 0 {
		return true
	}
	for _, each := range a {
		if each == elem {
			return true
		}
	}
	return false
}

func containsDate(d []DateRange, elem Date) bool {
	if d == nil || len(d) == 0 {
		return true
	}
	for _, each := range d {
		if each.Start <= elem && each.End >= elem {
			return true
		}
	}
	return false
}

func containsMoment(m []MomentRange, elem Moment) bool {
	if m == nil || len(m) == 0 {
		return true
	}
	for _, each := range m {
		if each.Start <= elem && each.End >= elem {
			return true
		}
	}
	return false
}

func containsDateStart(d []DateRange, elem Date) bool {
	for _, each := range d {
		if each.Start == elem {
			return true
		}
	}
	return false
}

func containsMomentStart(m []MomentRange, elem Moment) bool {
	for _, each := range m {
		if each.Start == elem {
			return true
		}
	}
	return false
}
