# package logx

`package logx` is a minimal log package inspired by [github.com/Sirupsen/logrus](logrus) and [github.com/uber-common/zap](zap) that follows the [https://socialpoint.atlassian.net/wiki/display/BAC/Logging+guidelines+for+Golang+applications](SP log guidelines).

Features:

-  pluggable io.Writer with the `WithWriter` decorator, default is os.Stdout
-  2 marshallers available, a logstash one and a human-readable one
-  support debug and info levels, and a `WithLevel` decorator to change the level
-  support structured fields, for now keys are strings and values can be string (`logx.S`) and interger (`logx.I`)
-  provides a Dummy logger for testing purposes

Things not implemented yet:

- colors
