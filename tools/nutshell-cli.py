#!/usr/bin/env python3
"""
nutshell-cli: Reference implementation for packing and unpacking Nutshell bundles.

Usage:
    nutshell pack   --dir <path> [--output <file>] [--budget <tokens>]
    nutshell unpack <file> [--output <path>]
    nutshell inspect <file>
    nutshell validate <file>
    nutshell init [--dir <path>]
"""

import argparse
import json
import os
import sys
import tarfile
import io
import hashlib
import uuid
from datetime import datetime, timezone
from pathlib import Path

NUTSHELL_VERSION = "0.1.0"
MAGIC_BYTES = b"NUT\x01"

# ─── ANSI colors ───
CYAN = "\033[96m"
GREEN = "\033[92m"
YELLOW = "\033[93m"
RED = "\033[91m"
DIM = "\033[2m"
RESET = "\033[0m"
BOLD = "\033[1m"

SHELL_ART = r"""
    ___  ___  _  _  _____  ___  _  _  ___  _     _
   | _ \| _ \| || ||_   _|/ __|| || || __|| |   | |
   |  _/|   /| __ |  | |  \__ \| __ || _| | |__ | |__
   |_|  |_|_\|_||_|  |_|  |___/|_||_||___||____||____|
             🐚  n u t s h e l l  🦞
"""


def cmd_init(args):
    """Initialize a new nutshell bundle directory."""
    base = Path(args.dir or ".")
    dirs = ["context", "files/src", "files/data", "files/assets", "apis/schemas",
            "credentials", "tests/scripts", "delivery/artifacts", "delivery/logs"]
    for d in dirs:
        (base / d).mkdir(parents=True, exist_ok=True)

    manifest = {
        "nutshell_version": NUTSHELL_VERSION,
        "bundle_type": "request",
        "id": f"nut-{uuid.uuid4()}",
        "created_at": datetime.now(timezone.utc).isoformat(),
        "expires_at": None,
        "task": {
            "title": "",
            "summary": "",
            "priority": "medium",
            "estimated_effort": "",
            "reward": {"amount": 0, "currency": "energy"}
        },
        "tags": {
            "skills_required": [],
            "domains": [],
            "data_sources": [],
            "custom": {}
        },
        "publisher": {
            "peer_id": "",
            "agent_name": "",
            "reputation": 0
        },
        "context": {
            "requirements": "context/requirements.md",
            "additional": []
        },
        "files": {"total_count": 0, "total_size_bytes": 0, "tree": []},
        "apis": {},
        "credentials": {},
        "acceptance": {
            "checklist": [],
            "auto_verifiable": False,
            "human_review_required": True
        },
        "harness": {
            "agent_type_hint": "execution",
            "context_budget_hint": 0.35,
            "execution_strategy": "incremental",
            "checkpoints": True,
            "constraints": []
        },
        "resources": {"repos": [], "docs": [], "images": [], "links": []},
        "compression": {"algorithm": "zstd"}
    }

    manifest_path = base / "nutshell.json"
    if not manifest_path.exists():
        with open(manifest_path, "w") as f:
            json.dump(manifest, f, indent=2)
        print(f"{GREEN}✓{RESET} Initialized nutshell bundle at {base}/")
        print(f"  Edit {CYAN}nutshell.json{RESET} to configure your task.")
    else:
        print(f"{YELLOW}⚠{RESET} nutshell.json already exists at {base}/")

    # Create template requirements.md
    req_path = base / "context" / "requirements.md"
    if not req_path.exists():
        with open(req_path, "w") as f:
            f.write("# Requirements\n\n## Objective\n\n(Describe the task objective)\n\n"
                    "## Functional Requirements\n\n- FR-1: ...\n\n"
                    "## Non-Functional Requirements\n\n- NFR-1: ...\n")


