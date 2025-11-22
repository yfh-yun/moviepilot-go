package actions

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// ActionParams 定义所有参数结构的最小约束。
type ActionParams interface {
	Validate() error
}

// BaseParams 提供空实现，嵌入到具体参数结构中即可。
type BaseParams struct{}

// Validate 默认返回 nil，由需要约束的字段自行覆盖。
func (BaseParams) Validate() error { return nil }

// DecodeParams 将任意入参解码为 ActionParams。
func DecodeParams(input any, out ActionParams) error {
	if out == nil {
		return fmt.Errorf("params target cannot be nil")
	}
	if input == nil {
		return out.Validate()
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		Result:           out,
		WeaklyTypedInput: true,
		ZeroFields:       true,
	})
	if err != nil {
		return err
	}

	if err := decoder.Decode(input); err != nil {
		return err
	}

	return out.Validate()
}
