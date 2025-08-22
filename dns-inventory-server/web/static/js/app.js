/**
 * Modern JavaScript Application - React-like Architecture
 * Component-based architecture with state management
 */

// Global App Object
window.App = (function() {
    'use strict';

    // State Management
    const state = {
        currentPage: 'domains',
        loading: false,
        user: null,
        notifications: [],
        modals: []
    };

    // State Management System
    const StateManager = {
        listeners: {},
        
        setState(key, value) {
            state[key] = value;
            this.notify(key, value);
        },
        
        getState(key) {
            return state[key];
        },
        
        subscribe(key, callback) {
            if (!this.listeners[key]) {
                this.listeners[key] = [];
            }
            this.listeners[key].push(callback);
        },
        
        notify(key, value) {
            if (this.listeners[key]) {
                this.listeners[key].forEach(callback => callback(value));
            }
        }
    };

    // HTTP Client
    const HTTP = {
        async request(url, options = {}) {
            const config = {
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            };

            try {
                const response = await fetch(url, config);
                const data = await response.json();
                
                if (!response.ok) {
                    throw new Error(data.message || `HTTP ${response.status}`);
                }
                
                return data;
            } catch (error) {
                console.error('HTTP Request failed:', error);
                throw error;
            }
        },

        get(url, options = {}) {
            return this.request(url, { method: 'GET', ...options });
        },

        post(url, data, options = {}) {
            return this.request(url, {
                method: 'POST',
                body: JSON.stringify(data),
                ...options
            });
        },

        put(url, data, options = {}) {
            return this.request(url, {
                method: 'PUT',
                body: JSON.stringify(data),
                ...options
            });
        },

        delete(url, options = {}) {
            return this.request(url, { method: 'DELETE', ...options });
        }
    };

    // Component Base Class
    class Component {
        constructor(element) {
            this.element = element;
            this.state = {};
            this.init();
        }

        init() {
            // Override in subclasses
        }

        setState(newState) {
            this.state = { ...this.state, ...newState };
            this.render();
        }

        render() {
            // Override in subclasses
        }

        destroy() {
            // Cleanup
            if (this.element && this.element.parentNode) {
                this.element.parentNode.removeChild(this.element);
            }
        }
    }

    // Toast Notification System
    const Toast = {
        container: null,

        init() {
            this.container = document.getElementById('toast-container');
            if (!this.container) {
                this.container = document.createElement('div');
                this.container.id = 'toast-container';
                this.container.className = 'fixed top-4 right-4 z-50 space-y-2';
                document.body.appendChild(this.container);
            }
        },

        show(message, type = 'info', duration = 5000) {
            const toast = document.createElement('div');
            toast.className = `toast ${type} flex items-center space-x-3 p-4 rounded-lg shadow-lg`;
            
            const icon = this.getIcon(type);
            toast.innerHTML = `
                ${icon}
                <div class="flex-1">
                    <p class="text-sm font-medium text-white">${message}</p>
                </div>
                <button class="toast-close text-gray-400 hover:text-white transition-colors duration-200">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                    </svg>
                </button>
            `;

            // Add close functionality
            const closeBtn = toast.querySelector('.toast-close');
            closeBtn.addEventListener('click', () => this.remove(toast));

            this.container.appendChild(toast);

            // Auto remove
            if (duration > 0) {
                setTimeout(() => this.remove(toast), duration);
            }

            return toast;
        },

        getIcon(type) {
            const icons = {
                success: `<svg class="w-5 h-5 text-green-400" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>`,
                error: `<svg class="w-5 h-5 text-red-400" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>`,
                warning: `<svg class="w-5 h-5 text-orange-400" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
                </svg>`,
                info: `<svg class="w-5 h-5 text-blue-400" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>`
            };
            return icons[type] || icons.info;
        },

        remove(toast) {
            toast.style.animation = 'slideOut 0.3s ease-in forwards';
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.parentNode.removeChild(toast);
                }
            }, 300);
        },

        success(message) {
            return this.show(message, 'success');
        },

        error(message) {
            return this.show(message, 'error');
        },

        warning(message) {
            return this.show(message, 'warning');
        },

        info(message) {
            return this.show(message, 'info');
        }
    };

    // Modal System
    const Modal = {
        container: null,

        init() {
            this.container = document.getElementById('modal-container');
            if (!this.container) {
                this.container = document.createElement('div');
                this.container.id = 'modal-container';
                document.body.appendChild(this.container);
            }
        },

        show(content, options = {}) {
            const modal = document.createElement('div');
            modal.className = 'modal-overlay fixed inset-0 flex items-center justify-center z-50';
            
            modal.innerHTML = `
                <div class="modal bg-gray-800 rounded-xl p-6 max-w-md w-full mx-4 border border-gray-700">
                    <div class="modal-header flex items-center justify-between mb-4">
                        <h3 class="text-lg font-semibold text-white">${options.title || 'Modal'}</h3>
                        <button class="modal-close text-gray-400 hover:text-white transition-colors duration-200">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                            </svg>
                        </button>
                    </div>
                    <div class="modal-content">
                        ${content}
                    </div>
                </div>
            `;

            // Add event listeners
            const closeBtn = modal.querySelector('.modal-close');
            closeBtn.addEventListener('click', () => this.close(modal));

            // Close on backdrop click
            modal.addEventListener('click', (e) => {
                if (e.target === modal) {
                    this.close(modal);
                }
            });

            // Close on escape key
            const escapeHandler = (e) => {
                if (e.key === 'Escape') {
                    this.close(modal);
                    document.removeEventListener('keydown', escapeHandler);
                }
            };
            document.addEventListener('keydown', escapeHandler);

            this.container.appendChild(modal);
            return modal;
        },

        close(modal) {
            modal.style.animation = 'modalSlideOut 0.3s ease-in forwards';
            setTimeout(() => {
                if (modal.parentNode) {
                    modal.parentNode.removeChild(modal);
                }
            }, 300);
        }
    };

    // Loading System
    const Loading = {
        overlay: null,

        init() {
            this.overlay = document.getElementById('loading-overlay');
        },

        show(message = 'Loading...') {
            if (this.overlay) {
                this.overlay.querySelector('span').textContent = message;
                this.overlay.classList.remove('hidden');
            }
        },

        hide() {
            if (this.overlay) {
                this.overlay.classList.add('hidden');
            }
        }
    };

    // Time Utilities
    const TimeUtils = {
        updateCurrentTime() {
            const timeElement = document.getElementById('current-time');
            if (timeElement) {
                const now = new Date();
                timeElement.textContent = now.toLocaleTimeString();
            }
        },

        startClock() {
            this.updateCurrentTime();
            setInterval(() => this.updateCurrentTime(), 1000);
        },

        formatDate(dateString) {
            const date = new Date(dateString);
            return date.toLocaleDateString();
        },

        formatDateTime(dateString) {
            const date = new Date(dateString);
            return date.toLocaleString();
        }
    };

    // Initialize App
    function init() {
        console.log('🚀 DNS Inventory - Initializing...');
        
        // Initialize systems
        Toast.init();
        Modal.init();
        Loading.init();
        TimeUtils.startClock();

        // Initialize global event listeners
        initGlobalEventListeners();

        console.log('✅ App initialized successfully');
    }

    function initGlobalEventListeners() {
        // Refresh button
        const refreshBtn = document.getElementById('refresh-btn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                window.location.reload();
            });
        }

        // Navigation
        document.addEventListener('click', (e) => {
            if (e.target.matches('.nav-link')) {
                e.preventDefault();
                const href = e.target.getAttribute('href');
                if (href) {
                    window.location.href = href;
                }
            }
        });
    }

    // Public API
    return {
        init,
        StateManager,
        HTTP,
        Component,
        Toast,
        Modal,
        Loading,
        TimeUtils
    };
})();

