// Application JavaScript for Intra-Oral Capture
const API_BASE_URL = window.location.origin;
let currentSession = null;
let currentAcquisition = null;
let websocketConnection = null;

document.addEventListener('DOMContentLoaded', function() {
    const acquireButton = document.getElementById('acquire-btn');
    
    if (acquireButton) {
        // Check protocol and display status
        displayProtocolStatus();
        
        // Check WebUSB compatibility on page load
        checkWebUSBCompatibility();
        acquireButton.addEventListener('click', handleAcquireClick);
    }
    
    // Setup periodic heartbeat for active sessions
    setInterval(sendHeartbeat, 30000); // Every 30 seconds
});

function displayProtocolStatus() {
    const protocolStatus = document.getElementById('protocol-status');
    const isHTTPS = window.location.protocol === 'https:';
    const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
    
    if (isHTTPS) {
        // HTTPS - WebUSB fully supported
        protocolStatus.className = 'protocol-status https';
        protocolStatus.innerHTML = `
            <span class="protocol-icon">🔒</span>
            <span>Secure Connection (HTTPS) - WebUSB Enabled</span>
        `;
    } else if (isLocalhost) {
        // HTTP localhost - WebUSB works but show info
        protocolStatus.className = 'protocol-status http';
        protocolStatus.innerHTML = `
            <span class="protocol-icon">⚠️</span>
            <span>Local Development (HTTP) - WebUSB Available</span>
        `;
    } else {
        // HTTP non-localhost - WebUSB blocked
        protocolStatus.className = 'protocol-status insecure';
        protocolStatus.innerHTML = `
            <span class="protocol-icon">🚫</span>
            <span>Insecure Connection - WebUSB Blocked. Switch to HTTPS: <a href="https://${window.location.host}:8443${window.location.pathname}" style="color: #FFFFFF; text-decoration: underline;">https://${window.location.host}:8443</a></span>
        `;
        
        // Also disable the acquire button for non-localhost HTTP
        const acquireButton = document.getElementById('acquire-btn');
        if (acquireButton && !navigator.usb) {
            acquireButton.disabled = true;
            acquireButton.textContent = 'HTTPS Required';
        }
    }
}

function checkWebUSBCompatibility() {
    if (!navigator.usb) {
        displayCompatibilityMessage();
        const acquireButton = document.getElementById('acquire-btn');
        acquireButton.disabled = true;
        acquireButton.textContent = 'WebUSB Not Supported';
    }
}

function displayCompatibilityMessage() {
    const container = document.querySelector('.container');
    const messageDiv = document.createElement('div');
    messageDiv.className = 'compatibility-message';
    messageDiv.innerHTML = `
        <div class="compatibility-icon">⚠️</div>
        <h3>WebUSB Not Supported</h3>
        <p>Your browser does not support WebUSB API. Please use a compatible browser like:</p>
        <ul>
            <li>Google Chrome (version 61+)</li>
            <li>Microsoft Edge (version 79+)</li>
            <li>Opera (version 48+)</li>
        </ul>
        <p class="note">Note: WebUSB requires HTTPS in production environments.</p>
    `;
    
    container.insertBefore(messageDiv, document.getElementById('acquire-btn'));
}

async function handleAcquireClick() {
    console.log('Acquire button clicked - initiating capture process');
    
    const acquireButton = document.getElementById('acquire-btn');
    
    // Check if device is already connected
    if (window.currentDevice) {
        // If device is connected, offer to disconnect
        if (confirm('Device is already connected. Would you like to disconnect?')) {
            await disconnectDevice();
        }
        return;
    }
    
    // Check if WebUSB is available
    if (!navigator.usb) {
        console.error('WebUSB is not supported in this browser');
        return;
    }
    
    const originalText = acquireButton.textContent;
    
    try {
        // Update button state to show device selection
        acquireButton.textContent = 'Selecting Device...';
        acquireButton.disabled = true;
        
        // Request WebUSB device - show all available devices for selection
        // The empty filters array will show all USB devices that support WebUSB
        const device = await navigator.usb.requestDevice({
            filters: [] // Empty filters array shows all available WebUSB devices
        });
        
        console.log('WebUSB device selected:', device);
        
        // Add device event listeners for disconnect handling
        navigator.usb.addEventListener('disconnect', (event) => {
            if (event.device === window.currentDevice) {
                console.log('Device disconnected unexpectedly');
                handleDeviceDisconnect();
            }
        });
        
        // First register device with server
        await registerDeviceWithServer(device);
        
        // Then initialize device
        await initializeDevice(device);
        
    } catch (error) {
        console.error('Error during device selection:', error);
        
        // Reset button state
        acquireButton.textContent = originalText;
        acquireButton.disabled = false;
        
        // Display error message based on error type
        if (error.name === 'NotFoundError') {
            displayErrorMessage('No device selected', 'Please select a device to continue.');
        } else if (error.name === 'SecurityError') {
            displayErrorMessage('Security Error', 'WebUSB access denied. Please ensure you are on a secure connection (HTTPS).');
        } else {
            displayErrorMessage('Connection Error', `Failed to connect to device: ${error.message}`);
        }
    }
}

