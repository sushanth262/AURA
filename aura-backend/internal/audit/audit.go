package audit

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"time"
)

type DecisionRecord struct {
	Timestamp      string `json:"timestamp"`
	DecisionID     string `json:"decision_id"`
	PolicyVersion  string `json:"policy_version"`
	Allowed        bool   `json:"allowed"`
	Reason         string `json:"reason"`
	Subject        string `json:"subject"`
	Action         string `json:"action"`
	Resource       string `json:"resource"`
	Route          string `json:"route,omitempty"`
	TraceID        string `json:"trace_id,omitempty"`
}

var lg = log.New(os.Stdout, "", 0)

func randomDecisionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func LogAuthz(r DecisionRecord) DecisionRecord {
	if r.DecisionID == "" {
		r.DecisionID = randomDecisionID()
	}
	if r.Timestamp == "" {
		r.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	b, _ := json.Marshal(r)
	lg.Println(string(b))
	return r
}
