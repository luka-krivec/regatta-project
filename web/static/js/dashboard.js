async function loadDashboardData() {
    // Check environment variable PORT
    const port = process.env.PORT || 8080; // Default to 8080 if PORT is not set
    console.log(`Server is running on port: ${port}`);
    
    try {
        // Load stats
        const statsResponse = await fetch(`http://localhost:${port}/api/dashboard/stats`);
        const stats = await statsResponse.json();
        
        // Update stat cards
        document.getElementById('activeRegattasCount').textContent = stats.activeRegattas || 0;
        document.getElementById('totalTeamsCount').textContent = stats.totalTeams || 0;
        document.getElementById('racesCompletedCount').textContent = stats.racesCompleted || 0;
        document.getElementById('upcomingRacesCount').textContent = stats.upcomingRaces || 0;

        // Load recent regattas
        const regattasResponse = await fetch(`http://localhost:${port}/api/regattas?limit=5`);
        const regattas = await regattasResponse.json();
        
        // Update recent regattas table
        const tbody = document.getElementById('recentRegattasTable').querySelector('tbody');
        tbody.innerHTML = regattas.map(regatta => `
            <tr>
                <td>${regatta.name}</td>
                <td>${formatDate(regatta.date)}</td>
                <td>${regatta.teamCount || 0}</td>
                <td>
                    <span class="badge bg-${getStatusBadgeClass(regatta.status)}">
                        ${regatta.status}
                    </span>
                </td>
            </tr>
        `).join('');
    } catch (error) {
        console.error('Error loading dashboard data:', error);
    }
}

function formatDate(dateString) {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString();
}

function getStatusBadgeClass(status) {
    switch (status?.toLowerCase()) {
        case 'active': return 'success';
        case 'upcoming': return 'primary';
        case 'completed': return 'secondary';
        default: return 'info';
    }
}

// Load dashboard data when page loads
document.addEventListener('DOMContentLoaded', loadDashboardData);