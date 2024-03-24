package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"

	"github.com/wendtek/kube-wan-dns-refresh/pkg/aws"
	"github.com/wendtek/kube-wan-dns-refresh/pkg/config"
	"github.com/wendtek/kube-wan-dns-refresh/pkg/wan"
)

func main() {
	ctx := context.Background()
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	// Parse the command line flags
	cfg := config.NewConfig()
	err := cfg.ParseFlags()
	if err != nil {
		logger.Error(fmt.Sprintf("Error parsing flags: %+v", err))
		return
	}

	// Read the configuration file
	_, err = cfg.ReadConfig(cfg.ConfigFilePath)
	if err != nil {
		logger.Error(fmt.Sprintf("Error reading config: %+v", err))
		return
	}

	// Load the AWS configuration and create r53 client
	awsCfg, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		func(*awsConfig.LoadOptions) error {
			return nil
		},
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Error creating route53 client: %+v", err))
		return
	}
	r53Client := route53.NewFromConfig(awsCfg)

	// Print the configuration
	logger.Info(fmt.Sprintf("Config: %s", cfg.ToString()))

	// Get the current WAN IP
	wanIP, err := wan.GetWanIP()
	if err != nil {
		logger.Error(fmt.Sprintf("Error getting WAN IP: %+v", err))
		return
	}
	logger.Info(fmt.Sprintf("WAN IP: %s", wanIP))

	// Update the Route53 records
	err = aws.SyncRecords(ctx, cfg, wanIP, r53Client)
	if err != nil {
		logger.Error(fmt.Sprintf("Error updating Route53 records: %+v", err))
		return
	}
}
