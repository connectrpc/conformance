// Copyright 2023 The Connect Authors
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

#!/usr/bin/env -S npx tsx

import {
  ClientCompatRequest,
  ClientCompatResponse,
  ClientResponseResult,
} from "./gen/proto/connectrpc/conformance/v1/client_compat_pb.js";
// import {
//   ClientCompatResponse as CCR,
//   ClientResponseResult,
// } from "./gen/proto/es/connectrpc/conformance/v1/client_compat_pb.js";
import { remote } from "webdriverio";
import type { ReadStream } from "node:tty";
import type { Readable } from "node:stream";
import invoke from "./invoke.js";
import * as esbuild from "esbuild";
import { readSync, writeSync } from "fs";
import { EventEmitter } from "node:events";

(async () => {
  async function main() {
    // const b = await remote({
    //   capabilities: {
    //     browserName: "chrome",
    //     "goog:chromeOptions": {
    //       args: ["headless", "disable-gpu"],
    //     },
    //   },
    // });

    let testCount = 0;
    try {
      while (testIo()) {
        testCount += 1;
      }
    } catch (e) {
      process.stderr.write(
        `conformance.ts: exiting after ${testCount} tests: ${String(e)}`,
      );
      process.exit(1);
    }
    const b = await remote({
      capabilities: {
        browserName: "chrome",
        "goog:chromeOptions": {
          args: ["headless", "disable-gpu"],
        },
      },
    });
    b.browserClose({});
    // console.error(b);
  }

  function testIo(): boolean {
    const request = readReqBuffer();
    if (!request) {
      return false;
    }

    const result = new ClientResponseResult();

    const response = new ClientCompatResponse();
    response.setTestName(request.getTestName());
    response.setResponse(result);

    // response.result = test(request);
    const serializedResponse = response.serializeBinary();
    const responseLengthBuf = Buffer.alloc(4);
    responseLengthBuf.writeUInt32BE(serializedResponse.length);
    writeBuffer(responseLengthBuf);
    writeBuffer(Buffer.from(serializedResponse));
    return true;
  }

  function readReqBuffer(): ClientCompatRequest | null {
    const requestLengthBuf = readBuffer(4);
    if (requestLengthBuf === "EOF") {
      return null;
    }
    const requestLength = requestLengthBuf.readUInt32BE(0);
    const serializedRequest = readBuffer(requestLength);
    if (serializedRequest === "EOF") {
      throw "Failed to read request.";
    }
    return ClientCompatRequest.deserializeBinary(serializedRequest);
  }

  function writeBuffer(buffer: Buffer): void {
    let totalWritten = 0;
    while (totalWritten < buffer.length) {
      totalWritten += writeSync(
        1,
        buffer,
        totalWritten,
        buffer.length - totalWritten,
      );
    }
  }

  function readBuffer(bytes: number): Buffer | "EOF" {
    const buf = Buffer.alloc(bytes);
    let read = 0;
    try {
      read = readSync(0, buf, 0, bytes, null);
    } catch (e) {
      throw `failed to read from stdin: ${String(e)}`;
    }
    if (read !== bytes) {
      if (read === 0) {
        return "EOF";
      }
      throw "premature EOF on stdin.";
    }
    return buf;
  }

  await main();
  console.error("returning");
})();
