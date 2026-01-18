/**
 * 0xDomainSnapshot - Navigation Component
 * Dynamically injects the navigation bar into all pages
 */

const NAV_CONFIG = {
    brand: {
        name: '0xDomainSnapshot',
        href: 'index.html'
    },
    items: [
        {
            title: 'DNS',
            items: [
                { title: 'Domains', href: 'dns/domains.html' },
                { title: 'Subdomains', href: 'dns/subdomains.html' }
            ]
        },
        {
            title: 'AWS',
            items: [
                { title: 'EC2 Instances', href: 'aws/ec2.html' },
                { title: 'EKS Clusters', href: 'aws/eks.html' },
                { title: 'ECS Clusters', href: 'aws/ecs.html' },
                { title: 'Lambda Functions', href: 'aws/lambda.html' },
                { title: 'App Runner', href: 'aws/apprunner.html' },
                { title: 'Elastic Beanstalk', href: 'aws/elasticbeanstalk.html' },
                { title: 'Lightsail', href: 'aws/lightsail.html' },
                { title: 'RDS Databases', href: 'aws/rds.html' },
                { title: 'DynamoDB', href: 'aws/dynamodb.html' },
                { title: 'ElastiCache', href: 'aws/elasticache.html' },
                { title: 'MemoryDB', href: 'aws/memorydb.html' },
                { title: 'DocumentDB', href: 'aws/documentdb.html' },
                { title: 'Neptune', href: 'aws/neptune.html' },
                { title: 'Redshift', href: 'aws/redshift.html' },
                { title: 'OpenSearch', href: 'aws/opensearch.html' },
                { title: 'S3 Buckets', href: 'aws/s3.html' },
                { title: 'EFS', href: 'aws/efs.html' },
                { title: 'FSx', href: 'aws/fsx.html' },
                { title: 'ECR', href: 'aws/ecr.html' },
                { title: 'IAM Users', href: 'aws/iam.html' },
                { title: 'Cognito', href: 'aws/cognito.html' },
                { title: 'Secrets Manager', href: 'aws/secrets.html' },
                { title: 'KMS Keys', href: 'aws/kms.html' },
                { title: 'VPCs', href: 'aws/vpc.html' },
                { title: 'Load Balancers', href: 'aws/elb.html' },
                { title: 'CloudFront', href: 'aws/cloudfront.html' },
                { title: 'API Gateway', href: 'aws/apigateway.html' },
                { title: 'AppSync', href: 'aws/appsync.html' },
                { title: 'SQS Queues', href: 'aws/sqs.html' },
                { title: 'SNS Topics', href: 'aws/sns.html' },
                { title: 'EventBridge', href: 'aws/eventbridge.html' },
                { title: 'Kinesis', href: 'aws/kinesis.html' },
                { title: 'MSK', href: 'aws/msk.html' },
                { title: 'Amazon MQ', href: 'aws/mq.html' },
                { title: 'Step Functions', href: 'aws/stepfunctions.html' },
                { title: 'Glue', href: 'aws/glue.html' },
                { title: 'Athena', href: 'aws/athena.html' },
                { title: 'EMR', href: 'aws/emr.html' },
                { title: 'SageMaker', href: 'aws/sagemaker.html' },
                { title: 'Batch', href: 'aws/batch.html' },
                { title: 'CodePipeline', href: 'aws/codepipeline.html' },
                { title: 'CodeBuild', href: 'aws/codebuild.html' },
                { title: 'CodeCommit', href: 'aws/codecommit.html' },
                { title: 'CodeDeploy', href: 'aws/codedeploy.html' },
                { title: 'CodeArtifact', href: 'aws/codeartifact.html' },
                { title: 'Amplify', href: 'aws/amplify.html' },
                { title: 'Backup', href: 'aws/backup.html' },
                { title: 'Transfer Family', href: 'aws/transfer.html' },
                { title: 'CloudWatch', href: 'aws/cloudwatch.html' },
                { title: 'WAF', href: 'aws/waf.html' },
                { title: 'ACM Certificates', href: 'aws/acm.html' }
            ]
        },
        {
            title: 'Azure',
            items: [
                { title: 'Virtual Machines', href: 'azure/vms.html' },
                { title: 'AKS Clusters', href: 'azure/aks.html' },
                { title: 'App Service', href: 'azure/appservice.html' },
                { title: 'Functions', href: 'azure/functions.html' },
                { title: 'SQL Databases', href: 'azure/sql.html' },
                { title: 'Cosmos DB', href: 'azure/cosmosdb.html' },
                { title: 'Redis Cache', href: 'azure/redis.html' },
                { title: 'Storage Accounts', href: 'azure/storage.html' },
                { title: 'Container Registry', href: 'azure/containerregistry.html' },
                { title: 'AD Users', href: 'azure/ad-users.html' },
                { title: 'Key Vault', href: 'azure/keyvault.html' },
                { title: 'CDN Profiles', href: 'azure/cdn.html' },
                { title: 'Front Door', href: 'azure/frontdoor.html' },
                { title: 'Load Balancer', href: 'azure/loadbalancer.html' },
                { title: 'Virtual Networks', href: 'azure/vnet.html' },
                { title: 'API Management', href: 'azure/apim.html' },
                { title: 'Service Bus', href: 'azure/servicebus.html' },
                { title: 'Event Hubs', href: 'azure/eventhub.html' },
                { title: 'Logic Apps', href: 'azure/logicapps.html' },
                { title: 'Data Factory', href: 'azure/datafactory.html' },
                { title: 'Synapse', href: 'azure/synapse.html' },
                { title: 'Databricks', href: 'azure/databricks.html' },
                { title: 'ML Workspace', href: 'azure/mlworkspace.html' },
                { title: 'Cognitive Services', href: 'azure/cognitiveservices.html' },
                { title: 'DevOps', href: 'azure/devops.html' }
            ]
        },
        {
            title: 'GCP',
            items: [
                { title: 'Compute Instances', href: 'gcp/compute.html' },
                { title: 'GKE Clusters', href: 'gcp/gke.html' },
                { title: 'Cloud Run', href: 'gcp/cloudrun.html' },
                { title: 'Cloud Functions', href: 'gcp/functions.html' },
                { title: 'Cloud SQL', href: 'gcp/cloudsql.html' },
                { title: 'BigQuery', href: 'gcp/bigquery.html' },
                { title: 'Bigtable', href: 'gcp/bigtable.html' },
                { title: 'Spanner', href: 'gcp/spanner.html' },
                { title: 'Memorystore', href: 'gcp/memorystore.html' },
                { title: 'Cloud Storage', href: 'gcp/gcs.html' },
                { title: 'Artifact Registry', href: 'gcp/artifactregistry.html' },
                { title: 'IAM Accounts', href: 'gcp/iam.html' },
                { title: 'Secret Manager', href: 'gcp/secretmanager.html' },
                { title: 'Load Balancing', href: 'gcp/loadbalancing.html' },
                { title: 'Pub/Sub', href: 'gcp/pubsub.html' },
                { title: 'Cloud Tasks', href: 'gcp/cloudtasks.html' },
                { title: 'Cloud Scheduler', href: 'gcp/cloudscheduler.html' },
                { title: 'Dataflow', href: 'gcp/dataflow.html' },
                { title: 'Dataproc', href: 'gcp/dataproc.html' },
                { title: 'Vertex AI', href: 'gcp/vertexai.html' },
                { title: 'Cloud Composer', href: 'gcp/composer.html' },
                { title: 'Cloud Build', href: 'gcp/cloudbuild.html' }
            ]
        },
        {
            title: 'Cloudflare',
            items: [
                { title: 'Workers', href: 'cloudflare/workers.html' },
                { title: 'Pages', href: 'cloudflare/pages.html' },
                { title: 'R2 Storage', href: 'cloudflare/r2.html' },
                { title: 'KV Namespaces', href: 'cloudflare/kv.html' },
                { title: 'D1 Databases', href: 'cloudflare/d1.html' },
                { title: 'Queues', href: 'cloudflare/queues.html' },
                { title: 'Durable Objects', href: 'cloudflare/durable-objects.html' },
                { title: 'Images', href: 'cloudflare/images.html' },
                { title: 'Stream', href: 'cloudflare/stream.html' },
                { title: 'Tunnels', href: 'cloudflare/tunnels.html' },
                { title: 'Access Apps', href: 'cloudflare/access.html' }
            ]
        },
        {
            title: 'DigitalOcean',
            items: [
                { title: 'Droplets', href: 'digitalocean/droplets.html' },
                { title: 'Kubernetes', href: 'digitalocean/kubernetes.html' },
                { title: 'App Platform', href: 'digitalocean/apps.html' },
                { title: 'Databases', href: 'digitalocean/databases.html' },
                { title: 'Spaces', href: 'digitalocean/spaces.html' }
            ]
        },
        {
            title: 'Oracle',
            items: [
                { title: 'Compute Instances', href: 'oracle/compute.html' },
                { title: 'OKE Clusters', href: 'oracle/oke.html' },
                { title: 'Databases', href: 'oracle/database.html' },
                { title: 'Object Storage', href: 'oracle/objectstorage.html' }
            ]
        },
        {
            title: 'Alibaba',
            items: [
                { title: 'ECS Instances', href: 'alibaba/ecs.html' },
                { title: 'ACK Clusters', href: 'alibaba/ack.html' },
                { title: 'RDS Databases', href: 'alibaba/rds.html' },
                { title: 'OSS Buckets', href: 'alibaba/oss.html' }
            ]
        },
        {
            title: 'Removed',
            items: [
                { title: 'DNS Removed', href: 'removed/dns-removed.html' },
                { title: 'AWS Removed', href: 'removed/aws-removed.html' },
                { title: 'Azure Removed', href: 'removed/azure-removed.html' },
                { title: 'GCP Removed', href: 'removed/gcp-removed.html' },
                { title: 'Cloudflare Removed', href: 'removed/cloudflare-removed.html' },
                { title: 'DigitalOcean Removed', href: 'removed/digitalocean-removed.html' },
                { title: 'Oracle Removed', href: 'removed/oracle-removed.html' },
                { title: 'Alibaba Removed', href: 'removed/alibaba-removed.html' }
            ]
        }
    ]
};

