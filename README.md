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