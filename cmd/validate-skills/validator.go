package main

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

var (
	frontMatter      = regexp.MustCompile(`(?s)\A---\r?\n(.*?)\r?\n---(?:\r?\n|\z)(.*)\z`)
	skillName        = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	brandColor       = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	inlineLinks      = regexp.MustCompile(`\[[^\]]*\]\(\s*(<[^>\n]+>|[^\s)]+)`)
	referenceLinks   = regexp.MustCompile(`(?m)^[ \t]{0,3}\[[^\]]+\]:[ \t]*(<[^>\n]+>|[^\s]+)`)
	bundledResources = regexp.MustCompile(`(?:^|[^\pL\pN_/])((?:references|scripts|assets)/[\pL\pN_./-]+\.[\pL\pN_-]+)`)
)

type validator struct {
	problems []string
}

func (v *validator) report(format string, args ...any) {
	v.problems = append(v.problems, fmt.Sprintf(format, args...))
}

func validate(root string) (int, []string) {
	v := &validator{}
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil && !os.IsNotExist(err) {
		v.report("%s: %v", filepath.Join(root, "skills"), err)
	}
	count := 0
	names := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() {
			count++
			v.checkSkill(filepath.Join(root, "skills", entry.Name()), names)
		}
	}
	if count == 0 {
		v.report("%s: no skill folders found", filepath.Join(root, "skills"))
	}
	files, err := filepath.Glob(filepath.Join(root, "*.md"))
	if err != nil {
		v.report("%s: %v", root, err)
	}
	for _, path := range files {
		v.checkMarkdown(path, root, "")
	}
	v.walk(filepath.Join(root, "docs"), func(path string) {
		if filepath.Ext(path) == ".md" {
			v.checkMarkdown(path, root, "")
		}
	})
	v.checkSyntax(root)
	return count, v.problems
}

func (v *validator) readText(path string) string {
	text, err := os.ReadFile(path)
	if err != nil {
		v.report("%s: %v", path, err)
		return ""
	}
	if !utf8.Valid(text) {
		v.report("%s: text must be valid UTF-8", path)
		return ""
	}
	return string(text)
}

func (v *validator) loadMapping(text, path string) map[string]any {
	decoder := yaml.NewDecoder(strings.NewReader(text))
	var value any
	if err := decoder.Decode(&value); err != nil {
		v.report("%s: invalid YAML: %v", path, err)
		return map[string]any{}
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		v.report("%s: expected one YAML document", path)
	}
	mapping, ok := value.(map[string]any)
	if !ok {
		v.report("%s: YAML must be a mapping with string keys", path)
		return map[string]any{}
	}
	return mapping
}

func (v *validator) checkSkill(skill string, names map[string]bool) {
	path := filepath.Join(skill, "SKILL.md")
	match := frontMatter.FindStringSubmatch(v.readText(path))
	if match == nil {
		v.report("%s: expected opening and closing YAML front matter delimiters", path)
		return
	}
	manifest := v.loadMapping(match[1], path)
	name, ok := manifest["name"].(string)
	if !ok || len(name) > 64 || !skillName.MatchString(name) {
		v.report("%s: name must be 1-64 lowercase letters, digits, or single hyphens", path)
	} else {
		if name != filepath.Base(skill) {
			v.report("%s: name must match folder %s", path, filepath.Base(skill))
		}
		if names[name] {
			v.report("%s: duplicate skill name: %s", path, name)
		}
		names[name] = true
	}
	description, ok := manifest["description"].(string)
	if !ok || strings.TrimSpace(description) == "" || utf8.RuneCountInString(description) > 1024 {
		v.report("%s: description must be a nonempty string of at most 1024 characters", path)
	}
	if strings.TrimSpace(match[2]) == "" {
		v.report("%s: instructions must not be empty", path)
	}
	if value, exists := manifest["disable-model-invocation"]; exists {
		if _, ok := value.(bool); !ok {
			v.report("%s: disable-model-invocation must be a boolean", path)
		}
	}
	v.checkMetadata(skill, manifest)
	v.walk(skill, func(markdown string) {
		if filepath.Ext(markdown) == ".md" {
			if strings.EqualFold(filepath.Base(markdown), "SKILL.md") && markdown != path {
				v.report("%s: each bundle must have only one SKILL.md, at its root", markdown)
			}
			v.checkMarkdown(markdown, skill, skill)
		}
	})
}

