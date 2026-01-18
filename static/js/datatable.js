/**
 * 0xDomainSnapshot - Enhanced DataTable Component
 * Provides sorting, filtering, pagination, and CSV export
 */
class DataTable {
    constructor(containerId, options = {}) {
        this.container = document.getElementById(containerId);
        if (!this.container) {
            console.error(`Container #${containerId} not found`);
            return;
        }

        this.options = {
            dataUrl: options.dataUrl || null,
            columns: options.columns || [],
            rowsPerPage: options.rowsPerPage || 50,
            pagingOptions: options.pagingOptions || [50, 100, 250, 500, -1],
            defaultSort: options.defaultSort || { column: 0, direction: 'asc' },
            rowClass: options.rowClass || null,
            onDataLoaded: options.onDataLoaded || null,
            emptyMessage: options.emptyMessage || 'No data available',
            showColumnToggle: options.showColumnToggle !== false,
            ...options
        };

        this.data = [];
        this.filteredData = [];
        this.eventListeners = [];
        this.hiddenColumns = new Set();
        this.accounts = [];
        this.selectedAccount = '';

        // Load state from URL parameters
        this.loadStateFromURL();

        this.init();
    }

    // Load state from URL parameters
    loadStateFromURL() {
        const urlParams = new URLSearchParams(window.location.search);

        this.currentPage = parseInt(urlParams.get('page')) || 0;
        this.sortColumn = parseInt(urlParams.get('sortCol')) || this.options.defaultSort.column;
        this.sortDirection = urlParams.get('sortDir') || this.options.defaultSort.direction;
        this.globalSearch = urlParams.get('search') || '';
        this.columnFilters = {};

        // Load column filters from URL
        urlParams.forEach((value, key) => {
            if (key.startsWith('filter_')) {
                const colIndex = parseInt(key.replace('filter_', ''));
                this.columnFilters[colIndex] = value;
            }
        });

        // Load rows per page
        const rowsParam = urlParams.get('rows');
        if (rowsParam) {
            this.options.rowsPerPage = parseInt(rowsParam);
        }

        // Load hidden columns from URL
        const hiddenCols = urlParams.get('hideCols');
        if (hiddenCols) {
            hiddenCols.split(',').forEach(col => {
                this.hiddenColumns.add(parseInt(col));
            });
        }

        // Load account filter from URL
        this.selectedAccount = urlParams.get('account') || '';
    }

    // Save state to URL parameters
    saveStateToURL() {
        const url = new URL(window.location);

        // Clear old DataTable params (preserve other params like 'service')
        ['page', 'sortCol', 'sortDir', 'search', 'rows', 'hideCols', 'account'].forEach(p => url.searchParams.delete(p));

        // Clear filter params
        const filterKeys = [];
        url.searchParams.forEach((value, key) => {
            if (key.startsWith('filter_')) filterKeys.push(key);
        });
        filterKeys.forEach(key => url.searchParams.delete(key));

        // Set new params (only if not default)
        if (this.currentPage > 0) url.searchParams.set('page', this.currentPage);
        if (this.sortColumn !== this.options.defaultSort.column) url.searchParams.set('sortCol', this.sortColumn);
        if (this.sortDirection !== 'asc') url.searchParams.set('sortDir', this.sortDirection);
        if (this.globalSearch) url.searchParams.set('search', this.globalSearch);
        if (this.options.rowsPerPage !== 50) url.searchParams.set('rows', this.options.rowsPerPage);

        // Column filters
        for (const [col, value] of Object.entries(this.columnFilters)) {
            if (value) url.searchParams.set(`filter_${col}`, value);
        }

        // Hidden columns
        if (this.hiddenColumns.size > 0) {
            url.searchParams.set('hideCols', Array.from(this.hiddenColumns).join(','));
        }

        // Account filter
        if (this.selectedAccount) {
            url.searchParams.set('account', this.selectedAccount);
        }

        window.history.replaceState({}, '', url);
    }

    // Clean up event listeners and DOM
    destroy() {
        // Remove all stored event listeners
        this.eventListeners.forEach(({ element, event, handler }) => {
            element.removeEventListener(event, handler);
        });
        this.eventListeners = [];

        // Clear container
        this.container.innerHTML = '';
    }

    // Helper to add event listener and track it for cleanup
    addEventHandler(element, event, handler) {
        element.addEventListener(event, handler);
        this.eventListeners.push({ element, event, handler });
    }

    async init() {
        this.render();
        if (this.options.data) {
            // Direct data passed
            this.setData(this.options.data);
        } else if (this.options.dataUrl) {
            await this.loadData(this.options.dataUrl);
        }
    }

