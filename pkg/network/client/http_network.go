package network_client

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"sync/atomic"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/sse"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	pe "github.com/pkg/errors"
)

type HttpNetworkOptions struct {
	BackendUrl      string
	ReceiveEndpoint string
	SendEndpoint    string
	Headers         map[string]string
}

type HttpNetwork struct {
	options    *HttpNetworkOptions
	httpClient *http.Client
	sseClient  *sse.SseClient
	ctx        context.Context
	cancel     context.CancelFunc
	msgHandler func(msg []byte)
	status     atomic.Int32
	statusEb   *util.EventBus[NetworkStatus]
}

var _ NetworkProvider = &HttpNetwork{}

func (n *HttpNetwork) ensureOptionsValid() {
	_, err := url.JoinPath(n.options.BackendUrl, n.options.ReceiveEndpoint)
	if err != nil {
		panic("invalid receive url " + err.Error())
	}

	_, err = url.JoinPath(n.options.BackendUrl, n.options.SendEndpoint)
	if err != nil {
		panic("invalid send url " + err.Error())
	}

	if n.options.Headers == nil {
		n.options.Headers = make(map[string]string)
		n.options.Headers["Content-Type"] = "application/octet-stream"
	}
}

func NewHttpNetwork(options *HttpNetworkOptions) *HttpNetwork {
	return NewHttpNetworkWithContext(options, context.Background())
}

func NewHttpNetworkWithContext(options *HttpNetworkOptions, ctx context.Context) *HttpNetwork {
	subCtx, cancel := context.WithCancel(ctx)

	n := &HttpNetwork{
		options:    options,
		httpClient: &http.Client{},
		sseClient:  nil, // init later
		msgHandler: nil,
		status:     atomic.Int32{},
		statusEb:   util.NewEventBus[NetworkStatus](),
		ctx:        subCtx,
		cancel:     cancel,
	}

	n.ensureOptionsValid()

	sseUrl, _ := url.JoinPath(options.BackendUrl, options.ReceiveEndpoint)

	n.sseClient = sse.NewSseClient(sseUrl)
	go n.syncStatusFromSseClient()
	for k, v := range n.options.Headers { // set sse headers
		n.sseClient.Headers[k] = v
	}

	return n
}

func (n *HttpNetwork) Connect() error {
	status := NetworkStatus(n.status.Load())
	if status == NetworkClosed {
		return pe.Errorf("network is already closed")
	}
	if status == NetworkReady {
		return pe.Errorf("network is already connected")
	}

	sseEventCh := make(chan *sse.SseEvent)
	err := n.sseClient.SubscribeChanWithContext(n.ctx, sseEventCh)
	if err != nil {
		return pe.Wrap(err, "failed to subscribe to sse")
	}

	go func() {
		defer close(sseEventCh)
		for {
			select {
			case <-n.ctx.Done():
				n.sseClient.Close()
				n.httpClient.CloseIdleConnections()
				n.setStatus(NetworkClosed)
				return
			case event := <-sseEventCh:
				n.handleSseEvent(event)
			}
		}
	}()

	return nil
}

func (n *HttpNetwork) Close() error {
	status := NetworkStatus(n.status.Load())
	if status == NetworkClosed {
		return nil
	}
	n.cancel()
	return nil
}

func (n *HttpNetwork) Send(msg []byte) error {
	status := NetworkStatus(n.status.Load())
	if status != NetworkReady {
		return pe.Errorf("network is not ready, current status: %v", status)
	}

	sendUrl, _ := url.JoinPath(n.options.BackendUrl, n.options.SendEndpoint)
	req, err := http.NewRequestWithContext(n.ctx, "POST", sendUrl, bytes.NewReader(msg))
	if err != nil {
		return pe.Errorf("failed to create request: %w", err)
	}

	for k, v := range n.options.Headers {
		req.Header.Set(k, v)
	}

	// log.Debugf("client send: %v", msg)

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return pe.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return pe.Errorf("request failed with status code %d", resp.StatusCode)
	}
	return nil
}

func (n *HttpNetwork) SetMsgHandler(handler func(msg []byte)) {
	n.msgHandler = handler
}

func (n *HttpNetwork) GetStatus() NetworkStatus {
	return NetworkStatus(n.status.Load())
}

func (n *HttpNetwork) SubscribeStatusChange() <-chan NetworkStatus {
	return n.statusEb.Subscribe()
}

func (n *HttpNetwork) UnsubscribeStatusChange(ch <-chan NetworkStatus) {
	n.statusEb.Unsubscribe(ch)
}

func (n *HttpNetwork) handleSseEvent(event *sse.SseEvent) {
	if n.msgHandler != nil {
		n.msgHandler(event.Data)
	} else {
		log.Warnf("no msg handler set, a event is ignored")
	}
}

func (n *HttpNetwork) syncStatusFromSseClient() {
	sseStatusCh := n.sseClient.SubscribeStatusChange()

	go func() {
		defer n.sseClient.UnsubscribeStatusChange(sseStatusCh)

		for {
			select {
			case <-n.ctx.Done():
				return
			case sseStatus := <-sseStatusCh:
				switch sseStatus {
				case sse.SSE_CLIENT_STATUS_CONNECTED:
					n.setStatus(NetworkReady)
				default:
					n.setStatus(NetworkNotReady)
				}
			}
		}
	}()
}

func (n *HttpNetwork) setStatus(status NetworkStatus) {
	oldStatus := n.status.Load()
	n.status.Store(int32(status))

	if oldStatus != int32(status) {
		log.Debugf("Client HttpNetwork status changed: %v -> %v", NetworkStatus(oldStatus), status)
		n.statusEb.Publish(status)
	}
}
