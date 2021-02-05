# knock #
The **knock** is the Go-based brute-force tool.

	+-------+  load config   +-----------+
	| knock | -------------> | word list | --------+
	+-------+                +-----------+         |
	  |   |      create runner                     | broker
	  |   +------------------------------------+   |
	  |                                        |   |
	  +----------------+                       v   v
	    create reducer |                     +--------+
	                   v                    /--------/|
	             +---------+    receiver   +--------+ | # NumWorker
	STDOUT <---  | reducer | < ==========  | Runner |/  (exit when channel closed)
	             +---------+               +--------+

