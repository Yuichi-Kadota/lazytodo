import React, { useState } from "react";
import { Box, Text, useInput } from "ink";
import TextInput from "ink-text-input";
import { useStore } from "../state/store";
import { Theme } from "../model";


type Field = "title" | "detail";
type Mode = "browse" | "edit";

export default function Modal({ theme, width }: { theme: Theme, width: number }) {
  const { todos, ui, closeModal, updateTodo } = useStore();
  const idx = ui.cursor;
  const t = todos[idx];

  const [title, setTitle] = useState<string>(t?.title || "");
  const [detail, setDetail] = useState<string>(t?.detail || "");
  const [selectedField, setSelectedField] = useState<Field>("title");
  const [mode, setMode] = useState<Mode>("browse");

  if (!ui.modalOpen || !t) return null;

  const handleSave = () => {
    updateTodo(idx, title, detail);
    closeModal();
  };

  const moveFieldDown = () => {
    if (selectedField === "title") setSelectedField("detail");
  };

  const moveFieldUp = () => {
    if (selectedField === "detail") setSelectedField("title");
  };

  useInput((input, key) => {
    if (key.ctrl && input === "s") {
      handleSave();
    } else if (key.escape) {
      if (mode === "edit") {
        setMode("browse");
      } else {
        closeModal();
      }
    } else if (mode === "browse") {
      if (input === "j" || key.downArrow) {
        moveFieldDown();
      } else if (input === "k" || key.upArrow) {
        moveFieldUp();
      } else if (input === " ") {
        setMode("edit");
      }
    } else if (mode === "edit") {
      if (key.return) {
        setMode("browse");
      }
    }
  });

  const getFieldColor = (field: Field) => {
    if (mode === "edit" && field === selectedField) return theme.accent;
    if (mode === "browse" && field === selectedField) return theme.selection;
    return theme.dim;
  };

  const getCursor = (field: Field) => {
    return mode === "browse" && field === selectedField ? "▶ " : "  ";
  };

  return (
    <Box flexDirection="column" borderStyle={theme.border} padding={1}>
      <Text color={theme.accent}>Edit Todo</Text>
      <Box marginTop={1} flexDirection="column">
        <Box>
          <Text>{getCursor("title")}</Text>
          <Text color={getFieldColor("title")}>Title: </Text>
          {mode === "edit" && selectedField === "title" ? (
            <TextInput
              value={title}
              onChange={setTitle}
              placeholder="Enter title"
              focus={true}
            />
          ) : (
            <Text>{title || "(empty)"}</Text>
          )}
        </Box>
        <Box marginTop={1}>
          <Text>{getCursor("detail")}</Text>
          <Text color={getFieldColor("detail")}>Detail: </Text>
          {mode === "edit" && selectedField === "detail" ? (
            <TextInput
              value={detail}
              onChange={setDetail}
              placeholder="Enter task details"
              focus={true}
            />
          ) : (
            <Text>{detail || "(none)"}</Text>
          )}
        </Box>
      </Box>
      <Box marginTop={1}>
        <Text dimColor>
          {mode === "browse"
            ? "j/k: move | Space: edit | Ctrl+S: save | Esc: cancel"
            : "Enter: done | Esc: cancel edit | Ctrl+S: save"}
        </Text>
      </Box>
    </Box>
  );
}