def cmd_pack(args):
    """Pack a directory into a .nut bundle."""
    source = Path(args.dir or ".")
    manifest_path = source / "nutshell.json"

    if not manifest_path.exists():
        print(f"{RED}✗{RESET} No nutshell.json found in {source}")
        sys.exit(1)

    with open(manifest_path) as f:
        manifest = json.load(f)

    # Collect all files
    files = []
    total_size = 0
    for root, _, filenames in os.walk(source):
        for fn in filenames:
            fpath = Path(root) / fn
            rel = fpath.relative_to(source)
            size = fpath.stat().st_size
            files.append({"path": str(rel), "size": size, "abs": str(fpath)})
            total_size += size

    # Update manifest file counts
    manifest["files"]["total_count"] = len(files)
    manifest["files"]["total_size_bytes"] = total_size
    manifest["compression"]["original_size_bytes"] = total_size

    # Determine output path
    task_slug = manifest.get("task", {}).get("title", "bundle")
    task_slug = task_slug.lower().replace(" ", "-")[:40] if task_slug else "bundle"
    output = args.output or f"{task_slug}.nut"

    # Create tar archive with zstd-like compression (using gzip as fallback)
    buf = io.BytesIO()
    buf.write(MAGIC_BYTES)  # Magic header

    with tarfile.open(fileobj=buf, mode="w:gz") as tar:
        # Write manifest first (always)
        manifest_bytes = json.dumps(manifest, indent=2).encode()
        info = tarfile.TarInfo(name="nutshell.json")
        info.size = len(manifest_bytes)
        tar.addfile(info, io.BytesIO(manifest_bytes))

        # Write all other files
        for f in files:
            if f["path"] == "nutshell.json":
                continue
            tar.add(f["abs"], arcname=f["path"])

    compressed_size = buf.tell()
    manifest["compression"]["compressed_size_bytes"] = compressed_size

    with open(output, "wb") as f:
        f.write(buf.getvalue())

    ratio = (1 - compressed_size / total_size) * 100 if total_size > 0 else 0
    print(f"{GREEN}✓{RESET} Packed {CYAN}{len(files)}{RESET} files into {BOLD}{output}{RESET}")
    print(f"  {DIM}Original: {total_size:,} bytes → Compressed: {compressed_size:,} bytes ({ratio:.1f}% reduction){RESET}")
    print(f"  {DIM}Bundle ID: {manifest['id']}{RESET}")


def cmd_unpack(args):
    """Unpack a .nut bundle."""
    nut_path = Path(args.file)
    if not nut_path.exists():
        print(f"{RED}✗{RESET} File not found: {nut_path}")
        sys.exit(1)

    output = Path(args.output or nut_path.stem)
    output.mkdir(parents=True, exist_ok=True)

    with open(nut_path, "rb") as f:
        magic = f.read(4)
        if magic != MAGIC_BYTES:
            print(f"{RED}✗{RESET} Invalid nutshell bundle (bad magic bytes)")
            sys.exit(1)

        # Read remaining bytes into buffer for gzip decompression
        data = f.read()

    with tarfile.open(fileobj=io.BytesIO(data), mode="r:gz") as tar:
        # Security: prevent path traversal
        for member in tar.getmembers():
            member_path = Path(member.name)
            if member_path.is_absolute() or ".." in member_path.parts:
                print(f"{RED}✗{RESET} Unsafe path in bundle: {member.name}")
                sys.exit(1)
        tar.extractall(path=output, filter="data")

    # Read manifest for summary
    manifest_path = output / "nutshell.json"
    if manifest_path.exists():
        with open(manifest_path) as mf:
            manifest = json.load(mf)
        title = manifest.get("task", {}).get("title", "Unknown")
        btype = manifest.get("bundle_type", "unknown")
        print(f"{GREEN}✓{RESET} Unpacked to {BOLD}{output}/{RESET}")
        print(f"  {DIM}Task: {title}{RESET}")
        print(f"  {DIM}Type: {btype}{RESET}")


