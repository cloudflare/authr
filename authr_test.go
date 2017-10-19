package authr

import (
	"testing"
)

func TestLooseEquality(t *testing.T) {
	scenarios := []struct {
		n    string
		a, b interface{}
		r    bool
	}{
		{
			n: "'5' == uint32(5) => true",
			a: "5",
			b: uint32(5),
			r: true,
		},
		{
			n: "'hi' == 'hi' => true",
			a: "hi",
			b: "hi",
			r: true,
		},
		{
			n: "'hi' != 'hello' => false",
			a: "hi",
			b: "hello",
			r: false,
		},
		{
			n: "'hi' == true => true",
			a: "hi",
			b: true,
			r: true,
		},
		{
			n: "float64(3.141519) == float32(3.141519) => true",
			a: float64(3.141519),
			b: float32(3.141519),
			r: true,
		},
		{
			n: "float32(0) == nil => true",
			a: float32(0),
			b: nil,
			r: true,
		},
		{
			n: "int32(1) == true => true",
			a: int32(1),
			b: true,
			r: true,
		},
		{
			n: "int16(0) == false => true",
			a: int16(0),
			b: false,
			r: true,
		},
		{
			n: "'' => nil => true",
			a: "",
			b: nil,
			r: true,
		},
		{
			n: "'hi' => nil => false",
			a: "hi",
			b: nil,
			r: false,
		},
	}

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