function handleDeviceDisconnect() {
    console.log('Handling unexpected device disconnect');
    
    // Clean up device references
    window.currentDevice = null;
    window.currentInterface = undefined;
    
    // Reset button state
    resetButtonState();
    
    // Show disconnection warning
    displayErrorMessage('Device Disconnected', 'The device was unexpectedly disconnected. Please reconnect to continue.');
}

function displayDeviceSelectedMessage(device) {
    // Remove any existing messages
    const existingMessage = document.querySelector('.device-message');
    if (existingMessage) {
        existingMessage.remove();
    }
    
    const container = document.querySelector('.container');
    const messageDiv = document.createElement('div');
    messageDiv.className = 'device-message success';
    messageDiv.innerHTML = `
        <div class="device-icon">✅</div>
        <h3>Device Connected</h3>
        <p><strong>Product:</strong> ${device.productName || 'Unknown Device'}</p>
        <p><strong>Manufacturer:</strong> ${device.manufacturerName || 'Unknown'}</p>
        <p><strong>Serial:</strong> ${device.serialNumber || 'N/A'}</p>
    `;
    
    container.appendChild(messageDiv);
    
    // Auto-hide success message after 5 seconds
    setTimeout(() => {
        if (messageDiv && messageDiv.parentNode) {
            messageDiv.remove();
        }
    }, 5000);
}

function displayErrorMessage(title, message) {
    // Remove any existing messages
    const existingMessage = document.querySelector('.device-message');
    if (existingMessage) {
        existingMessage.remove();
    }
    
    const container = document.querySelector('.container');
    const messageDiv = document.createElement('div');
    messageDiv.className = 'device-message error';
    messageDiv.innerHTML = `
        <div class="device-icon">❌</div>
        <h3>${title}</h3>
        <p>${message}</p>
    `;
    
    container.appendChild(messageDiv);
    
    // Auto-hide error message after 8 seconds
    setTimeout(() => {
        if (messageDiv && messageDiv.parentNode) {
            messageDiv.remove();
        }
    }, 8000);
}

function displayConnectionStatus(message) {
    // Remove any existing connection status messages
    const existingStatus = document.querySelector('.device-message.connecting');
    if (existingStatus) {
        existingStatus.remove();
    }
    
    const container = document.querySelector('.container');
    const messageDiv = document.createElement('div');
    messageDiv.className = 'device-message connecting';
    messageDiv.innerHTML = `
        <div class="connection-spinner"></div>
        <h3>Connecting</h3>
        <p>${message}</p>
    `;
    
    container.appendChild(messageDiv);
}

function displayConnectionSuccess(device) {
    // Remove any existing status messages
    const existingMessage = document.querySelector('.device-message');
    if (existingMessage) {
        existingMessage.remove();
    }
    
    const container = document.querySelector('.container');
    const messageDiv = document.createElement('div');
    messageDiv.className = 'device-message success';
    messageDiv.innerHTML = `
        <div class="device-icon">✅</div>
        <h3>Device Ready</h3>
        <p><strong>Product:</strong> ${device.productName || 'Unknown Device'}</p>
        <p><strong>Manufacturer:</strong> ${device.manufacturerName || 'Unknown'}</p>
        <p><strong>Serial:</strong> ${device.serialNumber || 'N/A'}</p>
        <p class="connection-details">Connection established and interface claimed</p>
    `;
    
    container.appendChild(messageDiv);
    
    // Update acquire button to show device is ready
    const acquireButton = document.getElementById('acquire-btn');
    acquireButton.textContent = 'Device Connected';
    acquireButton.classList.add('connected');
    acquireButton.disabled = false;
    
    // Auto-hide success message after 6 seconds
    setTimeout(() => {
        if (messageDiv && messageDiv.parentNode) {
            messageDiv.remove();
        }
    }, 6000);
}

