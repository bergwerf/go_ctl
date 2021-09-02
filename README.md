A simple CTL model checker in Go
================================
This is a minimal implementation of a CTL (Computation Tree Logic) model 
checker in Go using ROBDDs. The variable ordering is the same as the order
in which variables are defined. There is no intermediate expression format;
the interface to define transitions directly constructs an ROBDD.

The file `3_test.go` contains a more complex example of model checking to find 
deadlocks in packet switching networks. My implementation is not efficient 
enough to solve this problem. I suspect this is because there is no garbage 
collection over the BDD lookup table. This would require either weak references 
or reference counting. The former is not available in Go, but the latter could 
in theory be implemented.
