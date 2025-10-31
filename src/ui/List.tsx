import React, { useMemo } from "react";
import { Box, Text } from "ink";
import stringWidth from "string-width";
import { useStore } from "../state/store.js";
import { Theme } from "../model.js";


export function List({ theme, width, height, windowSize }: { theme: Theme, width: number, height: number, windowSize: number }) {
  const { todos, ui } = useStore();
  const cursor = ui.cursor;


  // 仮想スクロール: カーソル周辺だけ描画
  const start = Math.max(0, cursor - Math.floor(windowSize / 2));
  const end = Math.min(todos.length, start + windowSize);


  const rows = useMemo(() => todos.slice(start, end), [todos, start, end]);


  return (
    <Box flexDirection="column" >
      {
        rows.map((t, i) => {
          const idx = start + i;
          const selected = idx === cursor;
          const mark = selected ? "▶" : " ";
          const status = t.done ? "[x]" : "[ ]";
          const title = t.title.replace(/\n/g, " ");
          const pad = Math.max(0, width - 6 - stringWidth(title));
          return (
            <Text key={t.id} backgroundColor={selected ? theme.selection : undefined} >
              {mark} {status} {title} {" ".repeat(pad)}
            </Text>
          );
        })
      }
      {todos.length === 0 && <Text dimColor > Empty.Press 'a' to add.</Text>}
    </Box>
  );
}
