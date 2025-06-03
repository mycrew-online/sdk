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

// Fetch environmental data from API
async function updateWeather() {
    try {
        const response = await fetch('/api/weather');
        const data = await response.json();
        
        // Update Core Weather (Row 1)
        document.getElementById('temperature').textContent = data.temperature.toFixed(1);
        document.getElementById('pressure').textContent = data.pressure.toFixed(2);
        document.getElementById('windSpeed').textContent = data.windSpeed.toFixed(1);
        document.getElementById('windDirection').textContent = Math.round(data.windDirection);
        
        // Update Environmental Conditions (Row 2)
        document.getElementById('visibility').textContent = Math.round(data.visibility || 0);
        
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
        
        document.getElementById('densityAltitude').textContent = Math.round(data.densityAltitude || 0);
        document.getElementById('groundAltitude').textContent = Math.round(data.groundAltitude || 0);
        
        // Update Additional Environmental Data
        document.getElementById('magVar').textContent = (data.magVar || 0).toFixed(1);
        document.getElementById('seaLevelPress').textContent = (data.seaLevelPress || 0).toFixed(1);
        document.getElementById('ambientDensity').textContent = (data.ambientDensity || 0).toFixed(4);
        
        // Update timestamp
        document.getElementById('lastUpdate').textContent = data.lastUpdate;
        
    } catch (error) {
        console.error('Failed to fetch environmental data:', error);
    }
}

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    // Update weather data every 2 seconds
    updateWeather(); // Initial load
    setInterval(updateWeather, 2000);
});
