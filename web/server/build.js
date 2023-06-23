const esbuild = require("esbuild");
const { glob } = require("glob");

// Builds the server code into CJS using esbuild. The package.json file is CJS (i.e. not type=module)
// because setting to module will cause grpc-web code to break.
//
// In addition, our esbuild is done via the JavaScript API so that we can use glob syntax
(async () => {
  let result = await esbuild.build({
    entryPoints: await glob(["server/**/*.ts", "gen/**/*.ts"]),
    outdir: "dist",
    format: "cjs",
  });
})();