def cmd_inspect(args):
    """Inspect a .nut bundle without unpacking."""
    nut_path = Path(args.file)
    if not nut_path.exists():
        print(f"{RED}✗{RESET} File not found: {nut_path}")
        sys.exit(1)

    with open(nut_path, "rb") as f:
        magic = f.read(4)
        if magic != MAGIC_BYTES:
            print(f"{RED}✗{RESET} Invalid nutshell bundle")
            sys.exit(1)
        data = f.read()

    with tarfile.open(fileobj=io.BytesIO(data), mode="r:gz") as tar:
        # Find and read manifest
        manifest = None
        members = tar.getmembers()
        for m in members:
            if m.name == "nutshell.json":
                ef = tar.extractfile(m)
                if ef:
                    manifest = json.load(ef)
                break

    if not manifest:
        print(f"{RED}✗{RESET} No nutshell.json in bundle")
        sys.exit(1)

    print(SHELL_ART)
    print(f"  {BOLD}Bundle:{RESET} {nut_path.name}")
    print(f"  {BOLD}Version:{RESET} {manifest.get('nutshell_version', '?')}")
    print(f"  {BOLD}Type:{RESET} {manifest.get('bundle_type', '?')}")
    print(f"  {BOLD}ID:{RESET} {manifest.get('id', '?')}")
    print()

    task = manifest.get("task", {})
    print(f"  {CYAN}📋 Task:{RESET} {task.get('title', '?')}")
    print(f"  {DIM}{task.get('summary', '')}{RESET}")
    print(f"  Priority: {task.get('priority', '?')} | Effort: {task.get('estimated_effort', '?')}")
    reward = task.get("reward", {})
    print(f"  Reward: {reward.get('amount', 0)} {reward.get('currency', 'energy')}")
    print()

    tags = manifest.get("tags", {})
    skills = tags.get("skills_required", [])
    domains = tags.get("domains", [])
    print(f"  {CYAN}🏷️  Tags:{RESET} {', '.join(skills)}")
    if domains:
        print(f"  {DIM}Domains: {', '.join(domains)}{RESET}")
    print()

    creds = manifest.get("credentials", {})
    scopes = creds.get("scopes", [])
    if scopes:
        print(f"  {CYAN}🔑 Credentials:{RESET} {len(scopes)} scoped")
        for s in scopes:
            print(f"    • {s['name']} ({s['type']}) — {s.get('access_level', '?')}")
        print()

    files = manifest.get("files", {})
    print(f"  {CYAN}📦 Files:{RESET} {files.get('total_count', 0)} files, {files.get('total_size_bytes', 0):,} bytes")
    print(f"  {BOLD}Total entries in archive:{RESET} {len(members)}")

    comp = manifest.get("compression", {})
    if comp.get("context_tokens_estimate"):
        print(f"  {DIM}Est. tokens: ~{comp['context_tokens_estimate']:,}{RESET}")

    harness = manifest.get("harness", {})
    if harness:
        print(f"\n  {CYAN}⚙️  Harness Hints:{RESET}")
        print(f"    Agent type: {harness.get('agent_type_hint', '?')}")
        print(f"    Strategy: {harness.get('execution_strategy', '?')}")
        print(f"    Context budget: {harness.get('context_budget_hint', '?')}")
        constraints = harness.get("constraints", [])
        if constraints:
            print(f"    Constraints: {len(constraints)}")
            for c in constraints[:3]:
                print(f"      • {c}")


