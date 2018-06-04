package authr

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var (
	operators = map[string]operator{
		"=":    equals,
		"!=":   notequals,
		"$in":  in,
		"$nin": nin,
		"~=":   like,
		"~":    regexpOperatorFactory(regopts{ci: false, inv: false}),
		"~*":   regexpOperatorFactory(regopts{ci: true, inv: false}),
		"!~":   regexpOperatorFactory(regopts{ci: false, inv: true}),
		"!~*":  regexpOperatorFactory(regopts{ci: true, inv: true}),
	}
)

const Version = "1.1.2"

func init() {
	panic("the go implementation of github.com/cloudflare/authr is still untested. usage is actively discouraged.")
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
	Deny = "Deny"
)

type LogicalConjunction string

const (
	LogicalAnd LogicalConjunction = "$and"
	LogicalOr                     = "$or"
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

// Rule represents the basic building block of an access control system. Rules
// are entities which are said to "belong" to subjects in that they have been
// granted or applied to subjects based on the state of a datastore or the state
// of the subject themselves.
//
// Rules have a few sections which constitute a valid authr rule.
type Rule struct {
	access Access
	where  struct {
		resourceType  slugSet
		resourceMatch conditionSet
		action        slugSet
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

func (r Rule) Where(resourceType, action slugSet, conditions conditionSet) *Rule {
	r.where.resourceType = resourceType
	r.where.action = action
	r.where.resourceMatch = conditions
	return &r
}

func (r *Rule) reset() {
	*r = Rule{}
}

type slugSetMode int

const (
	whitelist slugSetMode = iota
	blacklist
)

type slugSet struct {
	mode     slugSetMode
	elements []string
}

func ResourceType(sset ...string) slugSet {
	ss := slugSet{}
	ss.elements = sset
	return ss
}

func Action(sset ...string) slugSet {
	return slugSet{
		elements: sset,
	}
}

func (s slugSet) Not() slugSet {
	s.mode = blacklist
	return s
}

func (s slugSet) contains(b string) (bool, error) {
	if s.mode == wildcard {
		return true, nil
	} else {
		var contained bool = false
		for _, a := range s.elements {
			if a == b {
				contained = true
				break
			}
		}
		if s.mode == blacklist {
			return !contained, nil
		} else if s.mode == whitelist {
			return contained, nil
		}
		panic(fmt.Sprintf("unknown slugset mode: '%v'", s.mode))
	}
}

type conditionSet struct {
	logicalConjunction LogicalConjunction
	evaluators         []Evaluator
}

// ResourceMatch is just a more readable way to start the rsrc_match section of
// a rule. It uses the implied logical conjunction AND.
func ResourceMatch(es ...Evaluator) conditionSet {
	return And(es...).(conditionSet)
}

// And returns an Evaluator that combines multiple Evaluators and will evaluate
// the set of evaluators with the logical conjunction AND. The behavior of the
// AND evaluator is to evaluate each sub-evaluator in order until one returns
// false or all return true. Once it finds a negative evaluator, it will halt
// and return — also known as short-circuiting.
func And(subEvaluators ...Evaluator) Evaluator {
	return conditionSet{
		logicalConjunction: LogicalAnd,
		evaluators:         subEvaluators,
	}
}

// Or returns an Evaluator that is just like And, except it evaluate with the OR
// logical conjunction. Meaning it will evaluate until a sub-evaluator returns
// true, and also short-circuit.
func Or(subEvaluators ...Evaluator) Evaluator {
	return conditionSet{
		logicalConjunction: LogicalOr,
		evaluators:         subEvaluators,
	}
}

func (c conditionSet) evaluate(r Resource) (bool, error) {
	result := true // Vacuous truth: https://en.wikipedia.org/wiki/Vacuous_truth
	for _, eval := range c.evaluators {
		subresult, err := eval.evaluate(r)
		if err != nil {
			return false, err
		}
		if c.logicalConjunction == LogicalOr {
			if subresult {
				return true // short-circuit
			}
			result = false
		} else if c.logicalConjunction == LogicalAnd {
			if !subresult {
				return false // short-cicuit
			}
			result = true
		}
	}
	return false, nil
}

func Can(s Subject, act string, r Resource) (bool, error) {
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
		if ok, err = rule.where.action.contains(act); err != nil {
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
	return _operator(left, right)
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

type operator func(left, right interface{}) (bool, error)

func equals(left, right interface{}) (bool, error) {
	return looseEquality(left, right)
}

func notequals(left, right interface{}) (bool, error) {
	eq, err := looseEquality(left, right)
	if err != nil {
		return false, err
	}
	return !eq, nil
}

func in(left, right interface{}) (bool, error) {
	rv := reflect.ValueOf(right)
	k := rv.Kind()
	if k != reflect.Array && k != reflect.Slice {
		return false, Error(fmt.Sprintf("$in operator expects the right operand to be an array or slice, received %T", right))
	}

	for i := 0; i < rv.Len(); i++ {
		eq, err := looseEquality(left, rv.Index(i).Interface())
		if err != nil {
			return false, err
		}
		if eq {
			return true, nil
		}
	}

	return false, nil
}

// \m/
func nin(left, right interface{}) (bool, error) {
	isin, err := in(left, right)
	if err != nil {
		return false, err
	}
	return !isin, nil
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
	return regexp.MustCompile("(?i)" + pleft + regexp.QuoteMeta(sr) + pright).MatchString(fmt.Sprintf("%v", left)), nil
}

type regopts struct {
	ci, inv bool
}

func operatorName(opts regopts) string {
	op := "~"
	name := []string{"regexp", "operator"}
	if opts.ci {
		op = op + "*"
		name = append([]string{"case-insensitive"}, name...)
	}
	if opts.inv {
		op = "!" + op
		name = append([]string{"inverse"}, name...)
	}

	return fmt.Sprintf("%s (%s)", strings.Join(name, " "), op)
}

func regexpOperatorFactory(opts regopts) operator {
	return func(left, right interface{}) (bool, error) {
		var pattern *regexp.Regexp
		if patstring, ok := right.(string); ok && len(patstring) > 0 {
			var err error
			if opts.ci {
				patstring = "(?i)" + patstring
			}
			pattern, err = regexp.Compile(patstring)
			if err != nil {
				return false, err
			}
		} else {
			return false, Error(fmt.Sprintf("right operand of the %s must be a non-empty string", operatorName(opts)))
		}

		ok := pattern.MatchString(fmt.Sprintf("%v", left))
		if opts.inv {
			return !ok, nil
		} else {
			return ok, nil
		}
	}
}

// Comparable is an abstract representation of a type which is capable of being
// compared for equality against arbitrary values. This is useful if you will
// be returning your own types from Resource.GetResourceAttribute and you want
// them to still be comparable.
type Comparable interface {
	EqualInterface(interface{}) (bool, error)
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
		case Comparable:
			return r.EqualInterface(left)
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
		case Comparable:
			return r.EqualInterface(left)
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
	case Comparable:
		return l.EqualInterface(right)
	default:
		return false, Error(fmt.Sprintf("unsupported type in loose equality check: '%T'", l))
	}
}

func boolstringequal(a bool, b string) bool {
	if !a {
		return b == "" || b == "0"
	} else {
		return len(b) > 0 && b != ""
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

func test() {
	new(Rule).
		Access(Allow).
		Where(
			ResourceType("zone"),
			Action("delete"),
			ResourceMatch(
				Cond("@id", "=", "123"),
				Or(
					Cond("@status", "=", "D"),
					Cond("@name", "$in", []string{"foo.com", "bar.net"}),
				),
			),
		).
		Meta(map[string]interface{}{
			"rule_id": 4431,
		})
	// x := new(Rule).
	// 	Access(Allow).
	// 	Where(
	// 		ResourceType("post").Not(),
	// 		Action("update"),
	// 		ResourceMatch(
	// 			Cond("@id", "=", "123"),
	// 			Or(
	// 				Cond("@name", "$in", []string{"foo", "bar"}),
	// 				Cond("@status", "$nin", []string{"A", "D"}),
	// 			),
	// 		),
	// 	)
}