    render() {
        const columnToggleHtml = this.options.showColumnToggle ? `
            <div class="column-toggle-container">
                <button class="column-toggle-btn" type="button">
                    Columns <span class="arrow">&#9662;</span>
                </button>
                <div class="column-toggle-dropdown">
                    ${this.options.columns.map((col, i) => `
                        <div class="column-toggle-item">
                            <input type="checkbox" id="col-toggle-${i}" data-column="${i}"
                                   ${!this.hiddenColumns.has(i) ? 'checked' : ''}>
                            <label for="col-toggle-${i}">${col.title}</label>
                        </div>
                    `).join('')}
                </div>
            </div>
        ` : '';

        // Check if data has account field
        const hasAccountColumn = this.options.columns.some(col => col.data === 'account');

        this.container.innerHTML = `
            <div class="table-container">
                <div class="table-toolbar">
                    ${hasAccountColumn ? `
                        <div class="account-filter-container">
                            <label for="account-filter">Account:</label>
                            <select id="account-filter" class="account-filter-select">
                                <option value="">All Accounts</option>
                            </select>
                        </div>
                    ` : ''}
                    ${columnToggleHtml}
                </div>
                <div class="table-wrapper">
                    <table id="${this.container.id}-table">
                        <thead>
                            <tr class="header-row">
                                ${this.options.columns.map((col, i) => `
                                    <th class="sortable ${i === this.sortColumn ? `sort-${this.sortDirection}` : ''} ${this.hiddenColumns.has(i) ? 'hidden-col' : ''}"
                                        data-column="${i}" ${this.hiddenColumns.has(i) ? 'style="display:none"' : ''}>${col.title}</th>
                                `).join('')}
                            </tr>
                            <tr class="filter-row">
                                ${this.options.columns.map((col, i) => `
                                    <th ${this.hiddenColumns.has(i) ? 'style="display:none"' : ''}>
                                        <input type="text" class="column-filter"
                                               data-column="${i}"
                                               placeholder="Filter ${col.title}...">
                                    </th>
                                `).join('')}
                            </tr>
                        </thead>
                        <tbody></tbody>
                    </table>
                </div>
                <div class="pagination-container">
                    <div class="pagination-info"></div>
                    <ul class="pagination"></ul>
                    <div class="page-size-select">
                        <span>Rows per page:</span>
                        <select class="page-size-selector">
                            ${this.options.pagingOptions.map(size => `
                                <option value="${size}" ${size === this.options.rowsPerPage ? 'selected' : ''}>
                                    ${size === -1 ? 'All' : size}
                                </option>
                            `).join('')}
                        </select>
                    </div>
                </div>
            </div>
        `;

        this.table = this.container.querySelector('table');
        this.tbody = this.table.querySelector('tbody');
        this.paginationInfo = this.container.querySelector('.pagination-info');
        this.paginationControls = this.container.querySelector('.pagination');

        // Restore column filter values from state
        this.table.querySelectorAll('.column-filter').forEach(input => {
            const colIndex = parseInt(input.dataset.column);
            if (this.columnFilters[colIndex]) {
                input.value = this.columnFilters[colIndex];
            }
        });

        // Restore global search value
        const globalSearch = document.getElementById('globalSearch');
        if (globalSearch && this.globalSearch) {
            globalSearch.value = this.globalSearch;
        }

        this.bindEvents();
    }

