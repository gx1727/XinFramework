// @ts-check

import mdx from "@astrojs/mdx";
import sitemap from "@astrojs/sitemap";
import { defineConfig, fontProviders } from "astro/config";

// https://astro.build/config
export default defineConfig({
  // site 和 base 用于绝对路径部署
  // 注释掉则使用相对路径（适合任意子目录部署）
  // site: "https://gx1727.github.io",
  // base: "/XinFramework",
  site: "https://gx1727.github.io",
  base: "/XinFramework",
  integrations: [mdx(), sitemap()],
  outDir: "./docs",
  fonts: [
    {
      provider: fontProviders.local(),
      name: "Atkinson",
      cssVariable: "--font-atkinson",
      fallbacks: ["sans-serif"],
      options: {
        variants: [
          {
            src: ["./src/assets/fonts/atkinson-regular.woff"],
            weight: 400,
            style: "normal",
            display: "swap",
          },
          {
            src: ["./src/assets/fonts/atkinson-bold.woff"],
            weight: 700,
            style: "normal",
            display: "swap",
          },
        ],
      },
    },
  ],
});
