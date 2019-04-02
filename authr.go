package authr

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var rcache regexpCache = &noopRegexpCache{}

var (
	operators = map[string]operator{
		"=":    operatorFunc(looseEquality),
		"!=":   negate(operatorFunc(looseEquality)),
		"$in":  in("$in", false),
		"$nin": in("$nin", true),
		"~=":   operatorFunc(like),
		"&":    intersect("&", false),
		"-":    intersect("-", true),
		"~":    &regexpOperator{ci: false, inv: false},
		"~*":   &regexpOperator{ci: true, inv: false},
		"!~":   &regexpOperator{ci: false, inv: true},
		"!~*":  &regexpOperator{ci: true, inv: true},
	}
)

const Version = "2.0.1"

func init() {
	rcache = newRegexpListCache(5)
}

// Error is used for any error that occurs during authr's evaluation. They are
// normally returned as a result of improperly constructed rules.
type Error string

func (e Error) Error() string {
	return string(e)
}

// Access represents a value which will distinguish a rule as either being
// a restricting rule or a permitting one.
type Access string

const (
	// Allow when set as the "access" on a rule will return true when the rule
	// is matched
	Allow Access = "allow"

	// Deny when set as the "access" on a rule will return false when the rule
	// is matched
	Deny Access = "deny"
)

// logicalConjunction is the representation of the logic that joins condition
// sets
type logicalConjunction string

func (l logicalConjunction) String() string {
	return string(l)
}

const (
	// logicalAnd is used as a single key in a map to denote a set of conditions
	// that should be evaluated and all values should be true, to return true
	logicalAnd logicalConjunction = "$and"

	// logicalOr is used as a single key in a map to denote a set of conditions
	// that should be evaluated and any values should be true to return true
	logicalOr logicalConjunction = "$or"

	// ImpliedConjunction is the default conjunction on condition sets that do
	// not have an explicit conjunction
	ImpliedConjunction = logicalAnd
)

// Subject is an abstract representation of an entity capable of performing
// actions on resources. It is distinguished by have a method which is supposed
// to return a list of rules that apply to the subject.
type Subject interface {
	// GetRules simply retrieves a list of rules. The ordering of these rules
	// does matter. The rules themselves can be retrieve by any means necessary —
	// whether it be from a database or a config file; whatever works.
	GetRules() ([]*Rule, error)
}

// Resource is an abstract representation of an entity the is the target of
// actions performed by subjects. Resources have a type and attributes.
//
// A "type" is what you might expect. If a blog were in need of an access
// control system, the resource type for a post would simply be "post" and the
// writers "author" perhaps.
//
// Attributes are any properties of a resource that can be evaluated. A post,
// for example, can have "tags", which when being retrieve with
// GetResourceAttribute() would return a slice of strings.
//
// Unknown or missing properties should simply return "nil" and not an error.
type Resource interface {
	GetResourceType() (string, error)
	GetResourceAttribute(string) (interface{}, error)
}

// Rule represents the basic building block of an access control system. They
// can be likened to a single statement in an access-control list (ACL). Rules
// are entities which are said to "belong" to subjects in that they have been
// granted or applied to subjects based on the state of a datastore or the state
// of the subject themselves.
//
// Building rules in Go (instead of say Unmarshaling from JSON) looks like this:
//     r := new(Rule).
//         Access(Allow).
//         Where(
//             Action("delete"),
//             Not(ResourceType("user")),
//             ResourceMatch(
//                 Cond("@id", "!=", "1"),
//                 Or(
//                     Cond("@status", "=", "active"),
//                     Cond("@deleted_date", "=", nil),
//                 ),
//             ),
//         )
// This can be quite verbose, externally. A suggestion to reduce the verbosity
// might be to have a dedicate .go file that specifies rules where you can dot
// import authr. (https://golang.org/ref/spec#Import_declarations)
type Rule struct {
	access Access
	where  struct {
		resourceType  SlugSet
		resourceMatch ConditionSet
		action        SlugSet
	}
	meta interface{}
}

func (r Rule) Access(at Access) *Rule {
	r.access = at
	return &r
}

func (r Rule) Meta(meta interface{}) *Rule {
	r.meta = meta
	return &r
}

func (r Rule) Where(action, resourceType SlugSet, conditions ConditionSet) *Rule {
	r.where.action = action
	r.where.resourceType = resourceType
	r.where.resourceMatch = conditions
	return &r
}

func (r *Rule) reset() {
	*r = Rule{}
}

type slugSetMode int

const (
	allowlist slugSetMode = iota
	blocklist
	wildcard
)

