import { expect, test } from "@playwright/test";
import assert from "node:assert/strict";
import { fileUrl, REPORTS } from "./paths";
import type { CostEstimate, Report } from "../src/report/schema.gen";

function expectedAmounts(cost: CostEstimate, period: "daily" | "weekly" | "monthly", currency: string) {
  const convert = (cents: number) => {
    const value = BigInt(cents);
    const magnitude = value < 0n ? -value : value;
    const rounded = period === "monthly" ? magnitude : (magnitude * (period === "daily" ? 24n : 168n) + 365n) / 730n;
    return value < 0n ? -rounded : rounded;
  };
  assert(cost.before_monthly_cents !== undefined && cost.after_monthly_cents !== undefined);
  const before = convert(cost.before_monthly_cents);
  const after = convert(cost.after_monthly_cents);
  return [before, after, after - before].map((amount, i) => {
    const magnitude = amount < 0n ? -amount : amount;
    return `${currency} ${amount < 0n ? "-" : i === 2 && amount > 0n ? "+" : ""}${magnitude / 100n}.${String(magnitude % 100n).padStart(2, "0")}`;
  });
}

test("unpriced reports omit pricing controls and inspector estimates", async ({ page }) => {
  await page.goto(fileUrl(REPORTS.sample));
  await expect(page.locator(".react-flow__node").first()).toBeVisible();
  await page.getByRole("button", { name: /review-events/ }).first().click();
  await expect(page.locator(".inspector")).toBeVisible();
  await expect(page.getByRole("region", { name: "Estimated resource-cost subtotal" })).toHaveCount(0);
  await expect(page.getByRole("combobox", { name: "Cost period" })).toHaveCount(0);
  await expect(page.getByRole("heading", { name: "Resource-cost estimate" })).toHaveCount(0);
});

test("one unavailable cost side never becomes zero or a delta", async ({ page }) => {
  await page.goto(fileUrl(REPORTS.pricingPartial));
  await page.getByRole("button", { name: /private_example/ }).first().click();
  const inspector = page.locator(".inspector");
  const amounts = inspector.locator(".pricing-amounts dd");
  await expect(amounts).toHaveText(["USD 0.37 →", "Not estimated", "Not estimated"]);
  await expect(inspector).toContainText("Create-first overlap: additional cost not estimated");
  await inspector.getByRole("combobox", { name: "Cost period" }).selectOption("daily");
  await expect(amounts).toHaveText(["USD 0.01 →", "Not estimated", "Not estimated"]);
});

for (const safe of [false, true]) {
  test(`real exporter pricing stays offline${safe ? " with safe-share" : ""}`, async ({ page }, info) => {
    const path = safe ? REPORTS.pricingSafe : REPORTS.pricing;
    const requests: string[] = [];
    const errors: string[] = [];
    await page.context().route(/^(?!file:)/, (route) => route.abort());
    page.on("request", (request) => requests.push(request.url()));
    page.on("pageerror", (error) => errors.push(error.message));
    page.on("console", (msg) => { if (["error", "warning"].includes(msg.type())) errors.push(msg.text()); });
    await page.addInitScript(() => document.addEventListener("securitypolicyviolation", (event) => console.error(`CSP: ${event.violatedDirective}`)));
    await page.goto(fileUrl(path));
    const report: Report = JSON.parse(await page.locator("#tfviz-report").textContent() ?? "");
    const pricing = report.pricing;
    assert(pricing);
    const summary = page.getByRole("region", { name: "Estimated resource-cost subtotal" });
    await expect(summary).toBeVisible();
    for (const text of [...expectedAmounts(pricing.summary, "monthly", pricing.currency), "1 estimated", "1 not estimated", "730 hours/month", "2026-01-01", "not a full bill", "usage-based charges excluded", "Discounts and taxes excluded"]) {
      await expect(summary).toContainText(text);
    }
    const selectedName = safe ? /EC2 instance 1/ : /private_example/;
    await page.getByRole("button", { name: selectedName }).first().click();
    const inspector = page.locator(".inspector");
    for (const text of ["Resource-cost estimate", "Steady-state", "USD 0.37", "USD 0.73", "USD +0.36", "us-east-1", "us-west-2", "Create-first overlap: additional cost not estimated"]) {
      await expect(inspector).toContainText(text);
    }
    const resource = report.resources.find((resource) => resource.cost?.before_monthly_cents !== undefined);
    assert(resource?.cost);
    await expect(summary.getByRole("combobox", { name: "Cost period" })).toHaveValue("monthly");
    // Both native controls share state, including changes initiated in the inspector.
    for (const period of ["daily", "weekly", "monthly"] as const) {
      const control = period === "weekly" ? inspector : summary;
      await control.getByRole("combobox", { name: "Cost period" }).selectOption(period);
      for (const container of [summary, inspector]) {
        await expect(container.getByRole("combobox", { name: "Cost period" })).toHaveValue(period);
      }
      for (const [container, cost] of [[summary, pricing.summary], [inspector, resource.cost]] as const) {
        const cells = container.locator(".pricing-amounts dd");
        const amounts = expectedAmounts(cost, period, pricing.currency);
        for (const [i, amount] of amounts.entries()) await expect(cells.nth(i)).toContainText(amount);
      }
    }
    // Sensitive evidence withholds both sides, even when one rate matches.
    const withheldName = safe ? /EC2 instance 2/ : /withheld_example/;
    await page.getByRole("button", { name: withheldName }).first().focus();
    await page.keyboard.press("Enter");
    await expect(inspector.locator(".pricing-amounts")).toContainText("Not estimated");
    await expect(inspector.locator(".pricing-amounts")).not.toContainText("USD 0.00");
    await expect(inspector).toContainText("Estimate withheld because pricing evidence is sensitive.");
    await summary.getByRole("combobox", { name: "Cost period" }).selectOption("daily");
    await expect(inspector.locator(".pricing-amounts dd")).toHaveText(["Not estimated →", "Not estimated", "Not estimated"]);
    expect(await page.locator("body").textContent()).not.toContain("CANARY-");
    if (safe) expect(await page.locator("body").textContent()).not.toContain("private_example");
    expect(errors).toEqual([]);
    expect(requests.filter((url) => url !== fileUrl(path))).toEqual([]);
    await page.screenshot({ path: info.outputPath("pricing.png"), fullPage: true });
  });
}
