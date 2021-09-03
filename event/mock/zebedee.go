// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-search-data-extractor/event"
	"sync"
)

var (
	lockZebedeeClientMockChecker          sync.RWMutex
	lockZebedeeClientMockGetPublishedData sync.RWMutex
)

// Ensure, that ZebedeeClientMock does implement ZebedeeClient.
// If this is not the case, regenerate this file with moq.
var _ event.ZebedeeClient = &ZebedeeClientMock{}

// ZebedeeClientMock is a mock implementation of event.ZebedeeClient.
//
//     func TestSomethingThatUsesZebedeeClient(t *testing.T) {
//
//         // make and configure a mocked event.ZebedeeClient
//         mockedZebedeeClient := &ZebedeeClientMock{
//             CheckerFunc: func(in1 context.Context, in2 *healthcheck.CheckState) error {
// 	               panic("mock out the Checker method")
//             },
//             GetPublishedDataFunc: func(ctx context.Context, uriString string) ([]byte, error) {
// 	               panic("mock out the GetPublishedData method")
//             },
//         }
//
//         // use mockedZebedeeClient in code that requires event.ZebedeeClient
//         // and then make assertions.
//
//     }
type ZebedeeClientMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(in1 context.Context, in2 *healthcheck.CheckState) error

	// GetPublishedDataFunc mocks the GetPublishedData method.
	GetPublishedDataFunc func(ctx context.Context, uriString string) ([]byte, error)

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// In1 is the in1 argument value.
			In1 context.Context
			// In2 is the in2 argument value.
			In2 *healthcheck.CheckState
		}
		// GetPublishedData holds details about calls to the GetPublishedData method.
		GetPublishedData []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UriString is the uriString argument value.
			UriString string
		}
	}
}

// Checker calls CheckerFunc.
func (mock *ZebedeeClientMock) Checker(in1 context.Context, in2 *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("ZebedeeClientMock.CheckerFunc: method is nil but ZebedeeClient.Checker was just called")
	}
	callInfo := struct {
		In1 context.Context
		In2 *healthcheck.CheckState
	}{
		In1: in1,
		In2: in2,
	}
	lockZebedeeClientMockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	lockZebedeeClientMockChecker.Unlock()
	return mock.CheckerFunc(in1, in2)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedZebedeeClient.CheckerCalls())
func (mock *ZebedeeClientMock) CheckerCalls() []struct {
	In1 context.Context
	In2 *healthcheck.CheckState
} {
	var calls []struct {
		In1 context.Context
		In2 *healthcheck.CheckState
	}
	lockZebedeeClientMockChecker.RLock()
	calls = mock.calls.Checker
	lockZebedeeClientMockChecker.RUnlock()
	return calls
}

// GetPublishedData calls GetPublishedDataFunc.
func (mock *ZebedeeClientMock) GetPublishedData(ctx context.Context, uriString string) ([]byte, error) {
	if mock.GetPublishedDataFunc == nil {
		panic("ZebedeeClientMock.GetPublishedDataFunc: method is nil but ZebedeeClient.GetPublishedData was just called")
	}
	callInfo := struct {
		Ctx       context.Context
		UriString string
	}{
		Ctx:       ctx,
		UriString: uriString,
	}
	lockZebedeeClientMockGetPublishedData.Lock()
	mock.calls.GetPublishedData = append(mock.calls.GetPublishedData, callInfo)
	lockZebedeeClientMockGetPublishedData.Unlock()
	return mock.GetPublishedDataFunc(ctx, uriString)
}

// GetPublishedDataCalls gets all the calls that were made to GetPublishedData.
// Check the length with:
//     len(mockedZebedeeClient.GetPublishedDataCalls())
func (mock *ZebedeeClientMock) GetPublishedDataCalls() []struct {
	Ctx       context.Context
	UriString string
} {
	var calls []struct {
		Ctx       context.Context
		UriString string
	}
	lockZebedeeClientMockGetPublishedData.RLock()
	calls = mock.calls.GetPublishedData
	lockZebedeeClientMockGetPublishedData.RUnlock()
	return calls
}
