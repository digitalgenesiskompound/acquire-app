package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"acquire-app/internal/models"
)

type SessionManager struct {
	sessions     map[string]*models.Session
	acquisitions map[string]*models.Acquisition
	mutex        sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions:     make(map[string]*models.Session),
		acquisitions: make(map[string]*models.Acquisition),
	}
}

// Session management methods
func (sm *SessionManager) CreateSession(deviceInfo models.DeviceInfo, capabilities models.DeviceCapabilities) (*models.Session, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sessionID := fmt.Sprintf("sess_%s", uuid.New().String()[:12])
	deviceID := fmt.Sprintf("dev_%s_%s", deviceInfo.ProductName, deviceInfo.SerialNumber)

	session := &models.Session{
		ID:              sessionID,
		DeviceID:        deviceID,
		Status:          "created",
		DeviceConnected: false,
		StartTime:       time.Now(),
		LastActivity:    time.Now(),
		Statistics: models.SessionStatistics{
			TotalDataTransferred: 0,
			AverageThroughput:    0,
			ErrorCount:           0,
			ReconnectionCount:    0,
		},
		DeviceHealth: models.DeviceHealth{
			Temperature:     0,
			BatteryLevel:    0,
			LastHealthCheck: time.Now(),
		},
		DeviceInfo:   deviceInfo,
		Capabilities: capabilities,
	}

	sm.sessions[sessionID] = session
	return session, nil
}

func (sm *SessionManager) GetSession(sessionID string) (*models.Session, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return session, nil
}

func (sm *SessionManager) UpdateSession(sessionID string, updates func(*models.Session)) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	updates(session)
	session.LastActivity = time.Now()
	return nil
}

func (sm *SessionManager) CloseSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.Status = "closed"
	session.DeviceConnected = false
	session.LastActivity = time.Now()

	// Clean up any active acquisitions for this session
	for _, acq := range sm.acquisitions {
		if acq.SessionID == sessionID && acq.Status == "active" {
			now := time.Now()
			acq.Status = "stopped"
			acq.EndTime = &now
		}
	}

	return nil
}

func (sm *SessionManager) DeleteSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.sessions[sessionID]; !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	delete(sm.sessions, sessionID)

	// Clean up acquisitions for this session
	for acqID, acq := range sm.acquisitions {
		if acq.SessionID == sessionID {
			delete(sm.acquisitions, acqID)
		}
	}

	return nil
}

// Acquisition management methods
func (sm *SessionManager) CreateAcquisition(sessionID string, params models.AcquisitionParams, metadata models.AcquisitionMetadata) (*models.Acquisition, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Verify session exists
	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	// Check if there's already an active acquisition for this session
	for _, acq := range sm.acquisitions {
		if acq.SessionID == sessionID && acq.Status == "active" {
			return nil, fmt.Errorf("session %s already has an active acquisition", sessionID)
		}
	}

	acquisitionID := fmt.Sprintf("acq_%s", uuid.New().String()[:8])

	acquisition := &models.Acquisition{
		ID:         acquisitionID,
		SessionID:  sessionID,
		Status:     "active",
		StartTime:  time.Now(),
		Parameters: params,
		Metadata:   metadata,
		Statistics: models.FinalStats{
			TotalChunks:     0,
			TotalBytes:      0,
			Duration:        0,
			AverageDataRate: 0,
		},
		DataPath: fmt.Sprintf("./data/acquisitions/%s", acquisitionID),
	}

	sm.acquisitions[acquisitionID] = acquisition
	
	// Update session with current acquisition
	session.CurrentAcquisition = acquisitionID
	session.LastActivity = time.Now()

	return acquisition, nil
}

func (sm *SessionManager) GetAcquisition(acquisitionID string) (*models.Acquisition, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	acquisition, exists := sm.acquisitions[acquisitionID]
	if !exists {
		return nil, fmt.Errorf("acquisition %s not found", acquisitionID)
	}

	return acquisition, nil
}

func (sm *SessionManager) StopAcquisition(acquisitionID string, reason string) (*models.Acquisition, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	acquisition, exists := sm.acquisitions[acquisitionID]
	if !exists {
		return nil, fmt.Errorf("acquisition %s not found", acquisitionID)
	}

	if acquisition.Status != "active" {
		return nil, fmt.Errorf("acquisition %s is not active", acquisitionID)
	}

	now := time.Now()
	acquisition.Status = "stopped"
	acquisition.EndTime = &now

	// Calculate final statistics
	duration := int(now.Sub(acquisition.StartTime).Seconds())
	acquisition.Statistics.Duration = duration

	if duration > 0 {
		acquisition.Statistics.AverageDataRate = acquisition.Statistics.TotalBytes / int64(duration)
	}

	// Update session to clear current acquisition
	if session, exists := sm.sessions[acquisition.SessionID]; exists {
		session.CurrentAcquisition = ""
		session.LastActivity = time.Now()
	}

	return acquisition, nil
}

func (sm *SessionManager) UpdateAcquisitionStats(acquisitionID string, totalChunks, totalBytes int64) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	acquisition, exists := sm.acquisitions[acquisitionID]
	if !exists {
		return fmt.Errorf("acquisition %s not found", acquisitionID)
	}

	acquisition.Statistics.TotalChunks = totalChunks
	acquisition.Statistics.TotalBytes = totalBytes

	// Update session statistics as well
	if session, exists := sm.sessions[acquisition.SessionID]; exists {
		session.Statistics.TotalDataTransferred += totalBytes
		session.LastActivity = time.Now()
	}

	return nil
}

// Heartbeat and health management
func (sm *SessionManager) ProcessHeartbeat(sessionID string, clientState models.ClientState) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	// Update session based on client state
	session.DeviceConnected = clientState.DeviceConnected
	session.LastActivity = time.Now()
	session.DeviceHealth.LastHealthCheck = time.Now()

	// Update statistics if needed
	// This could be expanded to track more detailed metrics

	return nil
}

// Utility methods
func (sm *SessionManager) GetActiveSessions() map[string]*models.Session {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	active := make(map[string]*models.Session)
	for id, session := range sm.sessions {
		if session.Status == "active" {
			active[id] = session
		}
	}
	return active
}

func (sm *SessionManager) GetActiveAcquisitions() map[string]*models.Acquisition {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	active := make(map[string]*models.Acquisition)
	for id, acquisition := range sm.acquisitions {
		if acquisition.Status == "active" {
			active[id] = acquisition
		}
	}
	return active
}

func (sm *SessionManager) CleanupExpiredSessions(timeout time.Duration) int {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	cleanupTime := time.Now().Add(-timeout)
	cleaned := 0

	for sessionID, session := range sm.sessions {
		if session.LastActivity.Before(cleanupTime) && session.Status != "closed" {
			session.Status = "expired"
			session.DeviceConnected = false
			cleaned++

			// Stop any active acquisitions
			for _, acq := range sm.acquisitions {
				if acq.SessionID == sessionID && acq.Status == "active" {
					now := time.Now()
					acq.Status = "expired"
					acq.EndTime = &now
				}
			}
		}
	}

	return cleaned
}
