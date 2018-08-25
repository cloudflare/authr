package authr

import (
	"math/rand"
	"runtime"
	"testing"
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
			if err != nil {
				t.Errorf("error: %s", err)
				return
			} else {
				if s.r != ok {
					t.Errorf("failed asserting that %t == %t", s.r, ok)
					return
				}
			}
			// flip the arguments
			ok, err = looseEquality(s.b, s.a)
			if err != nil {
				t.Errorf("error: %s", err)
				return
			} else {
				if s.r != ok {
					t.Errorf("failed asserting that %t == %t (arg flip)", s.r, ok)
					return
				}
			}
		})
	}
}

type testResource struct {
	rtype      string
	attributes map[string]interface{}
}

func (t testResource) GetResourceType() (string, error) {
	return t.rtype, nil
}

func (t testResource) GetResourceAttribute(key string) (interface{}, error) {
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
		if err != nil {
			t.Errorf("test failed with unexpected error: %s", err)
		} else if !ok {
			t.Errorf("test failed")
		}
	})
	t.Run("should return err when right operand is scalar", func(t *testing.T) {
		_, err := Cond("@id", "$in", 5).evaluate(tr)
		if err == nil {
			t.Errorf("test failed, expected error, found nil")
		}
	})
	t.Run("should evaluate to false when value not found", func(t *testing.T) {
		ok, err := Cond("foo", "$in", "@groups").evaluate(tr)
		if err != nil {
			t.Errorf("test failed with unexpected error: %s", err)
		} else if ok {
			t.Errorf("test failed with unexpected true return false")
		}
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
		if err != nil {
			t.Errorf("test failed with unexpected error: %s", err)
		} else if !ok {
			t.Errorf("test failed")
		}
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
	rules []*Rule
}

func (t testSubject) GetRules() ([]*Rule, error) {
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
