package rules

import (
	"sort"
	"strings"
	"unicode"

	"github.com/yoheimuta/go-protoparser/v4/parser"
	"github.com/yoheimuta/go-protoparser/v4/parser/meta"
	"github.com/yoheimuta/protolint/linter/report"
	"github.com/yoheimuta/protolint/linter/rule"
	"github.com/yoheimuta/protolint/linter/visitor"
)

// NoSharedRepeatedMessageRule checks that a message type is not used as a
// repeated field in multiple different messages. When a message type is shared
// as repeated across request/response boundaries, adding fields for one usage
// inadvertently affects the other.
type NoSharedRepeatedMessageRule struct{}

func NewNoSharedRepeatedMessageRule() NoSharedRepeatedMessageRule {
	return NoSharedRepeatedMessageRule{}
}

func (r NoSharedRepeatedMessageRule) ID() string {
	return "NO_SHARED_REPEATED_MESSAGE"
}

func (r NoSharedRepeatedMessageRule) Purpose() string {
	return `Verifies that a message type is not used as a repeated field in multiple different messages.`
}

// IsOfficial returns false since this is a custom rule.
func (r NoSharedRepeatedMessageRule) IsOfficial() bool {
	return false
}

func (r NoSharedRepeatedMessageRule) Severity() rule.Severity {
	return rule.SeverityWarning
}

func (r NoSharedRepeatedMessageRule) Apply(proto *parser.Proto) ([]report.Failure, error) {
	v := &noSharedRepeatedVisitor{
		BaseAddVisitor: visitor.NewBaseAddVisitor(r.ID(), string(r.Severity())),
		usages:         make(map[string][]usageInfo),
	}
	return visitor.RunVisitor(v, proto, r.ID())
}

// usageInfo records where a message type is used as a repeated field.
type usageInfo struct {
	parentMessage string
	pos           meta.Position
}

// noSharedRepeatedVisitor walks the proto AST collecting repeated message-type
// field usages and reports violations in Finally().
type noSharedRepeatedVisitor struct {
	*visitor.BaseAddVisitor
	// usages maps a type name to all locations where it appears as a repeated field.
	usages map[string][]usageInfo
}

// VisitMessage inspects direct fields of the message for repeated message-type
// references and returns true to continue visiting nested messages.
func (v *noSharedRepeatedVisitor) VisitMessage(msg *parser.Message) bool {
	for _, body := range msg.MessageBody {
		field, ok := body.(*parser.Field)
		if !ok {
			continue
		}
		if field.IsRepeated && isMessageType(field.Type) {
			v.usages[field.Type] = append(v.usages[field.Type], usageInfo{
				parentMessage: msg.MessageName,
				pos:           field.Meta.Pos,
			})
		}
	}
	return true
}

// Finally checks collected usages and reports violations for any message type
// referenced as repeated from two or more distinct parent messages.
func (v *noSharedRepeatedVisitor) Finally() error {
	for typeName, infos := range v.usages {
		parents := distinctParents(infos)
		if len(parents) < 2 {
			continue
		}
		sort.Strings(parents)
		parentList := `"` + strings.Join(parents, `", "`) + `"`
		for _, info := range infos {
			v.AddFailuref(
				info.pos,
				`%q is used as a repeated field in multiple messages (%s). Consider defining separate messages for each usage to allow independent evolution.`,
				typeName,
				parentList,
			)
		}
	}
	return nil
}

// isMessageType returns true if typeName refers to a message type rather than
// a scalar. Scalars (string, int32, bool, etc.) start with a lowercase letter
// and contain no dots. Qualified names (e.g. "foo.bar.Baz") are always message
// types.
func isMessageType(typeName string) bool {
	if len(typeName) == 0 {
		return false
	}
	if strings.Contains(typeName, ".") {
		return true
	}
	return unicode.IsUpper(rune(typeName[0]))
}

func distinctParents(infos []usageInfo) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, info := range infos {
		if _, ok := seen[info.parentMessage]; !ok {
			seen[info.parentMessage] = struct{}{}
			result = append(result, info.parentMessage)
		}
	}
	return result
}

// Ensure the rule implements the rule.Rule interface at compile time.
var _ rule.Rule = NoSharedRepeatedMessageRule{}
