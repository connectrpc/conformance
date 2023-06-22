#!/usr/bin/env node
var import_commander = require("commander");
var import_server = require("./server.js");
const program = new import_commander.Command();
function validateNumber(value) {
  const parsedValue = parseInt(value, 10);
  if (Number.isNaN(value)) {
    throw new import_commander.InvalidArgumentError("option must be a number.");
  }
  return parsedValue;
}
program.name("start").command("start").description("Start a Connect server using connect-node").requiredOption("--h1port <port>", "port for HTTP/1.1 traffic", validateNumber).requiredOption("--h2port <number>", "port for HTTP/2 traffic", validateNumber).option("--cert <cert>", "path to the TLS cert file").option("--key <key>", "path to the TLS key file").option("--insecure", "whether to server cleartext or TLS. HTTP/3 requires TLS").action((options) => {
  if (!options.insecure && (!options.key || !options.cert)) {
    console.error("error: either a 'cert' and 'key' combination or 'insecure' must be specified");
    return;
  }
  (0, import_server.start)(options);
});
program.parse();
