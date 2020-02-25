package gremcos

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/supplyon/gremcos/interfaces"
	mock_interfaces "github.com/supplyon/gremcos/test/mocks/interfaces"
	mock_metrics "github.com/supplyon/gremcos/test/mocks/metrics"
)

type payload struct {
	Data string `json:"data,omitempty"`
}

// predefinedUUIDGenerator is a helper that satisfies the uuidGeneratorFunc signature
// and returns the specified uuid
var predefinedUUIDGenerator = func(requestIDs []uuid.UUID) uuidGeneratorFunc {
	numIDs := len(requestIDs)
	idx := 0

	return func() (uuid.UUID, error) {
		id := requestIDs[idx]
		fmt.Printf("IIIID %d - %s\n", idx, id)
		idx = (idx + 1) % numIDs
		return id, nil
	}
}

func newResponse(requestID string, payload payload, code int) (interfaces.Response, []byte, error) {
	rawJSON, err := json.Marshal(payload)
	if err != nil {
		return interfaces.Response{}, nil, err
	}

	resp := interfaces.Response{
		RequestID: requestID,
		Status: interfaces.Status{
			Code: code,
		},
		Result: interfaces.Result{Data: rawJSON},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return interfaces.Response{}, nil, err
	}
	return resp, data, nil
}

func TestExecuteParallel(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
	metrics, metricMocks := NewMockedMetrics(mockCtrl)

	idleTimeout := time.Second * 12
	maxActiveConnections := 10
	username := "abcd"
	password := "xyz"

	numParallelExecuteCalls := 2
	var requestIDs []uuid.UUID
	for i := 0; i < numParallelExecuteCalls+1; i++ {
		requestID, err := uuid.NewV4()
		require.NoError(t, err)
		requestIDs = append(requestIDs, requestID)
	}

	cosmos, err := New("ws://host",
		ConnectionIdleTimeout(idleTimeout),
		NumMaxActiveConnections(maxActiveConnections),
		WithAuth(username, password),
		withMetrics(metrics),
	)
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	cosmos.dialer = mockedDialer
	cosmos.generateUUID = predefinedUUIDGenerator(requestIDs)

	mockedDialer.EXPECT().Connect().Return(nil).Times(numParallelExecuteCalls)
	mockedDialer.EXPECT().IsConnected().Return(true).Times(numParallelExecuteCalls)
	mockCount200 := mock_metrics.NewMockCounter(mockCtrl)
	mockCount200.EXPECT().Inc()
	metricMocks.statusCodeTotal.EXPECT().WithLabelValues("200").Return(mockCount200)
	metricMocks.serverTimePerQueryResponseAvgMS.EXPECT().Set(float64(0))
	metricMocks.serverTimePerQueryMS.EXPECT().Set(float64(0))
	metricMocks.requestChargePerQueryResponseAvg.EXPECT().Set(float64(0))
	metricMocks.requestChargePerQuery.EXPECT().Set(float64(0))
	metricMocks.requestChargeTotal.EXPECT().Add(float64(0))
	metricMocks.retryAfterMS.EXPECT().Set(float64(0))

	_, rawResponse, _ := newResponse(requestIDs[0].String(), payload{Data: "HELLO"}, interfaces.StatusSuccess)
	mockedDialer.EXPECT().Read().DoAndReturn(func() {
		//time.Sleep(time.Millisecond * 10)
	}).Return(1, rawResponse, nil).AnyTimes()
	_, rawResponse, _ = newResponse(requestIDs[1].String(), payload{Data: "HELLO"}, interfaces.StatusSuccess)
	mockedDialer.EXPECT().Read().DoAndReturn(func() {
		//time.Sleep(time.Millisecond * 10)
	}).Return(1, rawResponse, nil).AnyTimes()

	mockedDialer.EXPECT().Write(gomock.Any()).Return(nil)
	mockedDialer.EXPECT().Close().Return(nil)

	wg := sync.WaitGroup{}
	wg.Add(numParallelExecuteCalls)
	for i := 0; i < numParallelExecuteCalls; i++ {
		go func() {
			defer wg.Done()
			// WHEN
			responses, err := cosmos.Execute("g.V()")

			// THEN
			assert.NoError(t, err)
			assert.NotEmpty(t, responses)
		}()
	}

	// CLEANUP
	wg.Wait()
	err = cosmos.Stop()
	assert.NoError(t, err)
}

