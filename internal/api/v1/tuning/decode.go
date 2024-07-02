package tuning

import (
	"encoding/json"
	"io"

	definition "github.com/bigstack-oss/cube-cos-api/internal/definition/v1"
)

func decodeTuningReq(reqBody io.ReadCloser) (*definition.Tuning, error) {
	b, err := io.ReadAll(reqBody)
	if err != nil {
		return nil, err
	}

	tuning := definition.Tuning{}
	err = json.Unmarshal(b, &tuning)
	if err != nil {
		return nil, err
	}

	tuning.SetNodeInfo(
		definition.CurrentRole,
		definition.AdvertiseAddr,
	)

	return &tuning, nil
}

func decodeTuningsReq(reqBody io.ReadCloser) ([]definition.Tuning, error) {
	b, err := io.ReadAll(reqBody)
	if err != nil {
		return nil, err
	}

	tunings := []definition.Tuning{}
	err = json.Unmarshal(b, &tunings)
	if err != nil {
		return nil, err
	}

	for i := range tunings {
		tunings[i].SetNodeInfo(
			definition.CurrentRole,
			definition.AdvertiseAddr,
		)
	}

	return tunings, nil
}
