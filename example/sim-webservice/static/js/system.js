// Fetch system events data from API
async function updateSystemEvents() {
    try {
        const response = await fetch('/api/system');
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const data = await response.json();
        
        // Update simulator running status
        const simRunningEl = document.getElementById('simRunningStatus');
        if (data.simRunning) {
            simRunningEl.innerHTML = '<span class="status-indicator text-green-500">🟢</span> Running';
        } else {
            simRunningEl.innerHTML = '<span class="status-indicator text-red-500">🔴</span> Stopped';
        }
        
        // Update pause status
        const simPausedEl = document.getElementById('simPausedStatus');
        if (data.simPaused) {
            simPausedEl.innerHTML = '<span class="status-indicator text-yellow-500">⏸️</span> Paused';
        } else {
            simPausedEl.innerHTML = '<span class="status-indicator text-green-500">▶️</span> Active';
        }
        
        // Update last event name
        if (data.lastEventName) {
            document.getElementById('lastEventName').textContent = data.lastEventName;
        }
        
        // Update last event time
        if (data.lastEventTime) {
            const eventTime = new Date(data.lastEventTime);
            document.getElementById('lastEventTime').textContent = 
                eventTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
        }
        
    } catch (error) {
        console.error('Failed to fetch system events data:', error);
        // Don't update connection status here since main.js handles it
    }
}