class Navigation {
    constructor() {
        this.currentPath = window.location.pathname;
        this.basePath = this.getBasePath();
        this.isMobile = window.innerWidth <= 768;
    }

    getBasePath() {
        // Determine if we're in a subdirectory
        const path = window.location.pathname;
        if (path.includes('/aws/') || path.includes('/azure/') || path.includes('/gcp/') ||
            path.includes('/cloudflare/') || path.includes('/digitalocean/') ||
            path.includes('/oracle/') || path.includes('/alibaba/') ||
            path.includes('/dns/') || path.includes('/removed/')) {
            return '../';
        }
        return '';
    }

    resolvePath(href) {
        return this.basePath + href;
    }

    isActive(href) {
        const currentPage = this.currentPath.split('/').pop();
        const linkPage = href.split('/').pop();
        return currentPage === linkPage;
    }

    isDropdownActive(items) {
        return items.some(item => this.isActive(item.href));
    }

    // Get current page info for breadcrumb
    getCurrentPageInfo() {
        const path = this.currentPath;
        const pageName = path.split('/').pop().replace('.html', '');

        // Find current page in nav config
        for (const item of NAV_CONFIG.items) {
            if (item.items) {
                for (const subItem of item.items) {
                    if (this.isActive(subItem.href)) {
                        return {
                            category: item.title,
                            page: subItem.title,
                            href: subItem.href
                        };
                    }
                }
            }
        }

        // Check if it's the index page
        if (pageName === 'index' || pageName === '') {
            return { category: null, page: 'Dashboard', href: 'index.html' };
        }

        return { category: null, page: pageName, href: null };
    }

