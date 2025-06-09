# Production Deployment Guide

Comprehensive guide for deploying SimConnect applications in production environments with monitoring, scaling, and operational considerations.

## Table of Contents

- [Production Architecture](#production-architecture)
- [Configuration Management](#configuration-management)
- [Monitoring and Observability](#monitoring-and-observability)
- [Deployment Strategies](#deployment-strategies)
- [Scaling Considerations](#scaling-considerations)
- [Security Best Practices](#security-best-practices)
- [Operational Procedures](#operational-procedures)

## Production Architecture

### Microservices Architecture

```go
// main.go - Production entry point
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "your-app/internal/config"
    "your-app/internal/monitoring"
    "your-app/internal/services"
)

type Application struct {
    config      *config.Config
    monitor     *monitoring.Monitor
    services    map[string]services.Service
    healthCheck *services.HealthCheck
    ctx         context.Context
    cancel      context.CancelFunc
}

func main() {
    var configPath = flag.String("config", "config.yaml", "Configuration file path")
    var logLevel = flag.String("log-level", "info", "Logging level")
    flag.Parse()

    // Load configuration
    cfg, err := config.Load(*configPath)
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Initialize application
    app, err := NewApplication(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize application: %v", err)
    }

    // Start application
    if err := app.Start(); err != nil {
        log.Fatalf("Failed to start application: %v", err)
    }

    // Wait for shutdown signal
    app.WaitForShutdown()
}

func NewApplication(cfg *config.Config) (*Application, error) {
    ctx, cancel := context.WithCancel(context.Background())

    // Initialize monitoring
    monitor, err := monitoring.New(cfg.Monitoring)
    if err != nil {
        cancel()
        return nil, fmt.Errorf("failed to initialize monitoring: %w", err)
    }

    // Initialize services
    services := map[string]services.Service{
        "flight-monitor":    services.NewFlightMonitor(cfg.FlightMonitor),
        "data-collector":    services.NewDataCollector(cfg.DataCollector),
        "alert-manager":     services.NewAlertManager(cfg.AlertManager),
        "metrics-exporter":  services.NewMetricsExporter(cfg.MetricsExporter),
    }

    // Initialize health check
    healthCheck := services.NewHealthCheck(services)

    return &Application{
        config:      cfg,
        monitor:     monitor,
        services:    services,
        healthCheck: healthCheck,
        ctx:         ctx,
        cancel:      cancel,
    }, nil
}

func (app *Application) Start() error {
    log.Println("🚀 Starting production application...")

    // Start monitoring first
    if err := app.monitor.Start(app.ctx); err != nil {
        return fmt.Errorf("failed to start monitoring: %w", err)
    }

    // Start health check service
    if err := app.healthCheck.Start(app.ctx); err != nil {
        return fmt.Errorf("failed to start health check: %w", err)
    }

    // Start all services
    for name, service := range app.services {
        log.Printf("Starting service: %s", name)
        if err := service.Start(app.ctx); err != nil {
            return fmt.Errorf("failed to start service %s: %w", name, err)
        }
        
        // Record service startup
        app.monitor.RecordServiceStart(name)
    }

    log.Println("✅ All services started successfully")
    return nil
}

func (app *Application) WaitForShutdown() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    select {
    case sig := <-sigChan:
        log.Printf("🛑 Received signal %v, initiating shutdown...", sig)
        app.Shutdown()
    case <-app.ctx.Done():
        log.Println("🛑 Context cancelled, initiating shutdown...")
        app.Shutdown()
    }
}

func (app *Application) Shutdown() {
    log.Println("🔄 Starting graceful shutdown...")
    
    // Create shutdown context with timeout
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()

    // Stop services in reverse order
    serviceNames := []string{"metrics-exporter", "alert-manager", "data-collector", "flight-monitor"}
    for _, name := range serviceNames {
        if service, exists := app.services[name]; exists {
            log.Printf("Stopping service: %s", name)
            if err := service.Stop(shutdownCtx); err != nil {
                log.Printf("⚠️ Error stopping service %s: %v", name, err)
            } else {
                app.monitor.RecordServiceStop(name)
            }
        }
    }

    // Stop health check
    if err := app.healthCheck.Stop(shutdownCtx); err != nil {
        log.Printf("⚠️ Error stopping health check: %v", err)
    }

    // Stop monitoring last
    if err := app.monitor.Stop(shutdownCtx); err != nil {
        log.Printf("⚠️ Error stopping monitoring: %v", err)
    }

    // Cancel application context
    app.cancel()

    log.Println("✅ Graceful shutdown completed")
}
```

### Service Interface

```go
// internal/services/service.go
package services

import "context"

type Service interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() HealthStatus
    Name() string
}

type HealthStatus struct {
    Status      string                 `json:"status"`
    LastCheck   time.Time             `json:"last_check"`
    Details     map[string]interface{} `json:"details,omitempty"`
    Errors      []string              `json:"errors,omitempty"`
}

const (
    StatusHealthy   = "healthy"
    StatusUnhealthy = "unhealthy"
    StatusStarting  = "starting"
    StatusStopping  = "stopping"
)
```

### Flight Monitor Service

```go
// internal/services/flight_monitor.go
package services

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
    "github.com/mycrew-online/sdk/pkg/types"
    "your-app/internal/config"
    "your-app/internal/metrics"
)

type FlightMonitor struct {
    name           string
    config         *config.FlightMonitorConfig
    sdk            client.Connection
    metrics        *metrics.FlightMetrics
    isRunning      bool
    lastDataTime   time.Time
    dataCount      int64
    errorCount     int64
    mu             sync.RWMutex
    stopChan       chan struct{}
    wg             sync.WaitGroup
}

func NewFlightMonitor(cfg *config.FlightMonitorConfig) *FlightMonitor {
    return &FlightMonitor{
        name:     "flight-monitor",
        config:   cfg,
        sdk:      client.New("ProductionFlightMonitor"),
        metrics:  metrics.NewFlightMetrics(),
        stopChan: make(chan struct{}),
    }
}

func (fm *FlightMonitor) Start(ctx context.Context) error {
    fm.mu.Lock()
    defer fm.mu.Unlock()

    if fm.isRunning {
        return fmt.Errorf("flight monitor is already running")
    }

    // Connect to SimConnect with retry
    if err := fm.connectWithRetry(3); err != nil {
        return fmt.Errorf("failed to connect to SimConnect: %w", err)
    }

    // Register flight variables
    if err := fm.registerVariables(); err != nil {
        fm.sdk.Close()
        return fmt.Errorf("failed to register variables: %w", err)
    }

    // Start message processing
    fm.wg.Add(1)
    go fm.processMessages(ctx)

    // Start metrics collection
    fm.wg.Add(1)
    go fm.collectMetrics(ctx)

    fm.isRunning = true
    log.Printf("✅ Flight monitor started")
    return nil
}

func (fm *FlightMonitor) Stop(ctx context.Context) error {
    fm.mu.Lock()
    defer fm.mu.Unlock()

    if !fm.isRunning {
        return nil
    }

    log.Println("🔄 Stopping flight monitor...")

    // Signal stop
    close(fm.stopChan)

    // Wait for goroutines with timeout
    done := make(chan struct{})
    go func() {
        fm.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        log.Println("✅ Flight monitor stopped gracefully")
    case <-ctx.Done():
        log.Println("⚠️ Flight monitor stop timeout")
        return ctx.Err()
    }

    // Close SDK connection
    if err := fm.sdk.Close(); err != nil {
        log.Printf("⚠️ Error closing SDK connection: %v", err)
    }

    fm.isRunning = false
    return nil
}

func (fm *FlightMonitor) Health() HealthStatus {
    fm.mu.RLock()
    defer fm.mu.RUnlock()

    status := StatusHealthy
    var errors []string

    // Check if service is running
    if !fm.isRunning {
        status = StatusUnhealthy
        errors = append(errors, "service not running")
    }

    // Check data freshness
    if fm.isRunning && time.Since(fm.lastDataTime) > 30*time.Second {
        status = StatusUnhealthy
        errors = append(errors, fmt.Sprintf("no data received for %v", time.Since(fm.lastDataTime)))
    }

    // Check error rate
    if fm.errorCount > 0 && fm.dataCount > 0 {
        errorRate := float64(fm.errorCount) / float64(fm.dataCount)
        if errorRate > 0.1 { // 10% error rate threshold
            status = StatusUnhealthy
            errors = append(errors, fmt.Sprintf("high error rate: %.2f%%", errorRate*100))
        }
    }

    return HealthStatus{
        Status:    status,
        LastCheck: time.Now(),
        Details: map[string]interface{}{
            "data_count":     fm.dataCount,
            "error_count":    fm.errorCount,
            "last_data_time": fm.lastDataTime,
            "uptime":         time.Since(fm.lastDataTime),
        },
        Errors: errors,
    }
}

func (fm *FlightMonitor) Name() string {
    return fm.name
}

func (fm *FlightMonitor) connectWithRetry(maxRetries int) error {
    for attempt := 1; attempt <= maxRetries; attempt++ {
        if err := fm.sdk.Open(); err != nil {
            if attempt == maxRetries {
                return fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
            }
            log.Printf("Connection attempt %d failed: %v. Retrying...", attempt, err)
            time.Sleep(time.Duration(attempt) * time.Second)
            continue
        }
        return nil
    }
    return fmt.Errorf("unexpected retry loop exit")
}

func (fm *FlightMonitor) registerVariables() error {
    variables := []struct {
        id     uint32
        name   string
        units  string
        period types.SimConnectPeriod
    }{
        {1, "PLANE ALTITUDE", "feet", types.SIMCONNECT_PERIOD_SECOND},
        {2, "AIRSPEED INDICATED", "knots", types.SIMCONNECT_PERIOD_SECOND},
        {3, "HEADING INDICATOR", "degrees", types.SIMCONNECT_PERIOD_SECOND},
        {4, "FUEL TOTAL QUANTITY", "gallons", types.SIMCONNECT_PERIOD_SECOND},
    }

    for _, v := range variables {
        if err := fm.sdk.RegisterSimVarDefinition(v.id, v.name, v.units, types.SIMCONNECT_DATATYPE_FLOAT32); err != nil {
            return fmt.Errorf("failed to register variable %s: %w", v.name, err)
        }

        if err := fm.sdk.RequestSimVarDataPeriodic(v.id, v.id*100, v.period); err != nil {
            return fmt.Errorf("failed to request variable %s: %w", v.name, err)
        }
    }

    return nil
}

func (fm *FlightMonitor) processMessages(ctx context.Context) {
    defer fm.wg.Done()

    messages := fm.sdk.Listen()
    if messages == nil {
        log.Println("❌ Failed to get message channel")
        return
    }

    for {
        select {
        case <-ctx.Done():
            return
        case <-fm.stopChan:
            return
        case msg, ok := <-messages:
            if !ok {
                log.Println("⚠️ Message channel closed")
                return
            }
            fm.handleMessage(msg)
        }
    }
}

func (fm *FlightMonitor) handleMessage(msg any) {
    fm.mu.Lock()
    fm.dataCount++
    fm.lastDataTime = time.Now()
    fm.mu.Unlock()

    msgMap, ok := msg.(map[string]any)
    if !ok {
        fm.mu.Lock()
        fm.errorCount++
        fm.mu.Unlock()
        return
    }

    switch msgMap["type"] {
    case "SIMOBJECT_DATA":
        fm.handleSimObjectData(msgMap)
    case "EXCEPTION":
        fm.handleException(msgMap)
    }
}

func (fm *FlightMonitor) handleSimObjectData(msgMap map[string]any) {
    if data, exists := msgMap["parsed_data"]; exists {
        if simVar, ok := data.(*client.SimVarData); ok {
            // Update metrics
            fm.metrics.UpdateVariable(simVar.DefineID, simVar.Value)
            
            // Log critical values
            if simVar.DefineID == 1 { // Altitude
                if altitude, ok := simVar.Value.(float64); ok && altitude < 100 {
                    log.Printf("⚠️ Low altitude warning: %.0f feet", altitude)
                }
            }
        }
    }
}

func (fm *FlightMonitor) collectMetrics(ctx context.Context) {
    defer fm.wg.Done()

    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-fm.stopChan:
            return
        case <-ticker.C:
            fm.publishMetrics()
        }
    }
}

func (fm *FlightMonitor) publishMetrics() {
    fm.mu.RLock()
    dataCount := fm.dataCount
    errorCount := fm.errorCount
    fm.mu.RUnlock()

    // Publish metrics to monitoring system
    fm.metrics.PublishCounters(dataCount, errorCount)
}
```

## Configuration Management

### Configuration Structure

```go
// internal/config/config.go
package config

import (
    "fmt"
    "os"
    "time"
    
    "gopkg.in/yaml.v2"
)

type Config struct {
    App             AppConfig             `yaml:"app"`
    FlightMonitor   FlightMonitorConfig   `yaml:"flight_monitor"`
    DataCollector   DataCollectorConfig   `yaml:"data_collector"`
    AlertManager    AlertManagerConfig    `yaml:"alert_manager"`
    MetricsExporter MetricsExporterConfig `yaml:"metrics_exporter"`
    Monitoring      MonitoringConfig      `yaml:"monitoring"`
    Logging         LoggingConfig         `yaml:"logging"`
    Security        SecurityConfig        `yaml:"security"`
}

type AppConfig struct {
    Name        string        `yaml:"name"`
    Version     string        `yaml:"version"`
    Environment string        `yaml:"environment"`
    LogLevel    string        `yaml:"log_level"`
    ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type FlightMonitorConfig struct {
    Enabled         bool          `yaml:"enabled"`
    UpdateInterval  time.Duration `yaml:"update_interval"`
    SimConnectDLL   string        `yaml:"simconnect_dll"`
    RetryAttempts   int           `yaml:"retry_attempts"`
    Variables       []Variable    `yaml:"variables"`
}

type Variable struct {
    ID       uint32 `yaml:"id"`
    Name     string `yaml:"name"`
    Units    string `yaml:"units"`
    Critical bool   `yaml:"critical"`
}

type DataCollectorConfig struct {
    Enabled        bool          `yaml:"enabled"`
    OutputFormat   string        `yaml:"output_format"`
    OutputPath     string        `yaml:"output_path"`
    BufferSize     int           `yaml:"buffer_size"`
    FlushInterval  time.Duration `yaml:"flush_interval"`
    Compression    bool          `yaml:"compression"`
}

type AlertManagerConfig struct {
    Enabled     bool                    `yaml:"enabled"`
    SMTPConfig  SMTPConfig             `yaml:"smtp"`
    SlackConfig SlackConfig            `yaml:"slack"`
    Rules       []AlertRule            `yaml:"rules"`
}

type SMTPConfig struct {
    Server   string `yaml:"server"`
    Port     int    `yaml:"port"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
    From     string `yaml:"from"`
    To       []string `yaml:"to"`
}

type SlackConfig struct {
    WebhookURL string `yaml:"webhook_url"`
    Channel    string `yaml:"channel"`
    Username   string `yaml:"username"`
}

type AlertRule struct {
    Name      string            `yaml:"name"`
    Metric    string            `yaml:"metric"`
    Condition string            `yaml:"condition"`
    Threshold float64           `yaml:"threshold"`
    Duration  time.Duration     `yaml:"duration"`
    Severity  string            `yaml:"severity"`
    Actions   []string          `yaml:"actions"`
}

type MetricsExporterConfig struct {
    Enabled    bool   `yaml:"enabled"`
    Type       string `yaml:"type"` // prometheus, influxdb, etc.
    Address    string `yaml:"address"`
    Port       int    `yaml:"port"`
    Path       string `yaml:"path"`
    Interval   time.Duration `yaml:"interval"`
}

type MonitoringConfig struct {
    Enabled         bool          `yaml:"enabled"`
    HealthCheckPort int           `yaml:"health_check_port"`
    MetricsPort     int           `yaml:"metrics_port"`
    ProfilingPort   int           `yaml:"profiling_port"`
    CheckInterval   time.Duration `yaml:"check_interval"`
}

type LoggingConfig struct {
    Level      string `yaml:"level"`
    Format     string `yaml:"format"` // json, text
    Output     string `yaml:"output"` // stdout, file
    FilePath   string `yaml:"file_path"`
    MaxSize    int    `yaml:"max_size_mb"`
    MaxBackups int    `yaml:"max_backups"`
    MaxAge     int    `yaml:"max_age_days"`
    Compress   bool   `yaml:"compress"`
}

type SecurityConfig struct {
    TLSEnabled     bool   `yaml:"tls_enabled"`
    CertFile       string `yaml:"cert_file"`
    KeyFile        string `yaml:"key_file"`
    APIKeyEnabled  bool   `yaml:"api_key_enabled"`
    APIKey         string `yaml:"api_key"`
    RateLimiting   RateLimitConfig `yaml:"rate_limiting"`
}

type RateLimitConfig struct {
    Enabled     bool          `yaml:"enabled"`
    RequestsPerSecond int     `yaml:"requests_per_second"`
    BurstSize   int           `yaml:"burst_size"`
    WindowSize  time.Duration `yaml:"window_size"`
}

func Load(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    // Expand environment variables
    expandedData := os.ExpandEnv(string(data))

    var config Config
    if err := yaml.Unmarshal([]byte(expandedData), &config); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %w", err)
    }

    // Validate configuration
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }

    return &config, nil
}

func (c *Config) Validate() error {
    if c.App.Name == "" {
        return fmt.Errorf("app.name is required")
    }

    if c.FlightMonitor.Enabled && len(c.FlightMonitor.Variables) == 0 {
        return fmt.Errorf("flight_monitor.variables is required when enabled")
    }

    if c.AlertManager.Enabled && len(c.AlertManager.Rules) == 0 {
        return fmt.Errorf("alert_manager.rules is required when enabled")
    }

    return nil
}
```

### Sample Production Configuration

```yaml
# config.yaml
app:
  name: "simconnect-production"
  version: "1.0.0"
  environment: "production"
  log_level: "info"
  shutdown_timeout: "30s"

flight_monitor:
  enabled: true
  update_interval: "1s"
  simconnect_dll: "${SIMCONNECT_DLL_PATH}"
  retry_attempts: 5
  variables:
    - id: 1
      name: "PLANE ALTITUDE"
      units: "feet"
      critical: true
    - id: 2
      name: "AIRSPEED INDICATED"
      units: "knots"
      critical: true
    - id: 3
      name: "FUEL TOTAL QUANTITY"
      units: "gallons"
      critical: false

data_collector:
  enabled: true
  output_format: "json"
  output_path: "/var/log/simconnect/data"
  buffer_size: 10000
  flush_interval: "30s"
  compression: true

alert_manager:
  enabled: true
  smtp:
    server: "smtp.company.com"
    port: 587
    username: "${SMTP_USERNAME}"
    password: "${SMTP_PASSWORD}"
    from: "simconnect@company.com"
    to:
      - "ops-team@company.com"
      - "flight-ops@company.com"
  slack:
    webhook_url: "${SLACK_WEBHOOK_URL}"
    channel: "#flight-alerts"
    username: "SimConnect Bot"
  rules:
    - name: "Low Altitude Warning"
      metric: "altitude"
      condition: "less_than"
      threshold: 100
      duration: "10s"
      severity: "critical"
      actions: ["email", "slack"]
    - name: "High Error Rate"
      metric: "error_rate"
      condition: "greater_than"
      threshold: 0.1
      duration: "60s"
      severity: "warning"
      actions: ["slack"]

metrics_exporter:
  enabled: true
  type: "prometheus"
  address: "0.0.0.0"
  port: 9090
  path: "/metrics"
  interval: "15s"

monitoring:
  enabled: true
  health_check_port: 8080
  metrics_port: 9090
  profiling_port: 6060
  check_interval: "30s"

logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "/var/log/simconnect/app.log"
  max_size_mb: 100
  max_backups: 10
  max_age_days: 30
  compress: true

security:
  tls_enabled: true
  cert_file: "/etc/ssl/certs/simconnect.crt"
  key_file: "/etc/ssl/private/simconnect.key"
  api_key_enabled: true
  api_key: "${API_KEY}"
  rate_limiting:
    enabled: true
    requests_per_second: 100
    burst_size: 50
    window_size: "1m"
```

## Monitoring and Observability

### Prometheus Metrics

```go
// internal/metrics/prometheus.go
package metrics

import (
    "net/http"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type PrometheusMetrics struct {
    // Business metrics
    flightDataReceived prometheus.Counter
    flightAltitude     prometheus.Gauge
    flightSpeed        prometheus.Gauge
    flightFuel         prometheus.Gauge
    
    // System metrics
    simconnectConnections prometheus.Gauge
    messageProcessingTime prometheus.Histogram
    errorsByType         *prometheus.CounterVec
    serviceUptime        *prometheus.GaugeVec
    
    // Performance metrics
    messagesPerSecond    prometheus.Gauge
    memoryUsage         prometheus.Gauge
    cpuUsage            prometheus.Gauge
}

func NewPrometheusMetrics() *PrometheusMetrics {
    return &PrometheusMetrics{
        flightDataReceived: promauto.NewCounter(prometheus.CounterOpts{
            Name: "simconnect_flight_data_received_total",
            Help: "Total number of flight data messages received",
        }),
        
        flightAltitude: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "simconnect_flight_altitude_feet",
            Help: "Current aircraft altitude in feet",
        }),
        
        flightSpeed: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "simconnect_flight_speed_knots",
            Help: "Current aircraft speed in knots",
        }),
        
        flightFuel: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "simconnect_flight_fuel_gallons",
            Help: "Current fuel quantity in gallons",
        }),
        
        simconnectConnections: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "simconnect_active_connections",
            Help: "Number of active SimConnect connections",
        }),
        
        messageProcessingTime: promauto.NewHistogram(prometheus.HistogramOpts{
            Name:    "simconnect_message_processing_duration_seconds",
            Help:    "Time spent processing SimConnect messages",
            Buckets: prometheus.DefBuckets,
        }),
        
        errorsByType: promauto.NewCounterVec(prometheus.CounterOpts{
            Name: "simconnect_errors_total",
            Help: "Total number of errors by type",
        }, []string{"type", "severity"}),
        
        serviceUptime: promauto.NewGaugeVec(prometheus.GaugeOpts{
            Name: "simconnect_service_uptime_seconds",
            Help: "Service uptime in seconds",
        }, []string{"service"}),
        
        messagesPerSecond: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "simconnect_messages_per_second",
            Help: "Current message processing rate",
        }),
        
        memoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "simconnect_memory_usage_bytes",
            Help: "Current memory usage in bytes",
        }),
        
        cpuUsage: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "simconnect_cpu_usage_percent",
            Help: "Current CPU usage percentage",
        }),
    }
}

