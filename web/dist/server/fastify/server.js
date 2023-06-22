var __create = Object.create;
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __getProtoOf = Object.getPrototypeOf;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(isNodeMode || !mod || !mod.__esModule ? __defProp(target, "default", { value: mod, enumerable: true }) : target, mod));
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);
var server_exports = {};
__export(server_exports, {
  start: () => start
});
module.exports = __toCommonJS(server_exports);
var import_fs = require("fs");
var import_fastify = require("fastify");
var import_connect_fastify = require("@bufbuild/connect-fastify");
var import_cors = __toESM(require("@fastify/cors"));
var import_routes = __toESM(require("../routes.js"));
var import_interop = require("../interop.js");
var import_path = __toESM(require("path"));
const HOST = "0.0.0.0";
function getTLSConfig(key, cert) {
  return {
    key: (0, import_fs.readFileSync)(import_path.default.join(__dirname, "..", "..", "..", key), "utf-8"),
    cert: (0, import_fs.readFileSync)(import_path.default.join(__dirname, "..", "..", "..", cert), "utf-8")
  };
}
function createH1Server(opts) {
  const serverOpts = { https: null };
  if (!opts.insecure && opts.key && opts.cert) {
    serverOpts.https = getTLSConfig(opts.key, opts.cert);
  }
  return (0, import_fastify.fastify)(serverOpts);
}
function createH2Server(opts) {
  if (!opts.insecure && opts.key && opts.cert) {
    return (0, import_fastify.fastify)({
      http2: true,
      https: getTLSConfig(opts.key, opts.cert)
    });
  } else {
    return (0, import_fastify.fastify)({
      http2: true
    });
  }
}
async function start(opts) {
  const h1Server = createH1Server(opts);
  await h1Server.register(import_cors.default, import_interop.interop.corsOptions);
  await h1Server.register(import_connect_fastify.fastifyConnectPlugin, { routes: import_routes.default });
  await h1Server.listen({ host: HOST, port: opts.h1port });
  console.log(`Running ${opts.insecure ? "insecure" : "secure"} HTTP/1.1 server on `, h1Server.addresses());
  const h2Server = createH2Server(opts);
  await h2Server.register(import_cors.default, import_interop.interop.corsOptions);
  await h2Server.register(import_connect_fastify.fastifyConnectPlugin, { routes: import_routes.default });
  await h2Server.listen({ host: HOST, port: opts.h2port });
  console.log(`Running ${opts.insecure ? "insecure" : "secure"} HTTP/2 server on `, h2Server.addresses());
  return new Promise((resolve) => {
    resolve();
  });
}
