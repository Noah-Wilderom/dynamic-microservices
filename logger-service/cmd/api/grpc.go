package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Noah-Wilderom/dynamic-microservices/logger-service/data"
	"github.com/Noah-Wilderom/dynamic-microservices/shared/logs"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) GetVersionString(ctx context.Context, req *logs.VersionRequest) (*logs.VersionResponse, error) {
	type VersionConfig struct {
		Version string `json:"version"`
	}

	var config VersionConfig
	configFile, err := os.ReadFile("/config.json")
	if err != nil {
		log.Println("Cannot find config file")
		return nil, err
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Println("Cannot parse config file")
		return nil, err
	}

	return &logs.VersionResponse{
		Version: config.Version,
	}, nil
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	logEntry := data.LogEntry{
		Type: input.Type,
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		res := &logs.LogResponse{
			Result: "Failed",
		}
		return res, err
	}

	res := &logs.LogResponse{
		Result: "Logged",
	}

	return res, nil
}

func (app *Config) gRPCListen() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", app.GRPCPort))
	if err != nil {
		log.Fatalln("Failed to listen to gRPC on port", app.GRPCPort)
	}

	s := grpc.NewServer()

	logs.RegisterLogServiceServer(s, &LogServer{Models: app.Models})
	log.Println("gRPC Server started on port", app.GRPCPort)

	if err = s.Serve(lis); err != nil {
		log.Fatalln("Failed to listen to gRPC on port", app.GRPCPort)
	}
}
