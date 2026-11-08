import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { assert, describe, expect, it } from "vitest";
import { Header } from "../app/Header";
import { ReportIndexContext } from "../app/reportContext";
import { Inspector } from "../inspector/Inspector";
import { loadSample } from "../test/sample";
import { formatMonthly, periodAmounts, periodCents } from "./Pricing";
import type { Report } from "./schema.gen";
import { indexReport } from "./view";

function header(report: Report) {
  return renderToStaticMarkup(createElement(Header, {
    report, view: "changes", structure: "architecture", theme: "light",
    onViewChange() {}, onStructureChange() {}, onThemeChange() {},
  }));
}

function inspector(report: Report) {
  const index = indexReport(report);
  return renderToStaticMarkup(createElement(ReportIndexContext.Provider, { value: index },
    createElement(Inspector, { index, resource: report.resources[0], onSelect() {} })));
}

function priced(): Report {
  const report = loadSample();
  report.pricing = {
    currency: "ZZZ", as_of: "2026-11-01", hours_per_month: 730,
    estimated_resources: 1, unestimated_resources: report.resources.length - 1,
    summary: { before_monthly_cents: 12345, after_monthly_cents: 10000, delta_monthly_cents: -2345 },
  };
  const resource = report.resources[0];
  assert(resource);
  resource.cost = {
    ...report.pricing.summary, before_region: "us-east-1", after_region: "eu-west-1",
  };
  return report;
}

