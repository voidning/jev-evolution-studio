import {inspectProject,sha} from './source-core.mjs';
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

// Explicit IDs today; an AST injection transform can replace manifest() later.
export default function jevEditor() {
  let root;
  let serving;
  const bridge = fs.readFileSync(
    fileURLToPath(new URL("./selection-bridge.js", import.meta.url)),
    "utf8",
  );
  function manifest(){ return inspectProject(root); }
  return {
    name: "jev-explicit-editor",
    enforce: "post",
    configResolved(config) {
      root = fs.realpathSync(config.root);
      serving=config.command==='serve';
      if (config.base !== "/") throw Error("Jev MVP requires Vite base /");
      if (!inspectProject(root).tailwind && !fs.existsSync(path.join(root,"src/jev-edits.css"))) throw Error("Missing src/jev-edits.css");
      manifest();
    },
    transform(code,id) {
      const file=id.split('?')[0];
      if (serving && /\.[jt]sx$/.test(file)&&file.startsWith(root+path.sep)&&!file.includes('/node_modules/')){
        const revision=sha(fs.readFileSync(file,'utf8')),relative=path.relative(root,file).split(path.sep).join('/');
        return {code:code+'\n;globalThis.__JEV_REVISIONS__ ??= {}; globalThis.__JEV_REVISIONS__['+JSON.stringify(relative)+']='+JSON.stringify(revision)+';',map:null};
      }
    },
    transformIndexHtml: {
      order: "post",
      handler(_html, ctx) {
        return ctx.server
          ? [
              {
                tag: "script",
                attrs: { type: "module", src: "/__jev/client.js" },
                injectTo: "body",
              },
            ]
          : [];
      },
    },
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        if (req.url === "/__jev/meta") {
          try {
            res.setHeader("Content-Type", "application/json");
            res.setHeader("Cache-Control", "no-store");
            res.end(JSON.stringify(manifest()));
          } catch {
            res.statusCode = 422;
            res.end("Invalid explicit JSX identifiers");
          }
          return;
        }
        if (req.url === "/__jev/client.js") {
          res.setHeader("Content-Type", "text/javascript");
          res.end(
            "const projectRoot = " + JSON.stringify(root) + ";\n" + bridge,
          );
          return;
        }
        next();
      });
    },
  };
}
