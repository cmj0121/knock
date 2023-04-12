# Word Generator
The Producer is the interface to generate the word list by specified purpose.
It can used to generate several words by zero or more arguments.

For example, `--ip STRING` is used to generate the IP address with **IP/mask** format.


## Producer
The **knock** now support several producers to generate the specified word list.
Each producer may need zero or more arguments to generate words.


| optional                   | description                             |
|----------------------------|-----------------------------------------|
| -f FILE, --file=FILE       | read the external wordlist file         |
| -i STRING, --ip=STRING     | generate valid IP via IP/mask           |
| -r STRING, --regexp STRING | generate the word by regular expression |


### File
The producer **File** is used to read the word from external file, read line-by-line and
generate words to the worker

### IP
The producer **IP** is used to generate the sequency IP address, by pass the `IP/mask` format
and generate all the possible IP addressed.

For example, `--ip 127.0.0.1/31` will generate two IP address *127.0.0.0* and *127.0.0.1*. Also,
this producer also support IPv6, like `--ip ::1/63` will generate all possible ip addressed.

### Regular Expression
The producer **Regexp** is used to generate the word pattern related on the regexp. It can
generate the word list by the [Perl syntax][0]. The word will generate by the random sequency.

For example `--regexp \d` will generate the digest from 0 to 9, with random order.

[0]: https://pkg.go.dev/regexp/syntax#Flags
