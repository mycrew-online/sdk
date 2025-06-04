// Connection status management
let connectionState = {
    isConnected: true,
    retryCount: 0,
    maxRetries: 5,
    retryDelay: 1000, // Start with 1 second
    retryTimeout: null,
    bannerDismissed: false
};

// Tab functionality
function showTab(tabName) {
    // Hide all tab contents
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.add('hidden');
    });
    
    // Remove active class from all tabs
    document.querySelectorAll('.tab-button').forEach(button => {
        button.classList.remove('active');
    });
    
    // Show selected tab content
    document.getElementById(tabName + 'Content').classList.remove('hidden');
    
    // Add active class to selected tab    document.getElementById(tabName + 'Tab').classList.add('active');
}

// Connection status management functions
function updateConnectionStatus(isConnected) {
    const indicator = document.getElementById('connectionIndicator');
    const status = document.getElementById('connectionStatus');
    const banner = document.getElementById('connectionBanner');
    
    if (isConnected) {
        // Connected state
        indicator.className = 'w-3 h-3 bg-green-500 rounded-full mr-3';
        status.textContent = 'Connected to MSFS';
        
        // Hide banner if connection restored
        if (banner && !banner.classList.contains('hidden')) {
            banner.classList.add('hidden');
            connectionState.bannerDismissed = false;
        }
        
        // Reset retry state
        connectionState.retryCount = 0;
        connectionState.retryDelay = 1000;
        if (connectionState.retryTimeout) {
            clearTimeout(connectionState.retryTimeout);
            connectionState.retryTimeout = null;
        }
    } else {
        // Disconnected state
        indicator.className = 'w-3 h-3 bg-red-500 rounded-full mr-3 animate-pulse';
        status.textContent = 'Disconnected from Backend';
        
        // Show banner if not dismissed
        if (banner && !connectionState.bannerDismissed) {
            banner.classList.remove('hidden');
        }
        
        // Start retry mechanism
        scheduleRetry();
    }
    
    connectionState.isConnected = isConnected;
}

function scheduleRetry() {
    if (connectionState.retryCount >= connectionState.maxRetries) {
        console.log('Max retries reached. Manual refresh required.');
        return;
    }
    
    if (connectionState.retryTimeout) {
        clearTimeout(connectionState.retryTimeout);
    }
    
    connectionState.retryTimeout = setTimeout(() => {
        connectionState.retryCount++;
        console.log(`Retry attempt ${connectionState.retryCount}/${connectionState.maxRetries}`);
        
        // Exponential backoff: 1s, 2s, 4s, 8s, 16s
        connectionState.retryDelay = Math.min(connectionState.retryDelay * 2, 30000);
        
        // Try to fetch data to test connection
        testConnection();
    }, connectionState.retryDelay);
}

async function testConnection() {
    try {
        const response = await fetch('/api/monitor', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json'
            },
            timeout: 5000
        });
        
        if (response.ok) {
            updateConnectionStatus(true);
            // Resume normal data updates
            updateMonitorData();
        } else {
            throw new Error(`HTTP ${response.status}`);
        }
    } catch (error) {
        console.log('Connection test failed:', error);
        updateConnectionStatus(false);
    }
}

function initializeBannerControls() {
    const dismissButton = document.getElementById('dismissBanner');
    if (dismissButton) {
        dismissButton.addEventListener('click', () => {
            const banner = document.getElementById('connectionBanner');
            if (banner) {
                banner.classList.add('hidden');
                connectionState.bannerDismissed = true;
            }
        });
    }
}

