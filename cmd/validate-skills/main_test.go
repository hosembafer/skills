package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validManifest = "---\nname: example\ndescription: Handle example requests.\n---\n\nFollow the request.\n"
const validMetadata = "interface:\n  display_name: Example\n  short_description: Handle a focused example request\n  default_prompt: Use $example to handle this request.\n"

func writeFixture(t *testing.T, root, path, content string) {
	t.Helper()
	target := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFixture(t, root, "skills/example/SKILL.md", validManifest)
	writeFixture(t, root, "skills/example/agents/openai.yaml", validMetadata)
	return root
}

func checkResult(t *testing.T, root string, want int, contains string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	got := run([]string{root}, &stdout, &stderr)
	output := stdout.String() + stderr.String()
	if got != want || !strings.Contains(output, contains) {
		t.Fatalf("exit = %d, want %d; output must contain %q:\n%s", got, want, contains, output)
	}
	if strings.Contains(output, "panic:") {
		t.Fatalf("validator panicked: %s", output)
	}
}

func TestValidRepository(t *testing.T) {
	root := fixture(t)
	writeFixture(t, root, "skills/example/SKILL.md", strings.Replace(validManifest, "description:", "argument-hint: \"[notes]\"\nallowed-tools: Bash(git *)\ndisable-model-invocation: false\ndescription:", 1))
	checkResult(t, root, 0, "Validated 1 skills")
}

// Each case catches an omitted validation rule using an independently broken bundle.
func TestBrokenRepository(t *testing.T) {
	cases := []struct {
		name, path, content, diagnostic string
	}{
		{"missing_manifest", "skills/example/SKILL.md", "", "SKILL.md"},
		{"missing_metadata", "skills/example/agents/openai.yaml", "", "openai.yaml"},
		{"malformed_yaml", "skills/example/SKILL.md", "---\nname: [example\n---\nBody.\n", "invalid YAML"},
		{"duplicate_yaml_keys", "skills/example/SKILL.md", "---\nname: example\nname: other\ndescription: Example.\n---\nBody.\n", "invalid YAML"},
		{"multiple_yaml_documents", "skills/example/agents/openai.yaml", validMetadata + "---\ninterface: {}\n", "one YAML document"},
		{"non_mapping_yaml", "skills/example/SKILL.md", "---\n- example\n---\nBody.\n", "mapping"},
		{"folder_name_mismatch", "skills/example/SKILL.md", strings.Replace(validManifest, "name: example", "name: other", 1), "name must match"},
		{"invalid_skill_name", "skills/example/SKILL.md", strings.Replace(validManifest, "name: example", "name: Example", 1), "name must be"},
		{"empty_description", "skills/example/SKILL.md", strings.Replace(validManifest, "Handle example requests.", "''", 1), "description"},
		{"long_description", "skills/example/SKILL.md", strings.Replace(validManifest, "Handle example requests.", strings.Repeat("x", 1025), 1), "description"},
		{"empty_instructions", "skills/example/SKILL.md", "---\nname: example\ndescription: Example.\n---\n", "instructions"},
		{"invalid_claude_policy", "skills/example/SKILL.md", strings.Replace(validManifest, "description:", "disable-model-invocation: 'true'\ndescription:", 1), "boolean"},
		{"invalid_interface", "skills/example/agents/openai.yaml", "interface: []\n", "interface"},
		{"empty_display_name", "skills/example/agents/openai.yaml", strings.Replace(validMetadata, "display_name: Example", "display_name: ''", 1), "display_name"},
		{"short_ui_description", "skills/example/agents/openai.yaml", strings.Replace(validMetadata, "Handle a focused example request", "Example", 1), "25-64"},
		{"wrong_default_prompt", "skills/example/agents/openai.yaml", strings.Replace(validMetadata, "$example", "$other", 1), "default_prompt"},
		{"partial_default_prompt", "skills/example/agents/openai.yaml", strings.Replace(validMetadata, "$example", "$example-other", 1), "default_prompt"},
		{"invalid_brand_color", "skills/example/agents/openai.yaml", validMetadata + "  brand_color: red\n", "brand_color"},
		{"invalid_codex_policy", "skills/example/agents/openai.yaml", validMetadata + "policy:\n  allow_implicit_invocation: 'false'\n", "boolean"},
		{"claude_only_explicit_policy", "skills/example/SKILL.md", strings.Replace(validManifest, "description:", "disable-model-invocation: true\ndescription:", 1), "allow_implicit_invocation"},
		{"codex_only_explicit_policy", "skills/example/agents/openai.yaml", validMetadata + "policy:\n  allow_implicit_invocation: false\n", "disable-model-invocation"},
		{"nested_manifest", "skills/example/nested/SKILL.md", validManifest, "only one SKILL.md"},
		{"missing_supporting_reference", "skills/example/references/guide.md", "Read [details](missing.md).\n", "references/guide.md"},
		{"missing_script", "skills/example/SKILL.md", validManifest + "Run `bash scripts/missing.sh`.\n", "scripts/missing.sh"},
		{"escaping_reference", "skills/example/references/guide.md", "Read [outside](../../../README.md).\n", "leaves its bundle"},
		{"broken_documentation_link", "README.md", "See [usage](docs/missing.md).\n", "docs/missing.md"},
		{"malformed_url", "README.md", "See [guide](https://[bad).\n", "invalid resource"},
		{"invalid_bash", "skills/example/scripts/helper.sh", "#!/usr/bin/env bash\nif then\n", "Bash syntax"},
		{"shellcheck_warning", "skills/example/scripts/helper.sh", "#!/usr/bin/env bash\ncd /tmp\n", "SC2164"},
		{"invalid_go", "scripts/helper.go", "package main\nfunc {\n", "Go syntax"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := fixture(t)
			if tc.content == "" {
				if err := os.Remove(filepath.Join(root, filepath.FromSlash(tc.path))); err != nil {
					t.Fatal(err)
				}
			} else {
				writeFixture(t, root, tc.path, tc.content)
			}
			writeFixture(t, root, "README.md", "Outside the example bundle.\n")
			if tc.path == "README.md" {
				writeFixture(t, root, tc.path, tc.content)
			}
			checkResult(t, root, 1, tc.diagnostic)
		})
	}
}

