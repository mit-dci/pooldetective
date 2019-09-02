package util

import (
	"encoding/json"
	"fmt"
	"time"
)

type StratumJob struct {
	ID              string
	Difficulty      float64
	ExtraNonce1     string
	ExtraNonce2Size int
	JobData         []interface{}
}

func NewStratumJobFromJsonProperty(i interface{}) StratumJob {
	job := StratumJob{}
	b, _ := json.Marshal(i)
	json.Unmarshal(b, &job)
	return job
}

func (job StratumJob) AsJobData() []interface{} {
	return append([]interface{}{job.ID}, job.JobData[1:]...)
}

func (job StratumJob) AsJobDataWithNewTimestamp() []interface{} {
	j := job.AsJobData()
	return append(j[:7], fmt.Sprintf("%x", time.Now().Unix()), true)
}

type BlockHeader struct {
	Bits           string `json:"bits" bson:"bits"`
	Hash           string `json:"hash" bson:"hash"`
	Height         int64  `json:"height" bson:"height"`
	PrevHash       string `json:"previousblockhash" bson:"previousblockhash"`
	MedianTime     uint64 `json:"medianTime" bson:"medianTime"`
	Time           uint64 `json:"time" bson:"time"`
	Coin           string `json:"coin" bson:"coin"`
	TimeDiscovered uint64 `json:"timeDiscovered" bson:"timeDiscovered"`
}

type HubMessage struct {
	Type    string                 `json:"type"`
	Content map[string]interface{} `json:"content"`
}

type StratumClientIdentity struct {
	User             string
	Host             string
	Algorithm        string
	PoolName         string
	PoolCoin         string
	ObserverLocation string
}

type StratumServerIdentity struct {
	Algorithm string
}
