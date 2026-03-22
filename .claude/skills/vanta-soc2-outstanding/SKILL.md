---
name: vanta-soc2-outstanding
description: Report outstanding SOC2 compliance items grouped by owner
version: 1.0.0
---

# Vanta SOC2 Outstanding Items Sitrep

Generates `sitrep.md` — a report of all outstanding SOC2 compliance items grouped by owner.

## Steps

### 1. Fetch fresh data

Run each of these and save to `/tmp`:

```
vanta-cli tests --framework soc2 --status NEEDS_ATTENTION,IN_PROGRESS,INVALID  > /tmp/tests.json
vanta-cli documents --framework soc2 --status NEEDS_DOCUMENT,NEEDS_UPDATE > /tmp/documents.json
vanta-cli policies  > /tmp/policies.json
vanta-cli people    > /tmp/people.json
```

### 2. Filter items

**Tests** — include where:
- framework: SOC2 (all tests are SOC2)
- `status` is not `OK` and not `DEACTIVATED`

**Documents** — include where:
- framework: SOC2
- `uploadStatus` is `Needs document` or `Needs update`

**Policies** — include where:
- framework: SOC2
- `status` is `NEEDS_REMEDIATION`

### 3. Resolve owners

- Tests: use `test.owner.displayName`
- Documents: look up `document.ownerId` in people via `person.name.display`
- Policies: no owner field is available from `vanta-cli policies` — group under `Unassigned`

### 4. Resolve control associations

Tests have known control associations from the Vanta `/controls/{id}/tests` API. The current mappings are:

| Test ID | Controls |
| --- | --- |
| `aws-vpc-flow-logging-enabled` | MON-2 – Log management utilized |
| `logs-retained-for-twelve-months-config` | MON-2 – Log management utilized |
| `serverless-function-error-rate-monitored-aws` | MON-4 – Infrastructure performance monitored |
| `aws-storage-buckets-enforce-https` | NET-1 – Data transmission encrypted |
| `aws-security-group-restricted-default-group` | NET-3 – Network firewalls reviewed, NET-4 – Network firewalls utilized, NET-5 – Network and system hardening standards maintained |
| `risks-reviewed-annually` | RSK-1 – Risk assessment objectives specified, RSK-2 – Risk assessments performed |
| `high-severity-vendor-compliance-reports` | TPM-2 – Vendor management program established |
| `packages-checked-for-vulnerabilities-v2-records-closed-github-dependabot-high` | VPM-1 – Service infrastructure maintained, VPM-2 – Vulnerabilities scanned and remediated |
| `packages-checked-for-vulnerabilities-v2-records-closed-github-dependabot-medium` | VPM-1 – Service infrastructure maintained, VPM-2 – Vulnerabilities scanned and remediated |
| `packages-checked-for-vulnerabilities-v2-records-closed-github-dependabot-low` | VPM-1 – Service infrastructure maintained, VPM-2 – Vulnerabilities scanned and remediated |

Documents and policies: control associations are not exposed by `vanta-cli` — use `—` in the Controls column.

**Note:** New failing tests not in this table may appear over time. If a test ID is not in the table, check its control association via the Vanta API (`/controls/{id}/tests`) or mark Controls as `—` until confirmed.

### 5. Build links

- Tests: `https://app.vanta.com/tests/{id}`
- Documents: use the `url` field from the document object directly
- Policies: `https://app.vanta.com/policies/{id}`

### 6. Write sitrep.md

Output format: one `## Person Name` section per owner (alphabetical), each containing a markdown table:

```markdown
## Person Name

| Type | Name | Controls |
| --- | --- | --- |
| Test | [Test name (N items)](https://test_url) | [CTL-1 – Control name](https://control_url) |
| Document | [Document title (Needs document)](https://document_url) | — |
| Policy | [Policy name](https://policy_url) | — |
```

Write the result to `sitrep.md` in the project root.
