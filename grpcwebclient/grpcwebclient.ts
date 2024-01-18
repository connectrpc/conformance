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
import puppeteer from "puppeteer";
import * as esbuild from "esbuild";
import {
  ClientCompatRequest,
  ClientCompatResponse,
  ClientResponseResult,
  ClientErrorResult,
} from "./gen/proto/connectrpc/conformance/v1/client_compat_pb.js";

export async function run() {
  // Launch a browser. For a non-headless browser, pass false
  const browser = await puppeteer.launch({ headless: "new" });
  const page = await browser.newPage();

  await page.addScriptTag({
    type: "application/javascript",
    content: await buildBrowserScript(),
  });

  for await (const next of readReqBuffers(process.stdin)) {
    const req = ClientCompatRequest.deserializeBinary(next);
    const res = new ClientCompatResponse();
    res.setTestName(req.getTestName());

    try {
      // This will call the runTestCase function on the global scope
      // inside the browser. The arg to the function given to evaluate
      // is the test case we want to run
      const result = await page.evaluate(function (data) {
        // @ts-ignore
        return window.runTestCase(data);
      }, Array.from(next));

      const resp = ClientResponseResult.deserializeBinary(
        new Uint8Array(result),
      );
      res.setResponse(resp);
    } catch (e) {
      const err = new ClientErrorResult();
      err.setMessage((e as Error).message);
      res.setError(err);
    }

    const resData = res.serializeBinary();
    const resSize = Buffer.alloc(4);
    resSize.writeUInt32BE(resData.length);
    process.stdout.write(resSize);
    process.stdout.write(Buffer.from(resData));
  }

  await page.close();
  await browser.close();
}

async function buildBrowserScript() {
  const buildResult = await esbuild.build({
    entryPoints: ["grpcwebclient/browserscript.ts"],
    bundle: true,
    write: false,
  });
  if (buildResult.outputFiles.length !== 1) {
    throw new Error("Expected exactly one output file");
  }
  return buildResult.outputFiles[0].text;
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
