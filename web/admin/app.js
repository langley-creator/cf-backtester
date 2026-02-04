// API Configuration
const API_BASE = '/api/v1';

// Global state
const state = {
    strategies: [],
    instruments: [],
    backtests: [],
    currentStrategy: null,
    currentBacktest: null
};

// API Helper
async function apiCall(endpoint, options = {}) {
    try {
        const response = await fetch(`${API_BASE}${endpoint}`, {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        });
        
        if (!response.ok) {
            throw new Error(`API Error: ${response.statusText}`);
        }
        
        return await response.json();
    } catch (error) {
        console.error('API call failed:', error);
        showNotification(error.message, 'error');
        throw error;
    }
}

// Notification System
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 16px 24px;
        background: ${type === 'error' ? '#ff3b30' : type === 'success' ? '#34c759' : '#007aff'};
        color: white;
        border-radius: 12px;
        box-shadow: 0 4px 20px rgba(0,0,0,0.15);
        z-index: 10000;
        animation: slideIn 0.3s ease;
    `;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

// Tab Navigation
function initTabs() {
    const tabs = document.querySelectorAll('.tab');
    const sections = document.querySelectorAll('.section');
    
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const targetSection = tab.dataset.section;
            
            tabs.forEach(t => t.classList.remove('active'));
            sections.forEach(s => s.classList.remove('active'));
            
            tab.classList.add('active');
            document.getElementById(targetSection).classList.add('active');
            
            // Load section data
            loadSectionData(targetSection);
        });
    });
}

// Load Section Data
async function loadSectionData(section) {
    switch(section) {
        case 'dashboard':
            await loadDashboard();
            break;
        case 'strategies':
            await loadStrategies();
            break;
        case 'backtests':
            await loadBacktests();
            break;
        case 'instruments':
            await loadInstruments();
            break;
    }
}

// Dashboard
async function loadDashboard() {
    try {
        const [strategies, backtests] = await Promise.all([
            apiCall('/strategies'),
            apiCall('/backtests')
        ]);
        
        const dashboardHTML = `
            <div class="stats-grid">
                <div class="stat-card">
                    <div class="stat-value">${strategies.length}</div>
                    <div class="stat-label">Active Strategies</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value">${backtests.length}</div>
                    <div class="stat-label">Total Backtests</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value">${backtests.filter(b => b.status === 'completed').length}</div>
                    <div class="stat-label">Completed Tests</div>
                </div>
                <div class="stat-card">
                    <div class="stat-value">${backtests.filter(b => b.status === 'running').length}</div>
                    <div class="stat-label">Running Tests</div>
                </div>
            </div>
            
            <h2>Recent Backtests</h2>
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th>Strategy</th>
                            <th>Instrument</th>
                            <th>Period</th>
                            <th>Status</th>
                            <th>P&L</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${backtests.slice(0, 10).map(bt => `
                            <tr>
                                <td>${bt.strategy_name}</td>
                                <td>${bt.symbol}</td>
                                <td>${formatDate(bt.start_time)} - ${formatDate(bt.end_time)}</td>
                                <td><span class="status-badge ${bt.status}">${bt.status}</span></td>
                                <td class="${bt.total_pnl >= 0 ? 'positive' : 'negative'}">
                                    ${formatNumber(bt.total_pnl)}
                                </td>
                                <td>
                                    <button onclick="viewBacktest(${bt.id})" class="btn-icon">View</button>
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
        
        document.getElementById('dashboard').innerHTML = dashboardHTML;
    } catch (error) {
        console.error('Failed to load dashboard:', error);
    }
}

// Strategies
async function loadStrategies() {
    try {
        const strategies = await apiCall('/strategies');
        state.strategies = strategies;
        
        const strategiesHTML = `
            <div class="section-header">
                <h2>Trading Strategies</h2>
                <button onclick="showCreateStrategyModal()" class="btn-primary">+ New Strategy</button>
            </div>
            
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th>Name</th>
                            <th>Type</th>
                            <th>Parameters</th>
                            <th>Created</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${strategies.map(s => `
                            <tr>
                                <td><strong>${s.name}</strong></td>
                                <td>${s.strategy_type}</td>
                                <td><code>${JSON.stringify(s.config).substring(0, 50)}...</code></td>
                                <td>${formatDate(s.created_at)}</td>
                                <td>
                                    <button onclick="editStrategy(${s.id})" class="btn-icon">Edit</button>
                                    <button onclick="deleteStrategy(${s.id})" class="btn-icon btn-danger">Delete</button>
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
        
        document.getElementById('strategies').innerHTML = strategiesHTML;
    } catch (error) {
        console.error('Failed to load strategies:', error);
    }
}

// Backtests
async function loadBacktests() {
    try {
        const backtests = await apiCall('/backtests');
        state.backtests = backtests;
        
        const backtestsHTML = `
            <div class="section-header">
                <h2>Backtest Results</h2>
                <button onclick="showRunBacktestModal()" class="btn-primary">+ Run Backtest</button>
            </div>
            
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Strategy</th>
                            <th>Instrument</th>
                            <th>Period</th>
                            <th>Status</th>
                            <th>Total P&L</th>
                            <th>Win Rate</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${backtests.map(bt => `
                            <tr>
                                <td>#${bt.id}</td>
                                <td>${bt.strategy_name}</td>
                                <td>${bt.symbol}</td>
                                <td>
                                    ${formatDate(bt.start_time)}<br/>
                                    <small>to ${formatDate(bt.end_time)}</small>
                                </td>
                                <td><span class="status-badge ${bt.status}">${bt.status}</span></td>
                                <td class="${bt.total_pnl >= 0 ? 'positive' : 'negative'}">
                                    ${formatNumber(bt.total_pnl)}
                                </td>
                                <td>${(bt.win_rate * 100).toFixed(2)}%</td>
                                <td>
                                    <button onclick="viewBacktest(${bt.id})" class="btn-icon">View</button>
                                    <button onclick="downloadReport(${bt.id})" class="btn-icon">Export</button>
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
        
        document.getElementById('backtests').innerHTML = backtestsHTML;
    } catch (error) {
        console.error('Failed to load backtests:', error);
    }
}

// Instruments
async function loadInstruments() {
    try {
        const instruments = await apiCall('/instruments');
        state.instruments = instruments;
        
        const instrumentsHTML = `
            <div class="section-header">
                <h2>Trading Instruments</h2>
                <button onclick="showAddInstrumentModal()" class="btn-primary">+ Add Instrument</button>
            </div>
            
            <div class="table-container">
                <table>
                    <thead>
                        <tr>
                            <th>Symbol</th>
                            <th>Name</th>
                            <th>Type</th>
                            <th>Min Price</th>
                            <th>Max Price</th>
                            <th>Tick Size</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${instruments.map(i => `
                            <tr>
                                <td><strong>${i.symbol}</strong></td>
                                <td>${i.name}</td>
                                <td>${i.instrument_type}</td>
                                <td>${formatNumber(i.min_price)}</td>
                                <td>${formatNumber(i.max_price)}</td>
                                <td>${formatNumber(i.tick_size)}</td>
                                <td>
                                    <button onclick="editInstrument('${i.symbol}')" class="btn-icon">Edit</button>
                                    <button onclick="deleteInstrument('${i.symbol}')" class="btn-icon btn-danger">Delete</button>
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
        
        document.getElementById('instruments').innerHTML = instrumentsHTML;
    } catch (error) {
        console.error('Failed to load instruments:', error);
    }
}

// CRUD Operations
async function createStrategy(formData) {
    try {
        const data = {
            name: formData.get('name'),
            strategy_type: formData.get('strategy_type'),
            config: JSON.parse(formData.get('config')),
            description: formData.get('description')
        };
        
        await apiCall('/strategies', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        
        showNotification('Strategy created successfully', 'success');
        closeModal('strategyModal');
        await loadStrategies();
    } catch (error) {
        showNotification('Failed to create strategy', 'error');
    }
}

async function deleteStrategy(id) {
    if (!confirm('Are you sure you want to delete this strategy?')) return;
    
    try {
        await apiCall(`/strategies/${id}`, { method: 'DELETE' });
        showNotification('Strategy deleted successfully', 'success');
        await loadStrategies();
    } catch (error) {
        showNotification('Failed to delete strategy', 'error');
    }
}

async function runBacktest(formData) {
    try {
        const data = {
            strategy_id: parseInt(formData.get('strategy_id')),
            symbol: formData.get('symbol'),
            start_time: new Date(formData.get('start_time')).toISOString(),
            end_time: new Date(formData.get('end_time')).toISOString(),
            initial_balance: parseFloat(formData.get('initial_balance'))
        };
        
        await apiCall('/backtests', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        
        showNotification('Backtest started successfully', 'success');
        closeModal('backtestModal');
        await loadBacktests();
    } catch (error) {
        showNotification('Failed to start backtest', 'error');
    }
}

// Utility Functions
function formatDate(dateString) {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
}

function formatNumber(num) {
    if (num === null || num === undefined) return 'N/A';
    return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
    }).format(num);
}

function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) modal.remove();
}

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    loadDashboard();
});
