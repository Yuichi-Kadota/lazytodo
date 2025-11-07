export type ID = string;


export interface Todo {
  id: ID;
  title: string; // list はタイトルのみ表示
  detail?: string; // モーダルで表示/編集
  createdAt: number; // epoch ms
  done?: boolean; // 将来的な拡張
}


export interface KeyBinding {
  // キー名は Ink useInput の input/key を想定（例: 'j','k','g','G','enter','d','x','i','o','?','/','esc'）
  [action: string]: string | string[];
}


export interface Theme {
  name: string;
  accent: string; // e.g. 'cyan'
  fg: string;
  dim: string;
  danger: string;
  selection: string; // 選択中の行
  border: "single" | "double" | "round";
}


export interface AppConfig {
  dataPath: string; // JSON 永続化先（既定: ~/.config/todoq/data.json）
  exportDir?: string; // エクスポート先（md/csv）
  theme: Theme;
  keymap: KeyBinding; // 既定をマージ
  listWindowSize?: number; // 仮想スクロール: 表示ウィンドウ行数
}


export const DEFAULT_THEME_DARK: Theme = {
  name: "dark",
  accent: "cyan",
  fg: "white",
  dim: "gray",
  danger: "red",
  selection: "#1f2937",
  border: "round"
};


export const DEFAULT_THEME_LIGHT: Theme = {
  name: "light",
  accent: "blue",
  fg: "black",
  dim: "#6b7280",
  danger: "red",
  selection: "#e5e7eb",
  border: "round"
};


export const DEFAULT_KEYMAP: KeyBinding = {
  up: ["k", "upArrow"],
  down: ["j", "downArrow"],
  top: "g",
  bottom: "G",
  openModal: ["enter", "o"],
  toggleDone: ["x"],
  delete: ["d", "D"],
  add: "a",
  quit: "q",
  exportMd: "m",
  exportCsv: "c",
}; 
