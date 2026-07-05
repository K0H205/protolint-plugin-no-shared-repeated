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
//
// Type references are resolved within the file: qualified references such as
// "example.ItemInfo" or ".example.ItemInfo" are matched with the plain
// "ItemInfo" when they refer to the same definition, and nested types with the
// same short name in different scopes are treated as distinct types. Enum
// types defined in the same file are excluded. Types listed in allowedTypes
// are never reported, which is useful for types that are intentionally shared
// (e.g. google.protobuf.Timestamp).
type NoSharedRepeatedMessageRule struct {
	allowedTypes map[string]struct{}
}

// NewNoSharedRepeatedMessageRule creates the rule. allowedTypes is a list of
// type names (as written in the proto, leading dot optional) that are allowed
// to be shared as repeated fields across messages.
func NewNoSharedRepeatedMessageRule(allowedTypes []string) NoSharedRepeatedMessageRule {
	allowed := make(map[string]struct{}, len(allowedTypes))
	for _, t := range allowedTypes {
		t = strings.TrimPrefix(strings.TrimSpace(t), ".")
		if t != "" {
			allowed[t] = struct{}{}
		}
	}
	return NoSharedRepeatedMessageRule{allowedTypes: allowed}
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
		BaseAddVisitor:  visitor.NewBaseAddVisitor(r.ID(), string(r.Severity())),
		allowedTypes:    r.allowedTypes,
		definedMessages: make(map[string]struct{}),
		definedEnums:    make(map[string]struct{}),
	}
	return visitor.RunVisitor(v, proto, r.ID())
}

// repeatedUsage records where a type is referenced as a repeated field.
type repeatedUsage struct {
	// typeName is the reference exactly as written in the field.
	typeName string
	// scope is the path of enclosing messages, outermost first.
	scope []string
	pos   meta.Position
}

// noSharedRepeatedVisitor walks the proto AST collecting repeated message-type
// field usages together with the types defined in the file, then resolves the
// references and reports violations in Finally().
type noSharedRepeatedVisitor struct {
	*visitor.BaseAddVisitor
	allowedTypes map[string]struct{}
	packageName  string
	// definedMessages and definedEnums hold package-relative full names
	// (e.g. "Outer.Inner") of types defined in this file.
	definedMessages map[string]struct{}
	definedEnums    map[string]struct{}
	usages          []repeatedUsage
}

func (v *noSharedRepeatedVisitor) VisitPackage(p *parser.Package) bool {
	v.packageName = p.Name
	return false
}

func (v *noSharedRepeatedVisitor) VisitEnum(e *parser.Enum) bool {
	v.definedEnums[e.EnumName] = struct{}{}
	return false
}

// VisitMessage walks the message tree manually so that the enclosing scope of
// each field is known; returning false stops the default traversal.
func (v *noSharedRepeatedVisitor) VisitMessage(msg *parser.Message) bool {
	v.walkMessage(nil, msg)
	return false
}

func (v *noSharedRepeatedVisitor) walkMessage(parentScope []string, msg *parser.Message) {
	scope := make([]string, 0, len(parentScope)+1)
	scope = append(scope, parentScope...)
	scope = append(scope, msg.MessageName)
	fullName := strings.Join(scope, ".")

	v.definedMessages[fullName] = struct{}{}

	for _, body := range msg.MessageBody {
		switch b := body.(type) {
		case *parser.Field:
			if b.IsRepeated && isMessageType(b.Type) {
				v.usages = append(v.usages, repeatedUsage{
					typeName: b.Type,
					scope:    scope,
					pos:      b.Meta.Pos,
				})
			}
		case *parser.Message:
			v.walkMessage(scope, b)
		case *parser.Enum:
			v.definedEnums[fullName+"."+b.EnumName] = struct{}{}
		}
	}
}

// Finally resolves collected usages and reports violations for any type
// referenced as repeated from two or more distinct parent messages. Failures
// are emitted in file-position order so the output is deterministic.
func (v *noSharedRepeatedVisitor) Finally() error {
	grouped := make(map[string][]repeatedUsage)
	for _, u := range v.usages {
		resolved := v.resolveTypeName(u.typeName, u.scope)
		if v.isAllowed(u.typeName, resolved) {
			continue
		}
		if _, isEnum := v.definedEnums[resolved]; isEnum {
			continue
		}
		grouped[resolved] = append(grouped[resolved], u)
	}

	type violation struct {
		pos        meta.Position
		typeName   string
		parentList string
	}
	var violations []violation
	for typeName, infos := range grouped {
		parents := distinctParents(infos)
		if len(parents) < 2 {
			continue
		}
		sort.Strings(parents)
		parentList := `"` + strings.Join(parents, `", "`) + `"`
		for _, info := range infos {
			violations = append(violations, violation{
				pos:        info.pos,
				typeName:   typeName,
				parentList: parentList,
			})
		}
	}
	sort.Slice(violations, func(i, j int) bool {
		return violations[i].pos.Offset < violations[j].pos.Offset
	})

	for _, vi := range violations {
		v.AddFailuref(
			vi.pos,
			`%q is used as a repeated field in multiple messages (%s). Consider defining separate messages for each usage to allow independent evolution.`,
			vi.typeName,
			vi.parentList,
		)
	}
	return nil
}

// resolveTypeName resolves a type reference to the package-relative full name
// of a type defined in this file, following protobuf scoping: relative
// references are searched from the innermost enclosing message outward.
// References to types not defined in this file (e.g. imported types) are
// returned with any leading dot and own-package prefix stripped, so different
// spellings of the same reference still group together.
func (v *noSharedRepeatedVisitor) resolveTypeName(name string, scope []string) string {
	if strings.HasPrefix(name, ".") {
		return v.stripOwnPackage(strings.TrimPrefix(name, "."))
	}
	for i := len(scope); i > 0; i-- {
		candidate := strings.Join(scope[:i], ".") + "." + name
		if v.isDefined(candidate) {
			return candidate
		}
	}
	if v.isDefined(name) {
		return name
	}
	stripped := v.stripOwnPackage(name)
	return stripped
}

func (v *noSharedRepeatedVisitor) stripOwnPackage(name string) string {
	if v.packageName != "" && strings.HasPrefix(name, v.packageName+".") {
		return strings.TrimPrefix(name, v.packageName+".")
	}
	return name
}

func (v *noSharedRepeatedVisitor) isDefined(fullName string) bool {
	if _, ok := v.definedMessages[fullName]; ok {
		return true
	}
	_, ok := v.definedEnums[fullName]
	return ok
}

func (v *noSharedRepeatedVisitor) isAllowed(written, resolved string) bool {
	if _, ok := v.allowedTypes[strings.TrimPrefix(written, ".")]; ok {
		return true
	}
	_, ok := v.allowedTypes[resolved]
	return ok
}

// isMessageType returns true if typeName refers to a named type rather than
// a scalar. Scalars (string, int32, bool, etc.) start with a lowercase letter
// and contain no dots. Qualified names (e.g. "foo.bar.Baz") are always named
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

// distinctParents returns the unique full names of the messages that contain
// the given usages.
func distinctParents(infos []repeatedUsage) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, info := range infos {
		parent := strings.Join(info.scope, ".")
		if _, ok := seen[parent]; !ok {
			seen[parent] = struct{}{}
			result = append(result, parent)
		}
	}
	return result
}

// Ensure the rule implements the rule.Rule interface at compile time.
var _ rule.Rule = NoSharedRepeatedMessageRule{}