// SlugSet is an internal means of representing an arbitrary set of strings. The
// "rsrc_type" and "action" sections of a rule have this type.
type SlugSet struct {
	mode     slugSetMode
	elements []string
}

func newSlugSet(slugs []string) SlugSet {
	ss := SlugSet{}
	if len(slugs) == 1 && slugs[0] == "*" {
		ss.mode = wildcard
		slugs = []string{}
	}
	ss.elements = slugs
	return ss
}

// ResourceType allows for the specification of resource types in a rule. The
// default mode is an "allowlist". Use Not(Action(...)) to specify a "blocklist"
func ResourceType(sset ...string) SlugSet {
	return newSlugSet(sset)
}

// Action allows for the specification of actions in a rule. The default mode is
// an "allowlist". Use Not(Action(...)) to specify a "blocklist"
func Action(sset ...string) SlugSet {
	return newSlugSet(sset)
}

// Not will return a copy of the provided SlugSet that will operate in a blocklist
// mode. Meaning the elements if matched in a calculation will return "false"
func Not(s SlugSet) SlugSet {
	s.mode = blocklist
	return s
}

// Not is a way to turn a SlugSet into a blocklist instead of the default
// allowlist mode.
//
// DEPRECATED: this API is awkward, use the authr.Not(Action("foo", "bar"))
// method instead.
// TODO(nkcmr): remove this in v3 of authr
func (s SlugSet) Not() SlugSet {
	s.mode = blocklist
	return s
}

func (s SlugSet) contains(b string) (bool, error) {
	if s.mode == wildcard {
		return true, nil
	}
	contained := false
	for _, a := range s.elements {
		if a == b {
			contained = true
			break
		}
	}
	if s.mode == blocklist {
		return !contained, nil
	} else if s.mode == allowlist {
		return contained, nil
	}
	panic(fmt.Sprintf("unknown slugset mode: '%v'", s.mode))
}

type ConditionSet struct {
	conj       logicalConjunction
	evaluators []Evaluator
}

// ResourceMatch is just a more readable way to start the rsrc_match section of
// a rule. It uses the implied logical conjunction AND.
func ResourceMatch(es ...Evaluator) ConditionSet {
	return And(es...).(ConditionSet)
}

// And returns an Evaluator that combines multiple Evaluators and will evaluate
// the set of evaluators with the logical conjunction AND. The behavior of the
// AND evaluator is to evaluate each sub-evaluator in order until one returns
// false or all return true. Once it finds a negative evaluator, it will halt
// and return — also known as short-circuiting.
func And(subEvaluators ...Evaluator) Evaluator {
	return ConditionSet{
		conj:       logicalAnd,
		evaluators: subEvaluators,
	}
}

// Or returns an Evaluator that is just like And, except it evaluate with the OR
// logical conjunction. Meaning it will evaluate until a sub-evaluator returns
// true, and also short-circuit.
func Or(subEvaluators ...Evaluator) Evaluator {
	return ConditionSet{
		conj:       logicalOr,
		evaluators: subEvaluators,
	}
}

func (c ConditionSet) evaluate(r Resource) (bool, error) {
	result := true // Vacuous truth: https://en.wikipedia.org/wiki/Vacuous_truth
	for _, eval := range c.evaluators {
		subresult, err := eval.evaluate(r)
		if err != nil {
			return false, err
		}
		if c.conj == logicalOr {
			if subresult {
				return true, nil // short-circuit
			}
			result = false
		} else if c.conj == logicalAnd {
			if !subresult {
				return false, nil // short-circuit
			}
			result = true
		}
	}
	return result, nil
}

// Can is the core access control computation function. It takes in a subject,
// action, and resource. It will answer the question "Can this subject perform
// this action on this resource?".
func Can(s Subject, action string, r Resource) (bool, error) {
	var (
		err          error
		rules        []*Rule
		resourceType string
	)
	if rules, err = s.GetRules(); err != nil {
		return false, err
	}
	if resourceType, err = r.GetResourceType(); err != nil {
		return false, err
	}
	for _, rule := range rules {
		var (
			ok  bool
			err error
		)
		if ok, err = rule.where.resourceType.contains(resourceType); err != nil {
			return false, err
		}
		if !ok {
			continue
		}
		if ok, err = rule.where.action.contains(action); err != nil {
			return false, err
		}
		if !ok {
			continue
		}
		if ok, err = rule.where.resourceMatch.evaluate(r); err != nil {
			return false, err
		}
		if !ok {
			continue
		}

		if rule.access == Allow {
			return true, nil
		} else if rule.access == Deny {
			return false, nil
		}

		// unknown type!
		panic(fmt.Sprintf("authr: unknown access type: '%s'", rule.access))
	}

	// default to "deny all"
	return false, nil
}

