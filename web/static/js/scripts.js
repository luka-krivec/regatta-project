async function showResponse(response) {
    const responseDiv = document.getElementById('response');
    try {
        const data = await response.json();
        responseDiv.textContent = JSON.stringify(data, null, 2);
    } catch {
        responseDiv.textContent = 'Status: ' + response.status +
            (response.status === 204 ? ' (No Content)' : '');
    }
}

async function createRegatta() {
    const response = await fetch('/regattas', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            id: document.getElementById('createId').value,
            name: document.getElementById('createName').value,
            date: document.getElementById('createDate').value,
            location: document.getElementById('createLocation').value
        })
    });
    showResponse(response);
}

async function getAllRegattas() {
    const response = await fetch('/regattas');
    showResponse(response);
}

async function getRegatta() {
    const id = document.getElementById('getId').value;
    const response = await fetch('/regattas/' + id);
    showResponse(response);
}

async function updateRegatta() {
    const id = document.getElementById('updateId').value;
    const response = await fetch('/regattas/' + id, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            name: document.getElementById('updateName').value,
            date: document.getElementById('updateDate').value,
            location: document.getElementById('updateLocation').value
        })
    });
    showResponse(response);
}

async function deleteRegatta() {
    const id = document.getElementById('deleteId').value;
    const response = await fetch('/regattas/' + id, {
        method: 'DELETE'
    });
    showResponse(response);
}

// Load regattas into all select elements
async function loadRegattaSelects() {
    try {
        console.log('Loading regattas...');
        const response = await fetch('/regattas');
        console.log('Response status:', response.status);

        const regattas = await response.json();
        console.log('Loaded regattas:', regattas);

        const selects = ['teamRegattaSelect', 'resultRegattaSelect', 'standingsRegattaSelect'];
        selects.forEach(selectId => {
            const select = document.getElementById(selectId);
            if (!select) {
                console.error(`Select element ${selectId} not found`);
                return;
            }

            select.innerHTML = '<option value="">Select Regatta</option>';
            regattas.forEach(regatta => {
                console.log('Adding regatta:', regatta);
                select.innerHTML += `<option value="${regatta.id}">${regatta.name}</option>`;
            });
        });
    } catch (error) {
        console.error('Error loading regattas:', error);
    }
}

// Make sure the function is called when page loads
document.addEventListener('DOMContentLoaded', () => {
    console.log('Page loaded, calling loadRegattaSelects');
    loadRegattaSelects();
});

// Add team to regatta
async function addTeam() {
    const regattaId = document.getElementById('teamRegattaSelect').value;
    if (!regattaId) {
        alert('Please select a regatta');
        return;
    }

    const teamName = document.getElementById('teamName').value;
    if (!teamName) {
        alert('Please enter a team name');
        return;
    }

    const response = await fetch(`/regattas/${regattaId}/teams`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: teamName })
    });

    if (response.ok) {
        document.getElementById('teamName').value = '';
        loadTeams();
    }
    showResponse(response);
}

// Load teams for a regatta
async function loadTeams() {
    const regattaId = document.getElementById('teamRegattaSelect').value;
    if (!regattaId) return;

    const response = await fetch(`/regattas/${regattaId}/standings`);
    const standings = await response.json();

    const teamList = document.getElementById('teamList');
    teamList.innerHTML = '';
    standings.forEach(standing => {
        teamList.innerHTML += `
                <li class="list-group-item">
                    ${standing.name} (Total Points: ${standing.totalPoints})
                </li>`;
    });
}

// Load teams for race results
async function loadTeamsForResults() {
    const regattaId = document.getElementById('resultRegattaSelect').value;
    if (!regattaId) return;

    const response = await fetch(`/regattas/${regattaId}/standings`);
    const standings = await response.json();

    const positionInputs = document.getElementById('positionInputs');
    positionInputs.innerHTML = '';
    standings.forEach(standing => {
        positionInputs.innerHTML += `
                <div class="input-group mb-2">
                    <span class="input-group-text">${standing.name}</span>
                    <input type="number" class="form-control position-input" 
                           data-team-id="${standing.teamId}" placeholder="Position">
                </div>`;
    });
}

// Submit race results
async function submitRaceResults() {
    const regattaId = document.getElementById('resultRegattaSelect').value;
    const raceNumber = document.getElementById('raceNumber').value;

    if (!regattaId || !raceNumber) {
        alert('Please select a regatta and enter a race number');
        return;
    }

    const positionInputs = document.querySelectorAll('.position-input');
    const results = [];
    positionInputs.forEach(input => {
        const teamId = input.getAttribute('data-team-id');
        const position = input.value;
        results.push({ teamId, position });
    });

    const response = await fetch(`/regattas/${regattaId}/results`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ raceNumber, results })
    });

    showResponse(response);
}

// Load teams for scoring
async function loadTeamsForScoring() {
    const regattaId = document.getElementById('resultRegattaSelect').value;
    if (!regattaId) {
        document.getElementById('raceForm').style.display = 'none';
        return;
    }

    try {
        const response = await fetch(`/regattas/${regattaId}/teams`);
        const teams = await response.json();

        const teamScores = document.getElementById('teamScores');
        teamScores.innerHTML = teams.map(team => `
                <div class="mb-2">
                    <label class="form-label">${team.name}</label>
                    <input type="number" 
                           class="form-control team-score" 
                           data-team-id="${team.id}"
                           min="1"
                           placeholder="Position">
                </div>
            `).join('');

        document.getElementById('raceForm').style.display = 'block';
        loadCurrentStandings(regattaId);
    } catch (error) {
        console.error('Error loading teams:', error);
        alert('Error loading teams');
    }
}

// Submit race scores
async function submitRaceScores() {
    const regattaId = document.getElementById('resultRegattaSelect').value;
    const raceNumber = document.getElementById('raceNumber').value;

    if (!regattaId || !raceNumber) {
        alert('Please select a regatta and enter race number');
        return;
    }

    const scores = [];
    document.querySelectorAll('.team-score').forEach(input => {
        if (input.value) {
            scores.push({
                teamId: input.dataset.teamId,
                position: parseInt(input.value),
                points: parseInt(input.value) // points = position
            });
        }
    });

    if (scores.length === 0) {
        alert('Please enter at least one score');
        return;
    }

    try {
        const response = await fetch(`/regattas/${regattaId}/results`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                raceNumber: parseInt(raceNumber),
                scores: scores
            })
        });

        if (response.ok) {
            alert('Results saved successfully');
            document.getElementById('raceNumber').value = '';
            document.querySelectorAll('.team-score').forEach(input => input.value = '');
            loadCurrentStandings(regattaId);
        } else {
            alert('Error saving results');
        }
    } catch (error) {
        console.error('Error saving results:', error);
        alert('Error saving results');
    }
}

// Load current standings
async function loadCurrentStandings(regattaId) {
    try {
        const response = await fetch(`/regattas/${regattaId}/standings`);
        const standings = await response.json();

        const standingsDisplay = document.getElementById('currentStandings');
        standingsDisplay.innerHTML = standings.map(standing => `
                <li class="list-group-item">
                    ${standing.name} (Total Points: ${standing.totalPoints})
                </li>
            `).join('');
    } catch (error) {
        console.error('Error loading standings:', error);
        alert('Error loading standings');
    }
}