package kinesisx

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
)

// StreamWriter interface for writing to Kinesis-like streaming queues
type StreamWriter interface {
	//Push is only a helper for unit tests. Use MultiPush in production
	Push(data []byte, partitionKey string) error
	MultiPush(records []*kinesis.PutRecordsRequestEntry) (*kinesis.PutRecordsOutput, error)
}

// StreamReader interface for reading to Kinesis-like streaming queues
type StreamReader interface {
	Pop() ([]byte, error)
}

// NewStreamWriter creates a Kinesis write stream.
func NewStreamWriter(cli kinesisiface.KinesisAPI, streamName string) StreamWriter {
	return &streamWriter{
		client:     cli,
		streamName: streamName,
	}
}

type streamWriter struct {
	client     kinesisiface.KinesisAPI
	streamName string
}

func (k *streamWriter) Push(data []byte, partitionKey string) error {
	recordInput := kinesis.PutRecordInput{
		Data:         data,
		StreamName:   aws.String(k.streamName),
		PartitionKey: aws.String(partitionKey),
	}

	_, err := k.client.PutRecord(&recordInput)
	return err
}

func (k *streamWriter) MultiPush(records []*kinesis.PutRecordsRequestEntry) (*kinesis.PutRecordsOutput, error) {
	return k.client.PutRecords(&kinesis.PutRecordsInput{
		StreamName: aws.String(k.streamName),
		Records:    records,
	})
}

// NewSingleShardStreamReader creates a Kinesis read stream.
func NewSingleShardStreamReader(cli kinesisiface.KinesisAPI, streamName string, discardOldMessages bool) (StreamReader, error) {
	kin := &singleShardStreamReader{
		client:             cli,
		streamName:         streamName,
		discardOldMessages: discardOldMessages,
	}

	shardIt, err := kin.getAnyShardIterator()
	if err != nil {
		return nil, err
	}
	kin.shardIterator = *shardIt

	return kin, nil
}

type singleShardStreamReader struct {
	client             kinesisiface.KinesisAPI
	streamName         string
	discardOldMessages bool // is used by Read to define where to start reading
	shardIterator      string
}

// getAnyShardIterator gets the shard iterator of any shard
func (k *singleShardStreamReader) getAnyShardIterator() (*string, error) {
	shardIds, err := k.getShardIds()
	if err != nil {
		return nil, err
	}

	shardIteratorOut, err := k.client.GetShardIterator(&kinesis.GetShardIteratorInput{
		StreamName:        aws.String(k.streamName),
		ShardIteratorType: aws.String(k.getShardIteratorType()),
		ShardId:           aws.String(shardIds[0]),
	})
	return shardIteratorOut.ShardIterator, err
}

func (k *singleShardStreamReader) getShardIteratorType() string {
	if k.discardOldMessages {
		return kinesis.ShardIteratorTypeLatest
	}
	return kinesis.ShardIteratorTypeTrimHorizon

}

// getShardID gets the shardID of the specified shard
func (k *singleShardStreamReader) getShardIds() ([]string, error) {
	streamDesc, err := k.client.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: &k.streamName,
		Limit:      aws.Int64(1),
	})
	if err != nil {
		return nil, err
	}
	if len(streamDesc.StreamDescription.Shards) == 0 {
		return nil, fmt.Errorf("stream %s has 0 shards. Did you call WaitUntilStreamExists?", k.streamName)
	}

	var shardIds []string
	for _, shard := range streamDesc.StreamDescription.Shards {
		shardIds = append(shardIds, *shard.ShardId)
	}
	return shardIds, nil
}

// Pop reads 1 msg. So far only used for tests
// "Each shard can support up to 5 transactions per second for reads, up to a maximum total data read rate of 2 MB per second."
// -> so we should probably pop in big buckets in production code
// return nil slice when stream is empty
func (k *singleShardStreamReader) Pop() ([]byte, error) {
	var limit int64 = 1
	recordInput := kinesis.GetRecordsInput{
		ShardIterator: &k.shardIterator,
		Limit:         &limit,
	}
	output, err := k.client.GetRecords(&recordInput)
	if err != nil {
		return nil, err
	}
	if len(output.Records) == 0 {
		return nil, nil
	}
	k.shardIterator = *output.NextShardIterator

	return output.Records[0].Data, err
}
