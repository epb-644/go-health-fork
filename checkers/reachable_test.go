package checkers_test

import (
	"errors"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/InVisionApp/go-health/checkers"
	"github.com/InVisionApp/go-health/fakes"
	"github.com/InVisionApp/go-health/fakes/netfakes"
	"github.com/stretchr/testify/assert"
)

func TestReachableSuccess(t *testing.T) {
	assert := assert.New(t)
	dd := &fakes.FakeReachableDatadogIncrementer{}
	u, _ := url.Parse("http://example.com")
	cfg := &checkers.ReachableConfig{
		URL: u,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			return nil, nil
		},
		DatadogClient: dd,
		DatadogTags: []string{
			"dependency:test-service",
		},
	}
	c, err := checkers.NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status()
	assert.NoError(err)
	assert.Equal(0, dd.IncrCallCount())
}

func TestReachableError(t *testing.T) {
	assert := assert.New(t)
	u, _ := url.Parse("http://example.com")
	cfg := &checkers.ReachableConfig{
		URL: u,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			return nil, errors.New("Failed check")
		},
	}
	c, err := checkers.NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status()
	assert.Error(err)
}

func TestReachableConnError(t *testing.T) {
	assert := assert.New(t)
	u, _ := url.Parse("http://example.com")
	expectedErr := errors.New("Failed close")
	cfg := &checkers.ReachableConfig{
		URL: u,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			conn := &netfakes.FakeConn{}
			conn.CloseReturns(expectedErr)
			return conn, nil
		},
	}
	c, err := checkers.NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status()
	assert.EqualError(err, expectedErr.Error())
}

func TestReachableErrorWithDatadog(t *testing.T) {
	assert := assert.New(t)
	dd := &fakes.FakeReachableDatadogIncrementer{}
	ddTags := []string{
		"dependency:test-service",
	}
	u, _ := url.Parse("http://example.com")
	cfg := &checkers.ReachableConfig{
		URL: u,
		Dialer: func(network, address string, timeout time.Duration) (net.Conn, error) {
			return nil, errors.New("Failed check")
		},
		DatadogClient: dd,
		DatadogTags:   ddTags,
	}
	c, err := checkers.NewReachableChecker(cfg)
	assert.NoError(err)
	assert.NotNil(c)

	_, err = c.Status()
	assert.Error(err)
	assert.Equal(1, dd.IncrCallCount())
	name, tags, num := dd.IncrArgsForCall(0)
	assert.Equal(checkers.ReachableDDHealthErrors, name)
	assert.Equal(ddTags, tags)
	assert.Equal(1.0, num)
}
