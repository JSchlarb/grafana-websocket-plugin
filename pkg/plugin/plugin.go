package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sync"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/models"
)

// Make sure WebSocketDataSource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler      = (*WebSocketDataSource)(nil)
	_ backend.CheckHealthHandler    = (*WebSocketDataSource)(nil)
	_ backend.StreamHandler         = (*WebSocketDataSource)(nil)
	_ instancemgmt.InstanceDisposer = (*WebSocketDataSource)(nil)
)

type channelConfig struct {
	path        string
	queryParams map[string]string
}

// NewWebSocketDataSource creates a new datasource instance.
func NewWebSocketDataSource(_ context.Context, ds backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings, err := models.LoadPluginSettings(ds)
	if err != nil {
		return nil, fmt.Errorf("failed to read CustomSettings from the Query Request: %s", err.Error())
	}

	return &WebSocketDataSource{
		customHeaders:         settings.Headers,
		customQueryParameters: settings.QueryParameters,
		channelConfigs:        map[string]channelConfig{},
	}, nil
}

// WebSocketDataSource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type WebSocketDataSource struct {
	customHeaders         map[string]string
	customQueryParameters map[string]string
	channelConfigs        map[string]channelConfig
	mu                    sync.RWMutex
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewWebSocketDataSource factory function.
func (wsds *WebSocketDataSource) Dispose() {
	// Clean up datasource instance resources.
	log.DefaultLogger.Info("Dispose Method", "disposing instance")
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (wsds *WebSocketDataSource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := wsds.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	WsPath      string            `json:"path"`
	QueryParams map[string]string `json:"queryParams"`
}

func (wsds *WebSocketDataSource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}

	// Unmarshal the JSON into our queryModel.
	var qm queryModel
	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	// create data frame response.
	frame := data.NewFrame("response")

	// If query called with streaming on then return a channel
	// to subscribe on a client-side and consume updates from a plugin.
	// Feel free to remove this if you don't need streaming for your datasource.
	channel := live.Channel{
		Scope:     live.ScopeDatasource,
		Namespace: pCtx.DataSourceInstanceSettings.UID,
		Path:      path.Join(path.Clean(qm.WsPath), query.RefID),
	}
	frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
	// add the frames to the response.
	response.Frames = append(response.Frames, frame)
	wsds.setChannelConfig(channel.Path, qm.WsPath, wsds.mergeQueryParams(qm.QueryParams))

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (wsds *WebSocketDataSource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	// var jsonData map[string]string
	// if err := json.Unmarshal(req.PluginContext.AppInstanceSettings.JSONData, &jsonData); err != nil {
	// 	log.DefaultLogger.Error("Health Check", "Unmarshall JsonData check", err.Error())
	// 	return &backend.CheckHealthResult{
	// 		Status:  backend.HealthStatusError,
	// 		Message: err.Error(),
	// 	}, err
	// }

	// checkHost := func(host string) error {
	// 	_, err := url.Parse(host)
	// 	log.DefaultLogger.Error("Health Check func", "host check func", err.Error())
	// 	if err != nil {
	// 		return fmt.Errorf("host is not valid: %s", err.Error())
	// 	}
	// 	return nil
	// }

	// // if err := checkHost(string(jsonData["host"])); err != nil {
	// if err := checkHost("tes.com"); err != nil {
	// 	log.DefaultLogger.Error("Health Check", "host check", err.Error())
	// 	return &backend.CheckHealthResult{
	// 		Status:  backend.HealthStatusError,
	// 		Message: err.Error(),
	// 	}, err
	// }

	// else if err := checkCustomHeaders(); err != nil {
	// 	status = backend.HealthStatusError
	// 	message = err.Error()
	// } else if err := checkCustomQueryParameters(); err != nil {
	// 	status = backend.HealthStatusError
	// 	message = err.Error()
	// }

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (wsds *WebSocketDataSource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	status := backend.SubscribeStreamStatusOK

	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (wsds *WebSocketDataSource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "Path", req.Path)

	cfg, ok := wsds.getChannelConfig(req.Path)
	if !ok {
		err := fmt.Errorf("no channel config found for %s", req.Path)
		log.DefaultLogger.Error("RunStream config missing", "path", req.Path, "error", err)
		sendErrorFrame(err.Error(), sender)
		return err
	}

	wsDataProxy, err := NewWsDataProxy(req, sender, wsds, cfg)
	if err != nil {
		errCtx := "Starting WebSocket"

		log.DefaultLogger.Error(errCtx, "error", err.Error())

		sendErrorFrame(fmt.Sprintf("%s: %s", errCtx, err.Error()), sender)

		return err
	}

	go wsDataProxy.proxyMessage()

	go wsDataProxy.readMessage()

	select {
	case <-ctx.Done():

		wsDataProxy.done <- true

		log.DefaultLogger.Info("Closing Channel", "channel", req.Path)

		return nil
	case rError := <-wsDataProxy.readingErrors:
		log.DefaultLogger.Error("Error reading the websocket", "error", err.Error())
		sendErrorFrame(fmt.Sprintf("%s: %s", "Error reading the websocket", err.Error()), sender)

		log.DefaultLogger.Info("Closing Channel due an error to read websocket", "channel", req.Path)

		return rError
	}
}

// PublishStream is called when a client sends a message to the stream.
func (wsds *WebSocketDataSource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (wsds *WebSocketDataSource) mergeQueryParams(queryParams map[string]string) map[string]string {
	merged := map[string]string{}
	for k, v := range wsds.customQueryParameters {
		merged[k] = v
	}
	for k, v := range queryParams {
		merged[k] = v
	}
	return merged
}

func (wsds *WebSocketDataSource) setChannelConfig(channelPath string, basePath string, params map[string]string) {
	wsds.mu.Lock()
	defer wsds.mu.Unlock()
	wsds.channelConfigs[channelPath] = channelConfig{
		path:        basePath,
		queryParams: params,
	}
}

func (wsds *WebSocketDataSource) getChannelConfig(channelPath string) (channelConfig, bool) {
	wsds.mu.RLock()
	defer wsds.mu.RUnlock()
	cfg, ok := wsds.channelConfigs[channelPath]
	return cfg, ok
}
