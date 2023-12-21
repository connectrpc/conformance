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

import { remote } from "webdriverio";
import * as esbuild from "esbuild";
import * as net from "node:net";
import { join as joinPath } from "node:path";
import { mkdtempSync } from "node:fs";
import { tmpdir } from "node:os";
import { execFile } from "node:child_process";

export async function main() {
  // Start Webdriver
  const browser = await remote({
    capabilities: {
      browserName: "chrome",
      "goog:chromeOptions": {
        args: ["headless", "disable-gpu"],
      },
    },
  });

  const buildResult = await esbuild.build({
    entryPoints: ["./grpcwebclient/client-entry.ts"],
    bundle: true,
    write: false,
  });

  if (buildResult.outputFiles.length !== 1) {
    throw new Error("Expected exactly one output file");
  }
  const invokeScript = buildResult.outputFiles[0].text;

  const server = net.createServer((socket) => void run(socket, invokeScript));
  let runnerCloseResolve: () => void, runnerCloseReject: (err: unknown) => void;
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
    const runner = execFile("grpcwebclient/pipe.ts", [socketName], {});
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
}

async function run(socket: net.Socket, script) {
  setTimeout(() => {
    socket.write("socket to me");
    // process.stderr.write("reading req buffers");
  }, 2000);
  for await (const next of readReqBuffers(socket)) {
    // socket.write("socket to me");
    process.stderr.write("reading req buffers");
  }
}

async function* readReqBuffers(stream: Readable) {
  stream.once("error", (err) => {
    throw err;
  });
  for (; !stream.readableEnded; ) {
    const size = stream.read(4) as Buffer | null;
    if (size === null) {
      await new Promise<void>((resolve) => {
        stream.once("readable", resolve);
        stream.once("end", resolve);
      });
      continue;
    }
    let chunk: Buffer | null = null;
    // We are guaranteed to get the next chunk.
    for (;;) {
      chunk = stream.read(size.readUInt32BE()) as Buffer | null;
      if (chunk !== null) {
        break;
      }
      await new Promise((resolve) => stream.once("readable", resolve));
    }
    yield chunk;
  }
}

await main();
