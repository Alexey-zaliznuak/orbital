package routingrule

import (
	"regexp"
	"strings"
)

// MatchType определяет тип сопоставления паттерна с routing key.
type MatchType int

const (
	// MatchExact — точное совпадение.
	MatchExact MatchType = iota
	// MatchPrefix — routing key начинается с паттерна.
	MatchPrefix
	// MatchSuffix — routing key оканчивается на паттерн.
	MatchSuffix
	// MatchRegex — routing key соответствует регулярному выражению.
	MatchRegex
)

// RoutingRule связывает паттерн с PusherID.
type RoutingRule struct {
	// ID — уникальный идентификатор правила.
	ID string
	// Pattern — паттерн для сопоставления с routing key.
	Pattern string
	// MatchType — тип сопоставления.
	MatchType MatchType
	// PusherID — идентификатор пушера.
	PusherID string
	// Enabled — активно ли правило.
	Enabled bool
	// compiledRegex — скомпилированное регулярное выражение (для MatchRegex).
	// Не сериализуется, создаётся при загрузке правила.
	compiledRegex *regexp.Regexp `json:"-"`
}

// Match проверяет, соответствует ли routing key паттерну правила.
func (r *RoutingRule) Match(routingKey string) bool {
	switch r.MatchType {
	case MatchExact:
		return routingKey == r.Pattern
	case MatchPrefix:
		return strings.HasPrefix(routingKey, r.Pattern)
	case MatchSuffix:
		return strings.HasSuffix(routingKey, r.Pattern)
	case MatchRegex:
		if r.compiledRegex == nil {
			return false
		}
		return r.compiledRegex.MatchString(routingKey)
	default:
		return false
	}
}
