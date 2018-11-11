package authr

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	propAccess         = "access"
	propWhere          = "where"
	propWhereRsrcType  = "rsrc_type"
	propWhereRsrcMatch = "rsrc_match"
	propWhereAction    = "action"
	propMeta           = "$meta"

	jtypeBool   = "JSON boolean"
	jtypeNumber = "JSON number"
	jtypeString = "JSON string"
	jtypeArray  = "JSON array"
	jtypeObject = "JSON object"
	jtypeNull   = "JSON null"
)

func (r *Rule) UnmarshalJSON(data []byte) error {
	*r = Rule{}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o, ok := v.(map[string]interface{})
	if !ok {
		return Error(fmt.Sprintf("expecting %s for rule definition, got %s", jtypeObject, typename(v)))
	}
	if ai, ok := o[propAccess]; ok {
		a, ok := ai.(string)
		if !ok {
			return jsonInvalidType([]string{propAccess}, ai, jtypeString)
		}
		switch a {
		case "allow":
			r.access = Allow
		case "deny":
			r.access = Deny
		default:
			return jsonInvalidPropValue([]string{propAccess}, `"allow" or "deny"`, fmt.Sprintf(`"%s"`, a))
		}
	} else {
		return jsonMissingProperty([]string{propAccess})
	}
	if wi, ok := o[propWhere]; ok {
		w, ok := wi.(map[string]interface{})
		if !ok {
			return jsonInvalidType([]string{propWhere}, wi, jtypeObject)
		}
		var err error
		err = unmarshalSlugSet(&r.where.resourceType, propWhereRsrcType, w)
		if err != nil {
			return err
		}
		err = unmarshalSlugSet(&r.where.action, propWhereAction, w)
		if err != nil {
			return err
		}
		csi, ok := w[propWhereRsrcMatch]
		if !ok {
			return jsonMissingProperty([]string{propWhere, propWhereRsrcMatch})
		}
		r.where.resourceMatch, err = unmarshalConditionSet([]string{propWhere, propWhereRsrcMatch}, csi)
		if err != nil {
			return err
		}
	} else {
		return jsonMissingProperty([]string{propWhere})
	}
	if meta, ok := o[propMeta]; ok {
		r.meta = meta
	}
	return nil
}

func unmarshalConditionSet(path []string, csi interface{}) (ConditionSet, error) {
	cs := ConditionSet{}
	cs.evaluators = []Evaluator{}
	switch _cs := csi.(type) {
	case map[string]interface{}:
		logic, csinneri, err := unwrapKeywordMap(path, _cs, logicalAnd.String(), logicalOr.String())
		if err != nil {
			return ConditionSet{}, err
		}
		switch logic {
		case logicalAnd.String():
			cs.conj = logicalAnd
		case logicalOr.String():
			cs.conj = logicalOr
		}
		path = append(path, logic)
		switch csinner := csinneri.(type) {
		case []interface{}:
			var err error
			cs.evaluators, err = unmarshalNestedConditions(
				append(path, cs.conj.String()),
				csinner,
			)
			if err != nil {
				return ConditionSet{}, err
			}
		default:
			return ConditionSet{}, jsonInvalidType(path, csinneri, jtypeArray)
		}
	case []interface{}:
		cs.conj = logicalAnd
		var err error
		cs.evaluators, err = unmarshalNestedConditions(path, _cs)
		if err != nil {
			return ConditionSet{}, err
		}
	default:
		return ConditionSet{}, jsonInvalidType(path, csi, jtypeObject, jtypeArray)
	}
	return cs, nil
}

func unmarshalNestedConditions(path []string, csinner []interface{}) ([]Evaluator, error) {
	evals := make([]Evaluator, len(csinner))
	for i, v := range csinner {
		if jarr, ok := v.([]interface{}); ok && len(jarr) == 3 && isstring(jarr[1]) {
			// smells like a condition!
			evals[i] = Cond(jarr[0], jarr[1].(string), jarr[2])
			continue
		}
		var err error
		evals[i], err = unmarshalConditionSet(append(path, strconv.Itoa(i)), v)
		if err != nil {
			return nil, err
		}
	}
	return evals, nil
}