// Domains Page Component
window.DomainsPage = (function() {
    let pageData = {};
    let currentSort = { column: 'domain', direction: 'asc' };

    function init(data) {
        pageData = data;
        console.log('📄 Domains Page - Initializing...', data);
        
        initEventListeners();
        initPagination();
        
        console.log('✅ Domains Page initialized');
    }

    function initEventListeners() {
        // Search functionality
        const searchInput = document.getElementById('search-input');
        if (searchInput) {
            let searchTimeout;
            searchInput.addEventListener('input', (e) => {
                clearTimeout(searchTimeout);
                searchTimeout = setTimeout(() => {
                    search(e.target.value);
                }, 300);
            });
        }

        // Source filter
        const sourceFilter = document.getElementById('source-filter');
        if (sourceFilter) {
            sourceFilter.addEventListener('change', (e) => {
                filterBySource(e.target.value);
            });
        }

        // Collect domains button
        const collectBtn = document.getElementById('collect-domains-btn');
        if (collectBtn) {
            collectBtn.addEventListener('click', collectDomains);
        }

        // Export button
        const exportBtn = document.getElementById('export-domains-btn');
        if (exportBtn) {
            exportBtn.addEventListener('click', exportDomains);
        }

        // Table sorting
        document.querySelectorAll('.table-header[data-sort]').forEach(header => {
            header.addEventListener('click', () => {
                const column = header.getAttribute('data-sort');
                sortTable(column);
            });
        });

        // Assign user buttons
        document.addEventListener('click', (e) => {
            if (e.target.matches('.assign-user-btn')) {
                const domainId = e.target.getAttribute('data-domain-id');
                showAssignUserModal(domainId);
            }
        });

        // Pagination
        document.addEventListener('click', (e) => {
            if (e.target.matches('.pagination-btn')) {
                const page = parseInt(e.target.getAttribute('data-page'));
                if (page) {
                    loadPage(page);
                }
            }
        });
    }

    function initPagination() {
        const pageNumbers = document.getElementById('page-numbers');
        if (!pageNumbers) return;

        const totalPages = Math.ceil(pageData.total / pageData.limit);
        const currentPage = pageData.currentPage;
        
        let html = '';
        
        // Show up to 5 page numbers
        const startPage = Math.max(1, currentPage - 2);
        const endPage = Math.min(totalPages, startPage + 4);
        
        for (let i = startPage; i <= endPage; i++) {
            const isActive = i === currentPage;
            html += `
                <button class="pagination-btn px-3 py-2 rounded-lg ${isActive ? 'bg-blue-600 text-white' : 'bg-gray-700 hover:bg-gray-600 text-white'} transition-colors duration-200" 
                        data-page="${i}" ${isActive ? 'disabled' : ''}>
                    ${i}
                </button>
            `;
        }
        
        pageNumbers.innerHTML = html;
    }

    async function search(query) {
        try {
            App.Loading.show('Searching...');
            
            const params = new URLSearchParams({
                search: query,
                page: 1,
                limit: pageData.limit
            });
            
            const response = await App.HTTP.get(`/api/domains?${params}`);
            updateTable(response);
            
        } catch (error) {
            App.Toast.error('Search failed: ' + error.message);
        } finally {
            App.Loading.hide();
        }
    }

    async function filterBySource(source) {
        // Client-side filtering for now
        const rows = document.querySelectorAll('.domain-row');
        
        rows.forEach(row => {
            const sourceCell = row.children[1].textContent.toLowerCase();
            const shouldShow = !source || 
                (source === 'godaddy' && sourceCell.includes('godaddy')) ||
                (source === 'cloudflare' && sourceCell.includes('cloudflare')) ||
                (source === 'both' && sourceCell.includes('godaddy') && sourceCell.includes('cloudflare'));
            
            row.style.display = shouldShow ? '' : 'none';
        });
    }

    function sortTable(column) {
        if (currentSort.column === column) {
            currentSort.direction = currentSort.direction === 'asc' ? 'desc' : 'asc';
        } else {
            currentSort.column = column;
            currentSort.direction = 'asc';
        }

        // Update header indicators
        document.querySelectorAll('.table-header[data-sort]').forEach(header => {
            const svg = header.querySelector('svg');
            svg.style.transform = '';
            svg.style.opacity = '0.5';
        });

        const currentHeader = document.querySelector(`[data-sort="${column}"]`);
        if (currentHeader) {
            const svg = currentHeader.querySelector('svg');
            svg.style.opacity = '1';
            svg.style.transform = currentSort.direction === 'desc' ? 'rotate(180deg)' : '';
        }

        // Sort table rows
        const tbody = document.getElementById('domains-tbody');
        const rows = Array.from(tbody.querySelectorAll('.domain-row'));
        
        rows.sort((a, b) => {
            let aVal, bVal;
            
            switch (column) {
                case 'domain':
                    aVal = a.children[0].textContent.trim();
                    bVal = b.children[0].textContent.trim();
                    break;
                case 'source':
                    aVal = a.children[1].textContent.trim();
                    bVal = b.children[1].textContent.trim();
                    break;
                case 'status':
                    aVal = a.children[2].textContent.trim();
                    bVal = b.children[2].textContent.trim();
                    break;
                case 'date':
                    aVal = a.children[4].textContent.trim();
                    bVal = b.children[4].textContent.trim();
                    break;
                default:
                    return 0;
            }
            
            if (currentSort.direction === 'asc') {
                return aVal.localeCompare(bVal);
            } else {
                return bVal.localeCompare(aVal);
            }
        });

        // Re-append sorted rows
        rows.forEach(row => tbody.appendChild(row));
    }

    async function collectDomains() {
        try {
            App.Loading.show('Collecting domains...');
            
            const response = await App.HTTP.post('/api/collect-domains');
            App.Toast.success('Domain collection started in background');
            
            // Reload page after a delay to show updated data
            setTimeout(() => {
                window.location.reload();
            }, 3000);
            
        } catch (error) {
            App.Toast.error('Failed to start domain collection: ' + error.message);
        } finally {
            App.Loading.hide();
        }
    }

    function exportDomains() {
        window.open('/export/domains.csv', '_blank');
        App.Toast.info('Export started - download should begin shortly');
    }

    function showAssignUserModal(domainId) {
        const users = pageData.users || [];
        
        const userOptions = users.map(user => 
            `<option value="${user.id}">${user.name} (${user.group})</option>`
        ).join('');

        const content = `
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Select User</label>
                    <select id="user-select" class="input-primary w-full px-3 py-2 rounded-lg">
                        <option value="">Choose a user...</option>
                        ${userOptions}
                    </select>
                </div>
                <div class="flex justify-end space-x-3">
                    <button class="btn-secondary px-4 py-2 rounded-lg" onclick="App.Modal.close(this.closest('.modal-overlay'))">
                        Cancel
                    </button>
                    <button id="assign-confirm" class="btn-primary px-4 py-2 rounded-lg">
                        Assign
                    </button>
                </div>
            </div>
        `;

        const modal = App.Modal.show(content, { title: 'Assign User to Domain' });
        
        // Handle assignment
        const assignBtn = modal.querySelector('#assign-confirm');
        assignBtn.addEventListener('click', async () => {
            const userId = modal.querySelector('#user-select').value;
            
            if (!userId) {
                App.Toast.error('Please select a user');
                return;
            }

            try {
                await App.HTTP.post('/api/assign-domain', {
                    domain_id: parseInt(domainId),
                    user_id: parseInt(userId)
                });
                
                App.Toast.success('User assigned successfully');
                App.Modal.close(modal);
                
                // Reload to show updated assignments
                setTimeout(() => window.location.reload(), 1000);
                
            } catch (error) {
                App.Toast.error('Assignment failed: ' + error.message);
            }
        });
    }

    async function loadPage(page) {
        try {
            App.Loading.show('Loading page...');
            
            const params = new URLSearchParams({
                page,
                limit: pageData.limit,
                search: document.getElementById('search-input')?.value || ''
            });
            
            window.location.href = `/domains?${params}`;
            
        } catch (error) {
            App.Toast.error('Failed to load page: ' + error.message);
        } finally {
            App.Loading.hide();
        }
    }

    function updateTable(data) {
        // Update table with new data
        // This would re-render the table body with new data
        // For now, we'll just reload the page
        window.location.reload();
    }

    return {
        init
    };
})();

