package cmd

import (
	"dedicated-container-ingress-controller/pkg/server"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func serverCmd() *cobra.Command {
	serverArgs := server.DefaultArgs()

	cmd := &cobra.Command{
		Use:          "server",
		Short:        "Starts SampleApplicaiton as a server",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("%q is an invalid argument", args[0])
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := server.Run(serverArgs)
			if err != nil {
				log.Fatalf("Failed to run server.Run: %s\n", err.Error())
			}
		},
	}

	cmd.PersistentFlags().StringVarP(
		&serverArgs.APIAddress,
		"api-address",
		"",
		serverArgs.APIAddress,
		"Address to use API",
	)
	cmd.PersistentFlags().Int64VarP(
		&serverArgs.APIMaxConnections,
		"api-max-connections",
		"",
		serverArgs.APIMaxConnections,
		"Max connections of API",
	)
	cmd.PersistentFlags().StringVarP(
		&serverArgs.MonitorAddress,
		"monitor-address",
		"",
		serverArgs.MonitorAddress,
		"Address to use self-monitoring information",
	)
	cmd.PersistentFlags().Int64VarP(
		&serverArgs.MonitorMaxConnections,
		"monitor-max-connections",
		"",
		serverArgs.MonitorMaxConnections,
		"Max connections of self-monitoring information",
	)
	cmd.PersistentFlags().StringVarP(
		&serverArgs.MonitoringJaegerEndpoint,
		"monitoring-jaeger-endpoint",
		"",
		serverArgs.MonitoringJaegerEndpoint,
		"Address to use for distributed tracing",
	)
	cmd.PersistentFlags().BoolVarP(
		&serverArgs.EnableProfiling,
		"enable-profiling",
		"",
		serverArgs.EnableProfiling,
		"Enable profiling",
	)
	cmd.PersistentFlags().BoolVarP(
		&serverArgs.EnableTracing,
		"enable-tracing",
		"",
		serverArgs.EnableTracing,
		"Enable distributed tracing",
	)
	cmd.PersistentFlags().Float64VarP(
		&serverArgs.TracingSampleRate,
		"tracing-sample-rate",
		"",
		serverArgs.TracingSampleRate,
		"Tracing sample rate",
	)
	cmd.PersistentFlags().StringVarP(
		&serverArgs.RedisHost,
		"redis-host",
		"",
		serverArgs.RedisHost,
		"Target redis host",
	)
	cmd.PersistentFlags().Int64VarP(
		&serverArgs.RedisPort,
		"redis-port",
		"",
		serverArgs.RedisPort,
		"Target redis port",
	)
	cmd.PersistentFlags().IntVarP(
		&serverArgs.RedisDB,
		"redis-db",
		"",
		serverArgs.RedisDB,
		"Target redis db",
	)
	cmd.PersistentFlags().StringVarP(
		&serverArgs.CookieSecretKey,
		"cookie-secret-key",
		"",
		serverArgs.CookieSecretKey,
		"Secret key of CookieStore",
	)
	cmd.PersistentFlags().IntVarP(
		&serverArgs.CookieMaxAge,
		"cookie-max-age",
		"",
		serverArgs.CookieMaxAge,
		"Max age of CookieStore",
	)
	cmd.PersistentFlags().Int64VarP(
		&serverArgs.PodsLimit,
		"pods-limit",
		"",
		serverArgs.PodsLimit,
		"Max count of factory created pods",
	)
	cmd.PersistentFlags().BoolVarP(
		&serverArgs.KeepAlived,
		"enable-keep-alived",
		"",
		serverArgs.KeepAlived,
		"Enable HTTP KeepAlive",
	)
	cmd.PersistentFlags().BoolVarP(
		&serverArgs.ReUsePort,
		"enable-reuseport",
		"",
		serverArgs.ReUsePort,
		"Enable SO_REUSEPORT",
	)
	cmd.PersistentFlags().Int64VarP(
		&serverArgs.TCPKeepAliveInterval,
		"tcp-keep-alive-interval",
		"",
		serverArgs.TCPKeepAliveInterval,
		"Interval of TCP KeepAlive",
	)
	cmd.PersistentFlags().BoolVarP(
		&serverArgs.Verbose,
		"verbose",
		"",
		serverArgs.Verbose,
		"Verbose logging",
	)

	if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
		log.Fatalf("Failed to execute server command: %s\n", err.Error())
	}

	return cmd
}
