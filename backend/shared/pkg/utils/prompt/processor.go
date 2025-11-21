package prompt

import (
	"fmt"
	"regexp"

	"github.com/dongtandung2001/Doc-Agent/backend/shared/pkg/context"
)

// ProcessTemplateVariables replaces {{$key}} patterns with values from the context
// This is a pure utility function that can be used by any service
func ProcessTemplateVariables(prompt string, chatContext *context.ChatContext) string {
	var templateVarRegex = regexp.MustCompile(`\{\{\$([a-zA-Z0-9_]+)\}\}`)
	if chatContext == nil {
		return prompt
	}

	return templateVarRegex.ReplaceAllStringFunc(prompt, func(match string) string {
		// Extract key from the match without running regex again
		// match is "{{$key}}", so strip the {{$ prefix and }} suffix
		key := match[3 : len(match)-2]

		// Try to get the value from context
		if value, exists := chatContext.Get(key); exists {
			return fmt.Sprintf("%v", value)
		}

		// If key doesn't exist, return the original placeholder
		return match
	})
}
