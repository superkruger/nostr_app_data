package secrets

import (
	"encoding/json"
	"strings"
)

type mockService struct {
	source string
}

func NewMockService(source string) Service {
	return &mockService{source: source}
}

func (s *mockService) GetAndUnmarshal(_ string, v interface{}) error {
	return json.Unmarshal([]byte(s.source), v)
}

func (s *mockService) GetAndUnmarshalStrict(_ string, v interface{}) error {
	decoder := json.NewDecoder(strings.NewReader(s.source))
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

func (s *mockService) MustGetAndUnmarshal(_ string, v interface{}) {
	if err := json.Unmarshal([]byte(s.source), v); err != nil {
		panic(err)
	}
}