func (pm *PrometheusMetrics) RecordFlightData(altitude, speed, fuel float64) {
    pm.flightDataReceived.Inc()
    pm.flightAltitude.Set(altitude)
    pm.flightSpeed.Set(speed)
    pm.flightFuel.Set(fuel)
}

func (pm *PrometheusMetrics) RecordError(errorType, severity string) {
    pm.errorsByType.WithLabelValues(errorType, severity).Inc()
}

func (pm *PrometheusMetrics) RecordMessageProcessingTime(duration time.Duration) {
    pm.messageProcessingTime.Observe(duration.Seconds())
}

func (pm *PrometheusMetrics) UpdateServiceUptime(service string, uptime time.Duration) {
    pm.serviceUptime.WithLabelValues(service).Set(uptime.Seconds())
}

func (pm *PrometheusMetrics) UpdateSystemMetrics(messagesPerSec float64, memoryBytes uint64, cpuPercent float64) {
    pm.messagesPerSecond.Set(messagesPerSec)
    pm.memoryUsage.Set(float64(memoryBytes))
    pm.cpuUsage.Set(cpuPercent)
}

func (pm *PrometheusMetrics) StartServer(port int) error {
    http.Handle("/metrics", promhttp.Handler())
    return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
```

### Health Check Endpoint

```go
// internal/monitoring/health.go
package monitoring

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "your-app/internal/services"
)

