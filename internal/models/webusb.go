package models

import "time"

// Device registration structures
type DeviceInfo struct {
	VendorID         uint16 `json:"vendorId"`
	ProductID        uint16 `json:"productId"`
	ProductName      string `json:"productName"`
	ManufacturerName string `json:"manufacturerName"`
	SerialNumber     string `json:"serialNumber"`
	USBVersion       string `json:"usbVersion"`
}

type DeviceCapabilities struct {
	SupportedFormats  []string `json:"supportedFormats"`
	MaxDataRate       int64    `json:"maxDataRate"`
	HasCalibration    bool     `json:"hasCalibration"`
	FirmwareVersion   string   `json:"firmwareVersion"`
}

type EndpointInfo struct {
	Number        int    `json:"number"`
	Direction     string `json:"direction"`
	Type          string `json:"type"`
	MaxPacketSize int    `json:"maxPacketSize"`
}

type ConnectionDetails struct {
	InterfaceNumber int            `json:"interfaceNumber"`
	Endpoints       []EndpointInfo `json:"endpoints"`
}

type DeviceRegistrationRequest struct {
	DeviceInfo        DeviceInfo        `json:"deviceInfo"`
	Capabilities      DeviceCapabilities `json:"capabilities"`
	ConnectionDetails ConnectionDetails `json:"connectionDetails"`
}

// Server configuration response structures
type ServerConfig struct {
	BufferSize          int  `json:"bufferSize"`
	Timeout             int  `json:"timeout"`
	CompressionEnabled  bool `json:"compressionEnabled"`
}

type AcquisitionSettings struct {
	SampleRate int `json:"sampleRate"`
	BitDepth   int `json:"bitDepth"`
	Channels   int `json:"channels"`
}

type DeviceRegistrationResponse struct {
	Success             bool                `json:"success"`
	SessionID           string              `json:"sessionId"`
	DeviceID            string              `json:"deviceId"`
	ServerConfig        ServerConfig        `json:"serverConfig"`
	AcquisitionSettings AcquisitionSettings `json:"acquisitionSettings"`
}

// Connection status structures
type ConnectionStatus struct {
	Connected         bool      `json:"connected"`
	InterfaceClaimed  bool      `json:"interfaceClaimed"`
	ConfigurationSet  int       `json:"configurationSet"`
	Timestamp         time.Time `json:"timestamp"`
}

type DeviceState struct {
	Ready        bool    `json:"ready"`
	Calibrated   bool    `json:"calibrated"`
	BatteryLevel int     `json:"batteryLevel"`
	Temperature  float64 `json:"temperature"`
}

type DeviceConnectionRequest struct {
	SessionID        string           `json:"sessionId"`
	DeviceID         string           `json:"deviceId"`
	ConnectionStatus ConnectionStatus `json:"connectionStatus"`
	DeviceState      DeviceState      `json:"deviceState"`
}

type DeviceConnectionResponse struct {
	Success                   bool   `json:"success"`
	Message                   string `json:"message"`
	NextAction                string `json:"nextAction"`
	CalibrationRequired       bool   `json:"calibrationRequired"`
	EstimatedCalibrationTime  int    `json:"estimatedCalibrationTime"`
}

// Disconnection structures
type DeviceDisconnectionRequest struct {
	SessionID          string    `json:"sessionId"`
	DeviceID           string    `json:"deviceId"`
	DisconnectionReason string   `json:"disconnectionReason"`
	Timestamp          time.Time `json:"timestamp"`
}

type DeviceDisconnectionResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	SessionClosed bool   `json:"sessionClosed"`
	DataPreserved bool   `json:"dataPreserved"`
}

// Data acquisition structures
type AcquisitionParams struct {
	Mode        string `json:"mode"`
	Duration    int    `json:"duration"`
	Format      string `json:"format"`
	Compression bool   `json:"compression"`
	Quality     string `json:"quality"`
}

type AcquisitionMetadata struct {
	PatientID     string `json:"patientId"`
	ProcedureType string `json:"procedureType"`
	Operator      string `json:"operator"`
}

type AcquisitionStartRequest struct {
	SessionID         string              `json:"sessionId"`
	AcquisitionParams AcquisitionParams   `json:"acquisitionParams"`
	Metadata          AcquisitionMetadata `json:"metadata"`
}

type AcquisitionStartResponse struct {
	Success          bool   `json:"success"`
	AcquisitionID    string `json:"acquisitionId"`
	StreamEndpoint   string `json:"streamEndpoint"`
	ExpectedDataSize int64  `json:"expectedDataSize"`
	ChunkSize        int    `json:"chunkSize"`
}

