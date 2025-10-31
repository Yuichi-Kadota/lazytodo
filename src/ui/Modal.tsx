import React, { useState } from "react";
import { Box, Text } from "ink";
import { useStore } from "../state/store.js";
import { Theme } from "../model.js";


export default function Modal({ theme, width }: { theme: Theme, width: number }) {
  const { todos, ui, closeModal, updateDetailTags } = useStore();
  const idx = ui.cursor;
  const t = todos[idx];
  const [detail, setDetail] = useState<string>(t?.detail || "");
  const [tags, setTags] = useState<string>((t?.tags || []).join(", "));


  if (!ui.modalOpen || !t) return null;


  return (
    <Box flexDirection="column" borderStyle={theme.border} padding={1}>
      <Text color={theme.accent}>Edit: {t.title}</Text>
      <Text dimColor>Tags: {tags || "(none)"}</Text>
      <Text dimColor>Detail: {detail || "(none)"}</Text>
      <Text dimColor>Press Enter or Esc to close</Text>
    </Box>
  );
}
