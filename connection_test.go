package gremtune

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gorilla "github.com/gorilla/websocket"
	"github.com/schwartzmx/gremtune/interfaces"
	mock_interfaces "github.com/schwartzmx/gremtune/test/mocks/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDialer(t *testing.T) {
	// GIVEN
	errorChannel := make(chan error)

	// WHEN
	dialer, err := NewDialer("ws://localhost", errorChannel)

	// THEN
	assert.NotNil(t, dialer)
	assert.NoError(t, err)
}

func TestNewDialerFail(t *testing.T) {

	// WHEN - invalid host
	errorChannel := make(chan error)
	dialer, err := NewDialer("invalid host", errorChannel)
	assert.Nil(t, dialer)
	assert.Error(t, err)

	// WHEN - read buffer invalid
	dialer, err = NewDialer("ws://host", errorChannel, SetBufferSize(0, 10))
	assert.Nil(t, dialer)
	assert.Error(t, err)

	// WHEN - write buffer invalid
	dialer, err = NewDialer("ws://host", errorChannel, SetBufferSize(10, 0))
	assert.Nil(t, dialer)
	assert.Error(t, err)

	// WHEN - dialerFactory is nil
	dialer, err = NewDialer("ws://host", errorChannel, websocketDialerFactoryFun(nil))
	assert.Nil(t, dialer)
	assert.Error(t, err)

	// WHEN - errorChannel is nil
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)
	dialer, err = NewDialer("ws://host", nil, websocketDialerFactoryFun(mockedDialerFactory))
	assert.Nil(t, dialer)
	assert.Error(t, err)
}

func TestConnect(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)
	errorChannel := make(chan error)

	dialer, err := NewDialer("ws://localhost", errorChannel, websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = dialer.Connect()

	// THEN
	assert.NoError(t, err)
	assert.True(t, dialer.IsConnected())
}

func TestConnectFail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, true)
	errorChannel := make(chan error)

	dialer, err := NewDialer("ws://localhost", errorChannel, websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN
	err = dialer.Connect()

	// THEN
	assert.Error(t, err)
	assert.False(t, dialer.IsConnected())
}

func TestConnectReconnect(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)
	errorChannel := make(chan error)

	dialer, err := NewDialer("ws://localhost", errorChannel, websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN - first connect successful
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = dialer.Connect()

	// THEN
	assert.NoError(t, err)
	assert.True(t, dialer.IsConnected())

	// WHEN - second connect fails
	dialer.(*websocket).wsDialerFactory = newMockedDialerFactory(mockedWebsocketConnection, true)
	err = dialer.Connect()

	// THEN
	assert.Error(t, err)
	assert.False(t, dialer.IsConnected())
}

func TestConnectClose(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)
	errorChannel := make(chan error)

	dialer, err := NewDialer("ws://localhost", errorChannel, websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN connected
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = dialer.Connect()
	require.NoError(t, err)

	mockedWebsocketConnection.EXPECT().WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, "")).Return(nil)
	mockedWebsocketConnection.EXPECT().Close()
	err = dialer.Close()

	// THEN
	assert.NoError(t, err)
}

func TestConnectCloseOnNotConnectedWebsocket(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)
	errorChannel := make(chan error)

	dialer, err := NewDialer("ws://localhost", errorChannel, websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN - not connected
	err = dialer.Close()

	// THEN
	assert.NoError(t, err)
}

func TestConnectCloseFail(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)
	errorChannel := make(chan error)

	dialer, err := NewDialer("ws://localhost", errorChannel, websocketDialerFactoryFun(mockedDialerFactory))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN connected
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = dialer.Connect()
	require.NoError(t, err)

	mockedWebsocketConnection.EXPECT().WriteMessage(gorilla.CloseMessage, gorilla.FormatCloseMessage(gorilla.CloseNormalClosure, "")).Return(nil)
	mockedWebsocketConnection.EXPECT().Close()
	err = dialer.Close()
	require.NoError(t, err)

	// WHEN close is called again on a disposed websocket
	err = dialer.Close()

	// THEN
	assert.Error(t, err)
}

func TestPing(t *testing.T) {
	// GIVEN
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockedWebsocketConnection := mock_interfaces.NewMockWebsocketConnection(mockCtrl)
	mockedDialerFactory := newMockedDialerFactory(mockedWebsocketConnection, false)
	errorChannel := make(chan error)

	dialer, err := NewDialer("ws://localhost", errorChannel, websocketDialerFactoryFun(mockedDialerFactory), SetPingInterval(time.Millisecond*100))
	require.NoError(t, err)
	require.NotNil(t, dialer)

	// WHEN
	mockedWebsocketConnection.EXPECT().SetPongHandler(gomock.Any())
	err = dialer.Connect()
	require.NoError(t, err)

	mockedWebsocketConnection.EXPECT().WriteControl(gorilla.PingMessage, gomock.Any(), gomock.Any()).Return(nil)
	mockedWebsocketConnection.EXPECT().WriteControl(gorilla.PingMessage, gomock.Any(), gomock.Any()).Return(fmt.Errorf("ERR")).AnyTimes()
	errors := make(chan error, 5)
	go dialer.Ping(errors)

	// wait a bit to allow the ping timer to tick
	time.Sleep(time.Millisecond * 500)

	// THEN
	assert.False(t, dialer.IsConnected())
	assert.NotEmpty(t, errors)
}

func newMockedDialerFactory(websocketConnection interfaces.WebsocketConnection, fail bool) websocketDialerFactory {

	dialerFuncSuccess := func(urlStr string, requestHeader http.Header) (interfaces.WebsocketConnection, *http.Response, error) {
		return websocketConnection, nil, nil
	}

	dialerFuncError := func(urlStr string, requestHeader http.Header) (interfaces.WebsocketConnection, *http.Response, error) {
		return nil, nil, fmt.Errorf("Timeout")
	}

	// if needed return a dialer that can't create a connection successfully
	if fail {
		return func(wBufSize, rBifSize int, timeout time.Duration) websocketDialer {
			return dialerFuncError
		}
	}

	return func(wBufSize, rBifSize int, timeout time.Duration) websocketDialer {
		return dialerFuncSuccess
	}
}
