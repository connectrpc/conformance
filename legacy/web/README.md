web conformance
===============

The `web` test suite uses Jasmine and Karma to test `connect-web` and `grpc-web` against Connect
and gRPC servers. The test cases are at parity with [gRPC-web interop test cases][grpc-web-interop].

## Developers

To lint and format:

```
$ npm run lint
$ npm run format
```

## Requirements and Running the Tests

To run the web tests, use:

```
$ npm run test --host="<server-host-name>" --port="<server-port-name>" --implementation="<grpc-web, connect-web>"
```

[grpc-web-interop]: https://github.com/grpc/grpc-web/blob/master/doc/interop-test-descriptions.md