// Evaluator is an abstract representation of something that is capable of
// analyzing a Resource
type Evaluator interface {
	evaluate(Resource) (bool, error)
}

type condition struct {
	left, right interface{}
	op          string
}

// Cond is the basic unit of a resource match section of a rule. It represents
// a single condition to be evaluated against a Resource. Constructing a
// condition should be quite natural, like so:
//
//     Cond("@id", "=", "123")
//
// The above condition says that the "id" attribute on a resource MUST equal
// 123. References to resource attributes are prefixed with an "@" character
// to distinguish them from literal values. To specify multiple conditions, use
// the condition sets:
//
//     And(
//         Cond("@status", "=", "active"),
//         Cond("@name", "$in", []string{
//             "mike",
//             "jane",
//             "rachel",
//         }),
//     )
func Cond(left interface{}, op string, right interface{}) Evaluator {
	return condition{
		left:  left,
		right: right,
		op:    op,
	}
}

func (c condition) evaluate(r Resource) (bool, error) {
	var (
		_operator   operator
		ok          bool
		left, right interface{}
		err         error
	)
	if _operator, ok = operators[c.op]; !ok {
		return false, Error(fmt.Sprintf("unknown operator: '%s'", c.op))
	}
	left, err = determineValue(r, c.left)
	if err != nil {
		return false, err
	}
	right, err = determineValue(r, c.right)
	if err != nil {
		return false, err
	}
	return _operator.compute(left, right)
}

func determineValue(r Resource, a interface{}) (interface{}, error) {
	if str, ok := a.(string); ok && len(str) > 0 {
		if str[0] == '@' {
			return r.GetResourceAttribute(str[1:])
		}
		if len(str) >= 2 && str[0:2] == "\\@" {
			a = (str[1:])
		}
	}
	return a, nil
}

type operator interface {
	compute(left, right interface{}) (bool, error)
}

type operatorFunc func(left, right interface{}) (bool, error)

func (o operatorFunc) compute(left, right interface{}) (bool, error) {
	return o(left, right)
}

func negate(op operator) operator {
	return operatorFunc(func(left, right interface{}) (bool, error) {
		res, err := op.compute(left, right)
		if err != nil {
			return false, err
		}
		return !res, nil
	})
}

func intersect(opsym string, inv bool) operator {
	return operatorFunc(func(left, right interface{}) (bool, error) {
		lv, rv := reflect.ValueOf(left), reflect.ValueOf(right)
		if !isArrayIsh(lv) {
			return false, Error(fmt.Sprintf("%s operator expects both operands to be an array or slice, received %T for left operand", opsym, left))
		}
		if !isArrayIsh(rv) {
			return false, Error(fmt.Sprintf("%s operator expects both operands to be an array or slice, received %T for right operand", opsym, right))
		}
		for i := 0; i < lv.Len(); i++ {
			for j := 0; j < rv.Len(); j++ {
				ok, err := looseEquality(lv.Index(i).Interface(), rv.Index(j).Interface())
				if err != nil {
					return false, err
				}
				if ok {
					return !inv, nil
				}
			}
		}
		return inv, nil
	})
}

func isArrayIsh(v reflect.Value) bool {
	k := v.Kind()
	return k == reflect.Array || k == reflect.Slice
}

func in(opsym string, inv bool) operator {
	return operatorFunc(func(left, right interface{}) (bool, error) {
		rv := reflect.ValueOf(right)
		if !isArrayIsh(rv) {
			return false, Error(fmt.Sprintf("%s operator expects the right operand to be an array or slice, received %T", opsym, right))
		}
		for i := 0; i < rv.Len(); i++ {
			ok, err := looseEquality(left, rv.Index(i).Interface())
			if err != nil {
				return false, err
			}
			if ok {
				return !inv, nil
			}
		}
		return inv, nil
	})
}

func like(left, right interface{}) (bool, error) {
	sr, ok := right.(string)
	if !ok || len(sr) == 0 {
		return false, Error("right operand of the like operator (~=) must be a non-empty string")
	}
	var (
		pleft  string = "^"
		pright string = "$"
	)
	if sr[0] == '*' {
		pleft = ""
		sr = sr[1:]
	}
	if sr[len(sr)-1] == '*' {
		pright = ""
		sr = sr[0 : len(sr)-2]
	}
	patstring := "(?i)" + pleft + regexp.QuoteMeta(sr) + pright
	r, ok := rcache.find(patstring)
	if !ok {
		r = regexp.MustCompile(patstring)
		rcache.add(patstring, r)
	}
	switch lv := left.(type) {
	case string:
		return r.MatchString(lv), nil
	default:
		return r.MatchString(fmt.Sprintf("%v", left)), nil
	}
}