function resetButtonState() {
    const acquireButton = document.getElementById('acquire-btn');
    acquireButton.textContent = 'Acquire';
    acquireButton.disabled = false;
    acquireButton.classList.remove('connected');
    acquireButton.style.backgroundColor = ''; // Reset to default color
}

async function initializeDevice(device) {
    try {
        console.log('Initializing WebUSB device...');
        
        // Show connection status
        displayConnectionStatus('Opening device connection...');
        
        // Open the device connection
        await device.open();
        console.log('Device opened successfully');
        
        // Update connection status
        displayConnectionStatus('Configuring device interface...');
        
        // Select device configuration (usually configuration 1)
        if (device.configuration === null) {
            await device.selectConfiguration(1);
            console.log('Device configuration selected');
        }
        
        // Get the first interface (usually interface 0)
        const configuration = device.configuration;
        if (!configuration || configuration.interfaces.length === 0) {
            throw new Error('No interfaces found on device');
        }
        
        const interfaceNumber = configuration.interfaces[0].interfaceNumber;
        console.log(`Claiming interface ${interfaceNumber}`);
        
        // Update connection status
        displayConnectionStatus(`Claiming interface ${interfaceNumber}...`);
        
        // Claim the interface for exclusive access
        await device.claimInterface(interfaceNumber);
        console.log(`Interface ${interfaceNumber} claimed successfully`);
        
        // Store device reference for later use
        window.currentDevice = device;
        window.currentInterface = interfaceNumber;
        
        // Get endpoint information for communication
        const interface_ = configuration.interfaces[0];
        const alternate = interface_.alternates[0];
        
        console.log('Device endpoints:', alternate.endpoints.map(ep => ({
            endpointNumber: ep.endpointNumber,
            direction: ep.direction,
            type: ep.type,
            packetSize: ep.packetSize
        })));
        
        // Notify server of successful device connection
        await notifyDeviceConnection(device);
        
        // Display success with full connection details
        displayConnectionSuccess(device);
        
        console.log('Device initialized and ready for communication');
        
        // Start data acquisition automatically after successful connection
        setTimeout(async () => {
            try {
                await startDataAcquisition();
            } catch (error) {
                console.error('Auto-start acquisition failed:', error);
            }
        }, 2000); // Wait 2 seconds before auto-starting
        
    } catch (error) {
        console.error('Error initializing device:', error);
        
        // Reset button state on error
        resetButtonState();
        
        // Display specific error messages based on error type
        if (error.name === 'SecurityError') {
            displayErrorMessage('Security Error', 'Unable to access device. Please ensure proper permissions and HTTPS connection.');
        } else if (error.name === 'NetworkError') {
            displayErrorMessage('Network Error', 'Device disconnected or communication error occurred.');
        } else if (error.name === 'InvalidStateError') {
            displayErrorMessage('Device State Error', 'Device is in an invalid state. Please reconnect the device.');
        } else if (error.message.includes('interfaces')) {
            displayErrorMessage('Interface Error', 'No valid interfaces found on device. Device may not be supported.');
        } else {
            displayErrorMessage('Initialization Error', `Failed to initialize device: ${error.message}`);
        }
    }
}

