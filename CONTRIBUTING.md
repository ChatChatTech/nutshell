# Nutshell Development

## Project Structure

```
nutshell/
├── README.md                           # Project README
├── LICENSE                             # MIT License
├── spec/
│   └── nutshell-spec-v0.1.0.md        # Full specification
├── examples/
│   ├── 01-api-task/                    # Request bundle example
│   ├── 02-data-analysis/              # Another request example
│   └── 03-delivery/                    # Delivery bundle example
├── tools/
│   └── nutshell-cli.py                # Reference CLI implementation
└── docs/
    ├── harness-engineering-research.md # Research foundation
    └── clawnet-foundation.md          # ClawNet design summary
```

## CLI Usage

```bash
python3 tools/nutshell-cli.py init --dir my-task
python3 tools/nutshell-cli.py pack --dir my-task -o task.nut
python3 tools/nutshell-cli.py inspect task.nut
python3 tools/nutshell-cli.py validate task.nut
python3 tools/nutshell-cli.py unpack task.nut -o unpacked/
```

## Specification Version

Current: v0.1.0 (Draft)
