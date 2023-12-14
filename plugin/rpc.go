package plugin

import (
	"io"
	"net/rpc"

	goplugin "github.com/hashicorp/go-plugin"
	"github.com/reeveci/reeve-lib/schema"
)

type ReevePluginClient struct {
	client *rpc.Client
	broker *goplugin.MuxBroker
}

func (r *ReevePluginClient) Name() (resp string, err error) {
	err = r.client.Call("Plugin.Name", new(interface{}), &resp)
	return
}

func (r *ReevePluginClient) Register(settings map[string]string, api ReeveAPI) (resp Capabilities, err error) {
	apiServer := &ReeveAPIServer{impl: api}

	brokerID := r.broker.NextId()
	go func() {
		r.broker.AcceptAndServe(brokerID, apiServer)
		api.Close()
	}()

	err = r.client.Call("Plugin.Register", []interface{}{settings, brokerID}, &resp)
	return
}

func (r *ReevePluginClient) Unregister() error {
	return r.client.Call("Plugin.Unregister", new(interface{}), new(interface{}))
}

func (r *ReevePluginClient) Message(source string, message schema.Message) error {
	return r.client.Call("Plugin.Message", schema.FullMessage{Message: message, Source: source}, new(interface{}))
}

func (r *ReevePluginClient) Discover(trigger schema.Trigger) (resp []schema.Pipeline, err error) {
	err = r.client.Call("Plugin.Discover", trigger, &resp)
	return
}

func (r *ReevePluginClient) Resolve(env []string) (resp map[string]schema.Env, err error) {
	err = r.client.Call("Plugin.Resolve", env, &resp)
	return
}

func (r *ReevePluginClient) Notify(status schema.PipelineStatus) error {
	if status.Logs == nil || !status.Logs.Available() {
		return r.client.Call("Plugin.Notify", []interface{}{status}, new(interface{}))
	}

	logReaderProviderServer := &LogReaderProviderServer{impl: status.Logs, broker: r.broker}
	status.Logs = nil

	brokerID := r.broker.NextId()
	go r.broker.AcceptAndServe(brokerID, logReaderProviderServer)

	return r.client.Call("Plugin.Notify", []interface{}{status, brokerID}, new(interface{}))
}

func (r *ReevePluginClient) CLIMethod(method string, args []string) (resp string, err error) {
	params := make([]string, 1+len(args))
	params[0] = method
	copy(params[1:], args)

	err = r.client.Call("Plugin.CLIMethod", params, &resp)
	return
}

type ReevePluginServer struct {
	impl   Plugin
	broker *goplugin.MuxBroker
}

func (r *ReevePluginServer) Name(args *interface{}, resp *string) (err error) {
	*resp, err = r.impl.Name()
	return
}

func (r *ReevePluginServer) Register(args []interface{}, resp *Capabilities) error {
	conn, err := r.broker.Dial(args[1].(uint32))
	if err != nil {
		return err
	}

	api := &ReeveAPIClient{client: rpc.NewClient(conn)}

	*resp, err = r.impl.Register(args[0].(map[string]string), api)
	return err
}

func (r *ReevePluginServer) Unregister(args *interface{}, resp *interface{}) error {
	return r.impl.Unregister()
}

func (r *ReevePluginServer) Message(args schema.FullMessage, resp *interface{}) error {
	return r.impl.Message(args.Source, args.Message)
}

func (r *ReevePluginServer) Discover(args schema.Trigger, resp *[]schema.Pipeline) (err error) {
	*resp, err = r.impl.Discover(args)
	return
}

func (r *ReevePluginServer) Resolve(args []string, resp *map[string]schema.Env) (err error) {
	*resp, err = r.impl.Resolve(args)
	return
}

func (r *ReevePluginServer) Notify(args []interface{}, resp *interface{}) error {
	status := args[0].(schema.PipelineStatus)

	if len(args) == 1 {
		status.Logs = (*LogReaderProviderClient)(nil)
	} else {
		conn, err := r.broker.Dial(args[1].(uint32))
		if err != nil {
			return err
		}
		status.Logs = &LogReaderProviderClient{client: rpc.NewClient(conn), broker: r.broker, closed: make(chan bool)}
	}

	return r.impl.Notify(status)
}

func (r *ReevePluginServer) CLIMethod(args []string, resp *string) (err error) {
	*resp, err = r.impl.CLIMethod(args[0], args[1:])
	return
}

type ReeveAPIClient struct {
	client *rpc.Client
}

func (t *ReeveAPIClient) NotifyMessages(messages []schema.Message) error {
	return t.client.Call("Plugin.NotifyMessages", messages, new(interface{}))
}

func (t *ReeveAPIClient) NotifyTriggers(triggers []schema.Trigger) error {
	return t.client.Call("Plugin.NotifyTriggers", triggers, new(interface{}))
}

func (t *ReeveAPIClient) Close() error {
	return t.client.Close()
}

