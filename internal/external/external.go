package external

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/instill-ai/controller/config"
	"github.com/instill-ai/controller/internal/logger"
	"github.com/instill-ai/controller/internal/triton"

	connectorPB "github.com/instill-ai/protogen-go/vdp/connector/v1alpha"
	mgmtPB "github.com/instill-ai/protogen-go/vdp/mgmt/v1alpha"
	modelPB "github.com/instill-ai/protogen-go/vdp/model/v1alpha"
	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1alpha"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// InitTritonServiceClient initialises a TritonServiceClient instance
func InitEtcdServiceClient() (*etcdv3.Client) {
	logger, _ := logger.GetZapLogger()

	var clientDialOpts grpc.DialOption
	if config.Config.ConnectorBackend.HTTPS.Cert != "" && config.Config.ConnectorBackend.HTTPS.Key != "" {
		creds, err := credentials.NewServerTLSFromFile(config.Config.ConnectorBackend.HTTPS.Cert, config.Config.ConnectorBackend.HTTPS.Key)
		if err != nil {
			logger.Fatal(err.Error())
		}
		clientDialOpts = grpc.WithTransportCredentials(creds)
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	host := config.Config.Etcd.Host
	port := config.Config.Etcd.Port
	timeout := config.Config.Etcd.Timeout * time.Second
	// Create etcd client
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{fmt.Sprintf("%s:%s", host, port)}, //TODO: multiple nodes
		DialTimeout: timeout,
		DialOptions: []grpc.DialOption{clientDialOpts},
	})
	if err != nil {
		logger.Fatal(err.Error())
	}

	return client
}

// InitTritonServiceClient initialises a TritonServiceClient instance
func InitTritonServiceClient() (inferenceserver.GRPCInferenceServiceClient, *grpc.ClientConn) {
	logger, _ := logger.GetZapLogger()
	grpcUri := config.Config.TritonServer.GrpcURI
	// Connect to triton grpc server
	clientConn, err := grpc.Dial(grpcUri, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal(err.Error())
	}

	return inferenceserver.NewGRPCInferenceServiceClient(clientConn), clientConn
}

// InitConnectorPublicServiceClient initialises a ConnectorPublicServiceClient instance
func InitConnectorPublicServiceClient() (connectorPB.ConnectorPublicServiceClient, *grpc.ClientConn) {
	logger, _ := logger.GetZapLogger()

	var clientDialOpts grpc.DialOption
	if config.Config.ConnectorBackend.HTTPS.Cert != "" && config.Config.ConnectorBackend.HTTPS.Key != "" {
		creds, err := credentials.NewServerTLSFromFile(config.Config.ConnectorBackend.HTTPS.Cert, config.Config.ConnectorBackend.HTTPS.Key)
		if err != nil {
			logger.Fatal(err.Error())
		}
		clientDialOpts = grpc.WithTransportCredentials(creds)
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("%v:%v", config.Config.ConnectorBackend.Host, config.Config.ConnectorBackend.Port), clientDialOpts)
	if err != nil {
		logger.Error(err.Error())
		return nil, nil
	}

	return connectorPB.NewConnectorPublicServiceClient(clientConn), clientConn
}

// InitModelPublicServiceClient initialises a ModelPublicServiceClient instance
func InitModelPublicServiceClient() (modelPB.ModelPublicServiceClient, *grpc.ClientConn) {
	logger, _ := logger.GetZapLogger()

	var clientDialOpts grpc.DialOption
	if config.Config.ModelBackend.HTTPS.Cert != "" && config.Config.ModelBackend.HTTPS.Key != "" {
		creds, err := credentials.NewServerTLSFromFile(config.Config.ModelBackend.HTTPS.Cert, config.Config.ModelBackend.HTTPS.Key)
		if err != nil {
			logger.Fatal(err.Error())
		}
		clientDialOpts = grpc.WithTransportCredentials(creds)
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("%v:%v", config.Config.ModelBackend.Host, config.Config.ModelBackend.PublicPort), clientDialOpts)
	if err != nil {
		logger.Error(err.Error())
		return nil, nil
	}

	return modelPB.NewModelPublicServiceClient(clientConn), clientConn
}

// InitModelPrivateServiceClient initialises a ModelPrivateServiceClient instance
func InitModelPrivateServiceClient() (modelPB.ModelPrivateServiceClient, *grpc.ClientConn) {
	logger, _ := logger.GetZapLogger()

	var clientDialOpts grpc.DialOption
	if config.Config.ModelBackend.HTTPS.Cert != "" && config.Config.ModelBackend.HTTPS.Key != "" {
		creds, err := credentials.NewServerTLSFromFile(config.Config.ModelBackend.HTTPS.Cert, config.Config.ModelBackend.HTTPS.Key)
		if err != nil {
			logger.Fatal(err.Error())
		}
		clientDialOpts = grpc.WithTransportCredentials(creds)
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("%v:%v", config.Config.ModelBackend.Host, config.Config.ModelBackend.PrivatePort), clientDialOpts)
	if err != nil {
		logger.Error(err.Error())
		return nil, nil
	}

	return modelPB.NewModelPrivateServiceClient(clientConn), clientConn
}

// InitMgmtPublicServiceClient initialises a MgmtPublicServiceClient instance
func InitMgmtPublicServiceClient() (mgmtPB.MgmtPublicServiceClient, *grpc.ClientConn) {
	logger, _ := logger.GetZapLogger()

	var clientDialOpts grpc.DialOption
	var creds credentials.TransportCredentials
	var err error
	if config.Config.MgmtBackend.HTTPS.Cert != "" && config.Config.MgmtBackend.HTTPS.Key != "" {
		creds, err = credentials.NewServerTLSFromFile(config.Config.MgmtBackend.HTTPS.Cert, config.Config.MgmtBackend.HTTPS.Key)
		if err != nil {
			logger.Fatal(err.Error())
		}
		clientDialOpts = grpc.WithTransportCredentials(creds)
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("%v:%v", config.Config.MgmtBackend.Host, config.Config.MgmtBackend.PublicPort), clientDialOpts)
	if err != nil {
		logger.Error(err.Error())
		return nil, nil
	}

	return mgmtPB.NewMgmtPublicServiceClient(clientConn), clientConn
}

// InitPipelinePublicServiceClient initialises a PipelinePublicServiceClient instance
func InitPipelinePublicServiceClient() (pipelinePB.PipelinePublicServiceClient, *grpc.ClientConn) {
	logger, _ := logger.GetZapLogger()

	var clientDialOpts grpc.DialOption
	var creds credentials.TransportCredentials
	var err error
	if config.Config.PipelineBackend.HTTPS.Cert != "" && config.Config.PipelineBackend.HTTPS.Key != "" {
		creds, err = credentials.NewServerTLSFromFile(config.Config.PipelineBackend.HTTPS.Cert, config.Config.PipelineBackend.HTTPS.Key)
		if err != nil {
			logger.Fatal(err.Error())
		}
		clientDialOpts = grpc.WithTransportCredentials(creds)
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	clientConn, err := grpc.Dial(fmt.Sprintf("%v:%v", config.Config.PipelineBackend.Host, config.Config.PipelineBackend.Port), clientDialOpts)
	if err != nil {
		logger.Error(err.Error())
		return nil, nil
	}

	return pipelinePB.NewPipelinePublicServiceClient(clientConn), clientConn
}
