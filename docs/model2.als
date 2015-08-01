module model2
open util/ordering[A] as ao
open util/ordering[D] as do
open util/ordering[M] as mo
open util/integer

sig A {}
sig D {}
sig M {}

sig S {
	arr: Index -> lone Int} 
	{
	all idx: arr.univ | idx.i <= #A and idx.j <= #D and idx.k <= #M 
	no 0 & univ.arr
	all j: {i: Int | i >= 1 and i <= #D}, k: {i: Int | i >= 1 and i <= #M} | 	
		(sum idx : (arr.univ).filter[j, k] | arr[idx]) = 0
	}

sig Index { 
	i, j, k: Int } 
	{ 
		i > 0 and j > 0 and k > 0
	}

sig DI { start: D, end: D } { start.lte[end] }
sig MI { start: M, end: M } { start.lte[end] }

fact NoDuplicatedIntervals {
	no disj di, di': DI | di.start = di'.start and di.end = di'.end
	no disj mi, mi': MI | mi.start = mi'.start and mi.end = mi'.end
	}

fact NoDuplicatedIndexes {
	no disj idx, idx': Index | idx.i = idx'.i and idx.j = idx'.j and idx.k = idx'.k
	}

pred append (disj s, s', s'': S) {
	max[s.arr.univ.k] < min[s'.arr.univ.k]
	s''.arr = s.arr + s'.arr
	}

pred slice (disj s, s': S, a': set A, d: set DI, m: set MI) {
	s'.arr in s.arr
	all idx: s.arr.univ | 
		(some (s.row[idx].i & a'.pos)) and (idx.j in d.pos) and (idx.k in m.pos) <=>
			idx in s'.arr.univ
	}

pred projection (disj s, s': S, a': set A, d: set DI, m: set MI) {
	no disj di, di': d | di.end.gte[di'.start] and di.start.lte[di'.end]
	no disj mi, mi': m | mi.end.gte[mi'.start] and mi.start.lte[mi'.end]
	all idx: s'.arr.univ | 
		(some s'.row[idx].i & a'.pos) and 
		(idx.j in d.start.pos) and 
		(idx.k in m.start.pos)
	all idx': s'.arr.univ |
		some d': d, m': m | 
			let plane = { idx: s.arr.univ | 
					idx.i = idx'.i and idx.j in d'.pos and idx.k in m'.pos } |
				idx'.j = d'.start.pos and idx'.k = m'.start.pos and
					s'.arr[idx'] = (sum p: plane | s.arr[p])
	}

fact SumOfTheWholeSpaceIsZero {
	all s: S | (sum idx: s.arr.univ | s.arr[idx]) = 0
	}

assert NoNonEligibleSpacesHaveAnSlice {
	no s, s': S, a': set A, d: set DI, m: set MI |
		(no idx: s.arr.univ | idx.i in a'.pos and idx.j in d.pos and idx.k in m.pos) and
			slice[s, s', a', d, m] and some s.arr and some s'.arr
	}
check NoNonEligibleSpacesHaveAnSlice for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int

assert AllIndexesWithAccountsIn_a'_RemainInTheSlicedSpace {
	all s, s': S, a': set A, d: set DI, m: set MI |
		slice[s, s', a', d, m] and #d.pos = #D and #m.pos = #M =>
			(all idx: s.arr.univ | idx.i in a'.pos => idx in s'.arr.univ)
	}
check AllIndexesWithAccountsIn_a'_RemainInTheSlicedSpace for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int

assert AllTransactionsInTheSlicedSpaceMustHaveAnAccountIn_a' {
	all s, s': S, a': set A, d: set DI, m: set MI |
		slice[s, s', a', d, m] =>
			(all j': s'.arr.univ.j, k': s'.arr.univ.k | 
				let row = {idx: s'.arr.univ | idx.j = j' and idx.k = k'} |
					no row or some row.i & a'.pos)
	}
check AllTransactionsInTheSlicedSpaceMustHaveAnAccountIn_a' for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int

assert SlicedSpaceIsNoBiggerThanOriginalSpace {
	all s, s': S, a': set A, d: set DI, m: set MI |
		slice[s, s', a', d, m] => #s'.arr <= #s.arr
	}
check SlicedSpaceIsNoBiggerThanOriginalSpace for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int

assert SlicedSpaceValuesAreEqualToOriginalSpace {
	all s, s': S, a': set A, d: set DI, m: set MI |
		slice[s, s', a', d, m] => (all idx: s'.arr.univ | s.arr[idx] = s'.arr[idx])
	}
check SlicedSpaceValuesAreEqualToOriginalSpace for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int

assert NoNonEligibleSpacesHaveAnProjection {
	no s, s': S, a': set A, d: set DI, m: set MI |
		(no idx: s.arr.univ | idx.i in a'.pos and idx.j in d.pos and idx.k in m.pos) and
			projection[s, s', a', d, m] and some s.arr and some s'.arr
	}
check NoNonEligibleSpacesHaveAnProjection for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int

assert AllTransactionsInTheProjectedSpaceMustHaveAnAccountIn_s' {
	all s, s': S, a': set A, d: set DI, m: set MI |
		projection[s, s', a', d, m] => (all i': s'.arr.univ.i | one i' & s.arr.univ.i)
	}
check AllTransactionsInTheProjectedSpaceMustHaveAnAccountIn_s' for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int

assert ProjectedSpaceIsNoBiggerThanOriginalSpace {
	all s, s': S, a': set A, d: set DI, m: set MI |
		projection[s, s', a', d, m] => #s'.arr <= #s.arr
	}
check ProjectedSpaceIsNoBiggerThanOriginalSpace for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int

fun filter (indexes: set Index, j', k': Int): set Index {
	{ idx: indexes | idx.j = j' and idx.k = k' }
	}

fun pos (a: set A): set Int {
	{ i: Int | some a': a | i = 1.add[#a'.prevs] }
	}

fun pos (d: D): set Int {
	{ i: Int | some d': d | i = 1.add[#d'.prevs] }
	}

fun pos (m: M): set Int {
	{ i: Int | some m': m | i = 1.add[#m'.prevs] }
	}

fun pos (d: set DI): set Int {
	{i: Int | some d': d | i >= 1.add[#d'.start.prevs] and i <= 1.add[#d'.end.prevs]}
	}

fun pos (m: set MI): set Int {
	{i: Int | some m': m | i >= 1.add[#m'.start.prevs] and i <= 1.add[#m'.end.prevs]}
	}

fun row (s: S, idx: Index): set Index {
	{ idx': s.arr.univ | idx'.j = idx.j and idx'.k = idx.k }
	}

pred small(s: S) {
	#s.arr.univ.i <= 5
	#s.arr.univ.j <= 5
	#s.arr.univ.k <= 5
	all v: s.arr[univ] | v >= -5 and v <= 5 
	}

pred interesting(s: S) {
	some s.arr
	}

pred veryInteresting(s: S) {
	#s.arr.univ.j > 1
	#s.arr.univ.k > 1
	some disj idx, idx': s.arr.univ | idx.j != idx'.j and idx.k = idx'.k
	some disj idx, idx', idx'': s.arr.univ | 
		idx.j = idx'.j and idx.j = idx''.j and idx.k = idx'.k and idx.k = idx''.k
	}

pred allSmallAndInteresting {
	some S
	all s: S | s.small and s.interesting
	some s: S | s.small and s.veryInteresting
	}

pred showAppend(disj s, s', s'': S) {
	allSmallAndInteresting
	append[s, s', s'']
	}

pred showSlice(disj s, s': S, a': set A, d: set DI, m: set MI) {
	allSmallAndInteresting
	#a' > 0 and #a' < 3
	#d.pos = 1
	#m.pos = 1
	slice[s, s', a', d, m]
	#s'.arr != #s.arr
	}

pred showProjection(disj s, s': S, a': set A, d: set DI, m: set MI) {
	allSmallAndInteresting
	#a' = #A
	#d = 1 and #d.pos = #D
	#m = 1 and #m.pos = #M
	projection[s, s', a', d, m]
	#s'.arr != #s.arr
	}

run allSmallAndInteresting for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int
run showAppend for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int
run showSlice for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int
run showProjection for 7 but 3 A, 3 D, 3 M, 3 S, 6 Int
