package cachex

import "encoding/json"

type ValueCodec[V any] interface {
	Encode(v V) (string, error)
	Decode(raw string) (V, error)
}

type JSONCodec[V any] struct{}

func (c JSONCodec[V]) Encode(v V) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c JSONCodec[V]) Decode(raw string) (V, error) {
	var v V
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return v, err
	}
	return v, nil
}