// Fetch monitor data from API
async function updateMonitorData() {
    try {
        const response = await fetch('/api/monitor');
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        
        // Update connection status to connected if we got data successfully
        if (!connectionState.isConnected) {
            updateConnectionStatus(true);
        }// Update Core Environmental Data (Row 1)
        document.getElementById('temperature').textContent = data.temperature.toFixed(1);
          // Convert pressure from millibars (SEA LEVEL PRESSURE) to inHg for display
        const pressureInHg = (data.seaLevelPressure || 0) / 33.8639;
        document.getElementById('pressure').textContent = pressureInHg.toFixed(2);
        
        // Update pressure in hPa (millibars) - direct display since data.seaLevelPressure is now in millibars
        document.getElementById('pressureHpa').textContent = (data.seaLevelPressure || 0).toFixed(1);
        
        document.getElementById('windSpeed').textContent = data.windSpeed.toFixed(1);
        document.getElementById('windDirection').textContent = Math.round(data.windDirection);
        
        // Update Time & Simulation Variables
        document.getElementById('zuluTime').textContent = data.zuluTime || '--:--:--';
        document.getElementById('localTime').textContent = data.localTime || '--:--:--';
        document.getElementById('simulationTime').textContent = data.simulationTime || '--:--:--';
        document.getElementById('simulationRate').textContent = data.simulationRate || '--';
        
        // Update Environmental Conditions (Row 2)
        // Format visibility: show in km with 1 decimal if > 1000m, otherwise show in meters
        const visibility = data.visibility || 0;
        if (visibility > 1000) {
            document.getElementById('visibility').textContent = (visibility / 1000).toFixed(1);
            document.getElementById('visibilityUnit').textContent = 'km';
        } else {
            document.getElementById('visibility').textContent = Math.round(visibility);
            document.getElementById('visibilityUnit').textContent = 'm';
        }
        
        // Update precipitation with type and rate
        const precipState = data.precipState || 2;
        let precipType = "None";
        let precipIcon = "☀️";
        
        if (precipState === 4) {
            precipType = "Rain";
            precipIcon = "🌧️";
        } else if (precipState === 8) {
            precipType = "Snow";
            precipIcon = "❄️";
        } else {
            precipIcon = "☀️";
        }
        
        document.getElementById('precipType').textContent = precipType;
        document.getElementById('precipIcon').textContent = precipIcon;
        document.getElementById('precipRate').textContent = (data.precipRate || 0).toFixed(1);
        
        document.getElementById('densityAltitude').textContent = Math.round(data.densityAltitude || 0);        document.getElementById('groundAltitude').textContent = Math.round(data.groundAltitude || 0);
          // Update Additional Environmental Data
        document.getElementById('magVar').textContent = (data.magVar || 0).toFixed(1);        // barometerPressure is now BAROMETER PRESSURE in inHg - display directly
        document.getElementById('seaLevelPress').textContent = (data.barometerPressure || 0).toFixed(2);
        document.getElementById('ambientDensity').textContent = (data.ambientDensity || 0).toFixed(4);
        document.getElementById('realism').textContent = (data.realism || 0).toFixed(0);
        
        // Update Position & Navigation Data (Row 3)
        const lat = data.latitude || 0;
        const lng = data.longitude || 0;
        
        document.getElementById('latitude').textContent = lat.toFixed(6);
        document.getElementById('longitude').textContent = lng.toFixed(6);
        document.getElementById('altitude').textContent = Math.round(data.altitude || 0);
        document.getElementById('groundSpeed').textContent = (data.groundSpeed || 0).toFixed(1);
        document.getElementById('heading').textContent = Math.round(data.heading || 0);
        document.getElementById('verticalSpeed').textContent = (data.verticalSpeed || 0).toFixed(1);
        
        // Update Google Maps links
        if (lat !== 0 && lng !== 0) {
            const mapsUrl = `https://www.google.com/maps/search/?api=1&query=${lat},${lng}`;
            document.getElementById('mapsLink').href = mapsUrl;            document.getElementById('mapsLink2').href = mapsUrl;
        }
          // Update Airport & Navigation Info (Row 4)
        document.getElementById('nearestAirport').textContent = data.nearestAirport || "--";
        document.getElementById('airportDistance').textContent = Math.round(data.distanceToAirport || 0);
        document.getElementById('comFrequency').textContent = (data.comFrequency || 0).toFixed(3);
        document.getElementById('navFrequency').textContent = (data.nav1Frequency || 0).toFixed(3);
        document.getElementById('gpsDistance').textContent = Math.round(data.gpsDistance || 0);
        document.getElementById('gpsEte').textContent = Math.round(data.gpsEte || 0);        // Update External Power Status
        updateExternalPowerUI(data.externalPowerOn || 0);
          // Update External Power Availability
        const availabilityElement = document.getElementById('externalPowerAvailable');
        if (availabilityElement) {
            const isAvailable = (data.externalPowerAvailable === 1);
            availabilityElement.innerHTML = isAvailable ? 
                '<span class="text-green-600 font-bold">✅ Available</span>' : 
                '<span class="text-red-600 font-bold">❌ Not Available</span>';
        }        // Update Battery Systems
        updateBatteryUI(1, data.battery1Switch || 0, data.battery1Voltage || 0, data.battery1Charge || 0);
        updateBatteryUI(2, data.battery2Switch || 0, data.battery2Voltage || 0, data.battery2Charge || 0);
          // Update APU Systems
        updateApuUI('Master', data.apuMasterSwitch || 0);
        updateApuUI('Start', data.apuStartButton || 0);
        
        // Update Aircraft Control Systems
        updateAircraftControlUI('canopy', data.canopyOpen || 0);
        updateAircraftControlUI('noSmoking', data.cabinNoSmokingSwitch || 0);
        updateAircraftControlUI('seatbelts', data.cabinSeatbeltsSwitch || 0);
        
        // Update Flight Status (Row 5)
        document.getElementById('onGround').textContent = data.onGround ? "✅ Yes" : "❌ No";
        document.getElementById('onRunway').textContent = data.onRunway ? "✅ Yes" : "❌ No";
        document.getElementById('gpsActive').textContent = data.gpsActive ? "✅ Active" : "❌ Inactive";
        document.getElementById('autopilotMaster').textContent = data.autopilotMaster ? "✅ On" : "❌ Off";
        
        // Surface Type mapping
        const surfaceTypes = {
            0: "Concrete",
            1: "Grass",
            2: "Water",
            3: "Grass_bumpy",
            4: "Asphalt",
            5: "Short_grass",
            6: "Long_grass",
            7: "Hard_turf",
            8: "Snow",
            9: "Ice",
            10: "Urban",
            11: "Forest",
            12: "Dirt",
            13: "Coral",
            14: "Gravel",
            15: "Oil_treated",
            16: "Steel_mats",
            17: "Bituminus",
            18: "Brick",
            19: "Macadam",
            20: "Planks",
            21: "Sand",
            22: "Shale",
            23: "Tarmac",
            24: "Wright_flyer_track"
        };
        document.getElementById('surfaceType').textContent = surfaceTypes[data.surfaceType] || "Unknown";
        document.getElementById('indicatedSpeed').textContent = (data.indicatedSpeed || 0).toFixed(1);
          // Update timestamp
        document.getElementById('lastUpdate').textContent = data.lastUpdate;
        
    } catch (error) {
        console.error('Failed to fetch monitor data:', error);
        
        // Update connection status to disconnected
        if (connectionState.isConnected) {
            updateConnectionStatus(false);
        }
    }
}

