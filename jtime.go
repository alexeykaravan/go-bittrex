package bittrex

import (
	"encoding/json"
	"time"
)

//TIMEFORMAT const
const TIMEFORMAT = "2006-01-02T15:04:05.999Z"

type jTime struct {
	time.Time
}

func (jt *jTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse(TIMEFORMAT, s)
	if err != nil {
		return err
	}
	jt.Time = t
	return nil
}

func (jt jTime) MarshalJSON() ([]byte, error) {
	return json.Marshal((*time.Time)(&jt.Time).Format(TIMEFORMAT))
}