func isstring(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

func unwrapKeywordMap(path []string, msi map[string]interface{}, validKeys ...string) (string, interface{}, error) {
	if len(msi) == 1 {
		var k string
		for _k := range msi {
			k = _k
			break
		}
		for _, vk := range validKeys {
			if k == vk {
				return k, msi[k], nil
			}
		}
	}
	err := Error(
		fmt.Sprintf(
			`invalid value for property "%s": expected %s with only one of the these key(s): "%s"`,
			strings.Join(path, "."),
			jtypeObject,
			strings.Join(validKeys, `", "`),
		),
	)
	return "", nil, err
}

func unmarshalSlugSet(ss *SlugSet, prop string, w map[string]interface{}) error {
	path := []string{propWhere, prop}
	ssi, ok := w[prop]
	if !ok {
		return jsonMissingProperty(path)
	}
	switch _ss := ssi.(type) {
	case map[string]interface{}:
		_, ssni, err := unwrapKeywordMap(path, _ss, "$not")
		if err != nil {
			return err
		}
		path = append(path, "$not")
		ss.mode = blocklist
		switch ssn := ssni.(type) {
		case []interface{}:
			// empty slug set IS allowed if the slugset is a blocklist.
			err := unmarshalStringSlice(path, ss, ssn)
			if err != nil {
				return err
			}
		case string:
			ss.elements = []string{ssn}
		default:
			return jsonInvalidType(path, ssni, jtypeArray, jtypeString)
		}
	case []interface{}:
		// empty slug set is NOT allowed if the slug set is not a blocklist
		// the rule would never match anything
		if len(_ss) == 0 {
			return jsonInvalidPropValue(path, "non-empty array", "empty array")
		}
		err := unmarshalStringSlice(path, ss, _ss)
		if err != nil {
			return err
		}
	case string:
		if _ss == "*" {
			ss.mode = wildcard
			ss.elements = []string{}
		} else {
			ss.elements = []string{_ss}
		}
	default:
		return jsonInvalidType(path, ssi, jtypeObject, jtypeArray, jtypeString)
	}
	return nil
}

func unmarshalStringSlice(path []string, ss *SlugSet, jarr []interface{}) error {
	ss.elements = make([]string, len(jarr))
	for i, v := range jarr {
		// TODO(nick): check for empty strings
		s, ok := v.(string)
		if !ok {
			return jsonInvalidType(append(path, fmt.Sprintf("%v", i)), v, jtypeString)
		}
		ss.elements[i] = s
	}
	return nil
}

func lexicalJoin(a []string) string {
	switch len(a) {
	case 0:
		return ""
	case 1:
		return a[0]
	case 2:
		return a[0] + " or " + a[1]
	default:
		return strings.Join(a[:len(a)-1], ", ") + " or " + a[len(a)-1]
	}
}

type ruleUnmarshalError struct {
	path []string
}

type jsonInvalidTypeError struct {
	ruleUnmarshalError
	needTypes []string
	v         interface{}
}

func (j jsonInvalidTypeError) Error() string {
	return fmt.Sprintf(`expecting %s for property "%s", got %s`, lexicalJoin(j.needTypes), strings.Join(j.path, "."), typename(j.v))
}

type jsonMissingPropertyError struct {
	ruleUnmarshalError
}

func (j jsonMissingPropertyError) Error() string {
	return fmt.Sprintf(`invalid rule; missing required property "%s"`, strings.Join(j.path, "."))
}

func jsonMissingProperty(path []string) error {
	return jsonMissingPropertyError{
		ruleUnmarshalError: ruleUnmarshalError{path: path},
	}
}

type jsonInvalidPropertyValueError struct {
	ruleUnmarshalError
	expecting, got string
}

func (j jsonInvalidPropertyValueError) Error() string {
	return fmt.Sprintf(`invalid value for property "%s", expecting %s, got %s`, strings.Join(j.path, "."), j.expecting, j.got)
}

func jsonInvalidPropValue(path []string, e, g string) error {
	return jsonInvalidPropertyValueError{
		ruleUnmarshalError: ruleUnmarshalError{path: path},
		expecting:          e,
		got:                g,
	}
}

func jsonInvalidType(path []string, v interface{}, needType ...string) error {
	return jsonInvalidTypeError{ruleUnmarshalError: ruleUnmarshalError{path: path}, needTypes: needType, v: v}
}

func typename(v interface{}) string {
	switch v.(type) {
	case bool:
		return jtypeBool
	case float64:
		return jtypeNumber
	case string:
		return jtypeString
	case []interface{}:
		return jtypeArray
	case map[string]interface{}:
		return jtypeObject
	case nil:
		return jtypeNull
	}
	panic(fmt.Sprintf("unexpected go type found in unmarshaled value: %T", v))
}