func TestIconPaths(t *testing.T) {
	for _, icon := range []string{"//assets/missing.svg", "#missing", "/tmp/icon.svg", "http://[bad", "assets", "https://example.com/icon.svg"} {
		t.Run(icon, func(t *testing.T) {
			root := fixture(t)
			writeFixture(t, root, "skills/example/assets/icon.svg", "<svg/>\n")
			writeFixture(t, root, "skills/example/agents/openai.yaml", validMetadata+"  icon_small: '"+icon+"'\n")
			checkResult(t, root, 1, "openai.yaml")
		})
	}
	t.Run("existing_relative_icon", func(t *testing.T) {
		root := fixture(t)
		writeFixture(t, root, "skills/example/assets/icon.svg", "<svg/>\n")
		writeFixture(t, root, "skills/example/agents/openai.yaml", validMetadata+"  icon_small: ./assets/icon.svg\n")
		checkResult(t, root, 0, "Validated 1 skills")
	})
}

func TestMarkdownLinks(t *testing.T) {
	for _, link := range []string{"[guide](missing.md \"Guide\")", "[guide](<missing file.md>)", "[guide][details]\n\n[details]: missing.md"} {
		t.Run(link, func(t *testing.T) {
			root := fixture(t)
			writeFixture(t, root, "skills/example/references/guide.md", link+"\n")
			checkResult(t, root, 1, "missing local resource")
		})
	}
	t.Run("existing_links", func(t *testing.T) {
		root := fixture(t)
		writeFixture(t, root, "skills/example/references/details (advanced).md", "Details.\n")
		writeFixture(t, root, "skills/example/references/guide.md", "[guide](<details (advanced).md> \"Guide\")\n[guide][details]\n\n[details]: <details (advanced).md>\n[encoded](details%20%28advanced%29.md)\n[external](https://example.com/guide)\n")
		checkResult(t, root, 0, "Validated 1 skills")
	})
}

func TestMissingTools(t *testing.T) {
	root := fixture(t)
	t.Setenv("PATH", t.TempDir())
	checkResult(t, root, 1, "ShellCheck is missing")
	checkResult(t, root, 1, "Bash is missing")
}

func TestHelperIsNotExecuted(t *testing.T) {
	root := fixture(t)
	writeFixture(t, root, "skills/example/scripts/helper.sh", "#!/usr/bin/env bash\nexit 19\n")
	checkResult(t, root, 0, "Validated 1 skills")
}

func TestMissingSkills(t *testing.T) {
	checkResult(t, t.TempDir(), 1, "no skill folders")
}

func TestUnicodeResourceNames(t *testing.T) {
	root := fixture(t)
	writeFixture(t, root, "skills/example/SKILL.md", validManifest+"Read assets/画.png.\n")
	checkResult(t, root, 1, "assets/画.png")
	writeFixture(t, root, "skills/example/assets/画.png", "Example image.\n")
	checkResult(t, root, 0, "Validated 1 skills")
}

func TestSymlinkResources(t *testing.T) {
	for _, exists := range []bool{true, false} {
		name := "missing_outside_resource"
		if exists {
			name = "existing_outside_resource"
		}
		t.Run(name, func(t *testing.T) {
			root := fixture(t)
			if err := os.MkdirAll(filepath.Join(root, "outside", "nested"), 0o755); err != nil {
				t.Fatal(err)
			}
			writeFixture(t, root, "skills/example/references/outside.md", "Inside decoy.\n")
			if exists {
				writeFixture(t, root, "outside/outside.md", "Outside the installed bundle.\n")
			}
			if err := os.Symlink(filepath.Join(root, "outside", "nested"), filepath.Join(root, "skills", "example", "references", "link")); err != nil {
				t.Fatal(err)
			}
			writeFixture(t, root, "skills/example/SKILL.md", validManifest+"Read [outside](references/link/../outside.md).\n")
			checkResult(t, root, 1, "leaves its bundle")
		})
	}
}

func TestCLIUsage(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"--unknown"}, {"one", "two"}} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			want := 2
			if args[0] == "--help" {
				want = 0
			}
			if got := run(args, &stdout, &stderr); got != want || !strings.Contains(stderr.String(), "Usage:") {
				t.Fatalf("exit = %d, want %d; stderr = %s", got, want, stderr.String())
			}
		})
	}
}