def cmd_validate(args):
    """Validate a .nut bundle or directory against the spec."""
    target = Path(args.file)
    errors = []
    warnings = []

    # Load manifest
    if target.is_dir():
        manifest_path = target / "nutshell.json"
    elif target.suffix == ".nut":
        # Extract manifest from bundle
        with open(target, "rb") as f:
            magic = f.read(4)
            if magic != MAGIC_BYTES:
                print(f"{RED}✗{RESET} Invalid nutshell bundle")
                sys.exit(1)
            data = f.read()
        with tarfile.open(fileobj=io.BytesIO(data), mode="r:gz") as tar:
            for m in tar.getmembers():
                if m.name == "nutshell.json":
                    ef = tar.extractfile(m)
                    if ef:
                        manifest = json.load(ef)
                    break
            else:
                print(f"{RED}✗{RESET} No nutshell.json in bundle")
                sys.exit(1)
        manifest_path = None
    else:
        manifest_path = target

    if manifest_path:
        if not manifest_path.exists():
            print(f"{RED}✗{RESET} No nutshell.json found at {manifest_path}")
            sys.exit(1)
        with open(manifest_path) as f:
            manifest = json.load(f)

    # Required top-level fields
    required = ["nutshell_version", "bundle_type", "id", "task"]
    for field in required:
        if field not in manifest:
            errors.append(f"Missing required field: {field}")

    # Version check
    ver = manifest.get("nutshell_version", "")
    if ver and not ver.startswith("0."):
        warnings.append(f"Unknown spec version: {ver}")

    # Bundle type
    btype = manifest.get("bundle_type", "")
    if btype not in ("request", "delivery"):
        errors.append(f"Invalid bundle_type: '{btype}' (must be 'request' or 'delivery')")

    # Task fields
    task = manifest.get("task", {})
    if btype == "request":
        if not task.get("title"):
            errors.append("task.title is required")
        if not task.get("summary"):
            warnings.append("task.summary is empty")

    # Tags
    tags = manifest.get("tags", {})
    if btype == "request" and not tags.get("skills_required"):
        warnings.append("No skills_required tags — matching will be broad")

    # Credentials security
    creds = manifest.get("credentials", {})
    if creds.get("encryption") == "none":
        warnings.append("Credentials are unencrypted — not recommended for production")
    for scope in creds.get("scopes", []):
        if not scope.get("expires_at"):
            warnings.append(f"Credential '{scope.get('name')}' has no expiration")

    # Harness hints
    harness = manifest.get("harness", {})
    budget = harness.get("context_budget_hint", 0)
    if budget > 0.5:
        warnings.append(f"Context budget hint {budget} exceeds recommended 0.4 (40% rule)")

    # Print results
    print(f"\n  {BOLD}Validating:{RESET} {target}\n")
    if errors:
        for e in errors:
            print(f"  {RED}✗ ERROR:{RESET} {e}")
    if warnings:
        for w in warnings:
            print(f"  {YELLOW}⚠ WARN:{RESET}  {w}")

    if not errors and not warnings:
        print(f"  {GREEN}✓ All checks passed{RESET}")
    elif not errors:
        print(f"\n  {GREEN}✓ Valid{RESET} with {len(warnings)} warning(s)")
    else:
        print(f"\n  {RED}✗ Invalid{RESET} — {len(errors)} error(s), {len(warnings)} warning(s)")
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(
        prog="nutshell",
        description="🐚 Nutshell — Task packaging for AI agents"
    )
    sub = parser.add_subparsers(dest="command")

    # init
    p_init = sub.add_parser("init", help="Initialize a new bundle directory")
    p_init.add_argument("--dir", help="Target directory")

    # pack
    p_pack = sub.add_parser("pack", help="Pack directory into .nut bundle")
    p_pack.add_argument("--dir", help="Source directory")
    p_pack.add_argument("--output", "-o", help="Output .nut file path")
    p_pack.add_argument("--budget", type=int, help="Target token budget")

    # unpack
    p_unpack = sub.add_parser("unpack", help="Unpack a .nut bundle")
    p_unpack.add_argument("file", help="Path to .nut file")
    p_unpack.add_argument("--output", "-o", help="Output directory")

    # inspect
    p_inspect = sub.add_parser("inspect", help="Inspect bundle without unpacking")
    p_inspect.add_argument("file", help="Path to .nut file")

    # validate
    p_validate = sub.add_parser("validate", help="Validate bundle against spec")
    p_validate.add_argument("file", help="Path to .nut file or directory")

    args = parser.parse_args()
    if not args.command:
        parser.print_help()
        sys.exit(1)

    cmds = {
        "init": cmd_init,
        "pack": cmd_pack,
        "unpack": cmd_unpack,
        "inspect": cmd_inspect,
        "validate": cmd_validate,
    }
    cmds[args.command](args)


if __name__ == "__main__":
    main()
