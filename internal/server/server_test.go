package server

import (
	"context"
	"io/ioutil"
	"net"
	"testing"

	"github.com/sant470/distlogs/api/v1"
	"github.com/sant470/distlogs/internal/config"
	"github.com/sant470/distlogs/internal/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(
		t *testing.T,
		client api.LogClient,
		config *Config,
	){
		"produce/consume a message to/from the log succeeds": testProduceConsume,
		"produce/consume stream succeeds":                    testProduceConsumeStream,
		"consume past log boundary fails":                    testConsumePastBoundary,
	} {
		t.Run(scenario, func(t *testing.T) {
			client, config, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, client, config)
		})
	}
}

func setupTest(t *testing.T, fn func(*Config)) (api.LogClient, *Config, func()) {
	t.Helper()

	// Start server
	l, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)

	server, cfg, err := startServer(t, l, fn)
	require.NoError(t, err)

	// Start client
	client, cc, err := startClient(t, l)
	require.NoError(t, err)

	// Cleanup function
	return client, cfg, func() {
		server.Stop()
		cc.Close()
		l.Close()
	}
}

// startServer initializes the gRPC server with TLS
func startServer(t *testing.T, l net.Listener, fn func(*Config)) (*grpc.Server, *Config, error) {
	t.Helper()

	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		ServerAddress: l.Addr().String(),
	})
	require.NoError(t, err)

	serverCreds := credentials.NewTLS(serverTLSConfig)

	// Create temporary directory for log storage
	dir, err := ioutil.TempDir("", "server-test")
	require.NoError(t, err)

	clog, err := log.NewLog(dir, log.Config{})
	require.NoError(t, err)

	cfg := &Config{CommitLog: clog}

	server, err := NewGRPCServer(cfg, grpc.Creds(serverCreds))
	require.NoError(t, err)

	// Allow modifying config before starting the server
	if fn != nil {
		fn(cfg)
	}

	go func() {
		server.Serve(l)
	}()

	return server, cfg, nil
}

// startClient initializes the gRPC client with TLS
func startClient(t *testing.T, l net.Listener) (api.LogClient, *grpc.ClientConn, error) {
	t.Helper()

	// fmt.Println("Using CAFile:", config.CAFile)
	clientTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CAFile:   config.CAFile,
		CertFile: config.ClientCertFile,
		KeyFile:  config.ClientKeyFile,
	})
	require.NoError(t, err)

	clientCreds := credentials.NewTLS(clientTLSConfig)

	cc, err := grpc.NewClient(
		l.Addr().String(),
		grpc.WithTransportCredentials(clientCreds),
	)
	require.NoError(t, err)

	client := api.NewLogClient(cc)
	return client, cc, nil
}

func testProduceConsume(t *testing.T, client api.LogClient, config *Config) {
	ctx := context.Background()
	want := &api.Record{Value: []byte("hello world")}
	produce, err := client.Produce(ctx, &api.ProduceRequest{Record: want})
	require.NoError(t, err)
	consume, err := client.Consume(ctx, &api.ConsumeRequest{Offset: produce.Offset})
	require.NoError(t, err)
	require.Equal(t, want.Value, consume.Record.Value)
	require.Equal(t, want.Offset, consume.Record.Offset)
}

func testConsumePastBoundary(t *testing.T, client api.LogClient, config *Config) {
	ctx := context.Background()
	produce, err := client.Produce(ctx, &api.ProduceRequest{Record: &api.Record{Value: []byte("hello world!")}})
	require.NoError(t, err)
	consume, err := client.Consume(ctx, &api.ConsumeRequest{Offset: produce.Offset + 1})
	if consume != nil {
		t.Fatal("consume not nil")
	}
	got := grpc.Code(err)
	want := grpc.Code(api.ErrOffsetOutOfRange{}.GRPCStatus().Err())
	if got != want {
		t.Fatalf("got err: %v, want: %v", got, want)
	}
}

func testProduceConsumeStream(t *testing.T, client api.LogClient, config *Config) {
	ctx := context.Background()
	records := []*api.Record{{
		Value:  []byte("1st message"),
		Offset: 0,
	}, {
		Value:  []byte("2nd message"),
		Offset: 1,
	}}
	{
		stream, err := client.ProduceStream(ctx)
		require.NoError(t, err)
		for offset, record := range records {
			err = stream.Send(&api.ProduceRequest{Record: record})
			require.NoError(t, err)
			res, err := stream.Recv()
			require.NoError(t, err)
			if res.Offset != uint64(offset) {
				t.Fatalf("got offset: %d, want: %d", res.Offset, offset)
			}
		}
	}
	{
		stream, err := client.ConsumeStream(ctx, &api.ConsumeRequest{Offset: 0})
		require.NoError(t, err)
		for i, record := range records {
			res, err := stream.Recv()
			require.NoError(t, err)
			require.Equal(t, res.Record, &api.Record{Value: record.Value, Offset: uint64(i)})
		}
	}
}
