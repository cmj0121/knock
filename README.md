# knock #
[![GitHub Action][0]][1]

The *knock* is the Go-based brute-force tool.

The knock provides the binary tool when install the knock in your system.
It provides several sub-command for the specified target service.


## Install
You can install the binary tool by `go install` command, as easy as

```bash
go install github.com/cmj0121/knock
```
, or you can build and install from source code with the necessary build environment

```bash
> git clone git@github.com:cmj0121/knock.git
> cd knock
> make build && make install
```

## Example
You can execute **knock** and list all available sub-modules. Each module provides
a method for you to **brute-force** some information by your word generator.

### Support Worker
The Worker is the interface to serve the incoming words and execute some purpose task.
It can run task by the passed word and show the result.

| name  | description                            | optional argument  |
|-------|----------------------------------------|--------------------|
| dns   | list possible sub-domain DNS records   | target hostname    |
| s3    | list possible S3 bucket name           |                    |
| ssh   | list possible password of SSH protocol | target, username   |
| subp  | list possible URL path                 | target             |
| debug | just show the word on terminal         |                    |

### Support Word Generator

The producer is used to generator word by specified purpose. It can generate
the words by the optional arguments.

For example, `--ip STRING` is used to generate the IP address with **IP/mask** format.

| optional                   | description                             |
|----------------------------|-----------------------------------------|
| -f FILE, --file=FILE       | read the external wordlist file         |
| -i STRING, --ip=STRING     | generate valid IP via IP/mask           |
| -r STRING, --regexp STRING | generate the word by regular expression |


[0]: https://github.com/cmj0121/knock/actions/workflows/pipeline.yml/badge.svg
[1]: https://github.com/cmj0121/knock/actions
