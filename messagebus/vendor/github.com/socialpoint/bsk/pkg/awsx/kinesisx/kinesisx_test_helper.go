package kinesisx

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/socialpoint/bsk/pkg/awsx/awstest"
	"github.com/socialpoint/bsk/pkg/uuid"
	"github.com/stretchr/testify/require"
)

func init() {
	//to avoid that TestKinesis always chooses the same prefixes an get a collision in travis
	rand.Seed(time.Now().UnixNano())
}

// NewKinesisForTest creates a session discarding the old messages
func NewKinesisForTest(t *testing.T) *TestKinesis {
	sess := awstest.NewSession()
	stream := createStream(t, sess)
	kin := kinesis.New(sess)
	sw := NewStreamWriter(kin, stream)
	sr, err := NewSingleShardStreamReader(kin, stream, true)
	require.NoError(t, err)
	prefix := fmt.Sprintf("client%d_", rand.Intn(1000))
	return &TestKinesis{
		sw:         sw,
		sr:         sr,
		dataPrefix: []byte(prefix),
	}
}

// TestKinesis decorates a Kinesis by transparently prefixing messages so that several tests can simultaneously
// work on the same stream (creating 1 stream per test is not an option because it takes ~15s)
type TestKinesis struct {
	sw         StreamWriter
	sr         StreamReader
	dataPrefix []byte
}

// Push prefixes the data before pushing
func (k *TestKinesis) Push(data []byte, partitionKey string) error {
	return k.sw.Push(k.prefixData(data), partitionKey)
}

// MultiPush prefixes the data before pushing
func (k *TestKinesis) MultiPush(records []*kinesis.PutRecordsRequestEntry) (*kinesis.PutRecordsOutput, error) {
	dr := make([]*kinesis.PutRecordsRequestEntry, len(records))
	for i, r := range records {
		dr[i] = &kinesis.PutRecordsRequestEntry{
			Data:            k.prefixData(r.Data),
			PartitionKey:    r.PartitionKey,
			ExplicitHashKey: r.ExplicitHashKey}
	}
	return k.sw.MultiPush(dr)
}

// Pop reads messages until error or one is found starting with dataPrefix
func (k *TestKinesis) Pop() ([]byte, error) {
	for {
		data, err := k.sr.Pop()
		if err != nil || data == nil {
			return data, err
		}
		if bytes.HasPrefix(data, k.dataPrefix) {
			return data[len(k.dataPrefix):], err
		}
	}
}

// GetTestMessage creates a message ending with the specified index
func GetTestMessage(index int) string {
	return fmt.Sprintf("msg_%d", index)
}

func createStream(t *testing.T, s *session.Session) (stream string) {
	resource := awstest.GetResourceServiceByName(kinesis.ServiceName)
	streamStr, _ := resource.CreateResourceForSession(t, s, aws.String("integration-test-"+uuid.New()))
	stream = aws.StringValue(streamStr)
	return stream
}

func (k *TestKinesis) prefixData(data []byte) []byte {
	ret := append([]byte(nil), k.dataPrefix...)
	return append(ret, data...)
}
