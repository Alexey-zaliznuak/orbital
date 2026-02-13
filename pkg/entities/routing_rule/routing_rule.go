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
	ID string `json:"id"`
	// Pattern — паттерн для сопоставления с routing key.
	Pattern string `json:"pattern"`
	// MatchType — тип сопоставления.
	MatchType MatchType `json:"match_type"`
	// PusherID — идентификатор пушера.
	PusherID string `json:"pusher_id"`
	// Enabled — активно ли правило.
	Enabled bool `json:"enabled"`
	// compiledRegex — скомпилированное регулярное выражение (для MatchRegex).
	// Не сериализуется, создаётся при загрузке правила.
	compiledRegex *regexp.Regexp `json:"-"`
}

// SetCompiledRegex устанавливает скомпилированное регулярное выражение.
func (r *RoutingRule) SetCompiledRegex(re *regexp.Regexp) {
	r.compiledRegex = re
}

// CompileRegex компилирует паттерн как регулярное выражение.
// Возвращает ошибку если паттерн невалиден.
func (r *RoutingRule) CompileRegex() error {
	if r.MatchType != MatchRegex {
		return nil
	}
	compiled, err := regexp.Compile(r.Pattern)
	if err != nil {
		return err
	}
	r.compiledRegex = compiled
	return nil
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
