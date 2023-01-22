package server

import (
	"context"
	"fmt"
	platform "github.com/chenxuan520/goweb-platform"
	"github.com/chenxuan520/goweb-platform/logger"
	"github.com/chenxuan520/goweb-platform/mongo"
	"github.com/chenxuan520/goweb-platform/mysql"
	"github.com/chenxuan520/goweb-platform/redis"
	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	shutdownMaxAge = 15 * time.Second
	shutdownWait   = 1000 * time.Millisecond
)
const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

var (
	ApiOptions struct {
		flags.Options
		Environment     string `short:"e" long:"env" description:"Use ApiServer environment" default:"testing"`
		Version         bool   `short:"v" long:"verbose"  description:"Show ApiServer version"`
		EnablePProfile  bool   `short:"p" long:"enable-pprof"  description:"enable pprof"`
		PProfilePort    int    `short:"d" long:"pprof-port"  description:"pprof port" default:"8188"`
		HealthCheckURI  string `short:"i" long:"health-check-uri"  description:"health check uri" default:"/health" `
		HealthCheckPort int    `short:"f" long:"health-check-port"  description:"health check port" default:"8186"`
		ConfigFileName  string `short:"c" long:"config" description:"Use ApiServer config file" default:"main"`
	}
)

type Option func(c *platform.Config)

func WithMysql() Option {
	return func(c *platform.Config) {
		mysqlConfig := c.Mysql
		//db
		dsn := mysqlConfig.EmptyDsn()
		createSql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 ;", mysqlConfig.Dbname)
		if err := mysql.CreateDatabase(dsn, "mysql", createSql); err != nil {
			logger.GetLogger().Error(fmt.Sprintf("create mysql database failed , error:%s", err.Error()))
		}
		_, err := mysql.Init(mysqlConfig.Dsn())
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("api-server:init mysql failed , error:%s", err.Error()))
		} else {
			logger.GetLogger().Info("api-server:init mysql success")
		}
	}
}

func WithRedis() Option {
	return func(c *platform.Config) {
		redisConfig := c.Redis
		//reds
		_, err := redis.Init(redisConfig.Addr, redisConfig.Password, redisConfig.DB)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("api-server:init redis failed , error:%s", err.Error()))
		} else {
			logger.GetLogger().Info("api-server:init redis success")
		}
	}
}

func WithMongo() Option {
	return func(c *platform.Config) {
		mongoConfig := c.Mongo
		err := mongo.Init(mongoConfig.Host, mongoConfig.Port, mongoConfig.DBname, mongoConfig.User, mongoConfig.Password)
		if err != nil {
			logger.GetLogger().Error(fmt.Sprintf("api-server:init mongo failed , error:%s", err.Error()))
		} else {
			logger.GetLogger().Info("api-server:init mongo success")
		}
	}
}

type ApiServer struct {
	Engine      *gin.Engine
	HttpServer  *http.Server
	Addr        string
	mu          sync.Mutex
	doneChan    chan struct{}
	Routers     []func(*gin.Engine)
	Middlewares []func(*gin.Engine)
	Shutdowns   []func(*ApiServer)
	Services    []func(*ApiServer)
}

//get close Chan
func (srv *ApiServer) getDoneChan() <-chan struct{} {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	return srv.getDoneChanLocked()
}

func (srv *ApiServer) getDoneChanLocked() chan struct{} {
	if srv.doneChan == nil {
		srv.doneChan = make(chan struct{})
	}
	return srv.doneChan
}

func (srv *ApiServer) Shutdown(ctx context.Context) {
	//Give priority to business shutdown Hook
	if len(srv.Shutdowns) > 0 {
		for _, shutdown := range srv.Shutdowns {
			shutdown(srv)
		}
	}
	//wait for registry shutdown
	select {
	case <-time.After(shutdownWait):
	}
	// close the HttpServer
	srv.HttpServer.Shutdown(ctx)
}

