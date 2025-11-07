import fs from "node:fs";
import path from "node:path";
import type { Todo } from "./model";

export interface DataShape {
  todos: Todo[];
}

export function loadData(file: string): DataShape {
  if (!fs.existsSync(file)) {
    return { todos: [] };
  }
  try {
    const raw = fs.readFileSync(file, "utf8");
    const data = JSON.parse(raw) as DataShape;
    return data;
  } catch (err) {
    console.error("Error loading data:", err);
    return { todos: [] };
  }
}

export async function* loadDataLazy(file: string, chunkSize = 100): AsyncGenerator<Todo[]> {
  if (!fs.existsSync(file)) {
    return;
  }
  try {
    const raw = fs.readFileSync(file, "utf8");
    const data = JSON.parse(raw) as DataShape;
    const todos = data.todos;
    for (let i = 0; i < todos.length; i += chunkSize) {
      yield todos.slice(i, i + chunkSize);
    }
  } catch (err) {
    console.error("Error loading data lazily:", err);
  }
}

export function saveData(file: string, data: DataShape) {
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, JSON.stringify(data, null, 2), "utf8");
}

export function exportMarkdown(dir: string, todos: Todo[]) {
  const lines: string[] = [];
  lines.push(`# TODO Queue`);
  lines.push("");
  for (const t of todos) {
    const tagStr = t.tags?.length ? ` [${t.tags.join(', ')}]` : "";
    const detail = t.detail ? `\n${t.detail.replace(/\n/g, "\n")}` : "";
    lines.push(`- [${t.done ? "x" : " "}] ${t.title}${tagStr}${detail}`);
  }
  const out = path.join(dir, `todoq_${new Date().toISOString().slice(0, 10)}.md`);
  fs.writeFileSync(out, lines.join("\n"), "utf8");
  return out;
}

export function exportCsv(dir: string, todos: Todo[]) {
  const head = ["id", "title", "detail", "tags", "createdAt", "done"];
  const rows = todos.map(t => [
    t.id,
    escapeCsv(t.title),
    escapeCsv(t.detail || ""),
    escapeCsv((t.tags || []).join("|")),
    String(t.createdAt),
    t.done ? "1" : "0"
  ].join(","));
  const out = path.join(dir, `todoq_${Date.now()}.csv`);
  fs.writeFileSync(out, head.join(",") + "\n" + rows.join("\n"), "utf8");
  return out;
}

function escapeCsv(s: string) {
  if (s.includes(",") || s.includes("\"") || s.includes("\n")) {
    return '"' + s.replace(/"/g, '""') + '"';
  }
  return s;
}