async function disconnectDevice() {
    try {
        // Stop any active data acquisition first
        if (currentAcquisition) {
            await stopDataAcquisition();
        }
        
        // Notify server of disconnection
        await notifyDeviceDisconnection();
        
        if (window.currentDevice && window.currentInterface !== undefined) {
            console.log(`Releasing interface ${window.currentInterface}`);
            
            // Release the claimed interface
            await window.currentDevice.releaseInterface(window.currentInterface);
            
            // Close the device connection
            await window.currentDevice.close();
            
            // Clean up references
            window.currentDevice = null;
            window.currentInterface = undefined;
            
            console.log('Device disconnected successfully');
            
            // Reset UI state
            resetButtonState();
            
            // Show disconnection message
            const container = document.querySelector('.container');
            const messageDiv = document.createElement('div');
            messageDiv.className = 'device-message success';
            messageDiv.innerHTML = `
                <div class="device-icon">ℹ️</div>
                <h3>Device Disconnected</h3>
                <p>Device has been safely disconnected</p>
            `;
            
            container.appendChild(messageDiv);
            
            setTimeout(() => {
                if (messageDiv && messageDiv.parentNode) {
                    messageDiv.remove();
                }
            }, 3000);
        }
    } catch (error) {
        console.error('Error disconnecting device:', error);
        displayErrorMessage('Disconnection Error', `Failed to disconnect device properly: ${error.message}`);
    }
}

// ========== SERVER API INTEGRATION FUNCTIONS ==========

/**
 * Register device with server and create session
 */