    render() {
        const nav = document.createElement('nav');
        nav.className = 'navbar';
        nav.innerHTML = `
            <div class="navbar-container">
                <a href="${this.resolvePath(NAV_CONFIG.brand.href)}" class="navbar-brand">
                    ${NAV_CONFIG.brand.name}
                </a>
                <button class="mobile-menu-btn" aria-label="Toggle menu">&#9776;</button>
                <ul class="navbar-nav">
                    ${NAV_CONFIG.items.map(item => this.renderNavItem(item)).join('')}
                </ul>
                <div class="navbar-search">
                    <input type="text" id="globalSearch" placeholder="Search all columns... (Ctrl+F)">
                </div>
                <a href="${this.resolvePath('search.html')}" class="global-search-link" title="Global Search">&#128269;</a>
                <div class="theme-toggle">
                    <button class="theme-toggle-btn" id="theme-toggle" aria-label="Toggle theme" title="Toggle dark/light theme">
                        <span class="icon-sun">&#9728;</span>
                        <span class="icon-moon">&#9790;</span>
                    </button>
                </div>
            </div>
        `;

        // Insert at the beginning of body
        document.body.insertBefore(nav, document.body.firstChild);

        // Setup mobile menu toggle
        const menuBtn = nav.querySelector('.mobile-menu-btn');
        const navList = nav.querySelector('.navbar-nav');
        menuBtn.addEventListener('click', () => {
            navList.classList.toggle('show');
        });

        // Setup mobile dropdown toggles
        if (this.isMobile) {
            nav.querySelectorAll('.nav-item').forEach(item => {
                const link = item.querySelector('.nav-link');
                if (item.querySelector('.dropdown-menu')) {
                    link.addEventListener('click', (e) => {
                        e.preventDefault();
                        item.classList.toggle('open');
                    });
                }
            });
        }

        // Setup keyboard shortcut for search
        document.addEventListener('keydown', (e) => {
            if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
                e.preventDefault();
                const searchInput = document.getElementById('globalSearch');
                if (searchInput) searchInput.focus();
            }
        });

