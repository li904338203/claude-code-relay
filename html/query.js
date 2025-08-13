// API配置
const API_BASE_URL = window.location.origin; // 使用当前域名
const API_ENDPOINTS = {
    GET_API_KEY_INFO: '/api/v1/auth/api-key' // 根据路由配置确定的接口路径
};

// 全局变量
let currentApiKey = '';
let currentPage = 1;
let totalPages = 1;
let isLoading = false;

// DOM元素
const apiKeyInput = document.getElementById('apiKey');
const queryBtn = document.getElementById('queryBtn');
const queryBtnText = document.getElementById('queryBtnText');
const errorMessage = document.getElementById('errorMessage');
const resultsSection = document.getElementById('resultsSection');
const logsContainer = document.getElementById('logsContainer');
const pagination = document.getElementById('pagination');
const prevBtn = document.getElementById('prevBtn');
const nextBtn = document.getElementById('nextBtn');
const pageInfo = document.getElementById('pageInfo');

// 统计元素
const totalCalls = document.getElementById('totalCalls');
const totalCost = document.getElementById('totalCost');
const totalTokens = document.getElementById('totalTokens');
const avgCost = document.getElementById('avgCost');

// 初始化事件监听器
function initEventListeners() {
    queryBtn.addEventListener('click', handleQuery);
    apiKeyInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            handleQuery();
        }
    });
    
    apiKeyInput.addEventListener('input', function() {
        hideError();
        if (apiKeyInput.classList.contains('error')) {
            apiKeyInput.classList.remove('error');
        }
    });
    
    prevBtn.addEventListener('click', () => {
        if (currentPage > 1) {
            currentPage--;
            loadLogs();
        }
    });
    
    nextBtn.addEventListener('click', () => {
        if (currentPage < totalPages) {
            currentPage++;
            loadLogs();
        }
    });
}

// 验证API Key格式
function validateApiKey(apiKey) {
    if (!apiKey || apiKey.trim() === '') {
        return '请输入API Key';
    }
    
    if (!apiKey.startsWith('sk-')) {
        return 'API Key格式错误，应以"sk-"开头';
    }
    
    if (apiKey.length < 10) {
        return 'API Key长度过短';
    }
    
    return null;
}

// 显示错误信息
function showError(message) {
    errorMessage.textContent = message;
    errorMessage.style.display = 'block';
    apiKeyInput.classList.add('error');
}

// 隐藏错误信息
function hideError() {
    errorMessage.style.display = 'none';
    apiKeyInput.classList.remove('error');
}

// 设置加载状态
function setLoading(loading) {
    isLoading = loading;
    queryBtn.disabled = loading;
    
    if (loading) {
        queryBtnText.textContent = '查询中...';
        queryBtn.innerHTML = '<div class="loading-spinner"></div><span>查询中...</span>';
    } else {
        queryBtnText.textContent = '查询使用量';
        queryBtn.innerHTML = '<span>查询使用量</span>';
    }
}

// 处理查询请求
async function handleQuery() {
    if (isLoading) return;
    
    const apiKey = apiKeyInput.value.trim();
    const validationError = validateApiKey(apiKey);
    
    if (validationError) {
        showError(validationError);
        return;
    }
    
    hideError();
    currentApiKey = apiKey;
    currentPage = 1;
    
    try {
        setLoading(true);
        await loadApiKeyInfo();
    } catch (error) {
        console.error('查询失败:', error);
        showError(error.message || '查询失败，请稍后重试');
    } finally {
        setLoading(false);
    }
}

// 加载API Key信息
async function loadApiKeyInfo() {
    const url = `${API_BASE_URL}${API_ENDPOINTS.GET_API_KEY_INFO}/${encodeURIComponent(currentApiKey)}?page=${currentPage}&limit=20`;
    
    const response = await fetch(url, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        }
    });
    
    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    
    if (data.code !== 200 && data.code !== 0) {
        throw new Error(data.error || '查询失败');
    }
    
    displayResults(data.data);
}

// 显示查询结果
function displayResults(data) {
    // 显示结果区域
    resultsSection.style.display = 'block';
    
    // 更新统计数据
    updateStatsOverview(data.stats);
    
    // 显示日志
    displayLogs(data.logs, data.total);
    
    // 滚动到结果区域
    resultsSection.scrollIntoView({ behavior: 'smooth' });
}

