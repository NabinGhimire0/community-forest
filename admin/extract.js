import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

// Fix __dirname in ES module
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const ROOT_DIR = __dirname;

const OUTPUT_FILE = "v2.txt";

const EXCLUDED_FOLDERS = [
  "node_modules",
  ".git",
  ".dart_tool",
  ".idea",
  "build",
  "dist",
  "coverage",
  ".next",
  "vendor",
  "version"
];

const EXCLUDED_EXTENSIONS = [
  ".png",
  ".jpg",
  ".jpeg",
  ".gif",
  ".mp4",
  ".mp3",
  ".zip",
  ".exe",
  ".dll",
  ".ico",
  ".pdf"
];

let output = "";

function walkDir(dir, level = 0) {
  const files = fs.readdirSync(dir);

  files.forEach(file => {
    const fullPath = path.join(dir, file);
    const stat = fs.statSync(fullPath);

    if (stat.isDirectory() && EXCLUDED_FOLDERS.includes(file)) {
      return;
    }

    const indent = "  ".repeat(level);

    if (stat.isDirectory()) {
      output += `${indent}📁 ${file}\n`;
      walkDir(fullPath, level + 1);
    } else {
      const ext = path.extname(file);

      if (EXCLUDED_EXTENSIONS.includes(ext)) {
        return;
      }

      output += `${indent}📄 ${file}\n`;

      try {
        const content = fs.readFileSync(fullPath, "utf8");

        output += `${indent}${"═".repeat(80)}\n`;
        output += `${indent}FILE: ${fullPath}\n`;
        output += `${indent}${"─".repeat(80)}\n`;
        output += `${content}\n`;
        output += `${indent}${"═".repeat(80)}\n\n\n`;

      } catch (err) {
        output += `${indent}[Binary or unreadable file]\n\n`;
      }
    }
  });
}

walkDir(ROOT_DIR);

fs.writeFileSync(path.join(__dirname, OUTPUT_FILE), output);

console.log("✅ Project exported successfully!");