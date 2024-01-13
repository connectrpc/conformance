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

import { browser } from "@wdio/globals";
import * as esbuild from "esbuild";
import { execFile } from "node:child_process";
import type { Readable } from "node:stream";
import * as net from "node:net";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { join as joinPath } from "node:path";
import {
  ClientCompatRequest,
  ClientCompatResponse,
  ClientResponseResult,
  ClientErrorResult,
} from "./gen/proto/connectrpc/conformance/v1/client_compat_pb.js";

// We use a single unit test inside a Node process launched by Webdriver to
// execute the conformance tests. This file is executed by Webdriver and it then
// spawns a headless browser running the gRPC-web client.
//
// The communication between the two process is done via executeAsyncScript from
// Webdriver's browser API.

describe("Connect Conformance", () => {
  it("gRPC-web Client", async () => {
    const buildResult = await esbuild.build({
      entryPoints: ["./entry.ts"],
      bundle: true,
      write: false,
    });
    if (buildResult.outputFiles.length !== 1) {
      throw new Error("Expected exactly one output file");
    }
    const invokeScript = buildResult.outputFiles[0].text;
    const server = net.createServer((socket) => void run(socket, invokeScript));
    let runnerCloseResolve: () => void,
      runnerCloseReject: (err: unknown) => void;
    const runnerClose = new Promise<void>((resolve, reject) => {
      runnerCloseResolve = resolve;
      runnerCloseReject = reject;
    });
    server.on("error", (err) => {
      runnerCloseReject(err);
    });

    const socketName = joinPath(
      mkdtempSync(joinPath(tmpdir(), "connectconformance")),
      "socket",
    );
    server.listen(socketName, () => {
      const runner = execFile(
        "../.tmp/bin/connectconformance",
        [
          "--mode",
          "client",
          "--conf",
          "../testdata/grpc-web-client-impl-config.yaml",
          "--test-file",
          "../internal/app/connectconformance/testsuites/grpc-web-client.yaml",
          "-v",
          "./bin/pipe",
          socketName,
        ],
        {},
      );
      runner.stdout?.pipe(process.stdout);
      runner.stderr?.pipe(process.stderr);
      runner.on("error", (err) => {
        runnerCloseReject(err);
      });
      runner.on("close", (code) => {
        if (code !== 0) {
          runnerCloseReject(new Error(`Runner exited with code ${code}`));
        }
        runnerCloseResolve();
      });
    });
    await runnerClose.finally(() => server.close());

    expect(true).toBe(true);
  });
});

async function run(socket: net.Socket, invokeScript: string) {
  for await (const next of readReqBuffers(socket)) {
    const req = ClientCompatRequest.deserializeBinary(next);
    const res = new ClientCompatResponse();
    res.setTestName(req.getTestName());
    try {
      // Execute the JavaScript code that was built by esbuild inside the
      // browser launched by Webdriver. The arguments to executeAsyncScript
      // are the built JavaScript and an array of arguments to pass to the
      // script. Here, we pass one argument -- the ClientCompatRequest we
      // read from stdin
      //
      // The return value is a JSON object containing the results of the run
      const invokeResultJson = await browser.executeAsyncScript(invokeScript, [
        JSON.stringify(Array.from(next)),
      ]);

      // If type is 'data', then we have a successful run
      if (invokeResultJson.type === "data") {
        const result = ClientResponseResult.deserializeBinary(
          invokeResultJson.data,
        );
        res.setResponse(result);
      } else {
        // Otherwise, an error was returned
        const err = ClientErrorResult.deserializeBinary(invokeResultJson.error);
        res.setError(err);
      }
    } catch (e) {
      const err = new ClientErrorResult();
      err.setMessage((e as Error).message);
      res.setError(err);
    }
    const resData = res.serializeBinary();
    const resSize = Buffer.alloc(4);
    resSize.writeUInt32BE(resData.length);
    socket.write(resSize);
    socket.write(resData);
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
