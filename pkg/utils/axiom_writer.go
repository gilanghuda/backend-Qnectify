package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

type AxiomWriter struct{}

var _ io.Writer = (*AxiomWriter)(nil)

var EmitLogFunc = func(ctx context.Context, level, message string, attrs map[string]string) error {
	return nil
}

func (w *AxiomWriter) Write(p []byte) (n int, err error) {
	var data map[string]interface{}
	if err := json.Unmarshal(p, &data); err != nil {
		return len(p), nil
	}

	level := ""
	if lv, ok := data["level"].(string); ok {
		level = lv
	}
	msg := ""
	if m, ok := data["message"].(string); ok {
		msg = m
	}

	attrs := map[string]string{}
	for k, v := range data {
		if k == "level" || k == "message" || k == "time" {
			continue
		}
		switch val := v.(type) {
		case string:
			attrs[k] = val
		default:
			attrs[k] = fmt.Sprintf("%v", val)
		}
	}

	go func() {
		_ = EmitLogFunc(context.Background(), level, msg, attrs)
	}()

	return len(p), nil
}
