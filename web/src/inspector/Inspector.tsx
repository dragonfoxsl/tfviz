import { useState } from "react";
import type { Resource } from "../report/schema.gen";
import type { ReportIndex } from "../report/view";
import { ResourceIcon } from "../icons/ResourceIcon";
import { ChangeBadge } from "../graph/ChangeBadge";
import { describeReplaceOrder } from "../theme/changes";
import { MetadataTable } from "./MetadataTable";
import { RelationshipList } from "./RelationshipList";
import { ResourcePricing } from "../report/Pricing";

interface Props {
  index: ReportIndex;
  resource: Resource | undefined;
  onSelect: (id: string) => void;
}

export function Inspector({ index, resource, onSelect }: Props) {
  if (!resource) {
    return (
      <aside className="inspector" aria-label="Inspector">
        <p className="inspector-empty inspector-placeholder">Select a resource to see its details, changes and relationships.</p>
      </aside>
    );
  }
  const order = describeReplaceOrder(resource.change.replace_order);
  return (
    <aside className="inspector" aria-label={`Inspector: ${resource.label}`}>
      <header className="inspector-header">
        <span className="inspector-header__icon">
          <ResourceIcon resource={resource} size={20} />
        </span>
        <div>
          <h2>{resource.label}</h2>
          <p className="inspector-header__type">
            {resource.type}
            {resource.mode === "data" && " · data source"}
          </p>
        </div>
      </header>

      <div className="inspector-change">
        <ChangeBadge action={resource.change.action} />
        {order && <p>{order}</p>}
        {resource.change.replace_fields && resource.change.replace_fields.length > 0 && (
          <p>
            Replacement forced by <code>{resource.change.replace_fields.join(", ")}</code>.
          </p>
        )}
        {resource.support === "generic" && (
          <p className="inspector-note">This resource type is not supported yet, so tfviz shows limited detail and no placement.</p>
        )}
      </div>

      <ResourcePricing resource={resource} pricing={index.report.pricing} />

      <section aria-labelledby="inspector-identity">
        <h3 id="inspector-identity">Identity</h3>
        <dl className="field-list">
          <CopyRow label="Address" value={resource.address} />
          <CopyRow label="Module" value={resource.module || "Root module"} />
          <CopyRow label="Provider" value={resource.provider} />
        </dl>
      </section>

      <MetadataTable fields={resource.metadata} />

      <section aria-labelledby="inspector-relationships">
        <h3 id="inspector-relationships">Relationships</h3>
        <p className="inspector-hint">Recorded associations and the evidence for each. They do not show live traffic.</p>
        <RelationshipList index={index} resource={resource} onSelect={onSelect} />
      </section>
    </aside>
  );
}

function CopyRow({ label, value }: { label: string; value: string }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard?.writeText(value).then(
      () => {
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
      },
      () => setCopied(false),
    );
  };
  return (
    <div className="field-row">
      <dt>{label}</dt>
      <dd className="copy-row">
        <code>{value}</code>
        <button type="button" className="icon-button" onClick={copy} aria-label={`Copy ${label.toLowerCase()}`}>
          {copied ? "Copied" : "Copy"}
        </button>
      </dd>
    </div>
  );
}
