import React, { useEffect, useState, Suspense } from "react";
import { Box, Text, useApp, useInput, measureElement } from "ink";
import { useStore } from "./state/store.js";
import { Theme, AppConfig } from "./model.js";
import { List } from "./ui/List.js";
import { matchKey } from "./keymap.js";


const LazyModal = React.lazy(() => import("./ui/Modal.js"));


export default function App({ theme, config, onExportMd, onExportCsv }: {
  theme: Theme;
  config: AppConfig;
  onExportMd: () => void;
  onExportCsv: () => void;
}) {
  const { exit } = useApp();
  const { todos, ui, setCursor, add, removeAt, toggleDoneAt, openModal, closeModal } = useStore();
  const [boxSize, setBoxSize] = useState({ width: 80, height: 24 });


  useInput((input, key) => {
    const k = config.keymap;
    if (matchKey(k.quit, input, key)) exit();
    else if (ui.modalOpen) {
      if (key.return) {
        closeModal();
      } else if (key.escape) {
        closeModal();
      }
      return;
    } else {
      if (matchKey(k.down, input, key)) setCursor(useStore.getState().ui.cursor + 1);
      else if (matchKey(k.up, input, key)) setCursor(useStore.getState().ui.cursor - 1);
      else if (matchKey(k.top, input, key)) setCursor(0);
      else if (matchKey(k.bottom, input, key)) setCursor(todos.length - 1);
      else if (matchKey(k.toggleDone, input, key)) toggleDoneAt(useStore.getState().ui.cursor);
      else if (matchKey(k.delete, input, key)) removeAt(useStore.getState().ui.cursor);
      else if (matchKey(k.add, input, key)) addPrompt(add);
      else if (matchKey(k.openModal, input, key)) openModal();
      else if (matchKey(k.exportMd, input, key)) onExportMd();
      else if (matchKey(k.exportCsv, input, key)) onExportCsv();
    }
  });


  const ref = React.useRef<any>(null);
  useEffect(() => {
    if (!ref.current) return;
    const s = measureElement(ref.current);
    setBoxSize({ width: s.width, height: s.height });
  }, [ref.current]);


  return (
    <Box ref={ref} flexDirection="column">
      <Text color={theme.accent}>LazyQueue — single queue (v0.1)</Text>
      <Text dimColor>
        {"j/k:move g/G:top/bottom a:add x:toggle d:delete enter:o:modal :m md export :c csv export q:quit"}
      </Text>
      <Box borderStyle={theme.border} paddingX={1} paddingY={0}>
        <List theme={theme} width={boxSize.width - 4} height={boxSize.height - 6} windowSize={config.listWindowSize!} />
      </Box>
      {ui.modalOpen && (
        <Box marginTop={1}>
          <Suspense fallback={<Text dimColor>Loading modal...</Text>}>
            <LazyModal theme={theme} width={boxSize.width - 4} />
          </Suspense>
        </Box>
      )}
    </Box>
  );
}


function addPrompt(add: (title: string) => void) {
  add("New task");
}
