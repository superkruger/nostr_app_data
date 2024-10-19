package env

import (
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/pascaldekloe/goe/verify"
)

func TestGetStringsOrDefault(t *testing.T) {
	cases := map[string]struct {
		name  string
		value string
		def   []string
		want  []string
	}{
		"list": {"ENV_TEST", "aa,bb,cc", nil, []string{"aa", "bb", "cc"}},
		"def":  {"ENV_TEST", "", []string{"aa", "bb"}, []string{"aa", "bb"}},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			_ = os.Setenv(testCase.name, testCase.value)
			got := GetStringsOrDefault(testCase.name, testCase.def)
			if diff := deep.Equal(got, testCase.want); diff != nil {
				t.Errorf("incorrect result\n...got %+v\n..want %+v", got, testCase.want)
			}
		})
	}
	os.Clearenv()
}

func TestMustGetStrings(t *testing.T) {
	cases := map[string]struct {
		name  string
		value string
		want  []string
	}{
		"list": {"ENV_TEST", "aa,bb,cc", []string{"aa", "bb", "cc"}},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			_ = os.Setenv(testCase.name, testCase.value)
			got := MustGetStrings(testCase.name)
			if diff := deep.Equal(got, testCase.want); diff != nil {
				t.Errorf("incorrect result\n...got %+v\n..want %+v", got, testCase.want)
			}
		})
	}
	os.Clearenv()
}

func TestGetIntOrDefault(t *testing.T) {
	cases := map[string]struct {
		name    string
		value   string
		def     int
		want    int
		wantErr bool
	}{
		"val": {"ENV_TEST", "123", 10, 123, false},
		"def": {"ENV_TEST", "", 10, 10, false},
		"err": {"ENV_TEST", "foo", 10, 0, true},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			_ = os.Setenv(testCase.name, testCase.value)
			got, err := GetIntOrDefault(testCase.name, testCase.def)
			verify.Values(t, name, err != nil, testCase.wantErr)
			verify.Values(t, name, got, testCase.want)
		})
	}
	os.Clearenv()
}
