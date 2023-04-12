# Worker
The Worker is the interface to serve the incoming words and execute some purpose task.
It can run task by the passed word and show the result.

For example `debug` shows the word on terminal as the progress status.


| name  | description                            | optional argument  |
|-------|----------------------------------------|--------------------|
| dns   | list possible sub-domain DNS records   | target hostname    |
| s3    | list possible S3 bucket name           |                    |
| ssh   | list possible password of SSH protocol | target, username   |
| subp  | list possible URL path                 | target             |
| debug | just show the word on terminal         |                    |


### DNS
The **DNS** worker is used to check the target's DNS record, and the word is considered
as the sub-domain. This worker check the [CNAME][0], [MX][1], [NS][2] and [TXT][3] record.

### S3
The **S3** worker is used to check the public [s3 bucket][4], and the word is considered
as the bucket name. At this worker it will convert the word to lower case, validate
the word and only serve the word if it is the valid s3 bucket name.

### SSH
The **SSH** worker is used to check the password via SSH protocol. The word is considered
as the password and caller need to pass the target and the username.

### Sub-path
The **subp** worker is used to check the possible path for target URL. The word is considered
as the valid path and check possible HTTP method, include GET, POST, PUT and DELETE.

[0]: https://en.wikipedia.org/wiki/CNAME_record
[1]: https://en.wikipedia.org/wiki/MX_record
[2]: https://en.wikipedia.org/wiki/List_of_DNS_record_types#NS
[3]: https://en.wikipedia.org/wiki/TXT_record
[4]: https://aws.amazon.com/s3
