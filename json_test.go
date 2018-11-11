package authr

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type unmarshalScenario struct {
	n, d, err string
	r         *Rule
}

func unmarshalScenarios() []unmarshalScenario {
	return []unmarshalScenario{
		{
			n:   `should err; totally invalid JSON`,
			d:   `[{111`,
			err: "invalid character '1' looking for beginning of object key string",
		},
		{
			n:   `should err; invalid JSON type for rule def`,
			d:   `[1, 2, 3]`,
			err: "expecting JSON object for rule definition, got JSON array",
		},
		{
			n:   `should err; missing "where" property`,
			d:   `{"access":"allow"}`,
			err: `invalid rule; missing required property "where"`,
		},
		{
			n:   `should err; missing "where" property`,
			d:   `{"access":"deny"}`,
			err: `invalid rule; missing required property "where"`,
		},
		{
			n:   `should err; invalid json type for "where" prop`,
			d:   `{"access":"deny","where":4}`,
			err: `expecting JSON object for property "where", got JSON number`,
		},
		{
			n:   `should err; missing "where.rsrc_type" prop`,
			d:   `{"access":"deny","where":{}}`,
			err: `invalid rule; missing required property "where.rsrc_type"`,
		},
		{
			n:   `should err; invalid json type for "where.rsrc_type" prop`,
			d:   `{"access":"deny","where":{"rsrc_type":4}}`,
			err: `expecting JSON object, JSON array or JSON string for property "where.rsrc_type", got JSON number`,
		},
		{
			n:   `should err; invalid value for "where.rsrc_type" prop, missing "$not"`,
			d:   `{"access":"deny","where":{"rsrc_type":{}}}`,
			err: `invalid value for property "where.rsrc_type": expected JSON object with only one of the these key(s): "$not"`,
		},
		{
			n:   `should err; invalid value for "where.rsrc_type" prop, extra keys`,
			d:   `{"access":"deny","where":{"rsrc_type":{"$not":[],"$foo":3}}}`,
			err: `invalid value for property "where.rsrc_type": expected JSON object with only one of the these key(s): "$not"`,
		},
		{
			n:   `should err; invalid value for "where.rsrc_type.$not" prop`,
			d:   `{"access":"deny","where":{"rsrc_type":{"$not":4}}}`,
			err: `expecting JSON array or JSON string for property "where.rsrc_type.$not", got JSON number`,
		},
		{
			n:   `should err; invalid value for "where.rsrc_type.$not" prop, polymorphic array`,
			d:   `{"access":"deny","where":{"rsrc_type":{"$not":["foo",5]}}}`,
			err: `expecting JSON string for property "where.rsrc_type.$not.1", got JSON number`,
		},
		{
			n:   `should err; invalid value for "where.rsrc_type", polymorphic array`,
			d:   `{"access":"deny","where":{"rsrc_type":["foo",false]}}`,
			err: `expecting JSON string for property "where.rsrc_type.1", got JSON boolean`,
		},
		{
			n:   `should err; but "where.rsrc_type" prop ok, case 1`,
			d:   `{"access":"deny","where":{"rsrc_type":"zone"}}`,
			err: `invalid rule; missing required property "where.action"`,
		},
		{
			n:   `should err; but "where.rsrc_type" prop ok, case 2`,
			d:   `{"access":"deny","where":{"rsrc_type":["zone","dns_record"]}}`,
			err: `invalid rule; missing required property "where.action"`,
		},
		{
			n:   `should err; but "where.rsrc_type" prop ok, case 3`,
			d:   `{"access":"deny","where":{"rsrc_type":{"$not":"zone"}}}`,
			err: `invalid rule; missing required property "where.action"`,
		},
		{
			n:   `should err; missing "where.rsrc_match" prop`,
			d:   `{"access":"deny","where":{"action":"delete","rsrc_type":"zone"}}`,
			err: `invalid rule; missing required property "where.rsrc_match"`,
		},
		{
			n:   `should err; invalid value for "where.rsrc_match" prop`,
			d:   `{"access":"deny","where":{"action":"delete","rsrc_type":"zone","rsrc_match":4}}`,
			err: `expecting JSON object or JSON array for property "where.rsrc_match", got JSON number`,
		},
		{
			n:   `should err; invalid value for "where.rsrc_match" prop`,
			d:   `{"access":"deny","where":{"action":"delete","rsrc_type":"zone","rsrc_match":{}}}`,
			err: `invalid value for property "where.rsrc_match": expected JSON object with only one of the these key(s): "$and", "$or"`,
		},
		{
			n:   `should err; invalid value for "where.rsrc_match" prop`,
			d:   `{"access":"deny","where":{"action":"delete","rsrc_type":"zone","rsrc_match":{"$not":[]}}}`,
			err: `invalid value for property "where.rsrc_match": expected JSON object with only one of the these key(s): "$and", "$or"`,
		},
		{
			n:   `should err; missing "access" property`,
			d:   `{"where":{"action":"delete","rsrc_type":"zone","rsrc_match":[]}}`,
			err: `invalid rule; missing required property "access"`,
		},
		{
			n:   `should err; invalid "access" prop type`,
			d:   `{"access":2,"where":{"action":"delete","rsrc_type":"zone","rsrc_match":[]}}`,
			err: `expecting JSON string for property "access", got JSON number`,
		},
		{
			n:   `should err; invalid "access" property`,
			d:   `{"access":"allw","where":{"action":"delete","rsrc_type":"zone","rsrc_match":[]}}`,
			err: `invalid value for property "access", expecting "allow" or "deny", got "allw"`,
		},
		{
			n:   `should err; invalid "where.rsrc_type" prop`,
			d:   `{"access":"deny","where":{"action":"delete","rsrc_type":[],"rsrc_match":[["@id","&",[1,2,3]]]}}`,
			err: `invalid value for property "where.rsrc_type", expecting non-empty array, got empty array`,
		},
		{
			n: "ok case 1",
			d: `{"access":"deny","where":{"action":"delete","rsrc_type":"zone","rsrc_match":[["@id","&",[1,2,3]]]}}`,
			r: new(Rule).
				Access(Deny).
				Where(
					Action("delete"),
					ResourceType("zone"),
					ResourceMatch(
						Cond("@id", "&", []interface{}{
							float64(1),
							float64(2),
							float64(3),
						}),
					),
				),
		},
		{
			n: "ok case 2",
			d: `{"access":"allow","where":{"action":{"$not":["delete","update"]},"rsrc_type":"zone","rsrc_match":[{"$or":[["@id","&",[1,2,3]],["@status","$in",["A","V"]]]}]}}`,
			r: new(Rule).
				Access(Allow).
				Where(
					Action("delete", "update").Not(),
					ResourceType("zone"),
					ResourceMatch(
						Or(
							Cond("@id", "&", []interface{}{
								float64(1),
								float64(2),
								float64(3),
							}),
							Cond("@status", "$in", []interface{}{"A", "V"}),
						),
					),
				),
		},
	}
}

func TestRuleUnmarshalJSON(t *testing.T) {
	for _, s := range unmarshalScenarios() {
		t.Run(s.n, func(t *testing.T) {
			r := new(Rule)
			err := json.Unmarshal([]byte(s.d), r)
			if err != nil {
				if s.err == "" {
					t.Fatalf("unexpected error: %s", err.Error())
				} else {
					if s.err != err.Error() {
						t.Fatalf(`error expectation failed: "%s" != "%s"`, s.err, err.Error())
					}
				}
				return
			}
			require.Equal(t, s.r, r)
		})
	}
}
