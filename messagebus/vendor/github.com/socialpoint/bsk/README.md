# Backend Services Kit

[![Build Status](https://travis-ci.com/socialpoint/bsk.svg?token=YQABbkJWq7buTaP4qGqT&branch=master)](https://travis-ci.com/socialpoint/bsk) [![Coverage](https://coveralls.io/repos/github/socialpoint/bsk/badge.svg?branch=master&t=gTGL87)](https://coveralls.io/github/socialpoint/bsk)

BSK is a collection of common packages to build the foundations of backend-related services in Golang.

For a list of issues, milestones, etc..., see our [issue list](https://github.com/socialpoint/bsk/issues)

## Packages

### Extensions

Extensions are packages tha extends other packages from the standard library or from an official package.

| Package | Description|
| --- | --- |
| [`awsx`](pkg/awsx)   | Extends package `github.com/aws/aws-sdk-go` with testing utilities and easier to use helpers. |
| [`httpx`](pkg/httpx) | Extends package `net/http` with routing utilities, decorators, etc... |
| [`logx`](pkg/logx)   | Extends package `log` with minimal log package with structured logging and logstash support. |
| [`timex`](pkg/timex) | Extends `time` with other time functions and utilities |
| [`httpc`](pkg/httpc) | HTTP client extensions |

### Utilities

Utilities are general purpose packages that provides specific functionalities.

| Package | Description |
| --- | --- |
| [`jwt`](pkg/jwt)                   | JWT (JSON Web Tokens) |
| [`singleflight`](pkg/singleflight) | Single flight, duplicate function `call` suppression |
| [`throttler`](pkg/throttler) | An utility to limit the concurrent execution of actions |
| [`uuid`](pkg/uuid) | An utility package to generate time-ordered UUIDs |
| [`netutil`](pkg/netutil) | Net utilities to get free network ports |
| [`sk`](pkg/sk) | Services kit, utilities to work with services definitions |

### Social Point Related

Social Point related packages and game utilities.

| Package | Description |
| --- | --- |
| [`game`](pkg/game) | Game utilities related to the games in SP |

## Infrastructure

Infrastructure management tools and other DevOps tools has been moved to the [DevOps repository](https://github.com/socialpoint/devops).
