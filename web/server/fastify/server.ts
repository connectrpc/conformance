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

import { readFileSync } from "fs";
import { fastify, FastifyHttpsOptions } from "fastify";
import { fastifyConnectPlugin } from "@bufbuild/connect-fastify";
import fastifyCors from "@fastify/cors";
import routes from "../routes.js";
import { interop } from "../interop.js";
import https from "https";
import path from "path";

const HOST = "0.0.0.0";

export interface Options {
  h1port: number;
  h2port: number;
  cert?: string;
  key?: string;
  insecure?: boolean;
}

function getTLSConfig(key: string, cert: string) {
  return {
    key: readFileSync(path.join(__dirname, "..", "..", "..", key), "utf-8"),
    cert: readFileSync(path.join(__dirname, "..", "..", "..", cert), "utf-8"),
  };
}

function createH1Server(opts: Options) {
  const serverOpts: FastifyHttpsOptions<https.Server> = { https: null };
  if (!opts.insecure && opts.key && opts.cert) {
    serverOpts.https = getTLSConfig(opts.key, opts.cert);
  }

  return fastify(serverOpts);
}

function createH2Server(opts: Options) {
  if (!opts.insecure && opts.key && opts.cert) {
    return fastify({
      http2: true,
      https: getTLSConfig(opts.key, opts.cert),
    });
  } else {
    return fastify({
      http2: true,
    });
  }
}

export async function start(opts: Options) {
  const h1Server = createH1Server(opts);
  await h1Server.register(fastifyCors, interop.corsOptions);
  await h1Server.register(fastifyConnectPlugin, { routes });
  await h1Server.listen({ host: HOST, port: opts.h1port });
  console.log(
    `Running ${opts.insecure ? "insecure" : "secure"} HTTP/1.1 server on `,
    h1Server.addresses()
  );

  const h2Server = createH2Server(opts);
  await h2Server.register(fastifyCors, interop.corsOptions);
  await h2Server.register(fastifyConnectPlugin, { routes });
  await h2Server.listen({ host: HOST, port: opts.h2port });
  console.log(
    `Running ${opts.insecure ? "insecure" : "secure"} HTTP/2 server on `,
    h2Server.addresses()
  );
  return new Promise<void>((resolve) => {
    resolve();
  });
}
