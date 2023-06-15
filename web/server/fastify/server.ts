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
import { fastify } from "fastify";
import { fastifyConnectPlugin } from "@bufbuild/connect-fastify";
import { cors as connectCors } from "@bufbuild/connect";
import fastifyCors from "@fastify/cors";
import routes from "../routes";
import path from "path";
import url from "url";

const protocol = process.argv[2] ?? "h1";

const opts: any = {};

if (protocol === "h2" || protocol === "h2c") {
  opts.http2 = true;
  if (protocol === "h2") {
    const __filename = url.fileURLToPath(import.meta.url);
    const __dirname = path.dirname(__filename);
    opts.https = {
      key: readFileSync(
        path.join(__dirname, "..", "..", "..", "cert", "localhost.key")
      ),
      cert: readFileSync(
        path.join(__dirname, "..", "..", "..", "cert", "localhost.crt")
      ),
    };
  }
}

const server = fastify(opts);

// Options for configuring CORS. The @bufbuild/connect package exports
// convenience variables for configuring a CORS setup.
await server.register(fastifyCors, {
  // Reflects the request origin. This should only be used for development.
  // Production should explicitly specify an origin
  origin: true,
  methods: [...connectCors.allowedMethods],
  allowedHeaders: [...connectCors.allowedHeaders],
  exposedHeaders: [...connectCors.exposedHeaders],
});

await server.register(fastifyConnectPlugin, { routes });

await server.listen({ host: "localhost", port: 3000 });