        // Render breadcrumb after main content loads
        setTimeout(() => this.renderBreadcrumb(), 0);

        // Setup theme toggle
        this.setupThemeToggle(nav);
    }

    setupThemeToggle(nav) {
        const themeToggle = nav.querySelector('#theme-toggle');
        if (!themeToggle) return;

        themeToggle.addEventListener('click', () => {
            const currentTheme = document.documentElement.getAttribute('data-theme');
            const newTheme = currentTheme === 'light' ? 'dark' : 'light';

            if (newTheme === 'dark') {
                document.documentElement.removeAttribute('data-theme');
            } else {
                document.documentElement.setAttribute('data-theme', newTheme);
            }

            // Persist theme preference
            localStorage.setItem('theme', newTheme);
        });
    }

    renderBreadcrumb() {
        const pageInfo = this.getCurrentPageInfo();
        const main = document.querySelector('main');
        if (!main) return;

        // Don't show breadcrumb on dashboard
        if (!pageInfo.category && pageInfo.page === 'Dashboard') return;

        const breadcrumb = document.createElement('nav');
        breadcrumb.className = 'breadcrumb';
        breadcrumb.setAttribute('aria-label', 'Breadcrumb');

        let html = `<a href="${this.resolvePath('index.html')}">Dashboard</a>`;

        if (pageInfo.category) {
            html += `<span class="breadcrumb-separator">/</span>`;
            html += `<span class="breadcrumb-category">${pageInfo.category}</span>`;
        }

        html += `<span class="breadcrumb-separator">/</span>`;
        html += `<span class="breadcrumb-current">${pageInfo.page}</span>`;

        breadcrumb.innerHTML = html;

        // Insert before page header
        const pageHeader = main.querySelector('.page-header');
        if (pageHeader) {
            main.insertBefore(breadcrumb, pageHeader);
        } else {
            main.insertBefore(breadcrumb, main.firstChild);
        }
    }

    renderNavItem(item) {
        if (item.items) {
            // Dropdown menu
            const isActive = this.isDropdownActive(item.items);
            return `
                <li class="nav-item">
                    <a href="#" class="nav-link ${isActive ? 'active' : ''}">
                        ${item.title}
                        <span class="arrow">&#9662;</span>
                    </a>
                    <div class="dropdown-menu">
                        ${item.items.map(subItem => `
                            <a href="${this.resolvePath(subItem.href)}"
                               class="dropdown-item ${this.isActive(subItem.href) ? 'active' : ''}">
                                ${subItem.title}
                            </a>
                        `).join('')}
                    </div>
                </li>
            `;
        } else {
            // Simple link
            return `
                <li class="nav-item">
                    <a href="${this.resolvePath(item.href)}"
                       class="nav-link ${this.isActive(item.href) ? 'active' : ''}">
                        ${item.title}
                    </a>
                </li>
            `;
        }
    }
}

