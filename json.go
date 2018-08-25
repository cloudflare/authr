package authr

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	r.reset()
	br := bytes.NewReader(data)
	d := json.NewDecoder(br)
	var v interface{}
	err := d.Decode(&v)
	if err != nil {
		return err
	}
	o, ok := v.(map[string]interface{})
	if !ok {
		return Error(fmt.Sprintf("expecting %s for rule definition, got %s", jtypeObject, typename(v)))
	}
	if ai, ok := o[propAccess]; ok {
		if a, ok := ai.(string); !ok {
			return jsonInvalidType([]string{propAccess}, ai, jtypeString)
		} else {
			switch a {
			case "allow":
				r = r.Access(Allow)
			case "deny":
				r = r.Access(Deny)
			default:
				return jsonInvalidPropValue([]string{propAccess}, `"allow" or "deny"`, a)
			}
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
		err = unmarshalConditionSet([]string{propWhere, propWhereRsrcMatch}, &r.where.resourceMatch, csi)
		if err != nil {
			return err
		}
	} else {
		return jsonMissingProperty([]string{propWhere})
	}
	if meta, ok := o[propMeta]; ok {
		r = r.Meta(meta)
	}
	return nil
}

func unmarshalConditionSet(path []string, cs *conditionSet, csi interface{}) error {
	cs.evaluators = []Evaluator{}
	switch _cs := csi.(type) {
	case map[string]interface{}:
		logic, csinneri, err := unwrapKeywordMap(path, _cs, LogicalAnd.String(), LogicalOr.String())
		if err != nil {
			return err
		}
		switch logic {
		case LogicalAnd.String():
			cs.logicalConjunction = LogicalAnd
		case LogicalOr.String():
			cs.logicalConjunction = LogicalOr
		}
		path = append(path, logic)
		switch csinner := csinneri.(type) {
		case []interface{}:
			err := unmarshalNestedConditions(path, cs, csinner)
			if err != nil {
				return err
			}
		default:
			return jsonInvalidType(path, csinneri, jtypeArray)
		}
	case []interface{}:
		cs.logicalConjunction = LogicalAnd
		err := unmarshalNestedConditions(path, cs, _cs)
		if err != nil {
			return err
		}
	default:
		return jsonInvalidType(path, csi, jtypeObject, jtypeArray)
	}
	return nil
}

func unmarshalNestedConditions(path []string, cs *conditionSet, csinner []interface{}) error {
	for i, v := range csinner {
		if jarr, ok := v.([]interface{}); ok && len(jarr) == 3 && isstring(jarr[1]) {
			// smells like a condition!
			cs.evaluators = append(cs.evaluators, Cond(jarr[0], jarr[1].(string), jarr[2]))
		} else {
			nestedCs := &conditionSet{}
			err := unmarshalConditionSet(append(path, fmt.Sprintf("%v", i)), nestedCs, v)
			if err != nil {
				return err
			}
			cs.evaluators = append(cs.evaluators, nestedCs)
		}
	}
	return nil
}

func isstring(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

func unwrapKeywordMap(path []string, msi map[string]interface{}, validKeys ...string) (string, interface{}, error) {
	if len(msi) == 1 {
		var k string
		for _k, _ := range msi {
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

func unmarshalSlugSet(ss *slugSet, prop string, w map[string]interface{}) error {
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
		ss.mode = blacklist
		switch ssn := ssni.(type) {
		case []interface{}:
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
		err := unmarshalStringSlice(path, ss, _ss)
		if err != nil {
			return err
		}
	case string:
		if _ss == "*" {
			ss.mode = whitelist
		}
		ss.elements = []string{_ss}
	default:
		return jsonInvalidType(path, ssi, jtypeObject, jtypeArray, jtypeString)
	}
	return nil
}

func unmarshalStringSlice(path []string, ss *slugSet, jarr []interface{}) error {
	ss.elements = make([]string, len(jarr))
	for i, v := range jarr {
		// TODO(nick): check for zero items in this slice
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
	return fmt.Sprintf(`invalid value for property "%s", expecting %s, got "%s"`, strings.Join(j.path, "."), j.expecting, j.got)
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
