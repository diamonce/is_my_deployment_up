document.addEventListener('DOMContentLoaded', function() {
    fetchAllStatuses();
    setInterval(fetchAllStatuses, 30000);
});

function fetchAllStatuses() {
    fetch('/status')
        .then(response => response.json())
        .then(serviceIds => {
            clearTable();
            serviceIds.forEach(serviceId => {
                fetchServiceStatus(serviceId);
            });
        })
        .catch(error => console.error('Error fetching service IDs:', error));
}

function clearTable() {
    const tableBody = document.getElementById('statusTable').getElementsByTagName('tbody')[0];
    tableBody.innerHTML = '';
}

function fetchServiceStatus(serviceId) {
    fetch(`/status/${serviceId}`)
        .then(response => response.json())
        .then(data => {
            updateTable(serviceId, data);
        })
        .catch(error => console.error(`Error fetching status for service ${serviceId}:`, error));
}

function updateTable(serviceId, data) {
    const tableBody = document.getElementById('statusTable').getElementsByTagName('tbody')[0];
    const row = tableBody.insertRow();
    const serviceIdCell = row.insertCell(0);
    const serviceNameCell = row.insertCell(1);
    const statusCell = row.insertCell(2);

    serviceIdCell.textContent = serviceId;
    serviceNameCell.textContent = data.service_name;

    const isUp = data.status === 'up';
    const indicator = document.createElement('span');
    indicator.style.display = 'inline-block';
    indicator.style.width = '12px';
    indicator.style.height = '12px';
    indicator.style.borderRadius = '50%';
    indicator.style.marginRight = '8px';
    indicator.style.backgroundColor = isUp ? '#22c55e' : '#ef4444';

    statusCell.appendChild(indicator);
    statusCell.appendChild(document.createTextNode(isUp ? 'Up' : 'Down'));
}
