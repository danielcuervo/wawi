package kinesisx_test

import (
	"fmt"
	"testing"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/socialpoint/bsk/pkg/awsx"
	"github.com/socialpoint/bsk/pkg/awsx/kinesisx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	nonRetryableError = "nonRetryableError"
)

type TestCase struct {
	input []string
	//map keys are the input messages, values are the errorCodes provided by kinesis. nil when no error
	errorByInput    map[string]*string
	returnNilOutput bool
	maxRetries      int

	expectedRetries [][]string
	expectError     bool
}

// TestRetryDecorator verifies the MultiPush calls performed to the decorated writer and the decorator return values
func TestRetryDecorator(t *testing.T) {
	testCases := []TestCase{
		{
			input: []string{"a", "b"},
			errorByInput: map[string]*string{
				"a": nil,
				"b": nil},
			maxRetries: 2,

			expectedRetries: nil,
		},

		{
			input:           []string{"a"},
			maxRetries:      2,
			returnNilOutput: true,

			expectError: true,
		},

		{
			input:      []string{"a", "b"},
			maxRetries: 1,
			errorByInput: map[string]*string{
				"a": aws.String(kinesis.ErrCodeProvisionedThroughputExceededException)},

			expectError: true,
			expectedRetries: [][]string{
				{"a"},
			},
		},

		{
			input:      []string{"a", "b", "c"},
			maxRetries: 2,
			errorByInput: map[string]*string{
				"a": aws.String(nonRetryableError),
				"b": aws.String(kinesis.ErrCodeProvisionedThroughputExceededException),
				"c": nil},

			expectError: true,
		},

		{
			input:      []string{"a", "b", "c"},
			maxRetries: 2,
			errorByInput: map[string]*string{
				"a": aws.String(awsx.ErrCodeServiceUnavailable),
				"b": aws.String(kinesis.ErrCodeProvisionedThroughputExceededException),
				"c": nil},
			expectedRetries: [][]string{
				{"a", "b"},
				{"a", "b"},
			},
		},

		{
			input:      []string{"a", "b", "c"},
			maxRetries: 2,
			errorByInput: map[string]*string{
				"a": nil,
				"b": aws.String(kinesis.ErrCodeProvisionedThroughputExceededException),
				"c": aws.String(kinesis.ErrCodeProvisionedThroughputExceededException)},
			expectedRetries: [][]string{
				{"b", "c"},
				{"b", "c"},
			},
		},
	}
	for index, tc := range testCases {
		streamSpy := &writerSpy{
			tc: tc,
			t:  t,
		}
		retry := kinesisx.NewRetryDecorator(tc.maxRetries, 1*time.Millisecond)(streamSpy)
		t.Run(fmt.Sprintf("TestCase %d", index), func(t *testing.T) {
			out, err := retry.MultiPush(NewRecords(tc.input))

			if tc.expectError {
				require.Error(t, err)
				if out != nil {
					require.NotEqual(t, 0, int(*out.FailedRecordCount))
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, 0, int(*out.FailedRecordCount))
			}
			assert.Equal(t, len(tc.expectedRetries), streamSpy.callsCount-1)
		})
	}
}

func NewRecords(msgs []string) []*kinesis.PutRecordsRequestEntry {
	records := make([]*kinesis.PutRecordsRequestEntry, len(msgs))
	for i, m := range msgs {
		records[i] = &kinesis.PutRecordsRequestEntry{Data: []byte(m)}
	}
	return records
}

// writerSpy simulates a StreamWriter that behaves as described by a TestCase
type writerSpy struct {
	kinesisx.StreamWriter
	tc TestCase
	t  *testing.T

	callsCount int
}

//MultiPush will return the error codes from TestCase the first 2 times MultiPush is invoked
func (s *writerSpy) MultiPush(records []*kinesis.PutRecordsRequestEntry) (*kinesis.PutRecordsOutput, error) {
	s.callsCount++
	s.verifyExpectedCall(records)

	if s.callsCount > 2 {
		//happy path, no errors
		return &kinesis.PutRecordsOutput{
			FailedRecordCount: aws.Int64(0),
		}, nil
	}
	out, err := s.getErrorResultForInput(records)

	return out, err
}

func (s *writerSpy) verifyExpectedCall(records []*kinesis.PutRecordsRequestEntry) {
	expectedCall := s.getExpectedCall()
	assert.Len(s.t, records, len(expectedCall))
	for i, r := range records {
		assert.Equal(s.t, expectedCall[i], string(r.Data))
	}
}

func (s *writerSpy) getExpectedCall() []string {
	if s.callsCount == 1 {
		return s.tc.input
	}
	return s.tc.expectedRetries[s.callsCount-2]
}

func (s *writerSpy) getErrorResultForInput(records []*kinesis.PutRecordsRequestEntry) (*kinesis.PutRecordsOutput, error) {
	if s.tc.returnNilOutput {
		return nil, fmt.Errorf("error with no output")
	}
	out := &kinesis.PutRecordsOutput{
		Records: make([]*kinesis.PutRecordsResultEntry, len(records)),
	}
	var errorsCount int64
	for i, record := range records {
		out.Records[i] = &kinesis.PutRecordsResultEntry{}
		errorCode := s.tc.errorByInput[string(record.Data)]
		if errorCode != nil {
			out.Records[i].ErrorCode = errorCode
			out.Records[i].ErrorMessage = aws.String(fmt.Sprintf("Message %s", *errorCode))
			errorsCount++
		}
	}
	out.SetFailedRecordCount(errorsCount)
	if errorsCount > 0 {
		return out, fmt.Errorf("errors")
	}
	return out, nil
}
