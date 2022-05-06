web crosstests
===============

The `web` test suite uses Jasmine and Karma to test `connect-web` and `grpc-web` against gRPC
and Connect servers. The test cases are at parity with [gRPC-web interop test cases][grpc-web-interop].

## Developers

In order to have access to the private `@bufbuild/connect-web` package,
ensure that you have a `.npmrc` in your `$HOME` that sets an access token to the private
package in the `@bufbuild` scope.

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
