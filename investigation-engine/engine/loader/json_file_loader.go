package loader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"rootcause.ai/investigation-engine/pkg/investigation"
)

type JSONFileLoader struct{}

func NewJSONFileLoader() JSONFileLoader {
	return JSONFileLoader{}
}

func (JSONFileLoader) LoadIncident(ctx context.Context, path string) (investigation.Incident, error) {
	var incident investigation.Incident

	if err := ctx.Err(); err != nil {
		return incident, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return incident, fmt.Errorf("read incident file %q: %w", path, err)
	}

	if err := ctx.Err(); err != nil {
		return incident, err
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&incident); err != nil {
		return incident, fmt.Errorf("decode incident %q: %w", path, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return incident, fmt.Errorf("decode incident %q: unexpected trailing JSON value", path)
		}
		return incident, fmt.Errorf("decode incident %q: %w", path, err)
	}

	if err := incident.Validate(); err != nil {
		return incident, fmt.Errorf("validate incident %q: %w", path, err)
	}

	return incident, nil
}
