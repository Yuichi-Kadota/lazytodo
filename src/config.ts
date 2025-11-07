import fs from "node:fs";
import path from "node:path";
import os from "node:os";
import YAML from "yaml";
import { AppConfig, DEFAULT_KEYMAP, DEFAULT_THEME_DARK, DEFAULT_THEME_LIGHT, Theme } from "./model";


const XDG = process.env.XDG_CONFIG_HOME || path.join(os.homedir(), ".config");
const APPDIR = path.join(XDG, "todoq");
const DEFAULT_DATA = path.join(APPDIR, "data.json");
const DEFAULT_EXPORT = path.join(APPDIR, "export");
const CONFIG_PATH = path.join(APPDIR, "config.yaml");


export function ensureDirs() {
  fs.mkdirSync(APPDIR, { recursive: true });
  fs.mkdirSync(DEFAULT_EXPORT, { recursive: true });
}


export function loadConfig(cli?: Partial<Pick<AppConfig, "dataPath" | "exportDir">>): AppConfig {
  ensureDirs();
  let user: Partial<AppConfig> = {};
  if (fs.existsSync(CONFIG_PATH)) {
    const raw = fs.readFileSync(CONFIG_PATH, "utf8");
    user = YAML.parse(raw) as Partial<AppConfig>;
  }


  const theme: Theme = ((): Theme => {
    const t = user?.theme as any;
    if (t?.preset === "light") return DEFAULT_THEME_LIGHT;
    if (t?.preset === "dark") return DEFAULT_THEME_DARK;
    if (t?.name) return t as Theme;
    return DEFAULT_THEME_DARK;
  })();


  return {
    dataPath: cli?.dataPath || user?.dataPath || DEFAULT_DATA,
    exportDir: cli?.exportDir || user?.exportDir || DEFAULT_EXPORT,
    theme,
    keymap: { ...DEFAULT_KEYMAP, ...(user?.keymap || {}) },
    listWindowSize: user?.listWindowSize ?? 30,
  };
}


export function saveExampleConfig() {
  if (!fs.existsSync(CONFIG_PATH)) {
    const sample = YAML.stringify({
      dataPath: DEFAULT_DATA,
      exportDir: DEFAULT_EXPORT,
      theme: { preset: "dark" },
      keymap: { down: ["j", "downArrow"], up: ["k", "upArrow"] },
      listWindowSize: 30
    });
    fs.writeFileSync(CONFIG_PATH, sample, "utf8");
  }
}
