package authr

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

type equalitytestscen struct {
	n    string
	a, b interface{}
	r    bool
}

func getEqualityTestScenarios() []equalitytestscen {
	return []equalitytestscen{
		{
			n: `"5"==uint32(5)=>true`,
			a: "5",
			b: uint32(5),
			r: true,
		},
		{
			n: `"hi"=="hi"=>true`,
			a: "hi",
			b: "hi",
			r: true,
		},
		{
			n: `"hi"=="hello"=>false`,
			a: "hi",
			b: "hello",
			r: false,
		},
		{
			n: `"hi"==true=>true`,
			a: "hi",
			b: true,
			r: true,
		},
		{
			n: "float64(3.1415)==float32(3.1415)=>true",
			a: float64(3.1415),
			b: float32(3.1415),
			r: true,
		},
		{
			n: "float32(0)==nil=>true",
			a: float32(0),
			b: nil,
			r: true,
		},
		{
			n: "int32(1)==true=>true",
			a: int32(1),
			b: true,
			r: true,
		},
		{
			n: "int16(0)==false=>true",
			a: int16(0),
			b: false,
			r: true,
		},
		{
			n: `""==nil=>true`,
			a: "",
			b: nil,
			r: true,
		},
		{
			n: `"hi"==nil=>false`,
			a: "hi",
			b: nil,
			r: false,
		},
		{
			n: `true=="0"=>false`,
			a: true,
			b: "0",
			r: false,
		},
		{
			n: `true==true=>true`,
			a: true,
			b: true,
			r: true,
		},
		{
			n: `true==false=>false`,
			a: true,
			b: false,
			r: false,
		},
		{
			n: `true==nil=>false`,
			a: true,
			b: nil,
			r: false,
		},
		{
			n: `false==nil=>true`,
			a: false,
			b: nil,
			r: true,
		},
		{
			n: `nil==nil=>true`,
			a: nil,
			b: nil,
			r: true,
		},
		{
			n: `false=>""=>true`,
			a: false,
			b: "",
			r: true,
		},
		{
			n: `false=>"0"=>true`,
			a: false,
			b: "0",
			r: true,
		},
	}
}

func BenchmarkLooseEquality(b *testing.B) {
	for _, s := range getEqualityTestScenarios() {
		b.Run(s.n, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := looseEquality(s.a, s.b)
				if err != nil {
					b.Fatalf("unexpected error: %s", err)
				}
			}
		})
	}
}

type regexptestscen struct {
	n, op, p string
	v        interface{}
	r        bool
}

func getregexpscens() []regexptestscen {
	return []regexptestscen{
		{n: "int(33)~*^foo$=>false", op: "~*", p: "^foo$", v: 33, r: false},
		{n: `"foo-one"~*^Foo=>true`, op: "~*", p: "^Foo", v: "foo-one", r: true},
		{n: `"bar-two"~^Bar=>false`, op: "~", p: "^Bar", v: "bar-two", r: false},
	}
}

func BenchmarkRegexpOperatorSerial(b *testing.B) {
	regexpOperatorBenchmark(b, func(fn func()) func(*testing.B) {
		return func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				fn()
			}
		}
	})
}

func BenchmarkRegexpOperatorParallel(b *testing.B) {
	regexpOperatorBenchmark(b, func(fn func()) func(*testing.B) {
		return func(b *testing.B) {
			b.ReportAllocs()
			b.SetParallelism(runtime.NumCPU())
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					fn()
				}
			})
		}
	})
}

func regexpOperatorBenchmark(b *testing.B, fn func(func()) func(*testing.B)) {
	for _, s := range getregexpscens() {
		op, ok := operators[s.op]
		if !ok {
			b.Fatalf("unknown operator: %s", s.op)
		}
		b.Run(s.n, fn(func() {
			_, err := op.compute(s.v, s.p)
			if err != nil {
				b.Fatalf("unexpected error: %s", err)
			}
		}))
	}
}

// This should test how the regexp cache responds to random access and eviction
func BenchmarkRegexpOperatorThrash(b *testing.B) {
	tests := getregexpscens()
	l := len(tests)
	r := rand.New(rand.NewSource(5))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t := tests[r.Intn(l)]
		op, ok := operators[t.op]
		if !ok {
			b.Fatalf("unknown operator: %s", t.op)
		}
		op.compute(t.v, t.p)
	}
}