//func TestExecute(t *testing.T) {
//	// GIVEN
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//	mockedDialer := mock_interfaces.NewMockDialer(mockCtrl)
//	metrics, metricMocks := NewMockedMetrics(mockCtrl)
//
//	idleTimeout := time.Second * 12
//	maxActiveConnections := 10
//	username := "abcd"
//	password := "xyz"
//	requestID, err := uuid.NewV4()
//	require.NoError(t, err)
//
//	cosmos, err := New("ws://host",
//		ConnectionIdleTimeout(idleTimeout),
//		NumMaxActiveConnections(maxActiveConnections),
//		WithAuth(username, password),
//		withMetrics(metrics),
//	)
//	require.NoError(t, err)
//	require.NotNil(t, cosmos)
//	cosmos.dialer = mockedDialer
//	cosmos.generateUUID = predefinedUUIDGenerator(requestID)
//
//	mockedDialer.EXPECT().Connect().Return(nil)
//	mockedDialer.EXPECT().IsConnected().Return(true)
//	mockCount200 := mock_metrics.NewMockCounter(mockCtrl)
//	mockCount200.EXPECT().Inc()
//	metricMocks.statusCodeTotal.EXPECT().WithLabelValues("200").Return(mockCount200)
//	metricMocks.serverTimePerQueryResponseAvgMS.EXPECT().Set(float64(0))
//	metricMocks.serverTimePerQueryMS.EXPECT().Set(float64(0))
//	metricMocks.requestChargePerQueryResponseAvg.EXPECT().Set(float64(0))
//	metricMocks.requestChargePerQuery.EXPECT().Set(float64(0))
//	metricMocks.requestChargeTotal.EXPECT().Add(float64(0))
//	metricMocks.retryAfterMS.EXPECT().Set(float64(0))
//
//	_, rawResponse, _ := newResponse(requestID.String(), payload{Data: "HELLO"}, interfaces.StatusSuccess)
//	mockedDialer.EXPECT().Read().DoAndReturn(func() {
//		time.Sleep(time.Millisecond * 100)
//	}).Return(1, rawResponse, nil).Times(2) // one call for the read of the data and the second for the next blocking read
//	mockedDialer.EXPECT().Write(gomock.Any()).Return(nil)
//	mockedDialer.EXPECT().Close().Return(nil)
//
//	// WHEN
//	responses, err := cosmos.Execute("g.V()")
//
//	// THEN
//	assert.NoError(t, err)
//	assert.NotEmpty(t, responses)
//
//	// CLEANUP
//	err = cosmos.Stop()
//	assert.NoError(t, err)
//}

func TestNew(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)
	idleTimeout := time.Second * 12
	maxActiveConnections := 10
	username := "abcd"
	password := "xyz"

	// WHEN
	cosmos, err := New("ws://host",
		ConnectionIdleTimeout(idleTimeout),
		NumMaxActiveConnections(maxActiveConnections),
		WithAuth(username, password),
		withMetrics(metrics),
	)

	// THEN
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	assert.Equal(t, idleTimeout, cosmos.connectionIdleTimeout)
	assert.Equal(t, maxActiveConnections, cosmos.numMaxActiveConnections)
	assert.Equal(t, username, cosmos.username)
	assert.Equal(t, password, cosmos.password)
}

func TestStop(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	cosmos, err := New("ws://host", withMetrics(metrics))
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	cosmos.pool = mockedQueryExecutor
	mockedQueryExecutor.EXPECT().Close().Return(nil)

	// WHEN
	err = cosmos.Stop()

	// THEN
	assert.NoError(t, err)
}

