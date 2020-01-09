Computation Tree Logic model checker based on ROBDDs
====================================================
This is a minimal implementation of ROBDDs to execute model checking
(Computation Tree Logic; CTL). In efficient implementations the variable
ordering is very space sensitive, but since this is just a demo I use the order
in which variables are declared. There is no intermediate expression format,
the interface to define transitions directly constructs an ROBDD.