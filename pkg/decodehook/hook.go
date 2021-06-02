package decodehook

import (
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/durationpb"
	"reflect"
	"time"
)

// StringToTimeDurationHookFunc returns a DecodeHookFunc that converts
// strings to time.Duration.
func StringToTimeDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(&duration.Duration{}) {
			return data, nil
		}
		// Convert it by parsing
		d, err := time.ParseDuration(data.(string))
		if err == nil {
			return durationpb.New(d), nil
		}
		return data, err
	}
}
