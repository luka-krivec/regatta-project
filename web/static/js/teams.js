// Define API base URL
const API_BASE_URL = `http://localhost:${process.env.PORT || 8081}/api`;

async function loadTeamPage() {
    const select = document.getElementById('teamRegattaSelect');
    if (!select) {
        console.log('Select element not found');
        return;
    }

    try {
        console.log('Fetching regattas from:', `${API_BASE_URL}/regattas`);
        const response = await fetch(`${API_BASE_URL}/regattas`);
        const regattas = await response.json();
        console.log('Received regattas:', regattas);

        select.innerHTML = '<option value="">Select Regatta</option>';
        regattas.forEach(regatta => {
            select.innerHTML += `<option value="${regatta.id}">${regatta.name}</option>`;
        });

        // Add event listener for regatta selection
        select.addEventListener('change', async (e) => {
            const regattaId = e.target.value;
            console.log('Selected regattaId:', regattaId);
            
            if (regattaId) {
                try {
                    loadTeamList(regattaId);
                } catch (error) {
                    console.error('Error loading teams:', error);
                }
            } else {
                console.log('No regatta selected');
            }
        });
    } catch (error) {
        console.error('Error loading regattas:', error);
    }
}

// Add this helper function to display teams
function displayTeams(teams) {
    const teamsList = document.getElementById('teamList');
    if (!teamsList) {
        console.log('Teams list element not found');
        return;
    }

    console.log('Displaying teams:', teams);
    teamsList.innerHTML = '';
    teams.forEach(team => {
        teamsList.innerHTML += `
            <div class="team-item">
                <h3>${team.name}</h3>
            </div>
        `;
    });
}

async function loadTeamList(regattaId) {
    if (!regattaId) {
        console.log('No regattaId provided to loadTeamList');
        return;
    }

    try {
        const teamsUrl = `${API_BASE_URL}/regattas/${regattaId}/teams`;
        console.log('Fetching teams from:', teamsUrl);
        
        const response = await fetch(teamsUrl);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        const teams = await response.json();
        console.log('Received teams:', teams);

        // Check if teams is null or not an array
        if (!teams || !Array.isArray(teams)) {
            console.log('No teams data received:', teams);
            return; // Exit if teams data is empty
        }

        const teamsList = document.getElementById('teamList');
        if (!teamsList) {
            console.log('Teams list element not found');
            return;
        }

        teamsList.innerHTML = '';
        if (teams.length === 0) {
            teamsList.innerHTML = '<p>No teams found for this regatta</p>';
            return;
        }

        teamsList.innerHTML = teams.map(team => `
            <div class="list-group-item d-flex justify-content-between align-items-center" data-team-id="${team.id}">
                <span class="team-name">${team.name}</span>
                <div class="btn-group">
                    <button class="btn btn-primary btn-sm edit-btn" data-team-id="${team.id}">Edit</button>
                    <button class="btn btn-danger btn-sm delete-btn" data-team-id="${team.id}">Delete</button>
                </div>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error loading teams:', error);
        showToast('error', 'Failed to load teams: ' + error.message);
    }
}

async function addTeam() {
    const regattaSelect = document.getElementById('teamRegattaSelect');
    const teamName = document.getElementById('teamName').value.trim();
    const regattaId = regattaSelect.value;

    console.log('Adding team:', { teamName, regattaId });

    if (!teamName) {
        showToast('error', 'Please enter a team name');
        return;
    }

    if (!regattaId) {
        showToast('error', 'Please select a regatta');
        return;
    }

    const teamData = {
        name: teamName,
        regattaId: regattaId
    };

    try {
        console.log('Sending team data:', teamData);
        const response = await fetch(`${API_BASE_URL}/regattas/${regattaId}/teams`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(teamData)
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        const newTeam = await response.json();
        console.log('Team created:', newTeam);

        // Clear the input and refresh the list
        document.getElementById('teamName').value = '';
        showToast('success', 'Team added successfully');
        
        // Reload the team list
        await loadTeamList(regattaId);
    } catch (error) {
        console.error('Error adding team:', error);
        showToast('error', 'Failed to add team: ' + error.message);
    }
}

function showToast(type, message) {
    const toast = document.createElement('div');
    toast.className = 'toast-notification';
    toast.innerHTML = `
        <div class="toast-${type}">
            <div class="toast-message">${message}</div>
        </div>
    `;
    
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.remove();
    }, 3000);
}

async function deleteTeam(teamId) {
    if (!teamId) {
        console.error('No teamId provided to deleteTeam');
        return;
    }

    // Get the current regattaId from the select
    const regattaId = document.getElementById('teamRegattaSelect').value;
    if (!regattaId) {
        showToast('error', 'No regatta selected');
        return;
    }

    // Confirm deletion
    if (!confirm('Are you sure you want to delete this team?')) {
        return;
    }

    try {
        console.log('Deleting team:', { teamId, regattaId });
        const response = await fetch(`${API_BASE_URL}/regattas/${regattaId}/teams/${teamId}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
            }
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        console.log('Team deleted successfully');
        showToast('success', 'Team deleted successfully');
        
        // Refresh the team list
        await loadTeamList(regattaId);
    } catch (error) {
        console.error('Error deleting team:', error);
        showToast('error', 'Failed to delete team: ' + error.message);
    }
}

async function editTeam(teamId) {
    if (!teamId) {
        console.error('No teamId provided to editTeam');
        return;
    }

    const regattaId = document.getElementById('teamRegattaSelect').value;
    if (!regattaId) {
        showToast('error', 'No regatta selected');
        return;
    }

    // Get the current team name from the DOM
    const teamElement = document.querySelector(`[data-team-id="${teamId}"]`);
    if (!teamElement) {
        showToast('error', 'Team element not found');
        return;
    }

    const currentName = teamElement.querySelector('.team-name').textContent;
    const newName = prompt('Enter new team name:', currentName);

    if (!newName || newName.trim() === '' || newName === currentName) {
        return; // User cancelled or no change
    }

    try {
        const url = `${API_BASE_URL}/regattas/${regattaId}/teams/${teamId}`;
        console.log('Updating team at URL:', url); // Debug log
        console.log('Update payload:', { name: newName.trim() }); // Debug log

        const response = await fetch(url, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                name: newName.trim()
            })
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        console.log('Team updated successfully');
        showToast('success', 'Team updated successfully');
        
        // Refresh the team list
        await loadTeamList(regattaId);  // Pass regattaId here
    } catch (error) {
        console.error('Error updating team:', error);
        showToast('error', 'Failed to update team: ' + error.message);
    }
}

// Add event delegation for team actions
document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('teamList').addEventListener('click', async (e) => {
        const target = e.target;
        
        if (target.classList.contains('edit-btn')) {
            const teamId = target.dataset.teamId;
            await editTeam(teamId);
        }
        
        if (target.classList.contains('delete-btn')) {
            const teamId = target.dataset.teamId;
            await deleteTeam(teamId);
        }
    });

    loadTeamPage();
});