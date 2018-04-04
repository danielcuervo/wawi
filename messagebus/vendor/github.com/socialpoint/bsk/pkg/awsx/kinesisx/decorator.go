package kinesisx

import (
	"fmt"

	"time"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/socialpoint/bsk/pkg/awsx"
	"github.com/socialpoint/bsk/pkg/logx"
)

var retryableErrorCodes = map[string]struct{}{
	kinesis.ErrCodeProvisionedThroughputExceededException: {},
	//From https://aws.amazon.com/blogs/big-data/implementing-efficient-and-reliable-producers-with-the-amazon-kinesis-producer-library/
	awsx.ErrCodeInternalFailure:    {},
	awsx.ErrCodeServiceUnavailable: {},
}

//Decorator decorates a StreamWriter
type Decorator func(StreamWriter) StreamWriter

// NewRetryDecorator creates a RetryDecorator with backoff
func NewRetryDecorator(maxRetries int, backoff time.Duration) Decorator {
	return func(writer StreamWriter) StreamWriter {
		return &RetryDecorator{
			StreamWriter: writer,
			logger:       logx.New(),
			maxRetries:   maxRetries,
			backoff:      backoff}
	}
}

// RetryDecorator retries StreamWriter.MultiPush while there is any retryable error but no non-retryable errors
type RetryDecorator struct {
	StreamWriter
	retryRecords []*kinesis.PutRecordsRequestEntry
	logger       logx.Logger
	maxRetries   int
	backoff      time.Duration
}

// MultiPush retries the decorated writer logging the counts of each error code
func (r *RetryDecorator) MultiPush(records []*kinesis.PutRecordsRequestEntry) (*kinesis.PutRecordsOutput, error) {
	//from http://docs.aws.amazon.com/streams/latest/dev/developing-producers-with-sdk.html#kinesis-using-sdk-java-putrecords-handling-failures
	// and https://aws.amazon.com/blogs/big-data/implementing-efficient-and-reliable-producers-with-the-amazon-kinesis-producer-library/
	out, err := r.StreamWriter.MultiPush(records)
	if out == nil || *out.FailedRecordCount == 0 {
		return out, err
	}

	for retry := 0; *out.FailedRecordCount > 0 && retry < r.maxRetries; retry++ {
		abortDueTo := r.manageErrors(records, out)
		if abortDueTo != "" {
			r.logger.Info(fmt.Sprintf("Aborting due to \"" + abortDueTo + "\""))
			return out, err
		}
		records = r.retryRecords
		time.Sleep(r.backoff * time.Duration(retry))
		out, err = r.StreamWriter.MultiPush(records)
		if out == nil {
			return out, err
		}
	}
	return out, err
}

// manageErrors updates retryRecords field with only those which need to be retried
// If we need to abort, it will return the error message. Otherwise, returns nil
func (r *RetryDecorator) manageErrors(records []*kinesis.PutRecordsRequestEntry, out *kinesis.PutRecordsOutput) string {
	if len(r.retryRecords) < int(*out.FailedRecordCount) {
		r.retryRecords = make([]*kinesis.PutRecordsRequestEntry, *out.FailedRecordCount)
	}
	r.retryRecords = r.retryRecords[:0]
	var abortDueTo string
	var countByCode = make(map[string]int)
	for index, rec := range out.Records {
		if rec.ErrorCode == nil {
			continue
		}
		countByCode[*rec.ErrorCode]++
		_, retryCode := retryableErrorCodes[*rec.ErrorCode]
		if retryCode {
			r.retryRecords = append(r.retryRecords, records[index])
		} else {
			abortDueTo = *rec.ErrorCode + ": " + *rec.ErrorMessage
		}
	}
	for errorCode, count := range countByCode {
		r.logger.Info(fmt.Sprintf("%d messages failed due to %v", count, errorCode))
	}
	return abortDueTo
}