func TestIsHealthy(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)
	mockedQueryExecutor := mock_interfaces.NewMockQueryExecutor(mockCtrl)

	cosmos, err := New("ws://host", withMetrics(metrics))
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	cosmos.pool = mockedQueryExecutor

	// WHEN -- connected --> healthy
	mockedQueryExecutor.EXPECT().Ping().Return(nil)
	healthyWhenConnected := cosmos.IsHealthy()

	// WHEN -- not connected --> not healthy
	mockedQueryExecutor.EXPECT().Ping().Return(fmt.Errorf("Not connected"))
	healthyWhenNotConnected := cosmos.IsHealthy()

	// THEN
	assert.NoError(t, healthyWhenConnected)
	assert.Error(t, healthyWhenNotConnected)
}

func TestNewWithMetrics(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// WHEN
	cosmos, err := New("ws://host", MetricsPrefix("prefix"))

	// THEN
	require.NoError(t, err)
	require.NotNil(t, cosmos)
	assert.NotNil(t, cosmos.metrics)
}

func TestUpdateMetricsNoResponses(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, _ := NewMockedMetrics(mockCtrl)

	var responses []interfaces.Response

	// WHEN
	updateRequestMetrics(responses, metrics)

	// THEN
	// there should be no invocation on the metrics mock
	// since there where no responses
}

func TestUpdateMetricsZero(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, metricMocks := NewMockedMetrics(mockCtrl)

	noErrorNoAttrs := interfaces.Response{
		Status: interfaces.Status{
			Code: interfaces.StatusSuccess,
		},
	}

	noError := interfaces.Response{
		Status: interfaces.Status{
			Code: interfaces.StatusSuccess,
			Attributes: map[string]interface{}{
				"x-ms-status-code": 200,
			},
		},
	}

	responses := []interfaces.Response{noError, noErrorNoAttrs}

	// WHEN
	mockCount200 := mock_metrics.NewMockCounter(mockCtrl)
	mockCount200.EXPECT().Inc().Times(2)
	metricMocks.statusCodeTotal.EXPECT().WithLabelValues("200").Return(mockCount200).Times(2)
	metricMocks.serverTimePerQueryResponseAvgMS.EXPECT().Set(float64(0))
	metricMocks.serverTimePerQueryMS.EXPECT().Set(float64(0))
	metricMocks.requestChargePerQueryResponseAvg.EXPECT().Set(float64(0))
	metricMocks.requestChargePerQuery.EXPECT().Set(float64(0))
	metricMocks.requestChargeTotal.EXPECT().Add(float64(0))
	metricMocks.retryAfterMS.EXPECT().Set(float64(0))
	updateRequestMetrics(responses, metrics)

	// THEN
	// expect the calls on the metrics specified above
}

func TestUpdateMetricsFull(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	metrics, metricMocks := NewMockedMetrics(mockCtrl)

	withError := interfaces.Response{
		Status: interfaces.Status{
			Code: interfaces.StatusSuccess,
			Attributes: map[string]interface{}{
				"x-ms-status-code":          429,
				"x-ms-substatus-code":       3200,
				"x-ms-total-request-charge": 11,
				"x-ms-total-server-time-ms": 22,
				"x-ms-retry-after-ms":       "00:00:00.033",
			},
		},
	}

	responses := []interfaces.Response{withError}

	// WHEN
	mockCount200 := mock_metrics.NewMockCounter(mockCtrl)
	mockCount200.EXPECT().Inc()
	metricMocks.statusCodeTotal.EXPECT().WithLabelValues("429").Return(mockCount200)
	metricMocks.serverTimePerQueryResponseAvgMS.EXPECT().Set(float64(22))
	metricMocks.serverTimePerQueryMS.EXPECT().Set(float64(22))
	metricMocks.requestChargePerQueryResponseAvg.EXPECT().Set(float64(11))
	metricMocks.requestChargePerQuery.EXPECT().Set(float64(11))
	metricMocks.requestChargeTotal.EXPECT().Add(float64(11))
	metricMocks.retryAfterMS.EXPECT().Set(float64(33))
	updateRequestMetrics(responses, metrics)

	// THEN
	// expect the calls on the metrics specified above
}
