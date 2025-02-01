// Define API base URL
const API_BASE_URL = 'https://regatta-project-8fce866eb11d.herokuapp.com/api';

async function loadResultsPage() {
    const select = document.getElementById('resultRegattaSelect');
    if (!select) return; // Exit if element doesn't exist

    try {
        console.log('Fetching regattas from:', `${API_BASE_URL}/regattas`);
        const response = await fetch(`${API_BASE_URL}/regattas`);
        const regattas = await response.json();

        select.innerHTML = '<option value="">Select Regatta</option>';
        regattas.forEach(regatta => {
            select.innerHTML += `<option value="${regatta.id}">${regatta.name}</option>`;
        });
    } catch (error) {
        console.error('Error loading regattas:', error);
    }
}

// Load teams for scoring
async function loadTeamsForScoring() {
    const regattaId = document.getElementById('resultRegattaSelect').value;
    if (!regattaId) {
        document.getElementById('raceForm').style.display = 'none';
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/regattas/${regattaId}/teams`);
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

    // Log the selected regatta and race number
    console.log('Submitting race scores for Regatta ID:', regattaId, 'and Race Number:', raceNumber);

    if (!regattaId || !raceNumber) {
        alert('Please select a regatta and enter a race number');
        return;
    }

    const scoreInputs = document.querySelectorAll('.team-score'); // Use .team-score here
    const results = [];
    scoreInputs.forEach(input => {
        const teamId = input.getAttribute('data-team-id'); // Ensure this attribute is set in your HTML
        const position = input.value;

        // Log the team ID and position being processed
        console.log('Processing Team ID:', teamId, 'with Position:', position);

        // Ensure that position is a valid number
        if (teamId && position) {
            results.push({
                ID: teamId, // Assuming teamId is used as ID for the result
                RegattaID: regattaId,
                TeamID: teamId,
                RaceNumber: parseInt(raceNumber),
                Position: parseInt(position),
                Points: parseInt(position) // Assuming points are equal to position for this example
            });
        }
    });

    // Log the collected results
    console.log('Collected results:', results);

    // Check if results array is empty
    if (results.length === 0) {
        alert('Please enter at least one score');
        return;
    }

    try {
        console.log('Sending JSON: ', JSON.stringify({ raceNumber: parseInt(raceNumber), results }));
        const response = await fetch(`${API_BASE_URL}/regattas/${regattaId}/results`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ raceNumber: parseInt(raceNumber), results })
        });

        // Log the response status
        console.log('Response status:', response.status);

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        alert('Results saved successfully');
        document.getElementById('raceNumber').value = '';
        scoreInputs.forEach(input => input.value = '');
        loadCurrentStandings(regattaId);
    } catch (error) {
        console.error('Error saving results:', error);
        alert('Error saving results: ' + error.message);
    }
}

function clearRegattaResults() {
    // Assuming you have an element to display the results
    const resultsContainer = document.getElementById('currentStandings');
    if (resultsContainer) {
        resultsContainer.innerHTML = ''; // Clear the inner HTML
    }

    console.log('Regatta results cleared'); // Debug log
}

async function loadCurrentStandings(regattaId) {
    if (!regattaId) {
        console.error('No regattaId provided to loadCurrentStandings');
        return;
    }

    try {
        const standingsUrl = `${API_BASE_URL}/regattas/${regattaId}/standings`;
        console.log('Fetching current standings from:', standingsUrl);

        const response = await fetch(standingsUrl);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const standings = await response.json();
        console.log('Received standings:', standings);

        const standingsContainer = document.getElementById('currentStandings');
        if (!standingsContainer) {
            console.log('Standings container element not found');
            return;
        }

        standingsContainer.innerHTML = ''; // Clear previous standings

        if (!Array.isArray(standings) || standings.length === 0) {
            standingsContainer.innerHTML = '<p>No standings available for this regatta</p>';
            return;
        }

        standings.forEach(team => {
            standingsContainer.innerHTML += `
                <div class="standing-item">
                    <h4 class="team-name">${team.name}</h4>
                    <p class="total-points">Total Points: <strong>${team.totalPoints}</strong></p>
                    <h5>Results:</h5>
                    <ul class="results-list">
                        ${team.results.map(result => `
                            <li class="result-item">
                                <span class="race-number">Race Number: <strong>${result.raceNumber}</strong></span>, 
                                <span class="position">${getPositionIcon(result.position)} Position: <strong>${result.position}</strong></span>
                            </li>
                        `).join('')}
                    </ul>
                </div>
            `;
        });
    } catch (error) {
        console.error('Error loading current standings:', error);
        alert('Failed to load standings: ' + error.message);
    }
}

// Function to get the position icon based on the position number
function getPositionIcon(position) {
    switch (position) {
        case 1:
            return '<i class="fas fa-medal" style="color: gold;"></i>'; // Gold medal
        case 2:
            return '<i class="fas fa-medal" style="color: silver;"></i>'; // Silver medal
        case 3:
            return '<i class="fas fa-medal" style="color: #cd7f32;"></i>'; // Bronze medal
        default:
            return ''; // No icon for positions 4 and above
    }
}

document.addEventListener('DOMContentLoaded', loadResultsPage);