type ReeveAPIServer struct {
	impl ReeveAPI
}

func (t *ReeveAPIServer) NotifyMessages(args []schema.Message, resp *interface{}) error {
	return t.impl.NotifyMessages(args)
}

func (t *ReeveAPIServer) NotifyTriggers(args []schema.Trigger, resp *interface{}) error {
	return t.impl.NotifyTriggers(args)
}

func (t *ReeveAPIServer) Close(args *interface{}, resp *interface{}) error {
	return nil
}

type LogReaderProviderClient struct {
	client *rpc.Client
	broker *goplugin.MuxBroker
	closed chan bool
}

func (l *LogReaderProviderClient) Available() bool {
	return l != nil
}

func (l *LogReaderProviderClient) Reader() (schema.LogReader, error) {
	if !l.Available() {
		return nil, schema.ERROR_UNAVAILABLE
	}

	var id uint32
	err := l.client.Call("Plugin.Reader", new(interface{}), &id)
	if err != nil {
		return nil, err
	}

	conn, err := l.broker.Dial(id)
	if err != nil {
		return nil, err
	}

	provider := &LogReaderClient{client: rpc.NewClient(conn)}

	go func() {
		<-l.closed
		provider.client.Close()
	}()

	return provider, nil
}

func (l *LogReaderProviderClient) Close() error {
	if !l.Available() {
		return nil
	}

	close(l.closed)

	return l.client.Close()
}

type LogReaderProviderServer struct {
	impl   schema.LogReaderProvider
	broker *goplugin.MuxBroker
}

func (l *LogReaderProviderServer) Reader(args *interface{}, resp *uint32) error {
	reader, err := l.impl.Reader()
	if err != nil {
		return err
	}

	logReaderServer := &LogReaderServer{impl: reader}

	brokerID := l.broker.NextId()
	go func() {
		l.broker.AcceptAndServe(brokerID, logReaderServer)
		reader.Close()
	}()

	*resp = brokerID
	return nil
}

func (l *LogReaderProviderServer) Close(args *interface{}, resp *interface{}) error {
	return nil
}

type LogReaderClient struct {
	client *rpc.Client
}

func (l *LogReaderClient) Read(p []byte) (n int, err error) {
	var resp []byte
	err = l.client.Call("Plugin.Read", len(p), &resp)
	if err != nil {
		if resp != nil {
			n = len(resp)
		}
		if err.Error() == io.EOF.Error() {
			err = io.EOF
		}
		return
	}
	copy(p, resp)
	return len(resp), nil
}

func (l *LogReaderClient) Seek(offset int64, whence int) (n int64, err error) {
	err = l.client.Call("Plugin.Seek", []interface{}{offset, whence}, &n)
	return
}

func (l *LogReaderClient) ReadAt(p []byte, offset int64) (n int, err error) {
	var resp []byte
	err = l.client.Call("Plugin.ReadAt", []int64{int64(len(p)), offset}, &resp)
	if err != nil {
		if resp != nil {
			n = len(resp)
		}
		if err.Error() == io.EOF.Error() {
			err = io.EOF
		}
		return
	}
	copy(p, resp)
	return len(resp), nil
}

func (l *LogReaderClient) Size() (size int64, isClosed bool) {
	var resp []interface{}
	err := l.client.Call("Plugin.Size", new(interface{}), &resp)
	if err != nil {
		size = -1
		isClosed = true
		return
	}
	size = resp[0].(int64)
	isClosed = resp[1].(bool)
	return
}

func (l *LogReaderClient) Close() error {
	return l.client.Close()
}

type LogReaderServer struct {
	impl schema.LogReader
}

func (l *LogReaderServer) Read(args int, resp *[]byte) error {
	*resp = make([]byte, args)
	n, err := l.impl.Read(*resp)
	*resp = (*resp)[0:n]
	return err
}

func (l *LogReaderServer) Seek(args []interface{}, resp *int64) (err error) {
	*resp, err = l.impl.Seek(args[0].(int64), args[1].(int))
	return
}

func (l *LogReaderServer) ReadAt(args []int64, resp *[]byte) error {
	*resp = make([]byte, args[0])
	n, err := l.impl.ReadAt(*resp, args[1])
	*resp = (*resp)[0:n]
	return err
}

func (l *LogReaderServer) Size(args *interface{}, resp *[]interface{}) (err error) {
	size, isClosed := l.impl.Size()
	*resp = []interface{}{size, isClosed}
	return nil
}

func (l *LogReaderServer) Close(args *interface{}, resp *interface{}) error {
	return nil
}

type ReevePlugin struct {
	Impl Plugin
}

func (p *ReevePlugin) Server(b *goplugin.MuxBroker) (interface{}, error) {
	return &ReevePluginServer{impl: p.Impl, broker: b}, nil
}

var _ Plugin = (*ReevePluginClient)(nil)

func (ReevePlugin) Client(b *goplugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ReevePluginClient{client: c, broker: b}, nil
}
