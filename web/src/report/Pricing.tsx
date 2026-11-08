import { useContext } from "react";
import { CostPeriodContext, type CostPeriod } from "../app/reportContext";
import type { Report, Resource } from "./schema.gen";
import "./pricing.css";

type Pricing = NonNullable<Report["pricing"]>;
type Estimate = NonNullable<Resource["cost"]>;

/** Round absolute cents half-up, without floating-point intermediates. */
export function periodCents(cents: number | undefined, period: CostPeriod): bigint | undefined {
  if (cents === undefined || !Number.isSafeInteger(cents)) return undefined;
  const amount = BigInt(cents);
  if (period === "monthly") return amount;
  const magnitude = amount < 0n ? -amount : amount;
  const rounded = (magnitude * (period === "daily" ? 24n : 168n) + 365n) / 730n;
  return amount < 0n ? -rounded : rounded;
}

function formatAmount(amount: bigint | undefined, currency: string, delta = false): string {
  if (amount === undefined || !/^[A-Z]{3}$/.test(currency)) return "Not estimated";
  const magnitude = amount < 0n ? -amount : amount;
  const sign = amount < 0n ? "-" : delta && amount > 0n ? "+" : "";
  return `${currency} ${sign}${magnitude / 100n}.${String(magnitude % 100n).padStart(2, "0")}`;
}

export function formatMonthly(cents: number | undefined, currency: string, delta = false): string {
  return formatAmount(periodCents(cents, "monthly"), currency, delta);
}

export function periodAmounts(cost: Estimate | undefined, period: CostPeriod) {
  const before = periodCents(cost?.before_monthly_cents, period);
  const after = periodCents(cost?.after_monthly_cents, period);
  const delta = before !== undefined && after !== undefined ? after - before : periodCents(cost?.delta_monthly_cents, period);
  return { before, after, delta };
}

function PeriodSelect() {
  const { period, setPeriod } = useContext(CostPeriodContext);
  return (
    <label className="select-control">Cost period{" "}
      <select value={period} onChange={(event) => setPeriod(event.target.value as CostPeriod)}>
        <option value="daily">Daily</option>
        <option value="weekly">Weekly</option>
        <option value="monthly">Monthly</option>
      </select>
    </label>
  );
}

function Amounts({ cost, currency }: { cost: Estimate | undefined; currency: string }) {
  const { period } = useContext(CostPeriodContext);
  const { before, after, delta } = periodAmounts(cost, period);
  return (
    <dl className="pricing-amounts">
      <div><dt>Before</dt><dd>{formatAmount(before, currency)} <span aria-hidden="true">→</span></dd></div>
      <div><dt>After</dt><dd>{formatAmount(after, currency)}</dd></div>
      <div><dt>Delta</dt><dd>{formatAmount(delta, currency, true)}</dd></div>
    </dl>
  );
}

function UsageWarning({ cost }: { cost: Estimate | undefined }) {
  const missing = cost?.missing_usage ?? [];
  const before = cost?.before_monthly_cents;
  const after = cost?.after_monthly_cents;
  const comparable = before !== undefined && after !== undefined;
  if (missing.length > 0) {
    return <p className="pricing-warning">
      {comparable && !cost?.usage_included ? "Fixed resource charge only. " : ""}
      Price can vary with usage. Not included: <code>{missing.join(", ")}</code>. Add to a usage file to include.
    </p>;
  }
  if (comparable && !cost?.usage_included) {
    return <p className="pricing-warning">Price can vary with usage tfviz does not model, such as data transfer.</p>;
  }
  return null;
}

function Assumptions({ cost, pricing }: { cost: Estimate | undefined; pricing: Pricing }) {
  return <>
    <p>Normalized average cost, not a calendar bill. Monthly basis: 730 hours; daily: 24 hours; weekly: 168 hours.</p>
    {cost?.partial && <p>Partial coverage: only some resource costs are estimated.</p>}
    <p>{cost?.usage_included ? "Provided monthly usage included where estimated; unmodeled usage excluded."
      : pricing.usage_provided ? "Monthly usage provided; usage not included in this estimate."
      : "No monthly usage provided; usage-based charges excluded."} Discounts and taxes excluded.</p>
  </>;
}

export function PricingSummary({ pricing }: { pricing: Report["pricing"] }) {
  if (!pricing) return null;
  return (
    <section className="pricing-summary" aria-label="Estimated resource-cost subtotal">
      <strong>Estimated resource-cost subtotal</strong>{" "}<PeriodSelect />
      <Amounts cost={pricing.summary} currency={pricing.currency} />
      {pricing.usage_omitted_resources !== undefined && pricing.usage_omitted_resources > 0 &&
        <p className="pricing-warning">Price can vary with usage. Resources with usage charges not included: {pricing.usage_omitted_resources}.</p>}
      <p>{pricing.estimated_resources} estimated · {pricing.unestimated_resources} not estimated resources</p>
      <p>{pricing.currency} · rates as of {pricing.as_of} · {pricing.hours_per_month} hours/month</p>
      <p>Steady-state subtotal, not a full bill. Unsupported resource costs excluded.</p>
      <Assumptions cost={pricing.summary} pricing={pricing} />
    </section>
  );
}

export function ResourcePricing({ resource, pricing }: { resource: Resource; pricing: Pricing | undefined }) {
  if (!pricing) return null;
  const cost = resource.cost;
  const replacement = resource.change.action === "replace";
  const overlap = cost?.replacement_overlap || (replacement && resource.change.replace_order === "create_first");
  return (
    <section aria-labelledby="inspector-pricing">
      <h3 id="inspector-pricing">Resource-cost estimate</h3>
      <PeriodSelect />
      <p className="inspector-hint">{replacement ? "Steady-state · " : ""}{pricing.currency} · {pricing.hours_per_month} hours/month · rates as of {pricing.as_of}</p>
      <Amounts cost={cost} currency={pricing.currency} />
      <UsageWarning cost={cost} />
      <Assumptions cost={cost} pricing={pricing} />
      <dl className="field-list">
        <div className="field-row"><dt>Before region</dt><dd>{cost?.before_region ?? "Not recorded"}</dd></div>
        <div className="field-row"><dt>After region</dt><dd>{cost?.after_region ?? "Not recorded"}</dd></div>
      </dl>
      {cost?.reason && <p className="inspector-note">{cost.reason}</p>}
      {overlap && <p className="inspector-note">Create-first overlap: additional cost not estimated.</p>}
    </section>
  );
}