func (srv *ApiServer) setupSignal() {
	go func() {
		var sigChan = make(chan os.Signal, 1)
		signal.Notify(sigChan /*syscall.SIGUSR1,*/, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownMaxAge)
		defer shutdownCancel()

		for sig := range sigChan {
			if sig == syscall.SIGINT || sig == syscall.SIGHUP || sig == syscall.SIGTERM {
				logger.GetLogger().Error(fmt.Sprintf("Graceful shutdown:signal %v to stop api-server ", sig))
				srv.Shutdown(shutdownCtx)
			} else {
				logger.GetLogger().Info(fmt.Sprintf("Caught signal %v", sig))
			}
		}
		logger.Shutdown()
	}()
}

func NewApiServer(opts ...Option) (*ApiServer, error) {
	var parser = flags.NewParser(&ApiOptions, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}

		return nil, err
	}

	if ApiOptions.Version {
		//TODO
		// fmt.Printf("%s Version:%s\n", platform.ApiModule, platform.Version)
		os.Exit(0)
	}

	if ApiOptions.EnablePProfile {
		go func() {
			fmt.Printf("enable pprof http server at:%d\n", ApiOptions.PProfilePort)
			fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", ApiOptions.PProfilePort), nil))
		}()
	}

	var env = platform.Environment(ApiOptions.Environment)
	if env.Invalid() {
		var err error
		env, err = platform.NewGlobalEnvironment()
		if err != nil {
			return nil, err
		}
	}

	var configFile = ApiOptions.ConfigFileName
	if configFile == "" {
		configFile = "main"
	}
	defaultConfig, err := platform.LoadConfig(env.String(), configFile)
	if err != nil {
		fmt.Printf("api-server:init config error:%s", err.Error())
		return nil, err
	}
	logConfig := defaultConfig.Log
	//log
	logger.Init(logConfig.Level, logConfig.Format, logConfig.Prefix, logConfig.Director, logConfig.ShowLine, logConfig.EncodeLevel, logConfig.StacktraceKey, logConfig.LogInConsole)

	if len(opts) > 0 {
		for _, opt := range opts {
			opt(defaultConfig)
		}
	}
	apiServer := &ApiServer{
		Addr: fmt.Sprintf(":%d", defaultConfig.System.Addr),
	}

	apiServer.setupSignal()
	//set gin mode
	switch env {
	case platform.EnvProduction:
		gin.SetMode(gin.ReleaseMode)
	case platform.EnvTesting:
		gin.SetMode(gin.DebugMode)
	}
	return apiServer, nil
}

// ListenAndServe Listen And Serve()
func (srv *ApiServer) ListenAndServe() error {
	srv.Engine = gin.New()
	srv.Engine.Use(srv.apiRecoveryMiddleware())
	srv.Engine.Use(srv.cors())

	for _, service := range srv.Services {
		service(srv)
	}

	for _, middleware := range srv.Middlewares {
		middleware(srv.Engine)
	}

	for _, c := range srv.Routers {
		c(srv.Engine)
	}

	srv.HttpServer = &http.Server{
		Handler:        srv.Engine,
		Addr:           srv.Addr,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	logger.GetLogger().Info(fmt.Sprintf("api-server port run on %s ", srv.Addr))
	if err := srv.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Register Shutdown Handler
func (srv *ApiServer) RegisterShutdown(handlers ...func(*ApiServer)) {
	srv.Shutdowns = append(srv.Shutdowns, handlers...)
}

// Register Service Handler
func (srv *ApiServer) RegisterService(handlers ...func(*ApiServer)) {
	srv.Services = append(srv.Services, handlers...)
}

// Register Middleware Middleware
func (srv *ApiServer) RegisterMiddleware(middlewares ...func(engine *gin.Engine)) {
	srv.Middlewares = append(srv.Middlewares, middlewares...)
}

// RegisterRouters
func (srv *ApiServer) RegisterRouters(routers ...func(engine *gin.Engine)) *ApiServer {
	srv.Routers = append(srv.Routers, routers...)
	return srv
}
