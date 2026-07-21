package events

import (
	"strings"
	"unicode"

	"github.com/bigstack-oss/cube-cos-api/internal/definition/v1/events"
)

// filteredByKeyword: word-prefix match per keyword term over visible fields
// + metadata values ("CLU" matches "cluster", never "hacluster"); keeps time order
func (h *helper) filteredByKeyword(nonFilterEvents []events.Event) []events.Event {
	if !h.isKeywordRequired() {
		return nonFilterEvents
	}

	terms := strings.FieldsFunc(strings.ToLower(h.keyword), isNonToken)
	if len(terms) == 0 {
		return nonFilterEvents
	}

	// an event-id prefix ("CLU", "RUG00003") is the primary search token of
	// this table: when the keyword names ids, return only those
	if byId := filteredByIdPrefix(nonFilterEvents, strings.ToLower(h.keyword)); len(byId) > 0 {
		return byId
	}

	filtered := []events.Event{}
	for _, event := range nonFilterEvents {
		if matchAllTerms(event, terms) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

func filteredByIdPrefix(nonFilterEvents []events.Event, keyword string) []events.Event {
	filtered := []events.Event{}
	for _, event := range nonFilterEvents {
		if strings.HasPrefix(strings.ToLower(event.Id), keyword) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

func isNonToken(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsDigit(r)
}

func matchAllTerms(event events.Event, terms []string) bool {
	tokens := eventTokens(event)
	for _, term := range terms {
		if !anyTokenHasPrefix(tokens, term) {
			return false
		}
	}

	return true
}

func eventTokens(event events.Event) []string {
	fields := []string{
		event.Id,
		event.Description,
		event.Host,
		event.Category,
		event.Service,
		event.Severity,
		event.Message,
	}
	for _, value := range event.Metadata {
		if s, ok := value.(string); ok {
			fields = append(fields, s)
		}
	}

	tokens := []string{}
	for _, field := range fields {
		tokens = append(tokens, strings.FieldsFunc(strings.ToLower(field), isNonToken)...)
	}

	return tokens
}

func anyTokenHasPrefix(tokens []string, term string) bool {
	for _, token := range tokens {
		if strings.HasPrefix(token, term) {
			return true
		}
	}

	return false
}
