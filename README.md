# knock #
![Go test](https://github.com/cmj0121/knock/workflows/test/badge.svg) 

The **knock** is the Go-based brute-force tool.

It enumerates the information for the specified protocol with default word-lists.

```sh
usage: knock [OPTION]

option:
         -h, --help                  show this message
         -v, --version               show argparse version
             --log STR               log level [debug info verbose warn]
     -w INT, --worker INT            number of worker (default: 8)
     -W STR, --word-list STR         default word lists [passwords usernames wordlists] (default: wordlists)
     -t INT, --timeuot INT           global timeout based on seconds (default: 60)
             --wait INT              wait ms per each task

sub-command:
    demo                             list all the word-list
    info                             show the current system info
    scan                             scan via network protocol
    web                              web-related scanner
    dns                              scan DNS record
```
