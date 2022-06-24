// Copyright 2022 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

module.exports = function (config) {
  // determine test files by implementation flag, run all tests if undefined
  let testFiles = ["gen/**/*.ts"];
  switch (config.implementation) {
    case "connect-web":
    case "connect-grpc-web":
      testFiles.push("spec/connect-web.*.spec.ts");
      break;
    case "grpc-web":
      testFiles.push("spec/grpc-web.spec.ts");
      break;
    case undefined:
      testFiles.push("spec/**/*.ts");
      break;
    default:
      throw "unknown implementation flag for web test";
  }
  config.set({
    customLaunchers: {
      ChromeCustom: {
        base: "ChromeHeadless",
        // We ignore the certificate errors as the client certificates are managed by the browser.
        // We must disable the Chrome sandbox when running Chrome inside Docker (Chrome's sandbox needs
        // more permissions than Docker allows by default)
        flags: config.docker
          ? ["--no-sandbox", "--ignore-certificate-errors"]
          : ["--ignore-certificate-errors"],
      },
    },
    frameworks: ["jasmine"],
    files: testFiles,
    preprocessors: {
      "/**/*.ts": "esbuild",
    },
    reporters: ["progress"],
    browsers: ["ChromeCustom"],
    singleRun: true,
    esbuild: {
      target: "esnext",
      tsconfig: "./tsconfig.json",
    },
    client: {
      host: config.host,
      port: config.port,
      implementation: config.implementation,
    },
  });
};
