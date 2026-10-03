#!/usr/bin/env python3
"""Validate this repository's skill bundles and syntax without running helpers."""

import argparse
import ast
from pathlib import Path
import re
import shutil
import subprocess
import sys
from urllib.parse import unquote, urlsplit

try:
    import yaml
except ImportError:
    sys.exit("Install validation dependencies with: python3 -m pip install -r requirements-dev.txt")


class UniqueKeyLoader(yaml.SafeLoader):
    def construct_mapping(self, node, deep=False):
        mapping = super().construct_mapping(node, deep=deep)
        seen = set()
        for key_node, _ in node.value:
            key = self.construct_object(key_node, deep=deep)
            if key in seen:
                raise yaml.constructor.ConstructorError(
                    None, None, f"duplicate key: {key}", key_node.start_mark,
                )
            seen.add(key)
        return mapping


def read_text(path, errors):
    try:
        return path.read_text(encoding="utf-8")
    except (OSError, UnicodeError) as error:
        errors.append(f"{path}: {error}")
        return ""


def load_mapping(text, path, errors):
    try:
        value = yaml.load(text, Loader=UniqueKeyLoader)
    except yaml.YAMLError as error:
        errors.append(f"{path}: invalid YAML: {error}")
        return {}
    if not isinstance(value, dict):
        errors.append(f"{path}: YAML must be a mapping")
        return {}
    return value


def check_resource(source, destination, base, boundary, errors, asset=False):
    try:
        parsed = urlsplit(destination.strip("<>"))
    except ValueError as error:
        errors.append(f"{source}: invalid resource {destination}: {error}")
        return
    if asset and (parsed.scheme or parsed.netloc or not parsed.path or Path(parsed.path).is_absolute()):
        errors.append(f"{source}: icon must be a nonempty relative asset path: {destination}")
        return
    if parsed.scheme or parsed.netloc or not parsed.path:
        return
    target = (base / unquote(parsed.path)).resolve()
    if not target.is_relative_to(boundary.resolve()):
        errors.append(f"{source}: resource leaves its bundle: {destination}")
    elif not target.exists() or (asset and not target.is_file()):
        errors.append(f"{source}: missing local resource: {destination}")


def check_markdown(path, boundary, errors, skill=None):
    text = read_text(path, errors)
    # Destinations can use angle brackets for spaces, optional titles, or reference definitions.
    destinations = re.findall(r"\[[^\]]*\]\(\s*(<[^>\n]+>|[^\s)]+)", text)
    destinations += re.findall(r"^[ \t]{0,3}\[[^\]]+\]:[ \t]*(<[^>\n]+>|[^\s]+)", text, re.MULTILINE)
    for destination in set(destinations):
        check_resource(path, destination, path.parent, boundary, errors)
    if skill:
        resources = re.findall(r"(?<![\w/])(?:references|scripts|assets)/[\w./-]+\.[\w-]+", text)
        for resource in set(resources):
            check_resource(path, resource, skill, boundary, errors)


def check_metadata(skill, manifest, errors):
    path = skill / "agents" / "openai.yaml"
    data = load_mapping(read_text(path, errors), path, errors)
    interface = data.get("interface")
    if not isinstance(interface, dict):
        errors.append(f"{path}: interface must be a mapping")
        interface = {}
    for key in ("display_name", "short_description", "default_prompt"):
        value = interface.get(key)
        if not isinstance(value, str) or not value.strip():
            errors.append(f"{path}: interface.{key} must be a nonempty string")
    description = interface.get("short_description")
    if isinstance(description, str) and not 25 <= len(description) <= 64:
        errors.append(f"{path}: short_description must be 25-64 characters")
    prompt = interface.get("default_prompt")
    if isinstance(prompt, str) and not re.search(rf"\${re.escape(skill.name)}(?![\w-])", prompt):
        errors.append(f"{path}: default_prompt must mention ${skill.name}")
    for key in ("icon_small", "icon_large"):
        if key in interface:
            value = interface[key]
            if not isinstance(value, str) or not value.strip():
                errors.append(f"{path}: {key} must be a relative asset path")
            else:
                check_resource(path, value, skill, skill, errors, asset=True)
    if "brand_color" in interface and not re.fullmatch(r"#[0-9a-fA-F]{6}", str(interface["brand_color"])):
        errors.append(f"{path}: brand_color must be a six-digit hex color")
    policy = data.get("policy", {})
    if not isinstance(policy, dict):
        errors.append(f"{path}: policy must be a mapping")
        policy = {}
    invocation = policy.get("allow_implicit_invocation", True)
    if not isinstance(invocation, bool):
        errors.append(f"{path}: allow_implicit_invocation must be a boolean")
    if manifest.get("disable-model-invocation") is True and invocation is not False:
        errors.append(f"{path}: allow_implicit_invocation must be false for an explicit-only skill")
    if invocation is False and manifest.get("disable-model-invocation") is not True:
        errors.append(f"{skill / 'SKILL.md'}: disable-model-invocation must be true when Codex is explicit-only")


