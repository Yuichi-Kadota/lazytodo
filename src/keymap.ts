import { KeyBinding } from "./model";


export function matchKey(binding: string | string[], input: string, key: any): boolean {
  const arr = Array.isArray(binding) ? binding : [binding];
  for (const b of arr) {
    if (b === input) return true;
    if (b === "upArrow" && key.upArrow) return true;
    if (b === "downArrow" && key.downArrow) return true;
    if (b === "escape" && key.escape) return true;
    if (b === "ctrl+c" && key.ctrl && input === "c") return true;
  }
  return false;
}
