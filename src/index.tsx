#!/usr/bin/env node
import React from "react";
import { render } from "ink";
import { Command } from "commander";
import App from "./app";
import { loadConfig, saveExampleConfig } from "./config";
import { DEFAULT_THEME_DARK } from "./model";
import { useStore } from "./state/store";
import { exportCsv, exportMarkdown, loadData, loadDataLazy, saveData } from "./persistence";


const program = new Command();
program
  .name("lazytodo")
  .description("LazyTodo — Single-queue TODO (Ink)")
  .option("--data <path>", "data.json path")
  .option("--export-dir <dir>", "export directory")
  .option("--write-sample-config", "write ~/.config/todoq/config.yaml if missing")
  .parse(process.argv);


const opt = program.opts<{ data?: string; exportDir?: string; writeSampleConfig?: boolean }>();


if (opt.writeSampleConfig) saveExampleConfig();


const config = loadConfig({ dataPath: opt.data, exportDir: opt.exportDir });
const theme = config.theme || DEFAULT_THEME_DARK;


// Main async function to avoid top-level await
(async () => {
  let loaded = false;
  for await (const chunk of (loadDataLazy(config.dataPath))) {
    loaded = true;
    useStore.getState().setTodos([...useStore.getState().todos, ...chunk]);
  }
  if (!loaded) {
    // 既存ファイルなし等は空で開始
    const data = loadData(config.dataPath);
    useStore.getState().setTodos(data.todos);
  }

  // 同期保存：todos が変わるたびに即時保存
  useStore.subscribe((state) => {
    saveData(config.dataPath, { todos: state.todos });
  });

  function onExportMd() {
    const p = exportMarkdown(config.exportDir!, useStore.getState().todos);
    process.stdout.write(`
Exported Markdown: ${p}
`);
  }

  function onExportCsv() {
    const p = exportCsv(config.exportDir!, useStore.getState().todos);
    process.stdout.write(`
Exported CSV: ${p}
`);
  }

  const { waitUntilExit } = render(
    <App theme={theme} config={config} onExportMd={onExportMd} onExportCsv={onExportCsv} />
  );

  await waitUntilExit();
})();
