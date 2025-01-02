package main

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tsinghua-cel/attacker-service/config"
	"github.com/tsinghua-cel/attacker-service/dbmodel"
	"github.com/tsinghua-cel/attacker-service/docs"
	"github.com/tsinghua-cel/attacker-service/reward"
	"github.com/tsinghua-cel/attacker-service/server"
	"github.com/tsinghua-cel/attacker-service/types"
	"time"

	"os"
	"sync"
)

var logLevel string
var logPath string
var maxHackValIdx int
var minHackValIdx int
var timePerStrategyRun int
var configPath string
var strategies string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "attacker",
	Short: "The attacker command-line interface",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runNode()
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&logLevel, "loglevel", "debug", "log level")
	RootCmd.PersistentFlags().StringVar(&logPath, "logpath", "", "log path")
	RootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")
	RootCmd.PersistentFlags().StringVar(&strategies, "strategy", "", "choose the strategy to generate, split multi strategies by comma")
	RootCmd.PersistentFlags().IntVar(&timePerStrategyRun, "duration-per-strategy-run", 30, "time per strategy run (only when set multi strategies), unit: minute")
	RootCmd.PersistentFlags().IntVar(&maxHackValIdx, "max-hack-idx", -1, "max malicious validator index")
	RootCmd.PersistentFlags().IntVar(&minHackValIdx, "min-hack-idx", 0, "min malicious validator index")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	InitLog()

	if configPath != "" {
		_, err := config.ParseConfig(configPath)
		if err != nil {
			log.WithField("error", err).Fatal("parse config failed")
		} else {
			return
		}
	}

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		//log.Info("Using config file:", viper.ConfigFileUsed())
	} else {
		log.WithField("error", err).Fatal("Read config failed")
		return
	}

	conf, err := config.ParseConfig(viper.ConfigFileUsed())
	if err != nil {
		log.WithField("error", err).Fatal("parse config failed")
	}
	if len(conf.SwagHost) != 0 {
		docs.SwaggerInfo.Host = fmt.Sprintf("%s", conf.SwagHost)
	}
}

func runNode() {
	dbmodel.DbInit(config.GetConfig().DbConnect)
	bunnyFinder := server.NewServer(config.GetConfig(), types.StrategyGeneratorParam{
		Strategy:            strategies,
		DurationPerStrategy: int64(timePerStrategyRun),
		MinMaliciousIdx:     minHackValIdx,
		MaxMaliciousIdx:     maxHackValIdx,
	})
	bunnyFinder.Start()

	go getRewardBackgroud()

	wg := sync.WaitGroup{}
	wg.Add(1)

	wg.Wait()
}

func getLogLevel(level string) log.Level {
	switch level {
	case "info":
		return log.InfoLevel
	case "debug":
		return log.DebugLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}

func InitLog() {
	// standard setting
	log.SetLevel(getLogLevel(logLevel))
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05.000"})

	// file system logger setting
	if logPath != "" {
		localFilesystemLogger(logPath)
	}
}

func logWriter(logPath string) *rotatelogs.RotateLogs {
	logFullPath := logPath
	logwriter, err := rotatelogs.New(
		logFullPath+".%Y%m%d",
		rotatelogs.WithLinkName(logFullPath),
		rotatelogs.WithRotationSize(100*1024*1024), // 100MB
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		panic(err)
	}
	return logwriter
}

func localFilesystemLogger(logPath string) {
	lfHook := lfshook.NewHook(logWriter(logPath), &log.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05.000"})
	log.AddHook(lfHook)
}

func getRewardBackgroud() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.WithFields(log.Fields{
				"beacon": config.GetConfig().BeaconRpc,
			}).Debug("goto get reward")
			err := reward.GetRewardsToMysql(config.GetConfig().BeaconRpc)
			//err := reward.GetRewards(config.GetConfig().BeaconRpc, config.GetConfig().RewardFile)
			if err != nil {
				log.WithError(err).Error("collect reward failed")
			}
			reward.GetRewards(config.GetConfig().BeaconRpc, config.GetConfig().RewardFile)
		}
	}
}
