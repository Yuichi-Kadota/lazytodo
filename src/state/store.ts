import { create } from "zustand";
import { Todo, ID } from "../model";


export type UIState = {
  cursor: number; // 選択行
  modalOpen: boolean;
  filter: string; // 将来の検索
}


export type Store = {
  todos: Todo[];
  ui: UIState;
  setTodos: (xs: Todo[]) => void;
  add: (title: string) => void;
  removeAt: (idx: number) => void;
  toggleDoneAt: (idx: number) => void;
  setCursor: (i: number) => void;
  openModal: () => void;
  closeModal: () => void;
  updateDetailTags: (idx: number, detail: string, tags: string[]) => void;
  updateTodo: (idx: number, title: string, detail: string, tags: string[]) => void;
}


function newId(): ID { return Math.random().toString(36).slice(2); }


export const useStore = create<Store>((set, get) => ({
  todos: [],
  ui: { cursor: 0, modalOpen: false, filter: "" },
  setTodos: (xs) => set({ todos: xs }),
  add: (title) => set(s => ({ todos: [...s.todos, { id: newId(), title, createdAt: Date.now() }] })),
  removeAt: (idx) => set(s => ({ todos: s.todos.filter((_, i) => i !== idx), ui: { ...s.ui, cursor: Math.max(0, Math.min(s.ui.cursor, s.todos.length - 2)) } })),
  toggleDoneAt: (idx) => set(s => ({ todos: s.todos.map((t, i) => i === idx ? { ...t, done: !t.done } : t) })),
  setCursor: (i) => set(s => ({ ui: { ...s.ui, cursor: Math.max(0, Math.min(i, s.todos.length - 1)) } })),
  openModal: () => set(s => ({ ui: { ...s.ui, modalOpen: true } })),
  closeModal: () => set(s => ({ ui: { ...s.ui, modalOpen: false } })),
  updateDetailTags: (idx, detail, tags) => set(s => ({ todos: s.todos.map((t, i) => i === idx ? { ...t, detail, tags } : t) })),
  updateTodo: (idx, title, detail, tags) => set(s => ({ todos: s.todos.map((t, i) => i === idx ? { ...t, title, detail, tags } : t) }))
}));
