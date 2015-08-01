package deb

import (
	"reflect"
	"testing"
)

type SpaceTester int

type SpaceBuilder interface {
	NewSpace(Array, [][][]byte) Space
	NewSpaceWithOffset(Array, int, int, [][][]byte) Space
}

func (SpaceTester) TestTransactions(t *testing.T, b SpaceBuilder) {
	spaces := []Space{
		b.NewSpace(Array{{{1, -1}}}.Transposed(), [][][]byte{{{1}}}),
		b.NewSpaceWithOffset(Array{{{1, -1}}}.Transposed(), 1, 2, [][][]byte{{{2}}}),
	}
	cases := [][]Transaction{
		{Transaction{Moment(1), Date(1), Entries{Account(1): 1, Account(2): -1}, []byte{1}}},
		{Transaction{Moment(3), Date(2), Entries{Account(1): 1, Account(2): -1}, []byte{2}}},
	}
	assertSpaceTransactions(t, spaces, cases)
}

func (SpaceTester) TestAppend(t *testing.T, b SpaceBuilder) {
	spaces := []Space{
		b.NewSpace(Array{{{1, -1}}}.Transposed(), [][][]byte{{{1}}}),
	}
	arguments := []Space{
		b.NewSpaceWithOffset(Array{{{2, -2}}}.Transposed(), 0, 1, [][][]byte{{{2}}}),
	}
	cases := []Space{
		b.NewSpace(Array{{{1, -1}}, {{2, -2}}}.Transposed(), [][][]byte{{{1}, {2}}}),
	}
	spaces[0].Append(arguments[0])
	assertSpaces(t, spaces, cases)
}

func (SpaceTester) TestSlice(t *testing.T, b SpaceBuilder) {
	spaces := []Space{
		b.NewSpace(Array{{{1, -1}}, {{2, -2}}}.Transposed(), [][][]byte{{{1}, {2}}}),
	}
	arguments := []struct {
		a []Account
		d []DateRange
		m []MomentRange
	}{
		{[]Account{Account(1), Account(2)},
			[]DateRange{DateRange{1, 1}},
			[]MomentRange{MomentRange{2, 2}}},
	}
	cases := []Space{
		b.NewSpace(Array{{{0, 0}}, {{2, -2}}}.Transposed(), [][][]byte{{[]byte(nil), {2}}}),
	}
	for i := range spaces {
		spaces[i], _ = spaces[i].Slice(arguments[i].a, arguments[i].d, arguments[i].m)
	}
	assertSpaces(t, spaces, cases)
}

func (SpaceTester) TestProjection(t *testing.T, b SpaceBuilder) {
	spaces := []Space{
		b.NewSpace(Array{{{1, -1}}, {{2, -2}}}.Transposed(), [][][]byte{{{1}, {2}}}),
	}
	arguments := []struct {
		a []Account
		d []DateRange
		m []MomentRange
	}{
		{[]Account{Account(1), Account(2)}, []DateRange{DateRange{1, 1}},
			[]MomentRange{MomentRange{1, 2}}},
	}
	cases := []Space{
		b.NewSpace(Array{{{3, -3}}, {{0, 0}}}.Transposed(), nil),
	}
	for i := range spaces {
		spaces[i], _ = spaces[i].Projection(arguments[i].a, arguments[i].d, arguments[i].m)
	}
	assertSpaces(t, spaces, cases)
}

func assertSpaceTransactions(t *testing.T, spaces []Space, cases [][]Transaction) {
	for i, s := range spaces {
		j := 0
		c, _ := s.Transactions()
		for tx := range c {
			if !reflect.DeepEqual(*tx, cases[i][j]) {
				t.Errorf("%v != %v", *tx, cases[i][j])
			}
			j++
		}
		if j < len(cases[i]) {
			t.Errorf("len(transactions[%v]):%v != len(cases):%v", i, j, len(cases[i]))
		}
	}
}

func assertSpaces(t *testing.T, spaces []Space, cases []Space) {
	for i, s := range spaces {
		if s == nil {
			t.Errorf("Space %v is nil", i)
			continue
		}
		if cases[i] == nil {
			t.Errorf("Case %v is nil", i)
			continue
		}
		c1, _ := s.Transactions()
		c2, _ := cases[i].Transactions()
		for {
			tx1, ok1 := <-c1
			tx2, ok2 := <-c2
			if !ok1 && !ok2 {
				break
			}
			if !ok1 || !ok2 {
				t.Errorf("closed: %v != closed: %v", !ok1, !ok2)
			}
			if !reflect.DeepEqual(*tx1, *tx2) {
				t.Errorf("%v != %v", *tx1, *tx2)
			}
		}
	}
}