// Dark mode toggle functionality
function initializeThemeToggle() {
    const themeToggle = document.getElementById('themeToggle');
    
    if (themeToggle) {
        themeToggle.addEventListener('click', function() {
            const html = document.documentElement;
            const isDark = html.classList.contains('dark');
            
            if (isDark) {
                html.classList.remove('dark');
                localStorage.setItem('darkMode', 'false');
            } else {
                html.classList.add('dark');
                localStorage.setItem('darkMode', 'true');
            }
        });
    }
}

// Camera state toggle functionality
function initializeCameraToggle() {
    const cameraToggle = document.getElementById('cameraToggle');
    
    if (cameraToggle) {
        cameraToggle.addEventListener('click', async function() {            // Get current camera state
            const response = await fetch('/api/monitor');
            const data = await response.json();
            let currentState = data.cameraState || 2; // Default to cockpit view (2) if not available
            
            // Define the camera state sequence
            const cameraStates = [2, 3, 4, 5, 6]; // Cockpit, External, Drone, Fixed, Environment
            const cameraNames = {
                2: "Cockpit View", 
                3: "External View", 
                4: "Drone View", 
                5: "Fixed View", 
                6: "Environment View"
            };
            
            // Find current state in the sequence
            let currentIndex = cameraStates.indexOf(currentState);
            if (currentIndex === -1) currentIndex = 0;
            
            // Move to next state in the sequence
            let nextIndex = (currentIndex + 1) % cameraStates.length;
            let nextState = cameraStates[nextIndex];
            
            // Update tooltip with camera name
            cameraToggle.title = cameraNames[nextState] || "Camera View";
            
            // Send request to change camera state
            try {
                await fetch('/api/camera', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ state: nextState })
                });            } catch (error) {
                console.error('Failed to change camera state:', error);
            }
        });
    }
}