// DNS Page Component
window.DNSPage = (function() {
    let pageData = {};
    let currentSort = { column: 'domain', direction: 'asc' };

    function init(data) {
        pageData = data;
        console.log('📄 DNS Page - Initializing...', data);
        
        initEventListeners();
        initPagination();
        
        console.log('✅ DNS Page initialized');
    }

    function initEventListeners() {
        // Search functionality
        const searchInput = document.getElementById('search-input');
        if (searchInput) {
            let searchTimeout;
            searchInput.addEventListener('input', (e) => {
                clearTimeout(searchTimeout);
                searchTimeout = setTimeout(() => {
                    search(e.target.value);
                }, 300);
            });
        }

        // Filters
        const typeFilter = document.getElementById('type-filter');
        const sourceFilter = document.getElementById('source-filter');
        const statusFilter = document.getElementById('status-filter');
        
        [typeFilter, sourceFilter, statusFilter].forEach(filter => {
            if (filter) {
                filter.addEventListener('change', applyFilters);
            }
        });

        // Collect DNS button
        const collectBtn = document.getElementById('collect-dns-btn');
        if (collectBtn) {
            collectBtn.addEventListener('click', collectDNS);
        }

        // Export button
        const exportBtn = document.getElementById('export-dns-btn');
        if (exportBtn) {
            exportBtn.addEventListener('click', exportDNS);
        }

        // Table sorting
        document.querySelectorAll('.table-header[data-sort]').forEach(header => {
            header.addEventListener('click', () => {
                const column = header.getAttribute('data-sort');
                sortTable(column);
            });
        });

        // Assign user buttons
        document.addEventListener('click', (e) => {
            if (e.target.matches('.assign-user-btn')) {
                const recordId = e.target.getAttribute('data-record-id');
                showAssignUserModal(recordId);
            }
        });
    }

    function initPagination() {
        // Similar to domains page pagination
        const pageNumbers = document.getElementById('page-numbers');
        if (!pageNumbers) return;

        const totalPages = Math.ceil(pageData.total / pageData.limit);
        const currentPage = pageData.currentPage;
        
        let html = '';
        const startPage = Math.max(1, currentPage - 2);
        const endPage = Math.min(totalPages, startPage + 4);
        
        for (let i = startPage; i <= endPage; i++) {
            const isActive = i === currentPage;
            html += `
                <button class="pagination-btn px-3 py-2 rounded-lg ${isActive ? 'bg-blue-600 text-white' : 'bg-gray-700 hover:bg-gray-600 text-white'} transition-colors duration-200" 
                        data-page="${i}" ${isActive ? 'disabled' : ''}>
                    ${i}
                </button>
            `;
        }
        
        pageNumbers.innerHTML = html;
    }

    async function search(query) {
        try {
            App.Loading.show('Searching...');
            
            const params = new URLSearchParams({
                search: query,
                page: 1,
                limit: pageData.limit
            });
            
            const response = await App.HTTP.get(`/api/dns?${params}`);
            updateTable(response);
            
        } catch (error) {
            App.Toast.error('Search failed: ' + error.message);
        } finally {
            App.Loading.hide();
        }
    }

    function applyFilters() {
        const typeFilter = document.getElementById('type-filter').value;
        const sourceFilter = document.getElementById('source-filter').value;
        const statusFilter = document.getElementById('status-filter').value;
        
        const rows = document.querySelectorAll('.dns-row');
        
        rows.forEach(row => {
            const type = row.children[2].textContent.trim();
            const source = row.children[4].textContent.trim();
            const status = row.children[5].textContent.trim();
            
            const shouldShow = 
                (!typeFilter || type === typeFilter) &&
                (!sourceFilter || source === sourceFilter) &&
                (!statusFilter || status === statusFilter);
            
            row.style.display = shouldShow ? '' : 'none';
        });
    }

    function sortTable(column) {
        if (currentSort.column === column) {
            currentSort.direction = currentSort.direction === 'asc' ? 'desc' : 'asc';
        } else {
            currentSort.column = column;
            currentSort.direction = 'asc';
        }

        // Update header indicators
        document.querySelectorAll('.table-header[data-sort]').forEach(header => {
            const svg = header.querySelector('svg');
            svg.style.transform = '';
            svg.style.opacity = '0.5';
        });

        const currentHeader = document.querySelector(`[data-sort="${column}"]`);
        if (currentHeader) {
            const svg = currentHeader.querySelector('svg');
            svg.style.opacity = '1';
            svg.style.transform = currentSort.direction === 'desc' ? 'rotate(180deg)' : '';
        }

        // Sort table rows
        const tbody = document.getElementById('dns-tbody');
        const rows = Array.from(tbody.querySelectorAll('.dns-row'));
        
        rows.sort((a, b) => {
            let aVal, bVal;
            
            switch (column) {
                case 'domain':
                    aVal = a.children[0].textContent.trim();
                    bVal = b.children[0].textContent.trim();
                    break;
                case 'subdomain':
                    aVal = a.children[1].textContent.trim();
                    bVal = b.children[1].textContent.trim();
                    break;
                case 'type':
                    aVal = a.children[2].textContent.trim();
                    bVal = b.children[2].textContent.trim();
                    break;
                case 'source':
                    aVal = a.children[4].textContent.trim();
                    bVal = b.children[4].textContent.trim();
                    break;
                case 'status':
                    aVal = a.children[5].textContent.trim();
                    bVal = b.children[5].textContent.trim();
                    break;
                case 'date':
                    aVal = a.children[8].textContent.trim();
                    bVal = b.children[8].textContent.trim();
                    break;
                default:
                    return 0;
            }
            
            if (currentSort.direction === 'asc') {
                return aVal.localeCompare(bVal);
            } else {
                return bVal.localeCompare(aVal);
            }
        });

        rows.forEach(row => tbody.appendChild(row));
    }

    async function collectDNS() {
        try {
            App.Loading.show('Collecting DNS records...');
            
            const response = await App.HTTP.post('/api/collect-dns');
            App.Toast.success('DNS collection started in background');
            
            setTimeout(() => {
                window.location.reload();
            }, 3000);
            
        } catch (error) {
            App.Toast.error('Failed to start DNS collection: ' + error.message);
        } finally {
            App.Loading.hide();
        }
    }

    function exportDNS() {
        window.open('/export/dns.csv', '_blank');
        App.Toast.info('Export started - download should begin shortly');
    }

    function showAssignUserModal(recordId) {
        const users = pageData.users || [];
        
        const userOptions = users.map(user => 
            `<option value="${user.id}">${user.name} (${user.group})</option>`
        ).join('');

        const content = `
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Select User</label>
                    <select id="user-select" class="input-primary w-full px-3 py-2 rounded-lg">
                        <option value="">Choose a user...</option>
                        ${userOptions}
                    </select>
                </div>
                <div class="flex justify-end space-x-3">
                    <button class="btn-secondary px-4 py-2 rounded-lg" onclick="App.Modal.close(this.closest('.modal-overlay'))">
                        Cancel
                    </button>
                    <button id="assign-confirm" class="btn-primary px-4 py-2 rounded-lg">
                        Assign
                    </button>
                </div>
            </div>
        `;

        const modal = App.Modal.show(content, { title: 'Assign User to DNS Record' });
        
        const assignBtn = modal.querySelector('#assign-confirm');
        assignBtn.addEventListener('click', async () => {
            const userId = modal.querySelector('#user-select').value;
            
            if (!userId) {
                App.Toast.error('Please select a user');
                return;
            }

            try {
                await App.HTTP.post('/api/assign-dns', {
                    record_id: parseInt(recordId),
                    user_id: parseInt(userId)
                });
                
                App.Toast.success('User assigned successfully');
                App.Modal.close(modal);
                
                setTimeout(() => window.location.reload(), 1000);
                
            } catch (error) {
                App.Toast.error('Assignment failed: ' + error.message);
            }
        });
    }

    function updateTable(data) {
        window.location.reload();
    }

    return { init };
})();

