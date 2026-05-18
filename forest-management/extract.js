const fs = require("fs");
const path = require("path");

const ROOT_DIR = "./";

const OUTPUT_FILE = "project_dump.txt";

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

    // Skip excluded folders
    if (
      stat.isDirectory() &&
      EXCLUDED_FOLDERS.includes(file)
    ) {
      return;
    }

    const indent = "  ".repeat(level);

    // Folder
    if (stat.isDirectory()) {

      output += `${indent}📁 ${file}\n`;

      walkDir(fullPath, level + 1);

    } else {

      const ext = path.extname(file);

      // Skip excluded file types
      if (EXCLUDED_EXTENSIONS.includes(ext)) {
        return;
      }

      output += `${indent}📄 ${file}\n`;

      try {

        const content = fs.readFileSync(fullPath, "utf8");

        // Border Start
        output += `${indent}${"═".repeat(80)}\n`;

        output += `${indent}FILE: ${fullPath}\n`;

        output += `${indent}${"─".repeat(80)}\n`;

        output += `${content}\n`;

        // Border End
        output += `${indent}${"═".repeat(80)}\n\n\n`;

      } catch (err) {

        output += `${indent}[Binary or unreadable file]\n\n`;

      }
    }

  });

}

walkDir(ROOT_DIR);

fs.writeFileSync(OUTPUT_FILE, output);

console.log(`✅ Project exported successfully to ${OUTPUT_FILE}`);