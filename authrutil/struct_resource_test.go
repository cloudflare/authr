package authrutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStructResource(t *testing.T) {
	t.Run("should panic if given a non-struct value", func(t *testing.T) {
		require.Panics(t, func() {
			StructResource("a", 5)
		})
	})
	t.Run("should return the provided resource type", func(t *testing.T) {
		rt, err := StructResource("a", struct{}{}).GetResourceType()
		require.Nil(t, err)
		require.Equal(t, "a", rt)
	})
	t.Run("should retrieve correct values from struct", func(t *testing.T) {
		sr := StructResource("thing", struct {
			Foo int
			Bar string
		}{Foo: 5, Bar: "boom!"})
		avfoo, err := sr.GetResourceAttribute("Foo")
		require.Nil(t, err)
		require.Equal(t, 5, avfoo.(int))
		avbar, err := sr.GetResourceAttribute("Bar")
		require.Nil(t, err)
		require.Equal(t, "boom!", avbar.(string))
	})
	t.Run("should return <nil> for nonexistent struct fields", func(t *testing.T) {
		sr := StructResource("thing", struct {
			Foo int
			Bar string
		}{Foo: 7, Bar: "bam!"})
		avne, err := sr.GetResourceAttribute("Baz")
		require.Nil(t, err)
		require.Nil(t, avne)
	})
	t.Run("should not be able to read un-exported stuct fields", func(t *testing.T) {
		sr := StructResource("thing", struct {
			foo int
		}{foo: 9})
		avnil, err := sr.GetResourceAttribute("foo")
		require.Nil(t, err)
		require.Nil(t, avnil)
	})
}
