// Define API base URL
const API_BASE_URL = `http://localhost:${process.env.PORT || 8081}/api`;

async function loadStandings() {
    const select = document.getElementById('standingsRegattaSelect');
    if (!select) return; // Exit if element doesn't exist

    try {
        console.log('Fetching regattas from:', `${API_BASE_URL}/regattas`);
        const response = await fetch(`${API_BASE_URL}/regattas`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const regattas = await response.json();

        select.innerHTML = '<option value="">Select Regatta</option>'; // Clear previous options
        regattas.forEach(regatta => {
            select.innerHTML += `<option value="${regatta.id}">${regatta.name}</option>`;
        });

        // Load standings for the first regatta by default if available
        if (regattas.length > 0) {
            loadCurrentStandings(regattas[0].id);
        }

        // Set up the event listener for the dropdown
        select.addEventListener('change', function() {
            const selectedRegattaId = this.value;
            if (selectedRegattaId) {
                loadCurrentStandings(selectedRegattaId);
            } else {
                // Optionally clear the standings if no regatta is selected
                document.getElementById('currentStandings').innerHTML = '<p>Please select a regatta.</p>';
            }
        });
    } catch (error) {
        console.error('Error loading regattas:', error);
    }
}

// Function to load standings for a specific regatta
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

        const standingsContainer = document.getElementById('standingsTable');
        if (!standingsContainer) {
            console.log('Standings container element not found');
            return;
        }

        standingsContainer.innerHTML = ''; // Clear previous standings

        if (!Array.isArray(standings) || standings.length === 0) {
            standingsContainer.innerHTML = '<p>No standings available for this regatta</p>';
            return;
        }

        // Sort teams by total points (ascending)
        standings.sort((a, b) => a.totalPoints - b.totalPoints);

        // Create a table for standings
        let tableHTML = `
            <table class="standings-table">
                <thead>
                    <tr>
                        <th>Team</th>
                        <th>Total Points</th>
        `;

        // Assuming the first team has all the races, we can get the races from the first team's results
        const races = standings[0].results.map(result => result.raceNumber);
        races.forEach(race => {
            tableHTML += `<th>Race ${race}</th>`;
        });

        tableHTML += `
                    </tr>
                </thead>
                <tbody>
        `;

        standings.forEach(team => {
            tableHTML += `
                <tr>
                    <td>${team.name}</td>
                    <td>${team.totalPoints}</td>
            `;

            // Create a cell for each race
            races.forEach(race => {
                const result = team.results.find(r => r.raceNumber === race);
                const positionIcon = result ? getPositionIcon(result.position) : 'N/A'; // Display icon or N/A if no result
                tableHTML += `<td>${positionIcon} ${result ? result.position : 'N/A'}</td>`;
            });

            tableHTML += `
                </tr>
            `;
        });

        tableHTML += `
                </tbody>
            </table>
        `;

        standingsContainer.innerHTML = tableHTML; // Insert the table into the container

        // Set the dropdown to the currently selected regatta
        const select = document.getElementById('standingsRegattaSelect');
        if (select) {
            select.value = regattaId; // Set the dropdown to the current regatta
        }
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

// Initial call to load standings when the page loads
document.addEventListener('DOMContentLoaded', loadStandings); 