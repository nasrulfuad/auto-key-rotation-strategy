package apidoctest

import (
	"reflect"
	"strings"
	"testing"
)

func AssertShapeMatches(t *testing.T, mirror, real reflect.Type) {
	t.Helper()

	if mirror.Kind() != reflect.Struct || real.Kind() != reflect.Struct {
		if mirror != real {
			t.Fatalf("type mismatch: mirror %s != real %s", mirror, real)
		}
		return
	}

	if mirror.NumField() != real.NumField() {
		t.Fatalf("field count mismatch: mirror %s has %d field(s), real %s has %d",
			mirror, mirror.NumField(), real, real.NumField())
	}

	for i := 0; i < mirror.NumField(); i++ {
		mf, rf := mirror.Field(i), real.Field(i)

		mTag, rTag := jsonName(mf), jsonName(rf)
		if mTag != rTag {
			t.Fatalf("field %d json tag mismatch: mirror %q real %q", i, mTag, rTag)
		}

		switch mf.Type.Kind() {
		case reflect.Struct:
			AssertShapeMatches(t, mf.Type, rf.Type)
		case reflect.Slice, reflect.Pointer:
			AssertShapeMatches(t, mf.Type.Elem(), rf.Type.Elem())
		default:
			if mf.Type != rf.Type {
				t.Fatalf("field %q type mismatch: mirror %s real %s", mTag, mf.Type, rf.Type)
			}
		}
	}
}

func jsonName(f reflect.StructField) string {
	name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
	if name == "" {
		return f.Name
	}
	return name
}
