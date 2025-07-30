# go-connect

<a name="v1.23.0"></a>
## v1.23.0 - 2025-07-30
### Code Improvements
- move log level into opt
- naming
- option otel on cockroach
- lowerize function and const name
- lowerize function and const name
- remove grpc connection pool ([#19](https://github.com/kumparan/go-connect/issues/19))

### Fixes
- upgrade go/x/net dependencies vulnerability ([#41](https://github.com/kumparan/go-connect/issues/41))
- upgrade go version dan dependencies to fix vulnerability issue ([#40](https://github.com/kumparan/go-connect/issues/40))
- use applied http conn options which is not nullable ([#33](https://github.com/kumparan/go-connect/issues/33))
- remove rate limit information ([#24](https://github.com/kumparan/go-connect/issues/24))
- bump otel version ([#22](https://github.com/kumparan/go-connect/issues/22))
- move tracer into retryableInvoke
- prevent panic on span set attributes
- exclude ip configuration on rate limiter
- ignore private ip on rate limiter
- return original error ([#10](https://github.com/kumparan/go-connect/issues/10))
- missing nil
- fix double invoke ([#9](https://github.com/kumparan/go-connect/issues/9))
- redis-cluster url validation should be reverse of valid standalone ([#3](https://github.com/kumparan/go-connect/issues/3))

### Fixes
- create proper ping ([#36](https://github.com/kumparan/go-connect/issues/36))
- fix goredis connect on non clustered redis ([#2](https://github.com/kumparan/go-connect/issues/2))

### New Features
- update to latest dependencies version to resolve security issues
- customizeable span name in HTTP Request in trace ([#42](https://github.com/kumparan/go-connect/issues/42))
- add deployment.environment resource in tracer ([#39](https://github.com/kumparan/go-connect/issues/39))
- add mysql connector
- make keep-alive option configurable on http conn
- add RegisterHealthCheckService ([#29](https://github.com/kumparan/go-connect/issues/29))
- trace before return
- add param excluded ips and user agents
- grpc rate limiter
- recover from panic ([#27](https://github.com/kumparan/go-connect/issues/27))
- add asynq task tracer middleware ([#26](https://github.com/kumparan/go-connect/issues/26))
- enable exclusion of user agents on rate limiter ([#25](https://github.com/kumparan/go-connect/issues/25))
- echo redis rate limiter support redis v7
- upgrade library to support goredis v9 for redis7 ([#21](https://github.com/kumparan/go-connect/issues/21))
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
- add redis ip rate limiter middleware
- enable wait for connection when max active connection is rea… ([#11](https://github.com/kumparan/go-connect/issues/11))
- change max idle and max active default value ([#8](https://github.com/kumparan/go-connect/issues/8))
- add ReadOnly option in go-redis cluster ([#7](https://github.com/kumparan/go-connect/issues/7))
- Add gRPC Pool connection constructor ([#6](https://github.com/kumparan/go-connect/issues/6))
- add circuit breaker and retry wrapper for grpc UnaryClientInterceptor ([#5](https://github.com/kumparan/go-connect/issues/5))
- add config timeout (read, write, dial) in goredis ([#4](https://github.com/kumparan/go-connect/issues/4))
- init go-connect with http and redis connector ([#1](https://github.com/kumparan/go-connect/issues/1))

### Other Improvements
- add exclude dirs
- add github workflows
- upgrade dependencies
- use default health server
- fix register health check service
- circuit breaker can only open by certain error codes
- fix vulnerabilities library and fix lint ([#37](https://github.com/kumparan/go-connect/issues/37))
- upgrade go to 1.22
- upgrade dependencies ([#34](https://github.com/kumparan/go-connect/issues/34))
- up grpc & echo lib version
- add comment
- fix dependabot issue ([#13](https://github.com/kumparan/go-connect/issues/13))
- fix dependabot issue & upgrade to go 1.18 ([#12](https://github.com/kumparan/go-connect/issues/12))


[Unreleased]: https://github.com/kumparan/go-connect/compare/v1.23.0...HEAD
