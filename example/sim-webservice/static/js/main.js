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
    
    // Add active class to selected tab
    document.getElementById(tabName + 'Tab').classList.add('active');
}

// Fetch monitor data from API
async function updateMonitorData() {
    try {
        const response = await fetch('/api/monitor');
        const data = await response.json();        // Update Core Environmental Data (Row 1)
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
        document.getElementById('gpsEte').textContent = Math.round(data.gpsEte || 0);

        // Update External Power Status
        updateExternalPowerUI(data.externalPowerOn || 0);
        
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
        console.error('Failed to fetch environmental data:', error);
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

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    // Show default tab
    showTab('monitor');    // Update data every second
    updateMonitorData(); // Initial monitor data load
    updateSystemEvents(); // Initial system events load
    
    // Set intervals for updates
    setInterval(updateMonitorData, 1000);
    setInterval(updateSystemEvents, 1000);
    
    initializeThemeToggle();
    initializeCameraToggle();
});