    bindEvents() {
        // Header click for sorting
        this.table.querySelectorAll('thead .header-row th').forEach(th => {
            const handler = () => {
                const column = parseInt(th.dataset.column);
                this.sortBy(column);
                this.saveStateToURL();
            };
            this.addEventHandler(th, 'click', handler);
        });

        // Column filters
        this.table.querySelectorAll('.column-filter').forEach(input => {
            const handler = (e) => {
                const column = parseInt(e.target.dataset.column);
                this.columnFilters[column] = e.target.value.toLowerCase();
                this.applyFilters();
                this.saveStateToURL();
            };
            this.addEventHandler(input, 'input', handler);
        });

        // Page size selector
        const pageSizeSelector = this.container.querySelector('.page-size-selector');
        const pageSizeHandler = (e) => {
            this.options.rowsPerPage = parseInt(e.target.value);
            this.currentPage = 0;
            this.renderTable();
            this.saveStateToURL();
        };
        this.addEventHandler(pageSizeSelector, 'change', pageSizeHandler);

        // Global search (external)
        const globalSearch = document.getElementById('globalSearch');
        if (globalSearch) {
            const searchHandler = (e) => {
                this.globalSearch = e.target.value.toLowerCase();
                this.applyFilters();
                this.saveStateToURL();
            };
            this.addEventHandler(globalSearch, 'input', searchHandler);
        }

        // Account filter
        const accountFilter = this.container.querySelector('#account-filter');
        if (accountFilter) {
            const accountHandler = (e) => {
                this.selectedAccount = e.target.value;
                this.applyFilters();
                this.saveStateToURL();
            };
            this.addEventHandler(accountFilter, 'change', accountHandler);
        }

        // Column visibility toggle
        const columnToggleBtn = this.container.querySelector('.column-toggle-btn');
        const columnToggleDropdown = this.container.querySelector('.column-toggle-dropdown');

        if (columnToggleBtn && columnToggleDropdown) {
            // Toggle dropdown
            const toggleHandler = (e) => {
                e.stopPropagation();
                columnToggleDropdown.classList.toggle('show');
            };
            this.addEventHandler(columnToggleBtn, 'click', toggleHandler);

            // Close dropdown when clicking outside
            const closeHandler = (e) => {
                if (!columnToggleDropdown.contains(e.target) && e.target !== columnToggleBtn) {
                    columnToggleDropdown.classList.remove('show');
                }
            };
            this.addEventHandler(document, 'click', closeHandler);

            // Column toggle checkboxes
            columnToggleDropdown.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
                const checkboxHandler = (e) => {
                    const colIndex = parseInt(e.target.dataset.column);
                    if (e.target.checked) {
                        this.hiddenColumns.delete(colIndex);
                    } else {
                        this.hiddenColumns.add(colIndex);
                    }
                    this.updateColumnVisibility();
                    this.saveStateToURL();
                };
                this.addEventHandler(checkbox, 'change', checkboxHandler);
            });
        }
    }

    updateColumnVisibility() {
        // Update header cells
        this.table.querySelectorAll('thead .header-row th').forEach(th => {
            const colIndex = parseInt(th.dataset.column);
            th.style.display = this.hiddenColumns.has(colIndex) ? 'none' : '';
        });

        // Update filter cells
        this.table.querySelectorAll('thead .filter-row th').forEach((th, i) => {
            th.style.display = this.hiddenColumns.has(i) ? 'none' : '';
        });

        // Re-render table body
        this.renderTable();
    }

    async loadData(url) {
        try {
            this.showLoading();
            const response = await fetch(url);
            if (!response.ok) throw new Error(`HTTP ${response.status}`);
            const data = await response.json();
            this.setData(data);
        } catch (error) {
            console.error('Failed to load data:', error);
            this.showError(`Failed to load data: ${error.message}`);
        }
    }

    setData(data) {
        this.data = Array.isArray(data) ? data : [];

        // Apply initial filter if provided
        if (this.options.filter) {
            this.data = this.data.filter(this.options.filter);
        }

        // Extract unique accounts
        this.extractAccounts();

        this.filteredData = [...this.data];
        this.currentPage = 0;
        this.sort();
        this.applyFilters(); // Apply account filter if set from URL
        this.renderTable();

        if (this.options.onDataLoaded) {
            this.options.onDataLoaded(this.data);
        }
    }

    extractAccounts() {
        const accountSet = new Set();
        this.data.forEach(row => {
            if (row.account) {
                accountSet.add(row.account);
            }
        });
        this.accounts = Array.from(accountSet).sort();

        // Populate account filter dropdown
        const accountFilter = this.container.querySelector('#account-filter');
        if (accountFilter) {
            // Clear existing options except "All Accounts"
            accountFilter.innerHTML = '<option value="">All Accounts</option>';
            this.accounts.forEach(account => {
                const option = document.createElement('option');
                option.value = account;
                option.textContent = account;
                if (account === this.selectedAccount) {
                    option.selected = true;
                }
                accountFilter.appendChild(option);
            });
        }
    }

    showLoading() {
        // Generate skeleton rows for better UX
        const visibleCols = this.options.columns.filter((_, i) => !this.hiddenColumns.has(i));
        const skeletonRows = [];

        for (let i = 0; i < 10; i++) {
            const cells = visibleCols.map((col, idx) => {
                // Vary skeleton widths based on column position
                const widthClass = idx === 0 ? 'medium' : idx === visibleCols.length - 1 ? 'small' : '';
                return `<td><div class="skeleton skeleton-cell ${widthClass}"></div></td>`;
            }).join('');
            skeletonRows.push(`<tr>${cells}</tr>`);
        }

        this.tbody.innerHTML = skeletonRows.join('');
    }

    showError(message) {
        this.tbody.innerHTML = `
            <tr>
                <td colspan="${this.options.columns.length}">
                    <div class="empty-state">
                        <div class="empty-state-icon">!</div>
                        <div class="empty-state-title">Error</div>
                        <div class="empty-state-text">${message}</div>
                    </div>
                </td>
            </tr>
        `;
    }

    sortBy(column) {
        if (this.sortColumn === column) {
            this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
        } else {
            this.sortColumn = column;
            this.sortDirection = 'asc';
        }
        this.sort();
        this.renderTable();
    }

    sort() {
        const col = this.options.columns[this.sortColumn];
        const key = col.data;

        this.filteredData.sort((a, b) => {
            let valueA = this.getValue(a, key);
            let valueB = this.getValue(b, key);

            // Handle null/undefined
            if (valueA == null) valueA = '';
            if (valueB == null) valueB = '';

            // Convert to string for comparison
            valueA = String(valueA).toLowerCase();
            valueB = String(valueB).toLowerCase();

            // Try numeric comparison
            const numA = parseFloat(valueA);
            const numB = parseFloat(valueB);
            if (!isNaN(numA) && !isNaN(numB)) {
                return this.sortDirection === 'asc' ? numA - numB : numB - numA;
            }

            // String comparison
            if (valueA < valueB) return this.sortDirection === 'asc' ? -1 : 1;
            if (valueA > valueB) return this.sortDirection === 'asc' ? 1 : -1;
            return 0;
        });

        // Update sort indicators
        this.table.querySelectorAll('thead .header-row th').forEach((th, i) => {
            th.classList.remove('sort-asc', 'sort-desc');
            if (i === this.sortColumn) {
                th.classList.add(`sort-${this.sortDirection}`);
            }
        });
    }

    getValue(obj, path) {
        if (!path) return '';
        const keys = path.split('.');
        let value = obj;
        for (const key of keys) {
            if (value == null) return '';
            value = value[key];
        }
        return value;
    }

    applyFilters() {
        this.filteredData = this.data.filter(row => {
            // Account filter
            if (this.selectedAccount) {
                if (row.account !== this.selectedAccount) return false;
            }

            // Global search
            if (this.globalSearch) {
                const rowText = this.options.columns.map(col =>
                    String(this.getValue(row, col.data) || '')
                ).join(' ').toLowerCase();
                if (!rowText.includes(this.globalSearch)) return false;
            }

            // Column filters
            for (const [colIndex, filterValue] of Object.entries(this.columnFilters)) {
                if (!filterValue) continue;
                const col = this.options.columns[colIndex];
                const cellValue = String(this.getValue(row, col.data) || '').toLowerCase();
                if (!cellValue.includes(filterValue)) return false;
            }

            return true;
        });

        this.currentPage = 0;
        this.sort();
        this.renderTable();
    }

    renderTable() {
        this.tbody.innerHTML = '';
        const visibleColCount = this.options.columns.length - this.hiddenColumns.size;

        if (this.filteredData.length === 0) {
            this.tbody.innerHTML = `
                <tr>
                    <td colspan="${visibleColCount}">
                        <div class="empty-state">
                            <div class="empty-state-icon">-</div>
                            <div class="empty-state-title">${this.options.emptyMessage}</div>
                        </div>
                    </td>
                </tr>
            `;
            this.updatePagination(0, 0, 0);
            return;
        }

        const total = this.filteredData.length;
        const rowsPerPage = this.options.rowsPerPage === -1 ? total : this.options.rowsPerPage;
        const startIndex = this.currentPage * rowsPerPage;
        const endIndex = Math.min(startIndex + rowsPerPage, total);

        for (let i = startIndex; i < endIndex; i++) {
            const row = this.filteredData[i];
            const tr = document.createElement('tr');

            // Apply row class if defined
            if (this.options.rowClass) {
                const className = this.options.rowClass(row);
                if (className) tr.className = className;
            }

            this.options.columns.forEach((col, colIndex) => {
                const td = document.createElement('td');
                if (col.className) td.className = col.className;

                // Hide column if in hiddenColumns
                if (this.hiddenColumns.has(colIndex)) {
                    td.style.display = 'none';
                }

                if (col.render) {
                    td.innerHTML = col.render(this.getValue(row, col.data), row);
                } else {
                    td.textContent = this.getValue(row, col.data) || '';
                }

                tr.appendChild(td);
            });

            this.tbody.appendChild(tr);
        }

        this.updatePagination(total, startIndex, endIndex);
    }

    updatePagination(total, startIndex, endIndex) {
        // Update info text
        this.paginationInfo.textContent = total > 0
            ? `Showing ${startIndex + 1} to ${endIndex} of ${total} entries`
            : 'No entries to show';

        // Build pagination controls
        this.paginationControls.innerHTML = '';

        const rowsPerPage = this.options.rowsPerPage === -1 ? total : this.options.rowsPerPage;
        const totalPages = rowsPerPage === 0 ? 1 : Math.ceil(total / rowsPerPage);

        if (totalPages <= 1) return;

        // Previous button
        const prevLi = document.createElement('li');
        const prevBtn = document.createElement('button');
        prevBtn.textContent = 'Previous';
        prevBtn.disabled = this.currentPage === 0;
        prevBtn.addEventListener('click', () => {
            if (this.currentPage > 0) {
                this.currentPage--;
                this.renderTable();
                this.saveStateToURL();
            }
        });
        prevLi.appendChild(prevBtn);
        this.paginationControls.appendChild(prevLi);

        // Page numbers
        const maxVisible = 7;
        let startPage = Math.max(0, Math.min(this.currentPage - Math.floor(maxVisible / 2), totalPages - maxVisible));
        let endPage = Math.min(totalPages, startPage + maxVisible);

        if (startPage > 0) {
            this.addPageButton(0);
            if (startPage > 1) this.addEllipsis();
        }

        for (let i = startPage; i < endPage; i++) {
            this.addPageButton(i);
        }

        if (endPage < totalPages) {
            if (endPage < totalPages - 1) this.addEllipsis();
            this.addPageButton(totalPages - 1);
        }

        // Next button
        const nextLi = document.createElement('li');
        const nextBtn = document.createElement('button');
        nextBtn.textContent = 'Next';
        nextBtn.disabled = this.currentPage >= totalPages - 1;
        nextBtn.addEventListener('click', () => {
            if (this.currentPage < totalPages - 1) {
                this.currentPage++;
                this.renderTable();
                this.saveStateToURL();
            }
        });
        nextLi.appendChild(nextBtn);
        this.paginationControls.appendChild(nextLi);
    }

    addPageButton(pageIndex) {
        const li = document.createElement('li');
        const btn = document.createElement('button');
        btn.textContent = pageIndex + 1;
        btn.classList.toggle('active', pageIndex === this.currentPage);
        btn.addEventListener('click', () => {
            this.currentPage = pageIndex;
            this.renderTable();
            this.saveStateToURL();
        });
        li.appendChild(btn);
        this.paginationControls.appendChild(li);
    }

    addEllipsis() {
        const li = document.createElement('li');
        li.className = 'ellipsis';
        li.textContent = '...';
        this.paginationControls.appendChild(li);
    }

    // Export to CSV
    exportCSV(filename = 'export.csv') {
        const headers = this.options.columns.map(col => col.title);
        const rows = this.filteredData.map(row =>
            this.options.columns.map(col => {
                let value = this.getValue(row, col.data);
                if (value == null) value = '';
                value = String(value);
                // Escape quotes and wrap in quotes if contains comma/quote/newline
                if (value.includes(',') || value.includes('"') || value.includes('\n')) {
                    value = '"' + value.replace(/"/g, '""') + '"';
                }
                return value;
            })
        );

        const csv = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
        const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
        const link = document.createElement('a');
        link.href = URL.createObjectURL(blob);
        link.download = filename;
        link.click();
        URL.revokeObjectURL(link.href);
    }

    // Get stats
    getStats() {
        const total = this.data.length;
        const active = this.data.filter(r => r.status === 'active').length;
        const removed = this.data.filter(r => r.status === 'removed').length;
        return { total, active, removed };
    }

    // Refresh data
    async refresh() {
        if (this.options.dataUrl) {
            await this.loadData(this.options.dataUrl);
        }
    }
}

// Helper function to render badges
function renderBadge(value, type = null) {
    const badgeType = type || value.toLowerCase();
    return `<span class="badge badge-${badgeType}">${value}</span>`;
}

// Helper function to render tags
function renderTags(tags) {
    if (!tags || typeof tags !== 'object') return '';
    return Object.entries(tags).map(([key, value]) =>
        `<span class="tag"><span class="tag-key">${key}:</span> ${value}</span>`
    ).join('');
}

// Helper function to render provider icon
function renderProvider(provider) {
    const icons = {
        'aws': 'AWS',
        'azure': 'AZ',
        'gcp': 'GCP',
        'godaddy': 'GD',
        'cloudflare': 'CF'
    };
    const key = provider.toLowerCase();
    return `<span class="provider-icon provider-${key}">${icons[key] || provider}</span>${provider}`;
}

// Export for use in modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { DataTable, renderBadge, renderTags, renderProvider };
}