// 更新统计概览
function updateStatsOverview(stats) {
    if (!stats) {
        totalCalls.textContent = '0';
        totalCost.textContent = '$0.00';
        totalTokens.textContent = '0';
        avgCost.textContent = '$0.00';
        return;
    }
    
    // 计算总token数
    const totalTokenCount = (stats.total_input_tokens || 0) + (stats.total_output_tokens || 0);
    
    // 计算平均单次费用
    const avgCostValue = stats.total_requests > 0 ? stats.total_cost / stats.total_requests : 0;
    
    totalCalls.textContent = formatNumber(stats.total_requests || 0);
    totalCost.textContent = `$${(stats.total_cost || 0).toFixed(4)}`;
    totalTokens.textContent = formatNumber(totalTokenCount);
    avgCost.textContent = `$${avgCostValue.toFixed(4)}`;
}

// 显示日志列表
function displayLogs(logs, total) {
    if (!logs || logs.length === 0) {
        logsContainer.innerHTML = '<div class="empty-state">暂无调用日志数据</div>';
        pagination.style.display = 'none';
        return;
    }
    
    // 创建表格
    const table = document.createElement('table');
    table.className = 'logs-table';
    
    // 表头
    table.innerHTML = `
        <thead>
            <tr>
                <th>时间</th>
                <th>模型</th>
                <th>状态</th>
                <th>Input Token</th>
                <th>Output Token</th>
                <th>费用 (USD)</th>
            </tr>
        </thead>
        <tbody>
            ${logs.map(log => createLogRow(log)).join('')}
        </tbody>
    `;
    
    logsContainer.innerHTML = '';
    logsContainer.appendChild(table);
    
    // 更新分页
    updatePagination(total);
}

// 创建日志行
function createLogRow(log) {
    const statusClass = log.status_code >= 200 && log.status_code < 300 ? 'status-success' : 'status-error';
    const statusText = log.status_code >= 200 && log.status_code < 300 ? '成功' : '失败';
    
    return `
        <tr>
            <td>${formatDateTime(log.created_at)}</td>
            <td><span class="model-tag">${log.model || '-'}</span></td>
            <td><span class="status-badge ${statusClass}">${statusText} (${log.status_code})</span></td>
            <td>${formatNumber(log.prompt_tokens || 0)}</td>
            <td>${formatNumber(log.completion_tokens || 0)}</td>
            <td><span class="cost-value">$${(log.cost || 0).toFixed(4)}</span></td>
        </tr>
    `;
}

// 更新分页
function updatePagination(total) {
    const limit = 20;
    totalPages = Math.ceil(total / limit);
    
    if (totalPages <= 1) {
        pagination.style.display = 'none';
        return;
    }
    
    pagination.style.display = 'flex';
    prevBtn.disabled = currentPage <= 1;
    nextBtn.disabled = currentPage >= totalPages;
    pageInfo.textContent = `第 ${currentPage} 页，共 ${totalPages} 页`;
}

// 加载日志（用于分页）
async function loadLogs() {
    if (!currentApiKey) return;
    
    try {
        setLoading(true);
        await loadApiKeyInfo();
    } catch (error) {
        console.error('加载日志失败:', error);
        showError(error.message || '加载日志失败');
    } finally {
        setLoading(false);
    }
}

// 工具函数

// 格式化数字
function formatNumber(num) {
    if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
}

// 格式化日期时间
function formatDateTime(dateStr) {
    if (!dateStr) return '-';
    
    try {
        const date = new Date(dateStr);
        return date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    } catch (error) {
        return dateStr;
    }
}

// 复制API Key到剪贴板
function copyApiKey() {
    const apiKey = apiKeyInput.value.trim();
    if (apiKey) {
        navigator.clipboard.writeText(apiKey).then(() => {
            // 可以添加复制成功的提示
            console.log('API Key已复制到剪贴板');
        }).catch(err => {
            console.error('复制失败:', err);
        });
    }
}

// 清空表单
function clearForm() {
    apiKeyInput.value = '';
    hideError();
    resultsSection.style.display = 'none';
    currentApiKey = '';
    currentPage = 1;
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    initEventListeners();
    
    // 检查URL参数中是否有API Key
    const urlParams = new URLSearchParams(window.location.search);
    const apiKeyFromUrl = urlParams.get('api_key');
    if (apiKeyFromUrl) {
        apiKeyInput.value = apiKeyFromUrl;
        // 自动查询
        setTimeout(() => {
            handleQuery();
        }, 500);
    }
});

// 导出函数供其他脚本使用
window.QueryAPI = {
    validateApiKey,
    handleQuery,
    clearForm,
    copyApiKey
};