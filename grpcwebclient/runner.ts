// TODO - End result of all this is that we should end up with this file and browserscript.ts as the only files
// This runs the runner, the other one runs in Puppeteer

// Import puppeteer
import type { Readable } from "node:stream";
import puppeteer from "puppeteer";
import * as esbuild from "esbuild";
import {
  ClientCompatRequest,
  ClientCompatResponse,
  ClientResponseResult,
  ClientErrorResult,
} from "./gen/proto/connectrpc/conformance/v1/client_compat_pb.js";

export function run() {
  void main().catch((reason) => {
    console.error("ITS FAILING");
    // TODO - Write client error result back and exit
    throw reason;
  });
}

async function main() {
  const browser = await puppeteer.launch({ headless: "new" });
  const page = await browser.newPage();
  page.on("pageerror", (err) => {
    // If an error is raised here, write ClientErrorResult back and exit
    process.stderr.write(
      `Uncaught exception in browser client: ${err.stack ?? err}\n`,
    );
    process.exit(1);
  });

  await page.addScriptTag({
    type: "application/javascript",
    content: await buildBrowserScript(),
  });

  for await (const next of readReqBuffers(process.stdin)) {
    const req = ClientCompatRequest.deserializeBinary(next);
    const res = new ClientCompatResponse();
    res.setTestName(req.getTestName());

    const result = await page.evaluate(function (data) {
      // @ts-ignore
      return window.runTestCase(data);
    }, Array.from(next));

    const resData = new Uint8Array(result);
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
    entryPoints: ["./browserscript.ts"],
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
