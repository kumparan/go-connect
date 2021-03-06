# go-connect

<a name="v1.4.3"></a>
## [v1.4.3] - 2020-09-08
### Fixes
- should return original error without wrapping it


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


[Unreleased]: https://github.com/kumparan/go-connect/compare/v1.4.3...HEAD
[v1.4.3]: https://github.com/kumparan/go-connect/compare/v1.4.2...v1.4.3
[v1.4.2]: https://github.com/kumparan/go-connect/compare/v1.4.1...v1.4.2
[v1.4.1]: https://github.com/kumparan/go-connect/compare/v1.4.0...v1.4.1
[v1.4.0]: https://github.com/kumparan/go-connect/compare/v1.3.0...v1.4.0
[v1.3.0]: https://github.com/kumparan/go-connect/compare/v1.2.0...v1.3.0
[v1.2.0]: https://github.com/kumparan/go-connect/compare/v1.1.0...v1.2.0
[v1.1.0]: https://github.com/kumparan/go-connect/compare/v1.0.2...v1.1.0
[v1.0.2]: https://github.com/kumparan/go-connect/compare/v1.0.1...v1.0.2
[v1.0.1]: https://github.com/kumparan/go-connect/compare/v1.0.0...v1.0.1
