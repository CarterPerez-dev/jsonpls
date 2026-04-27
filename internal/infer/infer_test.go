// ===================
// © AngelaMos | 2026
// infer_test.go
// ===================

package infer

import "testing"

func TestInferColumns(t *testing.T) {
	cases := []struct {
		name    string
		header  string
		values  []string
		want    ColumnType
	}{
		{"plain ints", "views", []string{"100", "250", "0"}, TypeInt},
		{"ints with empty", "likes", []string{"100", "", "5"}, TypeInt},
		{"comma thousands ints", "reach", []string{"1,000", "2,500"}, TypeInt},
		{"big int -> id-string (overflow)", "post_id", []string{"17960906015933057"}, TypeIDString},
		{"id-hinted small ints -> id-string", "user_id", []string{"1", "2", "3"}, TypeIDString},
		{"id mid-name", "internal_id_value", []string{"1", "2"}, TypeIDString},
		{"floats with decimal", "rate", []string{"1.5", "2.0", "3.14"}, TypeFloat},
		{"pure ints not floats", "count", []string{"1", "2", "3"}, TypeInt},
		{"bool true/false", "active", []string{"true", "false", "true"}, TypeBool},
		{"bool yes/no", "opted_in", []string{"yes", "no", "Yes"}, TypeBool},
		{"iso date", "created_at", []string{"2026-04-24T06:05:00Z"}, TypeDate},
		{"us date with time", "publish_time", []string{"04/24/2026 06:05", "04/25/2026 12:00"}, TypeDate},
		{"date-only", "day", []string{"2026-04-24", "2026-04-25"}, TypeDate},
		{"mixed -> string", "label", []string{"100", "abc", "200"}, TypeString},
		{"empty col -> string", "notes", []string{"", "", ""}, TypeString},
		{"date mixed with literal -> string", "date", []string{"Lifetime", "Lifetime"}, TypeString},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := inferOne(c.header, c.values)
			if got != c.want {
				t.Fatalf("inferOne(%q, %v) = %v, want %v", c.header, c.values, got, c.want)
			}
		})
	}
}

func TestCast(t *testing.T) {
	cases := []struct {
		name  string
		value string
		typ   ColumnType
		want  any
	}{
		{"empty -> nil", "", TypeInt, nil},
		{"whitespace -> nil", "  ", TypeString, nil},
		{"int", "42", TypeInt, int64(42)},
		{"int with commas", "1,000", TypeInt, int64(1000)},
		{"float", "3.14", TypeFloat, 3.14},
		{"bool yes", "yes", TypeBool, true},
		{"bool N", "N", TypeBool, false},
		{"id-string keeps raw", "00123", TypeIDString, "00123"},
		{"date normalized", "04/24/2026 06:05", TypeDate, "2026-04-24T06:05:00Z"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Cast(c.value, c.typ)
			if got != c.want {
				t.Fatalf("Cast(%q, %v) = %v (%T), want %v (%T)", c.value, c.typ, got, got, c.want, c.want)
			}
		})
	}
}
