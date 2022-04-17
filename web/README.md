connect-web test suite
======================

The `connect-web` test suite is a small React web app that runs
the same test cases as [gRPC-web interop test cases][grpc-web-interop].

## Developers

In order to have access to private `@bufbuild/connect-web` and `@bufbuild/protobuf` packages,
ensure that you have a `.npmrc` in your `$HOME` that sets an access token to the private
package in the `@bufbuild` scope.

To build a new version:

```
$ npm run build
```

To preview the current build:

```
$ npm run preview
```

To lint and format:

```
$ npm run lint
$ npm run format
```

[grpc-web-interop]: https://github.com/grpc/grpc-web/blob/master/doc/interop-test-descriptions.md