// Users Page Component
window.UsersPage = (function() {
    let pageData = {};
    let currentSort = { column: 'name', direction: 'asc' };

    function init(data) {
        pageData = data;
        console.log('📄 Users Page - Initializing...', data);
        
        initEventListeners();
        
        console.log('✅ Users Page initialized');
    }

    function initEventListeners() {
        // Search functionality
        const searchInput = document.getElementById('search-input');
        if (searchInput) {
            let searchTimeout;
            searchInput.addEventListener('input', (e) => {
                clearTimeout(searchTimeout);
                searchTimeout = setTimeout(() => {
                    search(e.target.value);
                }, 300);
            });
        }

        // Filters
        const groupFilter = document.getElementById('group-filter');
        const statusFilter = document.getElementById('status-filter');
        
        [groupFilter, statusFilter].forEach(filter => {
            if (filter) {
                filter.addEventListener('change', applyFilters);
            }
        });

        // Add user button
        const addUserBtn = document.getElementById('add-user-btn');
        if (addUserBtn) {
            addUserBtn.addEventListener('click', showAddUserModal);
        }

        // Table sorting
        document.querySelectorAll('.table-header[data-sort]').forEach(header => {
            header.addEventListener('click', () => {
                const column = header.getAttribute('data-sort');
                sortTable(column);
            });
        });

        // Edit and delete buttons
        document.addEventListener('click', (e) => {
            if (e.target.matches('.edit-user-btn')) {
                const userId = e.target.getAttribute('data-user-id');
                showEditUserModal(userId);
            }
            
            if (e.target.matches('.delete-user-btn')) {
                const userId = e.target.getAttribute('data-user-id');
                showDeleteUserConfirm(userId);
            }
        });
    }

    function search(query) {
        const rows = document.querySelectorAll('.user-row');
        
        rows.forEach(row => {
            const name = row.children[0].textContent.toLowerCase();
            const group = row.children[1].textContent.toLowerCase();
            const email = row.children[2].textContent.toLowerCase();
            
            const shouldShow = !query || 
                name.includes(query.toLowerCase()) ||
                group.includes(query.toLowerCase()) ||
                email.includes(query.toLowerCase());
            
            row.style.display = shouldShow ? '' : 'none';
        });
    }

    function applyFilters() {
        const groupFilter = document.getElementById('group-filter').value;
        const statusFilter = document.getElementById('status-filter').value;
        
        const rows = document.querySelectorAll('.user-row');
        
        rows.forEach(row => {
            const group = row.children[1].textContent.trim();
            const status = row.children[3].textContent.trim().toLowerCase();
            
            const shouldShow = 
                (!groupFilter || group === groupFilter) &&
                (!statusFilter || status === statusFilter);
            
            row.style.display = shouldShow ? '' : 'none';
        });
    }

    function sortTable(column) {
        if (currentSort.column === column) {
            currentSort.direction = currentSort.direction === 'asc' ? 'desc' : 'asc';
        } else {
            currentSort.column = column;
            currentSort.direction = 'asc';
        }

        // Update header indicators
        document.querySelectorAll('.table-header[data-sort]').forEach(header => {
            const svg = header.querySelector('svg');
            if (svg) {
                svg.style.transform = '';
                svg.style.opacity = '0.5';
            }
        });

        const currentHeader = document.querySelector(`[data-sort="${column}"]`);
        if (currentHeader) {
            const svg = currentHeader.querySelector('svg');
            if (svg) {
                svg.style.opacity = '1';
                svg.style.transform = currentSort.direction === 'desc' ? 'rotate(180deg)' : '';
            }
        }

        // Sort table rows
        const tbody = document.getElementById('users-tbody');
        const rows = Array.from(tbody.querySelectorAll('.user-row'));
        
        rows.sort((a, b) => {
            let aVal, bVal;
            
            switch (column) {
                case 'name':
                    aVal = a.children[0].textContent.trim();
                    bVal = b.children[0].textContent.trim();
                    break;
                case 'group':
                    aVal = a.children[1].textContent.trim();
                    bVal = b.children[1].textContent.trim();
                    break;
                case 'status':
                    aVal = a.children[3].textContent.trim();
                    bVal = b.children[3].textContent.trim();
                    break;
                case 'created':
                    aVal = a.children[4].textContent.trim();
                    bVal = b.children[4].textContent.trim();
                    break;
                default:
                    return 0;
            }
            
            if (currentSort.direction === 'asc') {
                return aVal.localeCompare(bVal);
            } else {
                return bVal.localeCompare(aVal);
            }
        });

        rows.forEach(row => tbody.appendChild(row));
    }

    function showAddUserModal() {
        const groupOptions = pageData.groups.map(group => 
            `<option value="${group}">${group}</option>`
        ).join('');

        const content = `
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Name</label>
                    <input type="text" id="user-name" class="input-primary w-full px-3 py-2 rounded-lg" placeholder="Enter user name">
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Group</label>
                    <input type="text" id="user-group" list="group-list" class="input-primary w-full px-3 py-2 rounded-lg" placeholder="Enter or select group">
                    <datalist id="group-list">
                        ${groupOptions}
                    </datalist>
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Email (Optional)</label>
                    <input type="email" id="user-email" class="input-primary w-full px-3 py-2 rounded-lg" placeholder="Enter email address">
                </div>
                <div class="flex justify-end space-x-3">
                    <button class="btn-secondary px-4 py-2 rounded-lg" onclick="App.Modal.close(this.closest('.modal-overlay'))">
                        Cancel
                    </button>
                    <button id="create-user" class="btn-primary px-4 py-2 rounded-lg">
                        Create User
                    </button>
                </div>
            </div>
        `;

        const modal = App.Modal.show(content, { title: 'Add New User' });
        
        const createBtn = modal.querySelector('#create-user');
        createBtn.addEventListener('click', async () => {
            const name = modal.querySelector('#user-name').value.trim();
            const group = modal.querySelector('#user-group').value.trim();
            const email = modal.querySelector('#user-email').value.trim();
            
            if (!name || !group) {
                App.Toast.error('Name and group are required');
                return;
            }

            try {
                await App.HTTP.post('/api/users', {
                    name,
                    group,
                    email
                });
                
                App.Toast.success('User created successfully');
                App.Modal.close(modal);
                
                setTimeout(() => window.location.reload(), 1000);
                
            } catch (error) {
                App.Toast.error('Failed to create user: ' + error.message);
            }
        });
    }

    function showEditUserModal(userId) {
        const user = pageData.users.find(u => u.id == userId);
        if (!user) return;

        const groupOptions = pageData.groups.map(group => 
            `<option value="${group}" ${group === user.group ? 'selected' : ''}>${group}</option>`
        ).join('');

        const content = `
            <div class="space-y-4">
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Name</label>
                    <input type="text" id="user-name" class="input-primary w-full px-3 py-2 rounded-lg" value="${user.name}">
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Group</label>
                    <input type="text" id="user-group" list="group-list" class="input-primary w-full px-3 py-2 rounded-lg" value="${user.group}">
                    <datalist id="group-list">
                        ${groupOptions}
                    </datalist>
                </div>
                <div>
                    <label class="block text-sm font-medium text-gray-300 mb-2">Email</label>
                    <input type="email" id="user-email" class="input-primary w-full px-3 py-2 rounded-lg" value="${user.email || ''}">
                </div>
                <div>
                    <label class="flex items-center">
                        <input type="checkbox" id="user-active" class="mr-2" ${user.is_active ? 'checked' : ''}>
                        <span class="text-sm text-gray-300">Active User</span>
                    </label>
                </div>
                <div class="flex justify-end space-x-3">
                    <button class="btn-secondary px-4 py-2 rounded-lg" onclick="App.Modal.close(this.closest('.modal-overlay'))">
                        Cancel
                    </button>
                    <button id="update-user" class="btn-primary px-4 py-2 rounded-lg">
                        Update User
                    </button>
                </div>
            </div>
        `;

        const modal = App.Modal.show(content, { title: 'Edit User' });
        
        const updateBtn = modal.querySelector('#update-user');
        updateBtn.addEventListener('click', async () => {
            const name = modal.querySelector('#user-name').value.trim();
            const group = modal.querySelector('#user-group').value.trim();
            const email = modal.querySelector('#user-email').value.trim();
            const isActive = modal.querySelector('#user-active').checked;
            
            if (!name || !group) {
                App.Toast.error('Name and group are required');
                return;
            }

            try {
                await App.HTTP.put('/api/users', {
                    id: parseInt(userId),
                    name,
                    group,
                    email,
                    is_active: isActive
                });
                
                App.Toast.success('User updated successfully');
                App.Modal.close(modal);
                
                setTimeout(() => window.location.reload(), 1000);
                
            } catch (error) {
                App.Toast.error('Failed to update user: ' + error.message);
            }
        });
    }

    function showDeleteUserConfirm(userId) {
        const user = pageData.users.find(u => u.id == userId);
        if (!user) return;

        const content = `
            <div class="space-y-4">
                <p class="text-gray-300">Are you sure you want to delete the user <strong>${user.name}</strong>?</p>
                <p class="text-sm text-red-400">This action cannot be undone.</p>
                <div class="flex justify-end space-x-3">
                    <button class="btn-secondary px-4 py-2 rounded-lg" onclick="App.Modal.close(this.closest('.modal-overlay'))">
                        Cancel
                    </button>
                    <button id="confirm-delete" class="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-700 text-white font-medium transition-colors duration-200">
                        Delete User
                    </button>
                </div>
            </div>
        `;

        const modal = App.Modal.show(content, { title: 'Delete User' });
        
        const deleteBtn = modal.querySelector('#confirm-delete');
        deleteBtn.addEventListener('click', async () => {
            try {
                await App.HTTP.delete(`/api/users?id=${userId}`);
                
                App.Toast.success('User deleted successfully');
                App.Modal.close(modal);
                
                setTimeout(() => window.location.reload(), 1000);
                
            } catch (error) {
                App.Toast.error('Failed to delete user: ' + error.message);
            }
        });
    }

    return { init };
})();

// Add CSS animations for modals and toasts
const style = document.createElement('style');
style.textContent = `
    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
    
    @keyframes modalSlideOut {
        from {
            transform: translateY(0);
            opacity: 1;
        }
        to {
            transform: translateY(-50px);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);