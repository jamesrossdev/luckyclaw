package providers

import (
	"fmt"
	"regexp"
	"strings"
)

var thinkBlockRegex = regexp.MustCompile(`(?is)<think>.*?</think>`)
var thinkTagRegex = regexp.MustCompile(`(?is)</?think>`)

func StripThinkArtifacts(content string) string {
	cleaned := collapseDuplicateAroundThinkTag(content)
	cleaned = thinkBlockRegex.ReplaceAllString(cleaned, "")
	cleaned = thinkTagRegex.ReplaceAllString(cleaned, "")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

func BuildUserVisibleContent(content, reasoning string, showReasoning bool) string {
	if !showReasoning {
		return StripThinkArtifacts(content)
	}

	derivedReasoning, answer := ExtractReasoningFromContent(content)
	if strings.TrimSpace(reasoning) == "" {
		reasoning = derivedReasoning
	}
	if strings.TrimSpace(answer) == "" {
		answer = strings.TrimSpace(content)
	}

	reasoning = strings.TrimSpace(reasoning)
	answer = strings.TrimSpace(StripThinkArtifacts(answer))

	if reasoning == "" {
		return answer
	}
	if answer == "" {
		return fmt.Sprintf("Reasoning:\n%s", reasoning)
	}

	return fmt.Sprintf("Reasoning:\n%s\n\nAnswer:\n%s", reasoning, answer)
}

func ExtractReasoningFromContent(content string) (string, string) {
	if strings.TrimSpace(content) == "" {
		return "", ""
	}

	matches := thinkBlockRegex.FindAllString(content, -1)
	reasoningParts := make([]string, 0, len(matches))
	for _, block := range matches {
		clean := thinkTagRegex.ReplaceAllString(block, "")
		clean = strings.TrimSpace(clean)
		if clean != "" {
			reasoningParts = append(reasoningParts, clean)
		}
	}

	answer := thinkBlockRegex.ReplaceAllString(content, "")
	answer = thinkTagRegex.ReplaceAllString(answer, "")
	answer = strings.TrimSpace(answer)

	return strings.TrimSpace(strings.Join(reasoningParts, "\n\n")), answer
}

func collapseDuplicateAroundThinkTag(content string) string {
	lower := strings.ToLower(content)
	tag := "</think>"
	idx := strings.Index(lower, tag)
	if idx == -1 {
		return content
	}

	left := strings.TrimSpace(content[:idx])
	right := strings.TrimSpace(content[idx+len(tag):])
	if left != "" && right != "" && left == right {
		return left
	}

	return content
}
