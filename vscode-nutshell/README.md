# Nutshell — VS Code Extension

VS Code extension for editing, validating, and inspecting [Nutshell](https://github.com/ChatChatTech/nutshell) `.nut` bundles.

## Features

- **JSON Schema validation** — Auto-completion and inline validation for `nutshell.json`
- **Snippets** — Quick scaffolds for manifest sections (`nutshell-manifest`, `nut-task`, `nut-credentials`, etc.)
- **Commands** — Init, Pack, Unpack, Inspect, Validate, Check, Serve via command palette
- **Auto-validate on save** — Diagnostics panel shows errors when you save `nutshell.json`
- **Context menus** — Right-click `.nut` files to inspect/unpack, right-click `nutshell.json` to validate/pack
- **File icons** — Custom icon for `.nut` bundle files

## Requirements

- [Nutshell CLI](https://github.com/ChatChatTech/nutshell) installed and on `$PATH`
- VS Code 1.85+

## Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `nutshell.cliPath` | `"nutshell"` | Path to the nutshell CLI binary |
| `nutshell.autoValidate` | `true` | Auto-validate nutshell.json on save |

## Commands

| Command | Description |
|---------|-------------|
| `Nutshell: Init` | Create a new bundle scaffold in the workspace |
| `Nutshell: Pack` | Create .nut bundle from directory |
| `Nutshell: Unpack` | Extract .nut bundle to a directory |
| `Nutshell: Inspect` | View bundle manifest as JSON |
| `Nutshell: Validate` | Check bundle against spec |
| `Nutshell: Check Completeness` | Run completeness check |
| `Nutshell: Open Web Viewer` | Start local web viewer |

## Snippets

Type these prefixes in a `nutshell.json` file:

- `nutshell-manifest` — Full manifest scaffold
- `nut-task` — Task section
- `nut-credentials` — Credentials section
- `nut-acceptance` — Acceptance criteria section
- `nut-harness` — Harness hints section
- `nut-resources` — External resources section

## Development

```bash
cd vscode-nutshell
npm install
npm run compile
```

Press F5 in VS Code to launch the Extension Development Host.