type regexpOperator struct {
	ci, inv bool
}

func (r *regexpOperator) compute(left, right interface{}) (bool, error) {
	var pattern *regexp.Regexp
	if patstring, ok := right.(string); ok && len(patstring) > 0 {
		var (
			err error
			ok  bool
		)
		if r.ci {
			patstring = "(?i)" + patstring
		}
		pattern, ok = rcache.find(patstring)
		if !ok {
			pattern, err = regexp.Compile(patstring)
			if err != nil {
				return false, err
			}
			rcache.add(patstring, pattern)
		}
	} else {
		return false, Error(fmt.Sprintf("right operand of the %s must be a non-empty string", r.operatorName()))
	}

	var ok bool
	// so, we can potentially avoid a LOT of allocations if we simply see
	// our left value is a string before jamming it into fmt.Sprintf and
	// needing to allocate
	switch l := left.(type) {
	case string:
		ok = pattern.MatchString(l)
	default:
		ok = pattern.MatchString(fmt.Sprintf("%+v", l))
	}
	if r.inv {
		return !ok, nil
	} else {
		return ok, nil
	}
}

func (r *regexpOperator) operatorName() string {
	op := "~"
	name := []string{"regexp", "operator"}
	if r.ci {
		op = op + "*"
		name = append([]string{"case-insensitive"}, name...)
	}
	if r.inv {
		op = "!" + op
		name = append([]string{"inverse"}, name...)
	}

	return fmt.Sprintf("%s (%s)", strings.Join(name, " "), op)
}

func looseEquality(left, right interface{}) (bool, error) {
	switch l := left.(type) {
	case string:
		switch r := right.(type) {
		case string:
			return l == r, nil
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return l == fmt.Sprintf("%v", r), nil
		case bool:
			return boolstringequal(r, l), nil
		case nil:
			return l == "", nil
		default:
			return false, Error(fmt.Sprintf("unsupported type in loose equality check: '%T'", r))
		}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		switch r := right.(type) {
		case string:
			return fmt.Sprintf("%v", l) == r, nil
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return fmt.Sprintf("%v", l) == fmt.Sprintf("%v", r), nil
		case bool:
			n := numbertofloat64(l)
			if r {
				return n == float64(1), nil
			} else {
				return n == float64(0), nil
			}
		case nil:
			return numbertofloat64(l) == float64(0), nil
		default:
			return false, Error(fmt.Sprintf("unsupported type in loose equality check: '%T'", r))
		}
	case bool:
		switch r := right.(type) {
		case string:
			return boolstringequal(l, r), nil
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			n := numbertofloat64(r)
			if l {
				return n == float64(1), nil
			} else {
				return n == float64(0), nil
			}
		case bool:
			return l == r, nil
		case nil:
			return l == false, nil
		default:
			return false, Error(fmt.Sprintf("unsupported type in loose equality check: '%T'", r))
		}
	case nil:
		switch r := right.(type) {
		case string:
			return r == "", nil
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			return numbertofloat64(r) == float64(0), nil
		case bool:
			return r == false, nil
		case nil:
			return true, nil
		default:
			return false, Error(fmt.Sprintf("unsupported type in loose equality check: '%T'", r))
		}
	default:
		return false, Error(fmt.Sprintf("unsupported type in loose equality check: '%T'", l))
	}
}

func boolstringequal(a bool, b string) bool {
	if !a {
		return b == "" || b == "0"
	} else {
		return len(b) > 0 && b != "0"
	}
}

func numbertofloat64(n interface{}) float64 {
	switch _n := n.(type) {
	case int:
		return float64(_n)
	case int8:
		return float64(_n)
	case int16:
		return float64(_n)
	case int32:
		return float64(_n)
	case int64:
		return float64(_n)
	case uint:
		return float64(_n)
	case uint8:
		return float64(_n)
	case uint16:
		return float64(_n)
	case uint32:
		return float64(_n)
	case uint64:
		return float64(_n)
	case float32:
		return float64(_n)
	case float64:
		return _n
	}
	panic(fmt.Sprintf("numbertofloat64 received non-numeric type: %T", n))
}
