module.exports = function (config) {
  config.set({
    frameworks: ["jasmine"],
    files: ["spec/**/*.ts", "gen/**/*.ts"],
    preprocessors: {
      "/**/*.ts": "esbuild",
    },
    reporters: ["progress"],
    browsers: ["ChromeHeadless"],
    singleRun: true,
    esbuild: {
      target: "esnext",
      tsconfig: "./tsconfig.json",
    },
  });
};
