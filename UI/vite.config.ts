import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    // Windows 默认 localhost 同时解析 IPv4 + IPv6，IPv6 ::1 常因权限被拒。
    // 强制 127.0.0.1 只走 IPv4，避免 EACCES: permission denied ::1:5173。
    //
    // 端口选择：5173 / 5174 等 Vite 默认端口落在 Windows Hyper-V 排除范围
    // （5041-5140 / 5141-5240），绑定会 EACCES。改用 5241 避开保留段。
    host: "127.0.0.1",
    port: 5241,
    strictPort: true,
  },
})