func (v *validator) checkMetadata(skill string, manifest map[string]any) {
	path := filepath.Join(skill, "agents", "openai.yaml")
	data := v.loadMapping(v.readText(path), path)
	interfaceData, ok := data["interface"].(map[string]any)
	if !ok {
		v.report("%s: interface must be a mapping", path)
	}
	for _, key := range []string{"display_name", "short_description", "default_prompt"} {
		value, ok := interfaceData[key].(string)
		if !ok || strings.TrimSpace(value) == "" {
			v.report("%s: interface.%s must be a nonempty string", path, key)
		}
	}
	if description, ok := interfaceData["short_description"].(string); ok {
		length := utf8.RuneCountInString(description)
		if length < 25 || length > 64 {
			v.report("%s: short_description must be 25-64 characters", path)
		}
	}
	if prompt, ok := interfaceData["default_prompt"].(string); ok {
		mention := regexp.MustCompile(`\$` + regexp.QuoteMeta(filepath.Base(skill)) + `(?:$|[^\pL\pN_-])`)
		if !mention.MatchString(prompt) {
			v.report("%s: default_prompt must mention $%s", path, filepath.Base(skill))
		}
	}
	for _, key := range []string{"icon_small", "icon_large"} {
		if value, exists := interfaceData[key]; exists {
			icon, ok := value.(string)
			if !ok || strings.TrimSpace(icon) == "" {
				v.report("%s: %s must be a relative asset path", path, key)
			} else {
				v.checkResource(path, icon, skill, skill, true)
			}
		}
	}
	if color, exists := interfaceData["brand_color"]; exists {
		if value, ok := color.(string); !ok || !brandColor.MatchString(value) {
			v.report("%s: brand_color must be a six-digit hex color", path)
		}
	}
	policy := map[string]any{}
	if value, exists := data["policy"]; exists {
		if mapping, ok := value.(map[string]any); ok {
			policy = mapping
		} else {
			v.report("%s: policy must be a mapping", path)
		}
	}
	invocation, exists := policy["allow_implicit_invocation"]
	if !exists {
		invocation = true
	}
	allow, ok := invocation.(bool)
	if !ok {
		v.report("%s: allow_implicit_invocation must be a boolean", path)
	}
	explicit := manifest["disable-model-invocation"] == true
	if explicit && (!ok || allow) {
		v.report("%s: allow_implicit_invocation must be false for an explicit-only skill", path)
	}
	if ok && !allow && !explicit {
		v.report("%s: disable-model-invocation must be true when Codex is explicit-only", filepath.Join(skill, "SKILL.md"))
	}
}

func (v *validator) checkMarkdown(path, boundary, skill string) {
	text := v.readText(path)
	destinations := make(map[string]bool)
	for _, pattern := range []*regexp.Regexp{inlineLinks, referenceLinks} {
		for _, match := range pattern.FindAllStringSubmatch(text, -1) {
			if !destinations[match[1]] {
				v.checkResource(path, match[1], filepath.Dir(path), boundary, false)
				destinations[match[1]] = true
			}
		}
	}
	if skill != "" {
		for _, match := range bundledResources.FindAllStringSubmatch(text, -1) {
			v.checkResource(path, match[1], skill, boundary, false)
		}
	}
}

func (v *validator) checkResource(source, destination, base, boundary string, asset bool) {
	parsed, err := url.Parse(strings.Trim(destination, "<>"))
	if err != nil {
		v.report("%s: invalid resource %s: %v", source, destination, err)
		return
	}
	if asset && (parsed.Scheme != "" || parsed.Host != "" || parsed.Path == "" || filepath.IsAbs(parsed.Path)) {
		v.report("%s: icon must be a nonempty relative asset path: %s", source, destination)
		return
	}
	if parsed.Scheme != "" || parsed.Host != "" || parsed.Path == "" {
		return
	}
	// Keep '..' intact until symlinks are resolved; cleaning first changes the target.
	target := base + string(filepath.Separator) + filepath.FromSlash(parsed.Path)
	if filepath.IsAbs(parsed.Path) {
		target = parsed.Path
	}
	resolved, err := resolvePath(target)
	if err != nil {
		v.report("%s: invalid resource %s: %v", source, destination, err)
		return
	}
	limit, err := resolvePath(boundary)
	if err != nil {
		v.report("%s: %v", source, err)
		return
	}
	relative, err := filepath.Rel(limit, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		v.report("%s: resource leaves its bundle: %s", source, destination)
		return
	}
	info, err := os.Stat(target)
	if err != nil || (asset && !info.Mode().IsRegular()) {
		v.report("%s: missing local resource: %s", source, destination)
	}
}

// Resolve symlinked ancestors even when the final resource does not exist yet.
func resolvePath(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil || !os.IsNotExist(err) {
		return resolved, err
	}
	parent, name := filepath.Split(strings.TrimRight(path, string(filepath.Separator)))
	if parent == "" || name == "" {
		return "", err
	}
	resolved, err = resolvePath(parent)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolved, name), nil
}

func (v *validator) walk(directory string, visit func(string)) {
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if path == directory && os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if !entry.IsDir() {
			visit(path)
		}
		return nil
	})
	if err != nil {
		v.report("%s: %v", directory, err)
	}
}

func (v *validator) checkSyntax(root string) {
	bash, bashErr := exec.LookPath("bash")
	shellcheck, shellErr := exec.LookPath("shellcheck")
	if bashErr != nil {
		v.report("Bash is missing. Install Bash and make it available on PATH.")
	}
	if shellErr != nil {
		v.report("ShellCheck is missing. Install ShellCheck and make it available on PATH.")
	}
	var shellFiles []string
	for _, directory := range []string{"skills", "scripts", "cmd", "tests"} {
		v.walk(filepath.Join(root, directory), func(path string) {
			switch filepath.Ext(path) {
			case ".go":
				if _, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.AllErrors); err != nil {
					v.report("%s: Go syntax: %v", path, err)
				}
			case ".sh":
				shellFiles = append(shellFiles, path)
				if bashErr == nil {
					v.checkCommand("Bash syntax", bash, "-n", path)
				}
			}
		})
	}
	if shellErr == nil && len(shellFiles) > 0 {
		v.checkCommand("ShellCheck", shellcheck, append([]string{"--severity=warning"}, shellFiles...)...)
	}
}

func (v *validator) checkCommand(label, command string, args ...string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, command, args...).CombinedOutput()
	if err != nil {
		v.report("%s: %v: %s", label, err, strings.TrimSpace(string(output)))
	}
}
