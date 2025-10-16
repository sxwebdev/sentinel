package ptconverts

import (
	"encoding/json"
	"errors"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ConvertAnyFromProto converts a proto message to a specific type.
func ConvertAnyFromProto[T any](m proto.Message) (T, error) {
	var zero T
	if m == nil {
		return zero, errors.New("empty proto message")
	}

	b, err := protojson.Marshal(m)
	if err != nil {
		return zero, err
	}

	err = json.Unmarshal(b, &zero)
	if err != nil {
		return zero, err
	}

	return zero, nil
}

// ConvertAnyToProto converts a specific type to a proto message.
func ConvertAnyToProto[T any](v T, m proto.Message) error {
	if m == nil {
		return errors.New("empty proto message")
	}

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	err = protojson.Unmarshal(b, m)
	if err != nil {
		return err
	}

	return nil
}