async function registerDeviceWithServer(device) {
    try {
        displayConnectionStatus('Registering device with server...');
        
        // Build device info from WebUSB device
        const deviceInfo = {
            vendorId: device.vendorId,
            productId: device.productId,
            productName: device.productName || 'Unknown Device',
            manufacturerName: device.manufacturerName || 'Unknown Manufacturer',
            serialNumber: device.serialNumber || `DEV_${Date.now()}`,
            usbVersion: device.usbVersionMajor ? `${device.usbVersionMajor}.${device.usbVersionMinor}` : '2.0'
        };
        
        // Build capabilities based on device configuration
        const capabilities = {
            supportedFormats: ['raw'],
            maxDataRate: 1048576, // 1MB/s default
            hasCalibration: true,
            firmwareVersion: '1.0.0' // Default version
        };
        
        // Build connection details
        const connectionDetails = {
            interfaceNumber: 0,
            endpoints: []
        };
        
        const registrationData = {
            deviceInfo: deviceInfo,
            capabilities: capabilities,
            connectionDetails: connectionDetails
        };
        
        console.log('Registering device:', registrationData);
        
        const response = await fetch(`${API_BASE_URL}/api/webusb/devices/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(registrationData)
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.details || error.error || 'Registration failed');
        }
        
        const result = await response.json();
        console.log('Device registered successfully:', result);
        
        // Store session information
        currentSession = {
            sessionId: result.sessionId,
            deviceId: result.deviceId,
            serverConfig: result.serverConfig,
            acquisitionSettings: result.acquisitionSettings
        };
        
        displayConnectionStatus('Device registered with server');
        return result;
        
    } catch (error) {
        console.error('Error registering device with server:', error);
        displayErrorMessage('Registration Error', `Failed to register device with server: ${error.message}`);
        throw error;
    }
}

/**
 * Notify server of successful device connection
 */
async function notifyDeviceConnection(device) {
    if (!currentSession) {
        throw new Error('No active session found');
    }
    
    try {
        displayConnectionStatus('Confirming connection with server...');
        
        const connectionData = {
            sessionId: currentSession.sessionId,
            deviceId: currentSession.deviceId,
            connectionStatus: {
                connected: true,
                interfaceClaimed: true,
                configurationSet: device.configuration?.configurationValue || 1,
                timestamp: new Date().toISOString()
            },
            deviceState: {
                ready: true,
                calibrated: false,
                batteryLevel: 100, // Default value
                temperature: 25.0   // Default value
            }
        };
        
        const response = await fetch(`${API_BASE_URL}/api/webusb/devices/connect`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(connectionData)
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.details || error.error || 'Connection notification failed');
        }
        
        const result = await response.json();
        console.log('Connection confirmed with server:', result);
        
        if (result.calibrationRequired) {
            displayConnectionStatus(`Calibration required (${result.estimatedCalibrationTime}s)`);
            setTimeout(() => {
                displayConnectionStatus('Device ready for data acquisition');
            }, result.estimatedCalibrationTime * 1000);
        }
        
        return result;
        
    } catch (error) {
        console.error('Error notifying device connection:', error);
        displayErrorMessage('Connection Error', `Failed to confirm connection with server: ${error.message}`);
        throw error;
    }
}

/**
 * Start data acquisition session
 */
async function startDataAcquisition() {
    if (!currentSession) {
        throw new Error('No active session found');
    }
    
    try {
        displayConnectionStatus('Starting data acquisition...');
        
        const acquisitionData = {
            sessionId: currentSession.sessionId,
            acquisitionParams: {
                mode: 'continuous',
                duration: 0,
                format: 'raw',
                compression: false,
                quality: 'high'
            },
            metadata: {
                patientId: 'patient_demo',
                procedureType: 'demo_scan',
                operator: 'demo_user'
            }
        };
        
        const response = await fetch(`${API_BASE_URL}/api/webusb/acquisition/start`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(acquisitionData)
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.details || error.error || 'Failed to start acquisition');
        }
        
        const result = await response.json();
        console.log('Acquisition started:', result);
        
        currentAcquisition = {
            acquisitionId: result.acquisitionId,
            streamEndpoint: result.streamEndpoint,
            expectedDataSize: result.expectedDataSize,
            chunkSize: result.chunkSize
        };
        
        // Connect to WebSocket for real-time data streaming
        connectToWebSocket(result.streamEndpoint);
        
        displayConnectionStatus('Data acquisition active');
        return result;
        
    } catch (error) {
        console.error('Error starting data acquisition:', error);
        displayErrorMessage('Acquisition Error', `Failed to start data acquisition: ${error.message}`);
        throw error;
    }
}

/**
 * Stop data acquisition session
 */
async function stopDataAcquisition() {
    if (!currentAcquisition) {
        throw new Error('No active acquisition found');
    }
    
    try {
        displayConnectionStatus('Stopping data acquisition...');
        
        const stopData = {
            acquisitionId: currentAcquisition.acquisitionId,
            reason: 'user_requested'
        };
        
        const response = await fetch(`${API_BASE_URL}/api/webusb/acquisition/stop`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(stopData)
        });
        
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.details || error.error || 'Failed to stop acquisition');
        }
        
        const result = await response.json();
        console.log('Acquisition stopped:', result);
        
        // Disconnect WebSocket
        if (websocketConnection) {
            websocketConnection.close();
            websocketConnection = null;
        }
        
        currentAcquisition = null;
        
        displayConnectionStatus('Data acquisition stopped');
        return result;
        
    } catch (error) {
        console.error('Error stopping data acquisition:', error);
        displayErrorMessage('Stop Error', `Failed to stop data acquisition: ${error.message}`);
        throw error;
    }
}

/**
 * Connect to WebSocket for real-time data streaming
 */
function connectToWebSocket(streamEndpoint) {
    try {
        console.log('Connecting to WebSocket:', streamEndpoint);
        
        // Convert HTTP URL to WebSocket URL if needed
        const wsUrl = streamEndpoint.replace('http://', 'ws://').replace('https://', 'wss://');
        websocketConnection = new WebSocket(wsUrl);
        
        websocketConnection.onopen = function(event) {
            console.log('WebSocket connection opened');
            displayConnectionStatus('Real-time streaming connected');
        };
        
        websocketConnection.onmessage = function(event) {
            try {
                const message = JSON.parse(event.data);
                handleWebSocketMessage(message);
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        };
        
        websocketConnection.onerror = function(error) {
            console.error('WebSocket error:', error);
            displayErrorMessage('Streaming Error', 'WebSocket connection error occurred');
        };
        
        websocketConnection.onclose = function(event) {
            console.log('WebSocket connection closed:', event.code, event.reason);
            if (currentAcquisition) {
                displayErrorMessage('Streaming Disconnected', 'Real-time streaming connection was lost');
            }
        };
        
    } catch (error) {
        console.error('Error connecting to WebSocket:', error);
        displayErrorMessage('WebSocket Error', `Failed to connect to streaming service: ${error.message}`);
    }
}

/**
 * Handle incoming WebSocket messages
 */
function handleWebSocketMessage(message) {
    console.log('WebSocket message received:', message);
    
    switch (message.type) {
        case 'data_chunk':
            handleDataChunk(message);
            break;
        case 'status_update':
            handleStatusUpdate(message);
            break;
        case 'error':
            handleStreamingError(message);
            break;
        case 'processing_feedback':
            handleProcessingFeedback(message);
            break;
        default:
            console.log('Unknown WebSocket message type:', message.type);
    }
}

/**
 * Handle data chunk messages
 */
function handleDataChunk(message) {
    // Send acknowledgment back to server
    if (websocketConnection && websocketConnection.readyState === WebSocket.OPEN) {
        const ack = {
            type: 'ack',
            chunkId: message.chunkId,
            sequenceNumber: message.sequenceNumber
        };
        websocketConnection.send(JSON.stringify(ack));
    }
    
    // Update UI with progress if available
    if (message.metadata && message.metadata.progress) {
        updateAcquisitionProgress(message.metadata.progress);
    }
}

/**
 * Handle status update messages
 */
function handleStatusUpdate(message) {
    console.log('Status update:', message.status);
    displayConnectionStatus(`Status: ${message.status}`);
}

/**
 * Handle streaming error messages
 */
function handleStreamingError(message) {
    console.error('Streaming error:', message.error);
    displayErrorMessage('Streaming Error', message.error);
}

/**
 * Handle processing feedback messages
 */
function handleProcessingFeedback(message) {
    console.log('Processing feedback:', message);
    
    if (message.qualityMetrics) {
        updateQualityMetrics(message.qualityMetrics);
    }
    
    if (message.recommendations && message.recommendations.length > 0) {
        showRecommendations(message.recommendations);
    }
}

/**
 * Update acquisition progress in UI
 */
function updateAcquisitionProgress(progress) {
    const progressPercent = Math.round(progress * 100);
    displayConnectionStatus(`Acquiring data: ${progressPercent}%`);
}

/**
 * Update quality metrics in UI
 */
function updateQualityMetrics(metrics) {
    // For now, just log metrics - could be displayed in UI
    console.log('Quality metrics:', metrics);
}

/**
 * Show processing recommendations
 */
function showRecommendations(recommendations) {
    console.log('Processing recommendations:', recommendations);
    // Could display recommendations in UI
}

/**
 * Send periodic heartbeat to server
 */
async function sendHeartbeat() {
    if (!currentSession || !currentSession.sessionId) {
        return; // No active session
    }
    
    try {
        const heartbeatData = {
            clientState: {
                connected: window.currentDevice !== null,
                bufferUtilization: 0.5, // Default value
                lastActivity: new Date().toISOString()
            }
        };
        
        const response = await fetch(`${API_BASE_URL}/api/webusb/sessions/${currentSession.sessionId}/heartbeat`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(heartbeatData)
        });
        
        if (response.ok) {
            const result = await response.json();
            // Handle server instructions if any
            if (result.instructions && result.instructions.length > 0) {
                handleServerInstructions(result.instructions);
            }
        }
        
    } catch (error) {
        console.warn('Heartbeat failed:', error.message);
        // Don't show error to user for heartbeat failures
    }
}

/**
 * Handle server instructions from heartbeat response
 */
function handleServerInstructions(instructions) {
    instructions.forEach(instruction => {
        console.log('Server instruction:', instruction);
        
        switch (instruction.action) {
            case 'adjust_buffer_size':
                console.log('Server recommends buffer size adjustment to:', instruction.newSize);
                break;
            default:
                console.log('Unknown server instruction:', instruction.action);
        }
    });
}

/**
 * Notify server of device disconnection
 */
async function notifyDeviceDisconnection() {
    if (!currentSession) {
        return; // No active session
    }
    
    try {
        const disconnectionData = {
            sessionId: currentSession.sessionId,
            deviceId: currentSession.deviceId,
            disconnectionReason: 'user_initiated',
            timestamp: new Date().toISOString()
        };
        
        const response = await fetch(`${API_BASE_URL}/api/webusb/devices/disconnect`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(disconnectionData)
        });
        
        if (response.ok) {
            const result = await response.json();
            console.log('Disconnection confirmed with server:', result);
        }
        
    } catch (error) {
        console.warn('Failed to notify server of disconnection:', error.message);
    } finally {
        // Clean up session data
        currentSession = null;
        currentAcquisition = null;
        
        if (websocketConnection) {
            websocketConnection.close();
            websocketConnection = null;
        }
    }
}