// Initialize theme from localStorage before DOM loads to prevent flash
(function() {
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme === 'light') {
        document.documentElement.setAttribute('data-theme', 'light');
    }
})();

// Initialize navigation when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    const nav = new Navigation();
    nav.render();
});

// Update stats cards helper
function updateStats(stats) {
    const totalEl = document.getElementById('stat-total');
    const activeEl = document.getElementById('stat-active');
    const removedEl = document.getElementById('stat-removed');

    if (totalEl) totalEl.textContent = stats.total || 0;
    if (activeEl) activeEl.textContent = stats.active || 0;
    if (removedEl) removedEl.textContent = stats.removed || 0;
}

// Format relative time (e.g., "2 hours ago", "3 days ago")
function formatRelativeTime(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins} min ago`;
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    return date.toLocaleDateString();
}

// Get freshness status based on schedule and last updated time
function getFreshnessStatus(lastUpdated, schedule) {
    const date = new Date(lastUpdated);
    const now = new Date();
    const diffHours = (now - date) / 3600000;

    // Define staleness thresholds based on schedule
    const thresholds = {
        'hourly': { fresh: 2, stale: 4 },
        'daily': { fresh: 26, stale: 50 },
        'weekly': { fresh: 168, stale: 336 }
    };

    const threshold = thresholds[schedule] || thresholds['daily'];

    if (diffHours <= threshold.fresh) return 'fresh';
    if (diffHours <= threshold.stale) return 'stale';
    return 'old';
}

// Render last updated indicator
function renderLastUpdated(lastUpdated, schedule) {
    const status = getFreshnessStatus(lastUpdated, schedule);
    const relativeTime = formatRelativeTime(lastUpdated);

    return `
        <span class="last-updated">
            <span class="last-updated-dot ${status}"></span>
            Updated ${relativeTime}
        </span>
    `;
}

// Load and display last updated info for a service
async function loadLastUpdated(provider, service = null) {
    try {
        const basePath = window.location.pathname.includes('/') &&
            !window.location.pathname.endsWith('index.html') ? '../' : '';
        const response = await fetch(basePath + 'data/metadata.json');
        if (!response.ok) return null;

        const metadata = await response.json();
        const providerData = metadata.services[provider];

        if (!providerData) return null;

        if (service && providerData.services && providerData.services[service]) {
            return {
                lastUpdated: providerData.services[service].last_updated,
                schedule: providerData.schedule
            };
        }

        return {
            lastUpdated: providerData.last_updated,
            schedule: providerData.schedule
        };
    } catch (e) {
        console.warn('Could not load metadata:', e);
        return null;
    }
}

// Display last updated in page header
async function displayLastUpdated(provider, service = null) {
    const data = await loadLastUpdated(provider, service);
    if (!data) return;

    const pageHeader = document.querySelector('.page-header');
    if (!pageHeader) return;

    // Check if already exists
    if (pageHeader.querySelector('.last-updated')) return;

    const indicator = document.createElement('div');
    indicator.style.marginTop = '0.75rem';
    indicator.innerHTML = renderLastUpdated(data.lastUpdated, data.schedule);
    pageHeader.appendChild(indicator);
}

// Export for use in modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        Navigation,
        NAV_CONFIG,
        updateStats,
        formatRelativeTime,
        getFreshnessStatus,
        renderLastUpdated,
        loadLastUpdated,
        displayLastUpdated
    };
}
