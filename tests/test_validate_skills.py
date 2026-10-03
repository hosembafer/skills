"""Exercise the validator against small, real repositories with broken inputs."""

import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


VALIDATOR = Path(__file__).resolve().parents[1] / "scripts" / "validate-skills.py"


class ValidationTests(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.root = Path(self.scratch.name)
        self.skill = self.root / "skills" / "example"
        self.write("skills/example/SKILL.md", "---\nname: example\ndescription: Handle example requests.\n---\n\nFollow the request.\n")
        self.write("skills/example/agents/openai.yaml", 'interface:\n  display_name: "Example"\n  short_description: "Handle a focused example request"\n  default_prompt: "Use $example to handle this request."\n')

    def write(self, path, text):
        target = self.root / path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(text, encoding="utf-8")
        return target

    def run_validator(self, expected, filename=None, env=None):
        result = subprocess.run(
            [sys.executable, str(VALIDATOR), str(self.root)],
            text=True, capture_output=True, env=env,
        )
        output = result.stdout + result.stderr
        self.assertEqual(result.returncode, expected, output)
        self.assertNotIn("Traceback (most recent call last)", output)
        if filename:
            self.assertIn(filename, output)
        return output

    def test_accepts_valid_repository_and_claude_extensions(self):
        manifest = self.skill / "SKILL.md"
        manifest.write_text(manifest.read_text().replace("description:", 'argument-hint: "[notes]"\nallowed-tools: Bash(git *)\ndisable-model-invocation: false\ndescription:'))
        self.run_validator(0)

    def test_rejects_missing_manifest(self):
        (self.skill / "SKILL.md").unlink()
        self.run_validator(1, "SKILL.md")

    def test_rejects_malformed_yaml(self):
        self.write("skills/example/SKILL.md", "---\nname: [example\n---\nBody.\n")
        self.run_validator(1, "SKILL.md")

    def test_rejects_duplicate_yaml_keys(self):
        self.write("skills/example/SKILL.md", "---\nname: example\nname: other\ndescription: Example.\n---\nBody.\n")
        self.run_validator(1, "SKILL.md")

    def test_rejects_directory_name_mismatch(self):
        manifest = self.skill / "SKILL.md"
        manifest.write_text(manifest.read_text().replace("name: example", "name: other"))
        self.run_validator(1, "SKILL.md")

    def test_rejects_missing_reference_in_supporting_markdown(self):
        self.write("skills/example/references/guide.md", "Read [the missing details](missing.md).\n")
        self.run_validator(1, "references/guide.md")

    def test_rejects_missing_script_mentioned_in_instructions(self):
        with (self.skill / "SKILL.md").open("a") as manifest:
            manifest.write("Run `bash scripts/missing.sh`.\n")
        self.run_validator(1, "scripts/missing.sh")

    def test_rejects_resource_that_escapes_the_skill_bundle(self):
        self.write("outside.md", "This will not be installed with the skill.\n")
        with (self.skill / "SKILL.md").open("a") as manifest:
            manifest.write("Read [outside](../../outside.md).\n")
        self.run_validator(1, "outside.md")

    def test_rejects_missing_agent_metadata(self):
        (self.skill / "agents" / "openai.yaml").unlink()
        self.run_validator(1, "openai.yaml")

    def test_rejects_wrong_default_prompt(self):
        self.write("skills/example/agents/openai.yaml", 'interface:\n  display_name: "Example"\n  short_description: "Handle a focused example request"\n  default_prompt: "Use $other to handle this request."\n')
        self.run_validator(1, "openai.yaml")

    def test_reports_malformed_reference_without_traceback(self):
        self.write("README.md", "See [guide](https://[bad).\n")
        self.run_validator(1, "README.md")

    def test_rejects_invalid_icon_paths(self):
        icon = self.write("skills/example/assets/icon.svg", "<svg/>\n")
        metadata = self.skill / "agents" / "openai.yaml"
        original = metadata.read_text()
        for destination in ("//assets/missing.svg", "#missing", str(icon), "http://[bad", "assets"):
            with self.subTest(destination=destination):
                metadata.write_text(original + f'  icon_small: "{destination}"\n')
                self.run_validator(1, "openai.yaml")

    def test_accepts_existing_relative_icon(self):
        self.write("skills/example/assets/icon.svg", "<svg/>\n")
        metadata = self.skill / "agents" / "openai.yaml"
        metadata.write_text(metadata.read_text() + '  icon_small: "./assets/icon.svg"\n')
        self.run_validator(0)

    def test_rejects_broken_markdown_link_forms(self):
        for link in ('[guide](missing.md "Guide")', '[guide](<missing file.md>)', '[guide][details]\n\n[details]: missing.md'):
            with self.subTest(link=link):
                self.write("skills/example/references/guide.md", link + "\n")
                self.run_validator(1, "references/guide.md")

    def test_accepts_existing_markdown_links_with_spaces_and_titles(self):
        self.write("skills/example/references/details (advanced).md", "Details.\n")
        self.write("skills/example/references/guide.md", '[guide](<details (advanced).md> "Guide")\n\n[guide][details]\n\n[details]: <details (advanced).md>\n')
        self.run_validator(0)

    def test_rejects_inconsistent_manual_invocation_policy(self):
        manifest = self.skill / "SKILL.md"
        manifest.write_text(manifest.read_text().replace("description:", "disable-model-invocation: true\ndescription:"))
        self.run_validator(1, "allow_implicit_invocation")

    def test_rejects_manual_codex_policy_with_implicit_claude_policy(self):
        metadata = self.skill / "agents" / "openai.yaml"
        metadata.write_text(metadata.read_text() + "policy:\n  allow_implicit_invocation: false\n")
        self.run_validator(1, "disable-model-invocation")

    def test_rejects_invalid_python_without_executing_it(self):
        self.write("skills/example/scripts/helper.py", "print('must not execute')\nif\n")
        self.run_validator(1, "helper.py")

    def test_rejects_invalid_bash_without_executing_it(self):
        self.write("skills/example/scripts/helper.sh", "#!/usr/bin/env bash\nif then\n")
        self.run_validator(1, "helper.sh")

    def test_rejects_shellcheck_warning(self):
        self.write("skills/example/scripts/helper.sh", "#!/usr/bin/env bash\ncd /tmp\n")
        output = self.run_validator(1, "helper.sh")
        self.assertIn("SC2164", output)

    def test_reports_missing_shellcheck(self):
        env = dict(os.environ, PATH="")
        self.run_validator(1, "ShellCheck", env=env)

    def test_rejects_broken_repository_documentation_link(self):
        self.write("README.md", "See [usage examples](docs/missing.md).\n")
        self.run_validator(1, "docs/missing.md")


if __name__ == "__main__":
    unittest.main()
