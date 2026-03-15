# vanta-cli

A command-line tool for fetching data from the [Vanta API](https://developer.vanta.com/).

## Usage

### Authentication

vanta-cli authenticates using OAuth client credentials.

Follow the Vanta docs on [how to obtain these credentials](https://developer.vanta.com/docs/api-access-setup#creating-a-new-application).

Set the following environment variables before running any command:

```sh
export VANTA_CLIENT_ID=your-client-id
export VANTA_CLIENT_SECRET=your-client-secret
```

### Commands

Each resource is a subcommand.
All subcommands support a `--format` flag with the following options: `json` (default), `csv`, `tsv`, `markdown`.

```
vanta-cli [command] --format [json|csv|tsv|markdown]
```

Available commands:

| Command | Description |
| --- | --- |
| `controls` | Fetch controls |
| `discovered-vendors` | Fetch discovered vendors |
| `documents` | Fetch documents |
| `frameworks` | Fetch frameworks |
| `groups` | Fetch groups |
| `integrations` | Fetch integrations |
| `monitored-computers` | Fetch monitored computers |
| `people` | Fetch people |
| `policies` | Fetch policies |
| `risk-scenarios` | Fetch risk scenarios |
| `schema` | Emit machine-readable schema of subcommands and data models |
| `tests` | Fetch tests |
| `vendor-risk-attributes` | Fetch vendor risk attributes |
| `vendors` | Fetch vendors |
| `vulnerabilities` | Fetch vulnerabilities |
| `vulnerability-remediations` | Fetch vulnerability remediations |
| `vulnerable-assets` | Fetch vulnerable assets |

Pagination is handled automatically — all pages are fetched and concatenated before output.

### Examples

Fetch all vulnerabilities as JSON:

```sh
vanta-cli vulnerabilities
```

Fetch all vendors as CSV:

```sh
vanta-cli vendors --format csv
```

Count vulnerabilities using `jq`:

```sh
vanta-cli vulnerabilities | jq 'length'
```

Fetch all policies as a markdown table:

```sh
vanta-cli policies --format markdown
```

Emit the machine-readable schema (useful for LLM consumption):

```sh
vanta-cli schema
```

## Development

### Prerequisites

- [mise](https://mise.jdx.dev/) — manages the Go toolchain and task runner

Install dependencies:

```sh
mise install
```

### Building

```sh
mise run build
```

This produces a `./vanta-cli` binary in the current directory.

### Testing

```sh
mise run test
```

This runs three checks in sequence:

- **lint** — `golint` with exit-on-warnings
- **fmt** — `go fmt` (fails the pipeline if formatting changes are needed)
- **unit** — `go test ./...`
