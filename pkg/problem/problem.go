package problem

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
)

type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func New(typeStr string, title string, status int, detail, instance string) *Problem {
	return &Problem{typeStr, title, status, detail, instance}
}

func (p *Problem) Send(resp http.ResponseWriter) {
	log.Printf("### 💥 API %s", p.Error())
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(p.Status)
	_ = json.NewEncoder(resp).Encode(p)
}

func Wrap(status int, typeStr string, instance string, err error) *Problem {
	var p *Problem
	if err != nil {
		p = New(typeStr, MyCaller(), status, err.Error(), instance)
	} else {
		p = New(typeStr, MyCaller(), status, "Other error occurred", instance)
	}

	return p
}

func (p Problem) Error() string {
	return fmt.Sprintf("Problem: Type: '%s', Title: '%s', Status: '%d', Detail: '%s', Instance: '%s'",
		p.Type, p.Title, p.Status, p.Detail, p.Instance)
}

func getFrame(skipFrames int) runtime.Frame {
	targetFrameIndex := skipFrames + 2

	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}

	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])

		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()

			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}

func MyCaller() string {
	return getFrame(2).Function
}