type AcquisitionStopRequest struct {
	AcquisitionID string    `json:"acquisitionId"`
	Reason        string    `json:"reason"`
	Timestamp     time.Time `json:"timestamp"`
}

type FinalStats struct {
	TotalChunks     int64 `json:"totalChunks"`
	TotalBytes      int64 `json:"totalBytes"`
	Duration        int   `json:"duration"`
	AverageDataRate int64 `json:"averageDataRate"`
}

type AcquisitionStopResponse struct {
	Success      bool       `json:"success"`
	Message      string     `json:"message"`
	FinalStats   FinalStats `json:"finalStats"`
	DataLocation string     `json:"dataLocation"`
}

// Session management structures
type SessionStatistics struct {
	TotalDataTransferred int64 `json:"totalDataTransferred"`
	AverageThroughput    int64 `json:"averageThroughput"`
	ErrorCount           int   `json:"errorCount"`
	ReconnectionCount    int   `json:"reconnectionCount"`
}

type DeviceHealth struct {
	Temperature     float64   `json:"temperature"`
	BatteryLevel    int       `json:"batteryLevel"`
	LastHealthCheck time.Time `json:"lastHealthCheck"`
}

type SessionStatusResponse struct {
	SessionID           string            `json:"sessionId"`
	Status              string            `json:"status"`
	DeviceConnected     bool              `json:"deviceConnected"`
	CurrentAcquisition  string            `json:"currentAcquisition"`
	StartTime           time.Time         `json:"startTime"`
	LastActivity        time.Time         `json:"lastActivity"`
	Statistics          SessionStatistics `json:"statistics"`
	DeviceHealth        DeviceHealth      `json:"deviceHealth"`
}

// Heartbeat structures
type ClientState struct {
	DeviceConnected     bool      `json:"deviceConnected"`
	AcquisitionActive   bool      `json:"acquisitionActive"`
	BufferUtilization   float64   `json:"bufferUtilization"`
	LastDataTransfer    time.Time `json:"lastDataTransfer"`
}

type HeartbeatRequest struct {
	Timestamp   time.Time   `json:"timestamp"`
	ClientState ClientState `json:"clientState"`
}

type ServerState struct {
	ProcessingQueue     int     `json:"processingQueue"`
	StorageUtilization  float64 `json:"storageUtilization"`
	SystemHealth        string  `json:"systemHealth"`
}

type Instruction struct {
	Action  string `json:"action"`
	NewSize int    `json:"newSize,omitempty"`
}

type HeartbeatResponse struct {
	Success      bool          `json:"success"`
	ServerState  ServerState   `json:"serverState"`
	Instructions []Instruction `json:"instructions"`
}

// WebSocket message structures
type WSMessage struct {
	Type          string                 `json:"type"`
	AcquisitionID string                 `json:"acquisitionId,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

type DataChunkMessage struct {
	Type          string    `json:"type"`
	AcquisitionID string    `json:"acquisitionId"`
	ChunkIndex    int64     `json:"chunkIndex"`
	TotalChunks   int64     `json:"totalChunks"`
	Data          string    `json:"data"`
	Checksum      string    `json:"checksum"`
	Timestamp     time.Time `json:"timestamp"`
}

type StatusUpdateMessage struct {
	Type          string      `json:"type"`
	AcquisitionID string      `json:"acquisitionId"`
	Status        string      `json:"status"`
	Progress      float64     `json:"progress"`
	DeviceHealth  DeviceHealth `json:"deviceHealth"`
}

// Error response structure
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// Session store entry
type Session struct {
	ID                 string            `json:"id"`
	DeviceID          string            `json:"deviceId"`
	Status            string            `json:"status"`
	DeviceConnected   bool              `json:"deviceConnected"`
	CurrentAcquisition string           `json:"currentAcquisition,omitempty"`
	StartTime         time.Time         `json:"startTime"`
	LastActivity      time.Time         `json:"lastActivity"`
	Statistics        SessionStatistics `json:"statistics"`
	DeviceHealth      DeviceHealth      `json:"deviceHealth"`
	DeviceInfo        DeviceInfo        `json:"deviceInfo"`
	Capabilities      DeviceCapabilities `json:"capabilities"`
}

// Acquisition store entry
type Acquisition struct {
	ID          string              `json:"id"`
	SessionID   string              `json:"sessionId"`
	Status      string              `json:"status"`
	StartTime   time.Time           `json:"startTime"`
	EndTime     *time.Time          `json:"endTime,omitempty"`
	Parameters  AcquisitionParams   `json:"parameters"`
	Metadata    AcquisitionMetadata `json:"metadata"`
	Statistics  FinalStats          `json:"statistics"`
	DataPath    string              `json:"dataPath"`
}