func TestLooseEquality(t *testing.T) {
	scenarios := getEqualityTestScenarios()
	for _, s := range scenarios {
		t.Run(s.n, func(t *testing.T) {
			var (
				ok  bool
				err error
			)
			ok, err = looseEquality(s.a, s.b)
			require.Nil(t, err)
			require.Equal(t, s.r, ok)
			// flip the arguments
			ok, err = looseEquality(s.b, s.a)
			require.Nil(t, err)
			require.Equal(t, s.r, ok, "equality result was not equal when flipping arguments")
		})
	}
}

type testResource struct {
	rtype      string
	attributes map[string]interface{}

	rterr, raerr error // errors returned from either method
}

func (t testResource) GetResourceType() (string, error) {
	if t.rterr != nil {
		return "", t.rterr
	}
	return t.rtype, nil
}

func (t testResource) GetResourceAttribute(key string) (interface{}, error) {
	if t.raerr != nil {
		return nil, t.raerr
	}
	return t.attributes[key], nil
}

func TestInOperator(t *testing.T) {
	t.Parallel()
	tr := testResource{
		rtype: "user",
		attributes: map[string]interface{}{
			"id":     int32(23),
			"groups": []string{"alpha", "bravo"},
			"status": "active",
		},
	}
	t.Run("should loosely match id attribute in polymorphic slice", func(t *testing.T) {
		cond := Cond("@id", "$in", []interface{}{1, "31", "55", float64(23)})
		ok, err := cond.evaluate(tr)
		require.Nil(t, err, "unexpected error")
		require.True(t, ok)
	})
	t.Run("should return err when right operand is scalar", func(t *testing.T) {
		_, err := Cond("@id", "$in", 5).evaluate(tr)
		require.NotNil(t, err)
	})
	t.Run("should evaluate to false when value not found", func(t *testing.T) {
		ok, err := Cond("foo", "$in", "@groups").evaluate(tr)
		require.Nil(t, err, "unexpected error")
		require.False(t, ok)
	})
}

func TestNotInOperator(t *testing.T) {
	t.Parallel()
	tr := testResource{
		rtype: "post",
		attributes: map[string]interface{}{
			"tags":    []string{"one", "two"},
			"id":      int32(345),
			"user_id": int32(23),
		},
	}
	t.Run("should loosely match id attribute in polymorphic slice", func(t *testing.T) {
		ok, err := Cond("@id", "$nin", []interface{}{1, "31", "55", float64(23)}).evaluate(tr)
		require.Nil(t, err)
		require.True(t, ok)
	})
	t.Run("should return err when right operand is scalar", func(t *testing.T) {
		_, err := Cond("@user_id", "$nin", map[int]int{4: 2}).evaluate(tr)
		if err == nil {
			t.Errorf("test expected an error, got nil")
		}
	})
	t.Run("should evaluate to false when value found in array/slice", func(t *testing.T) {
		ok, err := Cond("two", "$nin", "@tags").evaluate(tr)
		if err != nil {
			t.Errorf("test failed with unexpected error: %s", err)
		} else if ok {
			t.Errorf("test failed")
		}
	})
}

func TestIntersectOperator(t *testing.T) {
	t.Parallel()
	r := testResource{
		rtype: "user",
		attributes: map[string]interface{}{
			"tags":       []string{"one", "two"},
			"is_serious": true,
		},
	}
	t.Run("should return false when arrays do not intersect", func(t *testing.T) {
		ok, err := Cond("@tags", "&", []interface{}{1.0, 2}).evaluate(r)
		assertNilError(t, err)
		assertNotOkay(t, ok)
	})
	t.Run("should return true when arrays do intersect", func(t *testing.T) {
		ok, err := Cond("@tags", "&", []interface{}{2, "one"}).evaluate(r)
		assertNilError(t, err)
		assertOkay(t, ok)
	})
	t.Run("should return err when left operand is not array-ish", func(t *testing.T) {
		_, err := Cond("@is_serious", "&", []int{1, 2}).evaluate(r)
		assertError(t, err)
	})
	t.Run("should return err when right operand is not array-ish", func(t *testing.T) {
		_, err := Cond([]int{2, 1}, "&", "@is_serious").evaluate(r)
		assertError(t, err)
	})
}