type HealthChecker struct {
    services map[string]services.Service
    port     int
}

func NewHealthChecker(services map[string]services.Service, port int) *HealthChecker {
    return &HealthChecker{
        services: services,
        port:     port,
    }
}

func (hc *HealthChecker) Start() error {
    mux := http.NewServeMux()
    mux.HandleFunc("/health", hc.healthHandler)
    mux.HandleFunc("/health/live", hc.livenessHandler)
    mux.HandleFunc("/health/ready", hc.readinessHandler)
    mux.HandleFunc("/health/detail", hc.detailedHealthHandler)

    server := &http.Server{
        Addr:    fmt.Sprintf(":%d", hc.port),
        Handler: mux,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }

    return server.ListenAndServe()
}

func (hc *HealthChecker) healthHandler(w http.ResponseWriter, r *http.Request) {
    overallStatus := "healthy"
    for _, service := range hc.services {
        health := service.Health()
        if health.Status != services.StatusHealthy {
            overallStatus = "unhealthy"
            break
        }
    }

    w.Header().Set("Content-Type", "application/json")
    if overallStatus == "unhealthy" {
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    response := map[string]interface{}{
        "status":    overallStatus,
        "timestamp": time.Now().UTC(),
        "version":   "1.0.0",
    }

    json.NewEncoder(w).Encode(response)
}

func (hc *HealthChecker) livenessHandler(w http.ResponseWriter, r *http.Request) {
    // Simple liveness check - application is running
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    response := map[string]interface{}{
        "status": "alive",
        "timestamp": time.Now().UTC(),
    }
    
    json.NewEncoder(w).Encode(response)
}

func (hc *HealthChecker) readinessHandler(w http.ResponseWriter, r *http.Request) {
    // Readiness check - can serve traffic
    ready := true
    for _, service := range hc.services {
        health := service.Health()
        if health.Status != services.StatusHealthy {
            ready = false
            break
        }
    }

    w.Header().Set("Content-Type", "application/json")
    if !ready {
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    status := "ready"
    if !ready {
        status = "not_ready"
    }

    response := map[string]interface{}{
        "status":    status,
        "timestamp": time.Now().UTC(),
    }

    json.NewEncoder(w).Encode(response)
}

func (hc *HealthChecker) detailedHealthHandler(w http.ResponseWriter, r *http.Request) {
    serviceHealths := make(map[string]services.HealthStatus)
    overallStatus := "healthy"

    for name, service := range hc.services {
        health := service.Health()
        serviceHealths[name] = health
        
        if health.Status != services.StatusHealthy {
            overallStatus = "unhealthy"
        }
    }

    w.Header().Set("Content-Type", "application/json")
    if overallStatus == "unhealthy" {
        w.WriteHeader(http.StatusServiceUnavailable)
    }

    response := map[string]interface{}{
        "status":    overallStatus,
        "timestamp": time.Now().UTC(),
        "services":  serviceHealths,
    }

    json.NewEncoder(w).Encode(response)
}
```

## Deployment Strategies

### Docker Deployment

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o simconnect-app ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy binary and configuration
COPY --from=builder /app/simconnect-app .
COPY --from=builder /app/config.yaml .

# Create log directory
RUN mkdir -p /var/log/simconnect && \
    chown -R appuser:appgroup /var/log/simconnect

# Switch to non-root user
USER appuser

EXPOSE 8080 9090 6060

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./simconnect-app", "-config", "config.yaml"]
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  simconnect-app:
    build: .
    container_name: simconnect-production
    restart: unless-stopped
    ports:
      - "8080:8080"   # Health check
      - "9090:9090"   # Metrics
      - "6060:6060"   # Profiling
    volumes:
      - ./config.yaml:/root/config.yaml:ro
      - ./logs:/var/log/simconnect
      - /c/MSFS 2024 SDK:/opt/simconnect:ro
    environment:
      - SIMCONNECT_DLL_PATH=/opt/simconnect/SimConnect SDK/lib/SimConnect.dll
      - API_KEY=${API_KEY}
      - SMTP_USERNAME=${SMTP_USERNAME}
      - SMTP_PASSWORD=${SMTP_PASSWORD}
      - SLACK_WEBHOOK_URL=${SLACK_WEBHOOK_URL}
    networks:
      - simconnect-network
    depends_on:
      - prometheus
      - grafana

  prometheus:
    image: prom/prometheus:latest
    container_name: simconnect-prometheus
    restart: unless-stopped
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'
    networks:
      - simconnect-network

  grafana:
    image: grafana/grafana:latest
    container_name: simconnect-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards:ro
      - ./grafana/datasources:/etc/grafana/provisioning/datasources:ro
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    networks:
      - simconnect-network

volumes:
  prometheus_data:
  grafana_data:

networks:
  simconnect-network:
    driver: bridge
```

### Kubernetes Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: simconnect-production
  labels:
    app: simconnect
    version: v1.0.0
spec:
  replicas: 2
  selector:
    matchLabels:
      app: simconnect
  template:
    metadata:
      labels:
        app: simconnect
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: simconnect-app
        image: your-registry/simconnect:v1.0.0
        ports:
        - containerPort: 8080
          name: health
        - containerPort: 9090
          name: metrics
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        env:
        - name: SIMCONNECT_DLL_PATH
          value: "/opt/simconnect/SimConnect.dll"
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: simconnect-secrets
              key: api-key
        volumeMounts:
        - name: config
          mountPath: /root/config.yaml
          subPath: config.yaml
        - name: simconnect-dll
          mountPath: /opt/simconnect
          readOnly: true
      volumes:
      - name: config
        configMap:
          name: simconnect-config
      - name: simconnect-dll
        hostPath:
          path: /c/MSFS 2024 SDK

---
apiVersion: v1
kind: Service
metadata:
  name: simconnect-service
spec:
  selector:
    app: simconnect
  ports:
  - name: health
    port: 8080
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

## Scaling Considerations

### Horizontal Scaling Strategy

```go
// internal/scaling/coordinator.go
package scaling

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/mycrew-online/sdk/pkg/client"
)

type ScalingCoordinator struct {
    instances    map[string]*Instance
    loadBalancer *LoadBalancer
    config       *ScalingConfig
    mu           sync.RWMutex
}

type Instance struct {
    ID       string
    SDK      client.Connection
    Load     float64
    Healthy  bool
    LastSeen time.Time
}

type ScalingConfig struct {
    MinInstances    int
    MaxInstances    int
    TargetLoad      float64
    ScaleUpThreshold float64
    ScaleDownThreshold float64
    CooldownPeriod time.Duration
}

func NewScalingCoordinator(config *ScalingConfig) *ScalingCoordinator {
    return &ScalingCoordinator{
        instances:    make(map[string]*Instance),
        loadBalancer: NewLoadBalancer(),
        config:       config,
    }
}

func (sc *ScalingCoordinator) AutoScale(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            sc.evaluateScaling()
        }
    }
}

func (sc *ScalingCoordinator) evaluateScaling() {
    sc.mu.Lock()
    defer sc.mu.Unlock()

    avgLoad := sc.calculateAverageLoad()
    instanceCount := len(sc.instances)

    if avgLoad > sc.config.ScaleUpThreshold && instanceCount < sc.config.MaxInstances {
        sc.scaleUp()
    } else if avgLoad < sc.config.ScaleDownThreshold && instanceCount > sc.config.MinInstances {
        sc.scaleDown()
    }
}

func (sc *ScalingCoordinator) calculateAverageLoad() float64 {
    if len(sc.instances) == 0 {
        return 0
    }

    totalLoad := 0.0
    healthyInstances := 0

    for _, instance := range sc.instances {
        if instance.Healthy {
            totalLoad += instance.Load
            healthyInstances++
        }
    }

    if healthyInstances == 0 {
        return 0
    }

    return totalLoad / float64(healthyInstances)
}

func (sc *ScalingCoordinator) scaleUp() {
    instanceID := fmt.Sprintf("instance-%d", time.Now().Unix())
    
    sdk := client.New(instanceID)
    if err := sdk.Open(); err != nil {
        log.Printf("❌ Failed to create new instance %s: %v", instanceID, err)
        return
    }

    instance := &Instance{
        ID:       instanceID,
        SDK:      sdk,
        Load:     0.0,
        Healthy:  true,
        LastSeen: time.Now(),
    }

    sc.instances[instanceID] = instance
    sc.loadBalancer.AddInstance(instance)

    log.Printf("✅ Scaled up: added instance %s (total: %d)", instanceID, len(sc.instances))
}

func (sc *ScalingCoordinator) scaleDown() {
    // Find least loaded instance
    var targetInstance *Instance
    minLoad := float64(1000000)

    for _, instance := range sc.instances {
        if instance.Load < minLoad {
            minLoad = instance.Load
            targetInstance = instance
        }
    }

    if targetInstance != nil {
        sc.removeInstance(targetInstance.ID)
        log.Printf("📉 Scaled down: removed instance %s (total: %d)", targetInstance.ID, len(sc.instances))
    }
}

func (sc *ScalingCoordinator) removeInstance(instanceID string) {
    if instance, exists := sc.instances[instanceID]; exists {
        instance.SDK.Close()
        sc.loadBalancer.RemoveInstance(instance)
        delete(sc.instances, instanceID)
    }
}
```

## Security Best Practices

### API Key Authentication

```go
// internal/security/auth.go
package security

import (
    "crypto/subtle"
    "net/http"
    "strings"
)

type APIKeyMiddleware struct {
    apiKey string
}

func NewAPIKeyMiddleware(apiKey string) *APIKeyMiddleware {
    return &APIKeyMiddleware{apiKey: apiKey}
}

func (akm *APIKeyMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip auth for health checks
        if strings.HasPrefix(r.URL.Path, "/health") {
            next.ServeHTTP(w, r)
            return
        }

        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
            return
        }

        // Expected format: "Bearer <api-key>"
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
            return
        }

        // Constant-time comparison to prevent timing attacks
        if subtle.ConstantTimeCompare([]byte(parts[1]), []byte(akm.apiKey)) != 1 {
            http.Error(w, "Invalid API key", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

### Rate Limiting

```go
// internal/security/ratelimit.go
package security

import (
    "net/http"
    "sync"
    "time"
    
    "golang.org/x/time/rate"
)

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(requestsPerSecond int, burst int) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     rate.Limit(requestsPerSecond),
        burst:    burst,
    }
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter, exists := rl.limiters[key]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[key] = limiter
    }

    return limiter
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Use client IP as the key
        clientIP := getClientIP(r)
        limiter := rl.getLimiter(clientIP)

        if !limiter.Allow() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func getClientIP(r *http.Request) string {
    // Check X-Forwarded-For header
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        return strings.Split(forwarded, ",")[0]
    }

    // Check X-Real-IP header
    realIP := r.Header.Get("X-Real-IP")
    if realIP != "" {
        return realIP
    }

    // Fall back to RemoteAddr
    return r.RemoteAddr
}
```

This production deployment guide provides a comprehensive foundation for deploying SimConnect applications in enterprise environments with proper monitoring, scaling, security, and operational procedures.
