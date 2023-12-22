# go-connect

<a name="v1.17.0"></a>
## [v1.17.0] - 2023-12-22
### Fixes
- use applied http conn options which is not nullable ([#33](https://github.com/kumparan/go-connect/issues/33))

### Other Improvements
- upgrade dependencies


<a name="v1.16.1"></a>
## [v1.16.1] - 2023-12-05
### Fixes
- use applied http conn options which is not nullable


<a name="v1.16.0"></a>
## [v1.16.0] - 2023-11-16
### New Features
- make keep-alive option configurable on http conn


<a name="v1.15.1"></a>
## [v1.15.1] - 2023-07-06
### Other Improvements
- up grpc & echo lib version


<a name="v1.15.0"></a>
## [v1.15.0] - 2023-04-05
### New Features
- add RegisterHealthCheckService ([#29](https://github.com/kumparan/go-connect/issues/29))


<a name="v1.14.0"></a>
## [v1.14.0] - 2023-04-03
### New Features
- trace before return
- add param excluded ips and user agents
- grpc rate limiter

### Other Improvements
- add comment


<a name="v1.13.0"></a>
## [v1.13.0] - 2023-03-27
### New Features
- recover from panic ([#27](https://github.com/kumparan/go-connect/issues/27))


<a name="v1.12.0"></a>
## [v1.12.0] - 2023-03-13
### New Features
- add asynq task tracer middleware ([#26](https://github.com/kumparan/go-connect/issues/26))


<a name="v1.11.0"></a>
## [v1.11.0] - 2023-03-10
### New Features
- enable exclusion of user agents on rate limiter ([#25](https://github.com/kumparan/go-connect/issues/25))


<a name="v1.10.1"></a>
## [v1.10.1] - 2023-03-07
### Fixes
- remove rate limit information ([#24](https://github.com/kumparan/go-connect/issues/24))


<a name="v1.10.0"></a>
## [v1.10.0] - 2023-03-06
### New Features
- echo redis rate limiter support redis v7


<a name="v1.9.1"></a>
## [v1.9.1] - 2023-02-22
### Fixes
- bump otel version ([#22](https://github.com/kumparan/go-connect/issues/22))


<a name="v1.9.0"></a>
## [v1.9.0] - 2023-02-22
### Fixes
- move tracer into retryableInvoke
- prevent panic on span set attributes

### New Features
- upgrade library to support goredis v9 for redis7 ([#21](https://github.com/kumparan/go-connect/issues/21))


<a name="v1.8.0"></a>
## [v1.8.0] - 2023-02-01
### Code Improvements
- move log level into opt
- naming
- option otel on cockroach
- lowerize function and const name
- lowerize function and const name

### New Features
- create InitializeCockroachConn
- handle opt use otel on NewElasticsearchClient
- create NewElasticsearchClient
- remove insecure option
- add InitTraceProvider
- fix panic
- fix panic
- rename elastic option into http option
- add elastic connect option
- add trace on go redis
- add UnaryServerInterceptor
- add UseOpenTelemetry field


<a name="v1.7.0"></a>
## [v1.7.0] - 2023-01-06
### Code Improvements
- remove grpc connection pool ([#19](https://github.com/kumparan/go-connect/issues/19))


<a name="v1.6.2"></a>
## [v1.6.2] - 2022-11-22
### Fixes
- exclude ip configuration on rate limiter


<a name="v1.6.1"></a>
## [v1.6.1] - 2022-11-21
### Fixes
- ignore private ip on rate limiter


<a name="v1.6.0"></a>
## [v1.6.0] - 2022-11-21
### New Features
- add redis ip rate limiter middleware


<a name="v1.5.2"></a>
## [v1.5.2] - 2022-07-01

<a name="v1.5.1"></a>
## [v1.5.1] - 2022-06-29
### Other Improvements
- fix dependabot issue ([#13](https://github.com/kumparan/go-connect/issues/13))
- fix dependabot issue & upgrade to go 1.18 ([#12](https://github.com/kumparan/go-connect/issues/12))


<a name="v1.5.0"></a>
## [v1.5.0] - 2022-03-23
### New Features
- enable wait for connection when max active connection is rea… ([#11](https://github.com/kumparan/go-connect/issues/11))


<a name="v1.4.3"></a>
## [v1.4.3] - 2020-09-08
### Fixes
- return original error ([#10](https://github.com/kumparan/go-connect/issues/10))


<a name="v1.4.2"></a>
## [v1.4.2] - 2020-09-07
### Fixes
- missing nil
- fix double invoke ([#9](https://github.com/kumparan/go-connect/issues/9))


<a name="v1.4.1"></a>
## [v1.4.1] - 2020-04-20
### New Features
- change max idle and max active default value ([#8](https://github.com/kumparan/go-connect/issues/8))
- add ReadOnly option in go-redis cluster ([#7](https://github.com/kumparan/go-connect/issues/7))


<a name="v1.4.0"></a>
## [v1.4.0] - 2020-04-06

<a name="v1.3.0"></a>
## [v1.3.0] - 2020-04-06
### New Features
- Add gRPC Pool connection constructor ([#6](https://github.com/kumparan/go-connect/issues/6))


<a name="v1.2.0"></a>
## [v1.2.0] - 2020-04-03
### New Features
- add circuit breaker and retry wrapper for grpc UnaryClientInterceptor ([#5](https://github.com/kumparan/go-connect/issues/5))


<a name="v1.1.0"></a>
## [v1.1.0] - 2020-03-18
### New Features
- add config timeout (read, write, dial) in goredis ([#4](https://github.com/kumparan/go-connect/issues/4))


<a name="v1.0.2"></a>
## [v1.0.2] - 2020-03-12
### Fixes
- redis-cluster url validation should be reverse of valid standalone ([#3](https://github.com/kumparan/go-connect/issues/3))


<a name="v1.0.1"></a>
## [v1.0.1] - 2020-03-12
### Fixes
- fix goredis connect on non clustered redis ([#2](https://github.com/kumparan/go-connect/issues/2))


<a name="v1.0.0"></a>
## v1.0.0 - 2020-03-11
### New Features
- init go-connect with http and redis connector ([#1](https://github.com/kumparan/go-connect/issues/1))


[Unreleased]: https://github.com/kumparan/go-connect/compare/v1.17.0...HEAD
[v1.17.0]: https://github.com/kumparan/go-connect/compare/v1.16.1...v1.17.0
[v1.16.1]: https://github.com/kumparan/go-connect/compare/v1.16.0...v1.16.1
[v1.16.0]: https://github.com/kumparan/go-connect/compare/v1.15.1...v1.16.0
[v1.15.1]: https://github.com/kumparan/go-connect/compare/v1.15.0...v1.15.1
[v1.15.0]: https://github.com/kumparan/go-connect/compare/v1.14.0...v1.15.0
[v1.14.0]: https://github.com/kumparan/go-connect/compare/v1.13.0...v1.14.0
[v1.13.0]: https://github.com/kumparan/go-connect/compare/v1.12.0...v1.13.0
[v1.12.0]: https://github.com/kumparan/go-connect/compare/v1.11.0...v1.12.0
[v1.11.0]: https://github.com/kumparan/go-connect/compare/v1.10.1...v1.11.0
[v1.10.1]: https://github.com/kumparan/go-connect/compare/v1.10.0...v1.10.1
[v1.10.0]: https://github.com/kumparan/go-connect/compare/v1.9.1...v1.10.0
[v1.9.1]: https://github.com/kumparan/go-connect/compare/v1.9.0...v1.9.1
[v1.9.0]: https://github.com/kumparan/go-connect/compare/v1.8.0...v1.9.0
[v1.8.0]: https://github.com/kumparan/go-connect/compare/v1.7.0...v1.8.0
[v1.7.0]: https://github.com/kumparan/go-connect/compare/v1.6.2...v1.7.0
[v1.6.2]: https://github.com/kumparan/go-connect/compare/v1.6.1...v1.6.2
[v1.6.1]: https://github.com/kumparan/go-connect/compare/v1.6.0...v1.6.1
[v1.6.0]: https://github.com/kumparan/go-connect/compare/v1.5.2...v1.6.0
[v1.5.2]: https://github.com/kumparan/go-connect/compare/v1.5.1...v1.5.2
[v1.5.1]: https://github.com/kumparan/go-connect/compare/v1.5.0...v1.5.1
[v1.5.0]: https://github.com/kumparan/go-connect/compare/v1.4.3...v1.5.0
[v1.4.3]: https://github.com/kumparan/go-connect/compare/v1.4.2...v1.4.3
[v1.4.2]: https://github.com/kumparan/go-connect/compare/v1.4.1...v1.4.2
[v1.4.1]: https://github.com/kumparan/go-connect/compare/v1.4.0...v1.4.1
[v1.4.0]: https://github.com/kumparan/go-connect/compare/v1.3.0...v1.4.0
[v1.3.0]: https://github.com/kumparan/go-connect/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/kumparan/go-connect/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/kumparan/go-connect/compare/v1.0.2...v1.1.0
[v1.0.2]: https://github.com/kumparan/go-connect/compare/v1.0.1...v1.0.2
[v1.0.1]: https://github.com/kumparan/go-connect/compare/v1.0.0...v1.0.1
