module model
open util/ordering[D] as do
open util/ordering[M] as mo
open util/integer

sig A {}
sig D {}
sig M {}

sig DI { start: D, end: D } { start.lte[end] }
sig MI { start: M, end: M } { start.lte[end] }

sig T { 
	m: MI, 
	d: DI, 
	e: A -> one Int } 
	{	
	#e.Int > 1
	not 0 in e[A]
	this.summation = 0
	}

fun summation (t: T) : Int {
	sum a : t.e.Int | t.e[a]
	}

sig S { 
	t: some T } {
	(sum t' : t | t'.summation) = 0
	}

fact noDistinctTransactionsWithSameDateAndMomentAllowed {
	no disj t, t': T | t.m = t'.m and t.d = t'.d
}

pred append (disj s, s', s'': S) {
	mo/max[s.t.m.end].lt[mo/min[s'.t.m.start]]
	s''.t = s.t + s'.t
	}

pred interestingSpace (s : S)  {
	#S > 1
	some t' : s.t | #t'.e.Int > 2 
	#s.t.d > 1
	#s.t.m > 1
	}

pred show(s: S) {
	all v: T.e[A] | v >= -5 and v <= 5
	interestingSpace[s]
	}
run show for 5 but 2 S, 3 A, 8 Int

pred showAppend(disj s, s', s'': S) {
	all v: T.e[A] | v >= -5 and v <= 5
	append[s, s', s'']
	interestingSpace[s'']
	}
run showAppend for 10 but 3 S, 6 T, 3 A, 2 D, 2 M, 2 DI, 2 MI, 8 Int
