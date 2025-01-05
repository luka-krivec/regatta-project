// Define API base URL
const API_BASE_URL = 'http://localhost:8081/api';

// Initialize page
document.addEventListener('DOMContentLoaded', () => {
    getAllRegattas();
    // Set default date to today
    const startDateInput = document.getElementById('regattaStartDate');
    if (startDateInput) {
        startDateInput.valueAsDate = new Date();
    }
    const endDateInput = document.getElementById('regattaEndDate');
    if (endDateInput) {
        endDateInput.valueAsDate = new Date();
    }
});

// Load all regattas
async function getAllRegattas() {
    try {
        console.log('Fetching from:', `${API_BASE_URL}/regattas`); // Debug log
        
        const response = await fetch(`${API_BASE_URL}/regattas`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const contentType = response.headers.get("content-type");
        if (!contentType || !contentType.includes("application/json")) {
            throw new Error("Response is not JSON");
        }
        
        const regattas = await response.json();
        const regattaList = document.getElementById('regattaList');
        
        if (!Array.isArray(regattas) || regattas.length === 0) {
            regattaList.innerHTML = '<div class="alert alert-info">No regattas found</div>';
            return;
        }

        regattaList.innerHTML = `
            <div class="table-responsive">
                <table class="table table-hover">
                    <thead class="table-dark">
                        <tr>
                            <th>Name</th>
                            <th>Start Date</th>
                            <th>End Date</th>
                            <th>Location</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${regattas.map(regatta => `
                            <tr>
                                <td>${regatta.name}</td>
                                <td>${formatDate(regatta.startDate)}</td>
                                <td>${formatDate(regatta.endDate)}</td>
                                <td>${regatta.location}</td>
                                <td>
                                    <button class="btn btn-sm btn-primary me-1" onclick="editRegatta('${regatta.id}')">
                                        <i class="bi bi-pencil"></i>
                                    </button>
                                    <button class="btn btn-sm btn-danger" onclick="deleteRegatta('${regatta.id}')">
                                        <i class="bi bi-trash"></i>
                                    </button>
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
    } catch (error) {
        console.error('Error loading regattas:', error);
        document.getElementById('regattaList').innerHTML = 
            '<div class="alert alert-danger">Error loading regattas. Please try again later.</div>';
        showToast('error', 'Error loading regattas');
    }
}

// Create new regatta
async function createRegatta() {
    const name = document.getElementById('regattaName').value.trim();
    const startDate = document.getElementById('regattaStartDate').value;
    const endDate = document.getElementById('regattaEndDate').value;
    const location = document.getElementById('regattaLocation').value.trim();

    if (!name || !startDate || !endDate || !location) {
        showToast('error', 'Please fill in all fields');
        return;
    }

    const regattaData = {
        regattaId: crypto.randomUUID(),
        name: name,
        startDate: startDate,
        endDate: endDate,
        location: location,
        status: "SCHEDULED"
    };

    console.log('Sending data:', regattaData);

    try {
        const response = await fetch(`${API_BASE_URL}/regattas`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(regattaData)
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        showToast('success', 'Regatta created successfully');
        clearForm();
        getAllRegattas();
    } catch (error) {
        console.error('Error creating regatta:', error);
        showToast('error', 'Error creating regatta: ' + error.message);
    }
}

// Clear form fields
function clearForm() {
    document.getElementById('regattaName').value = '';
    document.getElementById('regattaLocation').value = '';
    const startDateInput = document.getElementById('regattaStartDate');
    if (startDateInput) {
        startDateInput.valueAsDate = new Date();
    }
    const endDateInput = document.getElementById('regattaEndDate');
    if (endDateInput) {
        endDateInput.valueAsDate = new Date();
    }
}

// Edit regatta
async function editRegatta(id) {
    try {
        const response = await fetch(`${API_BASE_URL}/regattas/${id}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const regatta = await response.json();
        
        document.getElementById('editRegattaId').value = regatta.id;
        document.getElementById('editRegattaName').value = regatta.name;
        document.getElementById('editRegattaStartDate').value = regatta.date.split('T')[0];
        document.getElementById('editRegattaEndDate').value = regatta.date.split('T')[0];
        document.getElementById('editRegattaLocation').value = regatta.location;
        
        new bootstrap.Modal(document.getElementById('editRegattaModal')).show();
    } catch (error) {
        console.error('Error loading regatta:', error);
        showToast('error', 'Error loading regatta details');
    }
}

// Update regatta
async function updateRegatta() {
    const id = document.getElementById('editRegattaId').value;
    const name = document.getElementById('editRegattaName').value.trim();
    const startDate = document.getElementById('editRegattaStartDate').value;
    const endDate = document.getElementById('editRegattaEndDate').value;
    const location = document.getElementById('editRegattaLocation').value.trim();

    if (!name || !startDate || !endDate || !location) {
        showToast('error', 'Please fill in all fields');
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/regattas/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ name, startDate, endDate, location })
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        showToast('success', 'Regatta updated successfully');
        bootstrap.Modal.getInstance(document.getElementById('editRegattaModal')).hide();
        getAllRegattas();
    } catch (error) {
        console.error('Error updating regatta:', error);
        showToast('error', 'Error updating regatta: ' + error.message);
    }
}

// Delete regatta
async function deleteRegatta(id) {
    if (!await showConfirmDialog('Delete Regatta', 'Are you sure you want to delete this regatta?', 'danger')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/regattas/${id}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(error);
        }

        showToast('success', 'Regatta deleted successfully');
        getAllRegattas();
    } catch (error) {
        console.error('Error deleting regatta:', error);
        showToast('error', 'Error deleting regatta: ' + error.message);
    }
}

// Utility functions
function formatDate(dateString) {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
    });
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

function showConfirmDialog(title, message, type = 'primary') {
    return new Promise((resolve) => {
        const dialog = document.createElement('div');
        dialog.className = 'modal fade';
        dialog.innerHTML = `
            <div class="modal-dialog">
                <div class="modal-content">
                    <div class="modal-header bg-${type} text-white">
                        <h5 class="modal-title">${title}</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                    </div>
                    <div class="modal-body">
                        <p>${message}</p>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                        <button type="button" class="btn btn-${type}" id="confirmBtn">Confirm</button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(dialog);
        const modal = new bootstrap.Modal(dialog);
        
        dialog.querySelector('#confirmBtn').onclick = () => {
            modal.hide();
            resolve(true);
        };
        
        dialog.addEventListener('hidden.bs.modal', () => {
            document.body.removeChild(dialog);
            resolve(false);
        });
        
        modal.show();
    });
} 