describe("offline pricing", () => {
  it("converts periods with exact BigInt arithmetic and symmetric half-up rounding", () => {
    expect(periodCents(12345, "daily")).toBe(406n);
    expect(periodCents(12345, "weekly")).toBe(2841n);
    expect(periodCents(12345, "monthly")).toBe(12345n);
    expect(periodCents(-12345, "daily")).toBe(-406n);
    // Just below/above a half cent (exact ties cannot occur with integer cents and these ratios).
    expect(periodCents(15, "daily")).toBe(0n);
    expect(periodCents(16, "daily")).toBe(1n);
    expect(periodCents(-16, "daily")).toBe(-1n);
    expect(periodCents(9007199254740991, "daily")).toBe((9007199254740991n * 24n + 365n) / 730n);
    for (const cents of [undefined, NaN, Infinity, 1.5, 9007199254740992]) {
      expect(periodCents(cents, "weekly")).toBeUndefined();
    }
  });

  it("derives delta from rounded sides, including beyond safe integer differences", () => {
    const cost = { before_monthly_cents: 15, after_monthly_cents: 16, delta_monthly_cents: 1 };
    expect(periodAmounts(cost, "daily")).toEqual({ before: 0n, after: 1n, delta: 1n });
    expect(periodAmounts(undefined, "weekly")).toEqual({ before: undefined, after: undefined, delta: undefined });
    expect(periodAmounts({ delta_monthly_cents: -16 }, "daily")).toEqual({ before: undefined, after: undefined, delta: -1n });
    expect(periodAmounts({ before_monthly_cents: -9007199254740991, after_monthly_cents: 9007199254740991 }, "monthly").delta)
      .toBe(18014398509481982n);
  });

  it("plainly labels partial coverage and optional usage without exposing inputs", () => {
    const report = priced();
    assert(report.pricing && report.resources[0]?.cost);
    report.pricing.usage_provided = true;
    report.pricing.summary.partial = true;
    report.pricing.summary.usage_included = true;
    expect(header(report)).toContain("Partial coverage");
    expect(header(report)).toContain("Provided monthly usage included where estimated");
    expect(inspector(report)).toContain("Monthly usage provided; usage not included in this estimate");
    report.resources[0].cost.partial = true;
    report.resources[0].cost.usage_included = true;
    expect(inspector(report)).toContain("Partial coverage");
    expect(inspector(report)).toContain("Provided monthly usage included where estimated");
    expect(header(report)).not.toContain("Excludes storage");
  });
  it("warns that price can vary when usage is left out", () => {
    const report = priced();
    assert(report.pricing && report.resources[0]?.cost);
    report.resources[0].cost.missing_usage = ["processed_gb"];
    report.pricing.usage_omitted_resources = 3;
    const html = inspector(report);
    for (const text of ["Fixed resource charge only", "Price can vary with usage", "processed_gb", "usage file"]) {
      expect(html).toContain(text);
    }
    expect(header(report)).toContain("Resources with usage charges not included: 3");

    // Usage-only: nothing to show but the warning, and no claim of a fixed charge.
    report.resources[0].cost = { missing_usage: ["requests"] };
    expect(inspector(report)).toContain("requests");
    expect(inspector(report)).not.toContain("Fixed resource charge only");

    // Every modelled metric supplied: no warning.
    report.resources[0].cost = { before_monthly_cents: 1, after_monthly_cents: 2, usage_included: true };
    expect(inspector(report)).not.toContain("Price can vary");

    // Nothing modelled to supply, but usage still affects the real price.
    report.resources[0].cost = { before_monthly_cents: 1, after_monthly_cents: 2 };
    expect(inspector(report)).toContain("usage tfviz does not model");

    // A withheld estimate carries no warning.
    report.resources[0].cost = { reason: "Estimate withheld because pricing evidence is sensitive." };
    expect(inspector(report)).not.toContain("Price can vary");
  });

  it("formats integer cents exactly without Intl currency support or rounding", () => {
    expect(formatMonthly(12345, "ZZZ")).toBe("ZZZ 123.45");
    expect(formatMonthly(-1, "USD", true)).toBe("USD -0.01");
    expect(formatMonthly(1, "USD", true)).toBe("USD +0.01");
    expect(formatMonthly(0, "USD")).toBe("USD 0.00");
    expect(formatMonthly(9007199254740991, "USD")).toBe("USD 90071992547409.91");
    for (const cents of [undefined, NaN, Infinity, 1.5, 9007199254740992]) {
      expect(formatMonthly(cents, "USD")).toBe("Not estimated");
    }
    expect(formatMonthly(100, "usd")).toBe("Not estimated");
  });

  it("adds no pricing UI to unpriced reports, even with stray resource costs", () => {
    const report = priced();
    delete report.pricing;
    expect(header(report)).not.toContain("Estimated resource-cost subtotal");
    expect(header(report)).not.toContain("Cost period");
    expect(inspector(report)).not.toContain("Resource-cost estimate");
    expect(inspector(report)).not.toContain("ZZZ");
  });

  it("shows subtotal, coverage, both sides, signed delta and explicit assumptions", () => {
    const report = priced();
    const html = header(report);
    for (const text of ["Estimated resource-cost subtotal", "Before", "ZZZ 123.45", "After", "ZZZ 100.00",
      "Delta", "ZZZ -23.45", "Monthly", "Daily", "Weekly", "1 estimated", `${report.resources.length - 1} not estimated`,
      "ZZZ", "2026-11-01", "730 hours/month", "usage-based charges excluded", "Discounts", "taxes", "not a full bill", "not a calendar bill", "24 hours", "168 hours"]) {
      expect(html).toContain(text);
    }
  });

  it("shows unavailable amounts as Not estimated, never invented zeroes", () => {
    const report = priced();
    const resource = report.resources[0];
    assert(resource && report.pricing);
    report.pricing.summary = {};
    resource.cost = { reason: "Estimate withheld because pricing evidence is sensitive." };
    for (const html of [header(report), inspector(report)]) {
      expect(html).toContain("Not estimated");
      expect(html).not.toContain("ZZZ 0.00");
    }
    expect(inspector(report)).toContain("Estimate withheld because pricing evidence is sensitive.");
    delete resource.cost;
    expect(inspector(report)).toContain("Not estimated");
  });

  it("shows before/after regions and labels replacement steady-state and overlap", () => {
    const report = priced();
    const resource = report.resources[0];
    assert(resource?.cost);
    resource.change = { action: "replace", replace_order: "create_first", producer_actions: ["create", "delete"] };
    resource.cost.replacement_overlap = true;
    const html = inspector(report);
    for (const text of ["Before region", "us-east-1", "After region", "eu-west-1", "Steady-state", "Create-first overlap", "additional cost not estimated"]) {
      expect(html).toContain(text);
    }
    resource.change.replace_order = "delete_first";
    delete resource.cost.replacement_overlap;
    expect(inspector(report)).toContain("Steady-state");
    expect(inspector(report)).not.toContain("Create-first overlap");
  });
});