// External Power functionality
function updateExternalPowerUI(powerState) {
    const statusElement = document.getElementById('externalPowerStatus');
    const buttonElement = document.getElementById('externalPowerToggle');
    const buttonTextElement = document.getElementById('externalPowerButtonText');
    
    if (!statusElement || !buttonElement || !buttonTextElement) return;
    
    const isOn = powerState === 1;
    
    // Update status display
    if (isOn) {
        statusElement.innerHTML = '<span class="text-green-600 font-bold">⚡ ON</span>';
        buttonElement.className = 'px-4 py-2 rounded-lg font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 bg-red-500 hover:bg-red-600 text-white focus:ring-red-500';
        buttonTextElement.textContent = 'Turn OFF';
    } else {
        statusElement.innerHTML = '<span class="text-red-600 font-bold">❌ OFF</span>';
        buttonElement.className = 'px-4 py-2 rounded-lg font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 bg-green-500 hover:bg-green-600 text-white focus:ring-green-500';
        buttonTextElement.textContent = 'Turn ON';
    }
    
    // Enable the button
    buttonElement.disabled = false;
}

// Toggle external power
async function toggleExternalPower() {
    const buttonElement = document.getElementById('externalPowerToggle');
    
    if (!buttonElement) return;
    
    // Disable button temporarily
    buttonElement.disabled = true;
    
    try {
        const response = await fetch('/api/external-power', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        console.log('External power toggle sent successfully');
        
    } catch (error) {
        console.error('Failed to toggle external power:', error);
        // Re-enable button on error
        buttonElement.disabled = false;
    }
      // Button will be re-enabled when the next monitor update arrives
}

// Battery functionality
function updateBatteryUI(batteryNumber, switchState, voltage, charge) {
    const statusElement = document.getElementById(`battery${batteryNumber}Status`);
    const voltageElement = document.getElementById(`battery${batteryNumber}Voltage`);
    const chargeElement = document.getElementById(`battery${batteryNumber}Charge`);
    const buttonElement = document.getElementById(`battery${batteryNumber}Toggle`);
    const buttonTextElement = document.getElementById(`battery${batteryNumber}ButtonText`);
    
    if (!statusElement || !voltageElement || !buttonElement || !buttonTextElement) return;
    
    const isOn = switchState === 1;
    
    // Update status display
    if (isOn) {
        statusElement.innerHTML = '<span class="text-green-600 font-bold">⚡ ON</span>';
        buttonElement.className = 'px-4 py-2 rounded-lg font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 bg-red-500 hover:bg-red-600 text-white focus:ring-red-500';
        buttonTextElement.textContent = 'Turn OFF';
    } else {
        statusElement.innerHTML = '<span class="text-red-600 font-bold">❌ OFF</span>';
        buttonElement.className = 'px-4 py-2 rounded-lg font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 bg-green-500 hover:bg-green-600 text-white focus:ring-green-500';
        buttonTextElement.textContent = 'Turn ON';
    }
    
    // Update voltage display with health indicator
    const voltageValue = voltage || 0;
    if (voltageValue > 22) {
        voltageElement.innerHTML = `<span class="text-green-600">${voltageValue.toFixed(1)}</span>`;
    } else if (voltageValue > 20) {
        voltageElement.innerHTML = `<span class="text-yellow-600">${voltageValue.toFixed(1)}</span>`;
    } else {
        voltageElement.innerHTML = `<span class="text-red-600">${voltageValue.toFixed(1)}</span>`;
    }
      // Update charge display with health indicator
    if (chargeElement) {
        const chargeValue = charge || 0;
        if (chargeValue > 75) {
            chargeElement.innerHTML = `<span class="text-green-600">${chargeValue.toFixed(1)}</span>`;
        } else if (chargeValue > 25) {
            chargeElement.innerHTML = `<span class="text-yellow-600">${chargeValue.toFixed(1)}</span>`;
        } else {
            chargeElement.innerHTML = `<span class="text-red-600">${chargeValue.toFixed(1)}</span>`;
        }
    }
    
    // Enable the button
    buttonElement.disabled = false;
}

// Toggle battery 1
async function toggleBattery1() {
    const buttonElement = document.getElementById('battery1Toggle');
    
    if (!buttonElement) return;
    
    // Disable button temporarily
    buttonElement.disabled = true;
    
    try {
        const response = await fetch('/api/battery1', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        console.log('Battery 1 toggle sent successfully');
        
    } catch (error) {
        console.error('Failed to toggle battery 1:', error);
        // Re-enable button on error
        buttonElement.disabled = false;
    }
    
    // Button will be re-enabled when the next monitor update arrives
}

// Toggle battery 2
async function toggleBattery2() {
    const buttonElement = document.getElementById('battery2Toggle');
    
    if (!buttonElement) return;
    
    // Disable button temporarily
    buttonElement.disabled = true;
    
    try {
        const response = await fetch('/api/battery2', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
          console.log('Battery 2 toggle sent successfully');
        
    } catch (error) {
        console.error('Failed to toggle battery 2:', error);
        // Re-enable button on error
        buttonElement.disabled = false;
    }
    
    // Button will be re-enabled when the next monitor update arrives
}

// APU functionality
function updateApuUI(apuType, switchState) {
    const statusElement = document.getElementById(`apu${apuType}Status`);
    const buttonElement = document.getElementById(`apu${apuType}Toggle`);
    const buttonTextElement = document.getElementById(`apu${apuType}ButtonText`);
    
    if (!statusElement || !buttonElement || !buttonTextElement) return;
    
    const isOn = switchState === 1;
    
    // Update status display
    if (isOn) {
        statusElement.innerHTML = '<span class="text-green-600 font-bold">⚡ ON</span>';
        buttonElement.className = 'px-4 py-2 rounded-lg font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 bg-red-500 hover:bg-red-600 text-white focus:ring-red-500';
        buttonTextElement.textContent = 'Turn OFF';
    } else {
        statusElement.innerHTML = '<span class="text-red-600 font-bold">❌ OFF</span>';
        buttonElement.className = 'px-4 py-2 rounded-lg font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 bg-green-500 hover:bg-green-600 text-white focus:ring-green-500';
        buttonTextElement.textContent = 'Turn ON';
    }
    
    // Enable the button
    buttonElement.disabled = false;
}

// Toggle APU Master Switch
async function toggleApuMaster() {
    const buttonElement = document.getElementById('apuMasterToggle');
    
    if (!buttonElement) return;
    
    // Disable button temporarily
    buttonElement.disabled = true;
    
    try {
        const response = await fetch('/api/apu-master', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        console.log('APU Master Switch toggle sent successfully');
        
    } catch (error) {
        console.error('Failed to toggle APU Master Switch:', error);
        // Re-enable button on error
        buttonElement.disabled = false;
    }
    
    // Button will be re-enabled when the next monitor update arrives
}

// Toggle APU Start Button
async function toggleApuStart() {
    const buttonElement = document.getElementById('apuStartToggle');
    
    if (!buttonElement) return;
    
    // Disable button temporarily
    buttonElement.disabled = true;
    
    try {
        const response = await fetch('/api/apu-start', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        console.log('APU Start Button toggle sent successfully');
        
    } catch (error) {
        console.error('Failed to toggle APU Start Button:', error);
        // Re-enable button on error
        buttonElement.disabled = false;
    }
      // Button will be re-enabled when the next monitor update arrives
}

// Aircraft Control Systems functionality
function updateAircraftControlUI(controlType, switchState) {
    if (controlType === 'canopy') {
        // Handle canopy separately as it's still a single toggle
        const statusElement = document.getElementById('canopyStatus');
        const buttonElement = document.getElementById('canopyToggle');
        const buttonTextElement = document.getElementById('canopyButtonText');
        
        if (!statusElement || !buttonElement || !buttonTextElement) return;
        
        const isOn = switchState === 1;
        
        if (isOn) {
            statusElement.innerHTML = '<span class="text-green-600 font-bold">✅ OPEN</span>';
            buttonElement.className = 'px-4 py-2 rounded-lg font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 bg-red-500 hover:bg-red-600 text-white focus:ring-red-500';
            buttonTextElement.textContent = 'Close';
        } else {
            statusElement.innerHTML = '<span class="text-red-600 font-bold">❌ CLOSED</span>';
            buttonElement.className = 'px-4 py-2 rounded-lg font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 bg-green-500 hover:bg-green-600 text-white focus:ring-green-500';
            buttonTextElement.textContent = 'Open';
        }
        
        // Enable the button
        buttonElement.disabled = false;
    } else if (controlType === 'noSmoking' || controlType === 'seatbelts') {
        // Handle three-button interface for no smoking and seatbelts
        updateThreeStateControlUI(controlType, switchState);
    }
}

// Handle three-state controls (no smoking and seatbelts)
function updateThreeStateControlUI(controlType, switchState) {
    const statusElement = document.getElementById(`${controlType}Status`);
    const offBtn = document.getElementById(`${controlType}OffBtn`);
    const autoBtn = document.getElementById(`${controlType}AutoBtn`);
    const onBtn = document.getElementById(`${controlType}OnBtn`);
    
    if (!statusElement || !offBtn || !autoBtn || !onBtn) return;
    
    // Update status display
    let statusText, statusClass;
    switch(switchState) {
        case 2: // OFF
            statusText = '❌ OFF';
            statusClass = 'text-red-600 font-bold';
            break;
        case 1: // AUTO
            statusText = '🔄 AUTO';
            statusClass = 'text-yellow-600 font-bold';
            break;
        case 0: // ON
            statusText = '✅ ON';
            statusClass = 'text-green-600 font-bold';
            break;
        default:
            statusText = '❓ UNKNOWN';
            statusClass = 'text-gray-500 font-bold';
    }
    
    statusElement.innerHTML = `<span class="${statusClass}">${statusText}</span>`;
    
    // Enable all buttons first
    offBtn.disabled = false;
    autoBtn.disabled = false;
    onBtn.disabled = false;
    
    // Reset button styles to default
    const defaultOffClass = 'px-3 py-2 rounded-md text-sm font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 bg-red-500 hover:bg-red-600 text-white focus:ring-red-500';
    const defaultAutoClass = 'px-3 py-2 rounded-md text-sm font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 bg-yellow-500 hover:bg-yellow-600 text-white focus:ring-yellow-500';
    const defaultOnClass = 'px-3 py-2 rounded-md text-sm font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 bg-green-500 hover:bg-green-600 text-white focus:ring-green-500';
    const disabledClass = 'px-3 py-2 rounded-md text-sm font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed bg-gray-400 text-white';
    
    offBtn.className = defaultOffClass;
    autoBtn.className = defaultAutoClass;
    onBtn.className = defaultOnClass;
    
    // Disable and style the active button
    switch(switchState) {
        case 2: // OFF is active
            offBtn.disabled = true;
            offBtn.className = disabledClass;
            break;
        case 1: // AUTO is active
            autoBtn.disabled = true;
            autoBtn.className = disabledClass;
            break;
        case 0: // ON is active
            onBtn.disabled = true;
            onBtn.className = disabledClass;
            break;
    }
}

// Toggle Aircraft Exit (Canopy)
async function toggleAircraftExit() {
    const buttonElement = document.getElementById('canopyToggle');
    
    if (!buttonElement) return;
    
    // Disable button temporarily
    buttonElement.disabled = true;
    
    try {
        const response = await fetch('/api/aircraft-exit', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        console.log('Aircraft Exit toggle sent successfully');
        
    } catch (error) {
        console.error('Failed to toggle Aircraft Exit:', error);
        // Re-enable button on error
        buttonElement.disabled = false;
    }
    
    // Button will be re-enabled when the next monitor update arrives
}

// Toggle Cabin No Smoking Alert
async function toggleCabinNoSmoking() {
    const buttonElement = document.getElementById('noSmokingToggle');
    
    if (!buttonElement) return;
    
    // Disable button temporarily
    buttonElement.disabled = true;
    
    try {
        const response = await fetch('/api/cabin-no-smoking', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        console.log('Cabin No Smoking Alert toggle sent successfully');
        
    } catch (error) {
        console.error('Failed to toggle Cabin No Smoking Alert:', error);
        // Re-enable button on error
        buttonElement.disabled = false;
    }
    
    // Button will be re-enabled when the next monitor update arrives
}

// Toggle Cabin Seatbelts Alert
async function toggleCabinSeatbelts() {
    const buttonElement = document.getElementById('seatbeltsToggle');
    
    if (!buttonElement) return;
    
    // Disable button temporarily
    buttonElement.disabled = true;
    
    try {
        const response = await fetch('/api/cabin-seatbelts', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        console.log('Cabin Seatbelts Alert toggle sent successfully');
        
    } catch (error) {
        console.error('Failed to toggle Cabin Seatbelts Alert:', error);
        // Re-enable button on error
        buttonElement.disabled = false;
    }
    
    // Button will be re-enabled when the next monitor update arrives
}

// Set Cabin No Smoking Alert to specific state
async function setCabinNoSmoking(state) {
    const buttons = document.querySelectorAll('[id^="noSmokingO"][id$="Btn"], [id^="noSmokingA"][id$="Btn"]');
    
    // Disable all buttons temporarily
    buttons.forEach(btn => btn.disabled = true);
    
    try {
        const response = await fetch('/api/cabin-no-smoking-set', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ state: state })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        console.log(`Cabin No Smoking Alert set to state ${state} successfully`);
        
    } catch (error) {
        console.error(`Failed to set Cabin No Smoking Alert to state ${state}:`, error);
        // Re-enable buttons on error
        buttons.forEach(btn => btn.disabled = false);
    }
    
    // Buttons will be re-enabled when the next monitor update arrives
}

// Set Cabin Seatbelts Alert to specific state
async function setCabinSeatbelts(state) {
    const buttons = document.querySelectorAll('[id^="seatbeltsO"][id$="Btn"], [id^="seatbeltsA"][id$="Btn"]');
    
    // Disable all buttons temporarily
    buttons.forEach(btn => btn.disabled = true);
    
    try {
        const response = await fetch('/api/cabin-seatbelts-set', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ state: state })
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        console.log(`Cabin Seatbelts Alert set to state ${state} successfully`);
        
    } catch (error) {
        console.error(`Failed to set Cabin Seatbelts Alert to state ${state}:`, error);
        // Re-enable buttons on error
        buttons.forEach(btn => btn.disabled = false);
    }
    
    // Buttons will be re-enabled when the next monitor update arrives
}

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    // Show default tab
    showTab('monitor');

    // Initialize connection management
    initializeBannerControls();
    
    // Update data every second
    updateMonitorData(); // Initial monitor data load
    updateSystemEvents(); // Initial system events load
    
    // Set intervals for updates
    setInterval(updateMonitorData, 1000);
    setInterval(updateSystemEvents, 1000);
    
    initializeThemeToggle();
    initializeCameraToggle();
});
