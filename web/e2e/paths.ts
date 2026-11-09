import { fileURLToPath, pathToFileURL } from "node:url";

const dir = fileURLToPath(new URL("../test-results/e2e-reports/", import.meta.url));

export const REPORTS = {
  dir,
  sample: `${dir}sample.html`,
  hostile: `${dir}hostile.html`,
  large: `${dir}large.html`,
  safeShare: `${dir}safe-share.html`,
  icons: `${dir}icons.html`,
  state: `${dir}state.html`,
  pricing: `${dir}pricing.html`,
  pricingSafe: `${dir}pricing-safe.html`,
  pricingPartial: `${dir}pricing-partial.html`,
} as const;

export const fileUrl = (path: string) => pathToFileURL(path).href;
