Computation Tree Logic model checker based on ROBDDs
====================================================
This is a minimal implementation of ROBDDs to execute model checking
(Computation Tree Logic; CTL). In efficient implementations the variable
ordering is very space sensitive, but since this is just a demo I use the order
in which variables are declared. There is no intermediate expression format,
the interface to define transitions directly constructs an ROBDD.

In `3_test.go` I show a more complex example of model checking to find deadlocks
in packet switching networks. The current implementation does not yet merge
duplicate BDD nodes, or optimize binary operator application using hash tables,
so this test is not currently realistic. It would certainly be interesting to
learn which optimizations are most effective!

I considered using hash tables to keep track of duplicate nodes globally
(allowing pointer comparison to compare BDDs), but this presents a challence for
garbage collection. I tried to build something that keeps track of a stack of
BDD references, and can sweep the node memory, but this makes the code very
ugly. Then I thought I found exactly what I need: weak references (or a weak
hash map). Unfortunately Go does not have these :(.

I then tried using `sync.Map` without any GC to lookup existing BDD references,
and to globally cache `BDD.Apply`. This made the larger instance slightly faster
but not by a lot. I suspect the non-carry addition generates a very inefficient
BDD. A next step would be to look into inserting new variables in arbitrary
places in the ordering such that bits that are closely related in the binary
arithmetic formula are close to each other in the BDD (this heuristic is also
mentioned in the notes of a course I took). Another approach would be to port
the implementation to Rust and use reference counting.