def check_skill(skill, names, errors):
    path = skill / "SKILL.md"
    text = read_text(path, errors)
    match = re.match(r"\A---\r?\n(.*?)\r?\n---(?:\r?\n|\Z)(.*)\Z", text, re.DOTALL)
    if not match:
        errors.append(f"{path}: expected opening and closing YAML front matter delimiters")
        return
    manifest = load_mapping(match[1], path, errors)
    name = manifest.get("name")
    if not isinstance(name, str) or len(name) > 64 or not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", name):
        errors.append(f"{path}: name must be 1-64 lowercase letters, digits, or single hyphens")
    else:
        if name != skill.name:
            errors.append(f"{path}: name must match folder {skill.name}")
        if name in names:
            errors.append(f"{path}: duplicate skill name: {name}")
        names.add(name)
    description = manifest.get("description")
    if not isinstance(description, str) or not description.strip() or len(description) > 1024:
        errors.append(f"{path}: description must be a nonempty string of at most 1024 characters")
    if not match[2].strip():
        errors.append(f"{path}: instructions must not be empty")
    if "disable-model-invocation" in manifest and not isinstance(manifest["disable-model-invocation"], bool):
        errors.append(f"{path}: disable-model-invocation must be a boolean")
    check_metadata(skill, manifest, errors)
    for markdown in sorted(skill.rglob("*.md")):
        if markdown.name.lower() == "skill.md" and markdown != path:
            errors.append(f"{markdown}: each bundle must have only one SKILL.md, at its root")
        check_markdown(markdown, skill, errors, skill)


def check_syntax(root, errors):
    shellcheck = shutil.which("shellcheck")
    bash = shutil.which("bash")
    if not shellcheck:
        errors.append("ShellCheck is missing. Activate the environment installed from requirements-dev.txt.")
    if not bash:
        errors.append("Bash is missing. Install Bash and make it available on PATH.")
    directories = [root / "skills", root / "scripts", root / "tests"]
    for directory in directories:
        for path in sorted(directory.rglob("*.py")):
            try:
                ast.parse(read_text(path, errors), filename=str(path))
            except SyntaxError as error:
                errors.append(f"{path}:{error.lineno}: Python syntax: {error.msg}")
        shell_files = sorted(directory.rglob("*.sh"))
        if bash:
            for path in shell_files:
                result = subprocess.run([bash, "-n", str(path)], capture_output=True, text=True)
                if result.returncode:
                    errors.append(f"{path}: Bash syntax: {result.stderr.strip()}")
        if shellcheck and shell_files:
            result = subprocess.run(
                [shellcheck, "--severity=warning", *map(str, shell_files)],
                capture_output=True, text=True,
            )
            if result.returncode:
                errors.append(result.stdout.strip() or result.stderr.strip())


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("root", nargs="?", type=Path, default=Path(__file__).resolve().parents[1], help="repository to validate (default: this checkout)")
    root = parser.parse_args().root.resolve()
    errors = []
    skills = sorted(path for path in (root / "skills").glob("*") if path.is_dir())
    if not skills:
        errors.append(f"{root / 'skills'}: no skill folders found")
    names = set()
    for skill in skills:
        check_skill(skill, names, errors)
    for markdown in sorted(root.glob("*.md")) + sorted((root / "docs").rglob("*.md")):
        check_markdown(markdown, root, errors)
    check_syntax(root, errors)
    if errors:
        for error in errors:
            print(f"error: {error}", file=sys.stderr)
        return 1
    print(f"Validated {len(skills)} skills, agent metadata, local references, and script syntax.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
