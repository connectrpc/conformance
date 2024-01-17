// Copyright 2023-2024 The Connect Authors
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

import type { Readable } from "node:stream";
import {
  ClientCompatRequest,
  ClientCompatResponse,
  ClientResponseResult,
} from "./gen/proto/connectrpc/conformance/v1/client_compat_pb.js";
import { fork, spawn } from "node:child_process";
import path from "path";
import { remote } from "webdriverio";

export async function run() {
  const browser = await remote({
    logLevel: "silent",
    // outputDir: "./timo",
    capabilities: {
      browserName: "chrome",
      "goog:chromeOptions": {
        args: ["headless", "disable-gpu"],
      },
    },
  });
  // process.stderr.write("executing script\n");
  // const done = await browser.executeAsyncScript(
  //   "arguments[0]('script executed');",
  //   [],
  // );
  // process.stderr.write(done + "\n");

  // console.log("test");

  process.stderr.write("reading from stdin\n");
  for await (const next of readReqBuffers(process.stdin)) {
    const req = ClientCompatRequest.deserializeBinary(next);
    const res = new ClientCompatResponse();
    res.setTestName(req.getTestName());

    // Send to webdriver here

    const result = new ClientResponseResult();
    res.setResponse(result);

    const resData = res.serializeBinary();
    const resSize = Buffer.alloc(4);
    resSize.writeUInt32BE(resData.length);
    process.stdout.write(resSize);
    process.stdout.write(Buffer.from(resData));
  }
}

async function* readReqBuffers(stream: Readable) {
  stream.once("error", (err) => {
    throw err;
  });
  for (; !stream.readableEnded; ) {
    // Read 4 bytes from the stream into a buffer, which contains the size of the message to read
    const sizeBuffer = stream.read(4) as Buffer | null;
    if (sizeBuffer === null) {
      await new Promise<void>((resolve) => {
        stream.once("readable", () => {
          resolve();
        });
        stream.once("end", () => {
          resolve();
        });
      });
      continue;
    }
    let bytes: Buffer | null = null;
    // We are guaranteed to get the next chunk.
    for (;;) {
      let bytesRead = 0;
      let chunks: Buffer[] = [];

      // Determine how big the message is we need to read (i.e. how many bytes to read from the buffer)
      const sizeOfMessage = sizeBuffer.readUint32BE();
      // Continue reading until we've read the whole message
      // This is needed for large messages greater than the highWaterMark of the buffer
      while (bytesRead < sizeOfMessage) {
        await new Promise((resolve) => {
          stream.once("readable", resolve);
        });

        // If the number of bytes left to read for the message is greater than the
        // buffer capacity, then just read the entire buffer (i.e. the readable length).
        // Otherwise, read the remaining bytes we need for the message.
        const toRead = Math.min(
          stream.readableLength,
          sizeOfMessage - bytesRead,
        );
        let chunk = stream.read(toRead) as Buffer | null;
        if (chunk !== null) {
          bytesRead += chunk.length;
          chunks.push(chunk);
        } else {
          break;
        }
      }

      // Assemble the chunks we've read and return them
      if (chunks.length > 0) {
        bytes = Buffer.concat(chunks);
        break;
      }
      await new Promise((resolve) => {
        stream.once("readable", resolve);
      });
    }
    yield bytes;
  }
}