func TestDifferenceOperator(t *testing.T) {
	t.Parallel()
	r := testResource{
		rtype: "account",
		attributes: map[string]interface{}{
			"groups":  []string{"pro", "22.56"},
			"balance": float64(23.123),
		},
	}
	t.Run("should return true when arrays do not intersect", func(t *testing.T) {
		ok, err := Cond("@groups", "-", []string{"ent"}).evaluate(r)
		assertNilError(t, err)
		assertOkay(t, ok)
	})
	t.Run("should return false when arrays do intersect", func(t *testing.T) {
		ok, err := Cond("@groups", "-", []interface{}{float32(22.56)}).evaluate(r)
		assertNilError(t, err)
		assertNotOkay(t, ok)
	})
	t.Run("should return err when left operand is not array-sh", func(t *testing.T) {
		_, err := Cond("@balance", "-", []string{"23.123"}).evaluate(r)
		assertError(t, err)
	})
	t.Run("should return err when right operand is not array-ish", func(t *testing.T) {
		_, err := Cond([]string{"pop"}, "-", "@balance").evaluate(r)
		assertError(t, err)
	})
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func assertNilError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func assertNotOkay(t *testing.T, ok bool) {
	t.Helper()
	if ok {
		t.Fatalf("unexpected okay-ness")
	}
}

func assertOkay(t *testing.T, ok bool) {
	t.Helper()
	if !ok {
		t.Fatalf("unexpected non-okay-ness")
	}
}

func TestLikeOperator(t *testing.T) {
	t.Parallel()
	tr := testResource{
		rtype: "cart",
		attributes: map[string]interface{}{
			"name": "linda's cart",
			"tag":  "wish_list",
		},
	}
	t.Run("should match beginning of string", func(t *testing.T) {
		ok, err := Cond("@name", "~=", "Linda*").evaluate(tr)
		if err != nil {
			t.Errorf("test failed with unexpected error: %s", err)
		} else if !ok {
			t.Errorf("test failed")
		}
	})
	t.Run("should not match a string that does NOT end with a specifed pattern", func(t *testing.T) {
		ok, err := Cond("@tag", "~=", "*bla").evaluate(tr)
		if err != nil {
			t.Errorf("test failed with unexpected error: %s", err)
		} else if ok {
			t.Errorf("test failed")
		}
	})
}

type testSubject struct {
	err   error
	rules []*Rule
}

func (t testSubject) GetRules() ([]*Rule, error) {
	if t.err != nil {
		return nil, t.err
	}
	return t.rules, nil
}

func TestFull(t *testing.T) {
	actor := &testSubject{
		rules: []*Rule{
			new(Rule).Access(Deny).Where(
				Action("delete"),
				ResourceType("zone"),
				ResourceMatch(Cond("@attr", "!=", nil)),
			),
			new(Rule).Access(Allow).Where(
				Action("delete"),
				ResourceType("zone"),
				ResourceMatch(
					Or(
						Cond("@id", "=", 321),
						Cond("@zone_name", "~*", `\.com$`),
					),
				),
			),
		},
	}
	resource := testResource{
		rtype: "zone",
		attributes: map[string]interface{}{
			"id":        "123",
			"zone_name": "example.com",
		},
	}
	ok, err := Can(actor, "delete", resource)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !ok {
		t.Fatalf("unexpected access denial")
	}
}

type testCan_case struct {
	g, s     string
	subject  Subject
	act      string
	resource Resource
	errcheck func(error) bool
	ok       bool

	noskip bool // for debugging tests
}

func testCan_getCases() []testCan_case {
	sub := func(r []*Rule) Subject {
		return testSubject{rules: r}
	}
	jsonlist := func(r ...string) Subject {
		rules := make([]*Rule, len(r))
		for i := 0; i < len(r); i++ {
			rule := new(Rule)
			if err := json.Unmarshal([]byte(r[i]), rule); err != nil {
				panic(err.Error())
			}
			rules[i] = rule
		}
		if rules[0].access == "" {
			panic("something went wrong when unmarshaling JSON rules for tests")
		}
		return sub(rules)
	}
	msi := func(a ...interface{}) map[string]interface{} {
		o := make(map[string]interface{})
		for i, v := range a {
			if i%2 != 0 {
				o[a[i-1].(string)] = v
			}
		}
		return o
	}
	testerr := errors.New("testerr")
	return []testCan_case{
		{
			g:        "an error being returned from subject.GetRules()",
			s:        "return error",
			subject:  testSubject{err: testerr},
			act:      "testcan1",
			resource: testResource{rtype: "thing"},
			errcheck: func(e error) bool { return e == testerr },
		},
		{
			g:        "an error being returned from resource.GetResourceType()",
			s:        "return error",
			subject:  testSubject{rules: []*Rule{}},
			act:      "testcan2",
			resource: testResource{rterr: testerr},
			errcheck: func(e error) bool { return e == testerr },
		},
		{
			g:        "a subject with NO rules",
			s:        "default to deny all",
			subject:  testSubject{rules: []*Rule{}},
			act:      "testcan3",
			resource: testResource{rtype: "thing", attributes: msi("id", 5)},
			ok:       false,
		},
		{
			g: "a subject with no rsrc_type matching rule",
			s: "deny",
			subject: jsonlist(
				`{"access":"allow","where":{"rsrc_type":"thing","rsrc_match":[],"action":"testcan4"}}`,
			),
			act:      "testcan4",
			resource: testResource{rtype: "widget" /* <- different! */, attributes: msi("id", 5)},
			ok:       false,
		},
		{
			g: "a subject with no action matching rule",
			s: "deny",
			subject: jsonlist(
				`{"access":"allow","where":{"rsrc_type":"thing","rsrc_match":[],"action":"NOTtestcan6"}}`,
			),
			act:      "testcan6",
			resource: testResource{rtype: "thing" /* <- same! */, attributes: msi("id", 5)},
			ok:       false,
		},
		{
			g: "a subject with no resource attribute matching rule",
			s: "deny",
			subject: jsonlist(
				`{"access":"allow","where":{"rsrc_type":"thing","rsrc_match":[["@id","=",3]],"action":"testcan7"}}`,
			),
			act:      "testcan7",
			resource: testResource{rtype: "thing", attributes: msi("id", 5)},
			ok:       false,
		},
		{
			g: "a subject with a matching rule that denies",
			s: "deny",
			subject: jsonlist(
				`{"access":"deny","where":{"rsrc_type":"thing","rsrc_match":[["@id","=",5]],"action":"testcan7"}}`,
			),
			act:      "testcan7",
			resource: testResource{rtype: "thing", attributes: msi("id", 5)},
			ok:       false,
		},
		{
			g: "a subject with a matching rule that allows",
			s: "allow",
			subject: jsonlist(
				`{"access":"allow","where":{"rsrc_type":"thing","rsrc_match":[["@id","=",5]],"action":"testcan7"}}`,
			),
			act:      "testcan7",
			resource: testResource{rtype: "thing", attributes: msi("id", 5)},
			ok:       true,
		},
		{
			g: "a subject with a blocklist",
			s: "not match the provided action",
			subject: sub([]*Rule{
				new(Rule).
					Access(Allow).
					Where(
						Not(Action("delete")),
						ResourceType("thing"),
						ResourceMatch(Cond("@id", "=", 5)),
					),
			}),
			act:      "delete",
			resource: testResource{rtype: "thing", attributes: msi("id", 5)},
			ok:       false,
		},
	}
}

func BenchmarkCan(b *testing.B) {
	for _, c := range testCan_getCases() {
		b.Run(fmt.Sprintf("given %s, Can() should %s", c.g, c.s), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Can(c.subject, c.act, c.resource)
			}
		})
	}
}

func TestCan(t *testing.T) {
	for _, c := range testCan_getCases() {
		t.Run(fmt.Sprintf("given %s, Can() should %s", c.g, c.s), func(t *testing.T) {
			// Set this env var and put the "noskip: true" on whatever test you
			// want to concentrate on :)
			if os.Getenv("TEST_CAN_SKIP") != "" && !c.noskip {
				t.SkipNow()
				return
			}
			ok, err := Can(c.subject, c.act, c.resource)
			if err != nil {
				require.NotNil(t, c.errcheck, "unexpected error returned: %s", err.Error())
				require.True(t, c.errcheck(err), "error returned from Can() did not match expected error")
				require.False(t, ok, "Can() returned an error AND true, this should never happen")
				return
			}
			require.Nil(t, c.errcheck, "expected error to be returned, none returned")
			require.Equal(t, c.ok, ok, "Can() returned wrong result (no error)")
		})
	}
}
