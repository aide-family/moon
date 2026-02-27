window.addEventListener('DOMContentLoaded', () => {
    const docEl = document.getElementById('thedoc');

    // 存储当前选中的 headers（保留变量以兼容后续扩展）
    let currentSelectedHeaders = [];

    docEl.addEventListener('before-try', (e) => {
        // 获取当前选中的 Header 组
        const select = document.getElementById('headerSetSelect');
        const selectedHeaderSetId = select ? select.value : localStorage.getItem('selectedHeaderSet');

        if (selectedHeaderSetId && currentSelectedHeaders.length > 0) {
            const systemHeaders = ['content-type', 'accept', 'user-agent'];
            currentSelectedHeaders.forEach(header => {
                if (header.key && header.value) {
                    const key = header.key.trim().toLowerCase();
                    if (!systemHeaders.includes(key)) {
                        e.detail.request.headers.set(header.key.trim(), header.value.trim());
                    }
                }
            });
        } else if (selectedHeaderSetId) {
            // 备用方案：从 localStorage 读取
            const headerSets = JSON.parse(localStorage.getItem('headerSets') || '[]');
            const headerSet = headerSets.find(set => set.id === selectedHeaderSetId);
            if (headerSet && headerSet.headers) {
                const systemHeaders = ['content-type', 'accept', 'user-agent'];
                headerSet.headers.forEach(header => {
                    if (header.key && header.value) {
                        const key = header.key.trim().toLowerCase();
                        if (!systemHeaders.includes(key)) {
                            e.detail.request.headers.set(header.key.trim(), header.value.trim());
                        }
                    }
                });
            }
        }

        // 兼容旧的 JWT 存储方式
        const jwt = localStorage.getItem('jwt');
        if (jwt && !e.detail.request.headers.has('Authorization')) {
            e.detail.request.headers.set('Authorization', 'Bearer ' + jwt);
        }
    });
});

let currentHeaderSetId = null;
let currentHeaders = [];
let editingHeaderIndex = -1;

// 初始化
window.addEventListener('DOMContentLoaded', () => {
    createTestHeaderSet();
    loadHeaderSetList();
    const selectedSet = localStorage.getItem('selectedHeaderSet');
    if (selectedSet) {
        document.getElementById('headerSetSelect').value = selectedSet;
        loadHeaderSet();
    }
    updateHeaderStats();
});

function createTestHeaderSet() {
    const headerSets = JSON.parse(localStorage.getItem('headerSets') || '[]');
    const testSetExists = headerSets.some(set => set.name === '测试组');
    if (!testSetExists) {
        const testSet = {
            id: 'test_set_' + Date.now(),
            name: '测试组',
            headers: [
                { key: 'Authorization', value: 'Bearer test-token-123' },
                { key: 'X-Namespace', value: 'default' },
            ],
            createdAt: new Date().toISOString()
        };
        headerSets.push(testSet);
        localStorage.setItem('headerSets', JSON.stringify(headerSets));
        console.log('Created test header set:', testSet);
    }
}

function toggleHeaderSection() {
    const section = document.getElementById('headerSection');
    section.style.display = section.style.display === 'none' ? 'flex' : 'none';
}

function updateHeaderStats() {
    const stats = document.getElementById('headerStats');
    const headerSets = JSON.parse(localStorage.getItem('headerSets') || '[]');
    const currentSet = headerSets.find(set => set.id === currentHeaderSetId);
    const setName = currentSet ? currentSet.name : '未选择';
    const headerCount = currentHeaders.length;
    stats.textContent = `当前组: ${setName} | Headers: ${headerCount}`;
}

function loadHeaderSetList() {
    const headerSets = JSON.parse(localStorage.getItem('headerSets') || '[]');
    const select = document.getElementById('headerSetSelect');
    select.innerHTML = '<option value="">-- 选择 Header 组 --</option>';
    headerSets.forEach(set => {
        const option = document.createElement('option');
        option.value = set.id;
        option.textContent = `${set.name} (${set.headers.length} headers)`;
        select.appendChild(option);
    });
}

function loadHeaderSet() {
    const select = document.getElementById('headerSetSelect');
    const setId = select.value;
    currentHeaderSetId = setId;
    if (!setId) {
        currentHeaders = [];
        renderHeaders();
        document.getElementById('editSetForm').style.display = 'none';
        updateHeaderStats();
        localStorage.removeItem('selectedHeaderSet');
        return;
    }
    const headerSets = JSON.parse(localStorage.getItem('headerSets') || '[]');
    const headerSet = headerSets.find(set => set.id === setId);
    if (headerSet) {
        currentHeaders = [...headerSet.headers];
        renderHeaders();
        document.getElementById('editSetForm').style.display = 'flex';
        localStorage.setItem('selectedHeaderSet', setId);
        updateHeaderStats();
    } else {
        select.value = '';
        currentHeaderSetId = null;
        currentHeaders = [];
        renderHeaders();
        document.getElementById('editSetForm').style.display = 'none';
        updateHeaderStats();
        localStorage.removeItem('selectedHeaderSet');
    }
}

function showNewHeaderSet() {
    document.getElementById('newSetForm').classList.add('active');
    document.getElementById('editSetForm').style.display = 'none';
    document.getElementById('newSetName').value = '';
    document.getElementById('headerSetSelect').value = '';
    currentHeaderSetId = null;
    currentHeaders = [];
    renderHeaders();
    updateHeaderStats();
}

function createNewHeaderSet() {
    const name = document.getElementById('newSetName').value.trim();
    if (!name) { alert('请输入组名称'); return; }
    const headerSets = JSON.parse(localStorage.getItem('headerSets') || '[]');
    const newId = 'headerSet_' + Date.now();
    const newSet = { id: newId, name: name, headers: [], createdAt: new Date().toISOString() };
    headerSets.push(newSet);
    localStorage.setItem('headerSets', JSON.stringify(headerSets));
    document.getElementById('newSetForm').classList.remove('active');
    document.getElementById('editSetForm').style.display = 'flex';
    document.getElementById('headerSetSelect').value = newId;
    currentHeaderSetId = newId;
    currentHeaders = [];
    renderHeaders();
    loadHeaderSetList();
    updateHeaderStats();
    alert('创建成功');
}

function addHeader() {
    const key = document.getElementById('headerKey').value.trim();
    const value = document.getElementById('headerValue').value.trim();
    if (!key) { alert('请输入 Header Key'); return; }
    if (editingHeaderIndex >= 0) { currentHeaders[editingHeaderIndex] = { key, value }; editingHeaderIndex = -1; }
    else { currentHeaders.push({ key, value }); }
    document.getElementById('headerKey').value = '';
    document.getElementById('headerValue').value = '';
    renderHeaders();
    updateHeaderStats();
}

function editHeader(index) {
    const header = currentHeaders[index];
    document.getElementById('headerKey').value = header.key;
    document.getElementById('headerValue').value = header.value;
    editingHeaderIndex = index;
}

function deleteHeader(index) {
    if (confirm('确定要删除这个 Header 吗？')) {
        currentHeaders.splice(index, 1);
        renderHeaders();
        updateHeaderStats();
    }
}

function renderHeaders() {
    const container = document.getElementById('headerItems');
    container.innerHTML = '';
    if (currentHeaders.length === 0) {
        container.innerHTML = '<div style="padding: 20px; text-align: center; color: #999;">暂无 Header，请添加</div>';
        return;
    }
    currentHeaders.forEach((header, index) => {
        const item = document.createElement('div');
        item.className = 'header-item';
        item.innerHTML = `
			<div class="header-item-key">${escapeHtml(header.key)}</div>
			<div class="header-item-value">${escapeHtml(header.value)}</div>
			<div class="header-item-actions">
				<button class="btn-small btn-edit" onclick="editHeader(${index})">编辑</button>
				<button class="btn-small btn-delete" onclick="deleteHeader(${index})">删除</button>
			</div>
		`;
        container.appendChild(item);
    });
}

function saveHeaderSet() {
    if (!currentHeaderSetId) { alert('请先选择一个 Header 组或创建新组'); return; }
    const headerSets = JSON.parse(localStorage.getItem('headerSets') || '[]');
    const headerSetIndex = headerSets.findIndex(set => set.id === currentHeaderSetId);
    if (headerSetIndex >= 0) {
        headerSets[headerSetIndex].headers = currentHeaders;
        headerSets[headerSetIndex].updatedAt = new Date().toISOString();
        localStorage.setItem('headerSets', JSON.stringify(headerSets));
        localStorage.setItem('selectedHeaderSet', currentHeaderSetId);
        loadHeaderSetList();
        updateHeaderStats();
        alert('保存成功');
    }
}

function deleteHeaderSet() {
    const select = document.getElementById('headerSetSelect');
    const setId = select.value;
    if (!setId) { alert('请先选择一个 Header 组'); return; }
    if (!confirm('确定要删除这个 Header 组吗？')) { return; }
    const headerSets = JSON.parse(localStorage.getItem('headerSets') || '[]');
    const filteredSets = headerSets.filter(set => set.id !== setId);
    localStorage.setItem('headerSets', JSON.stringify(filteredSets));
    if (localStorage.getItem('selectedHeaderSet') === setId) { localStorage.removeItem('selectedHeaderSet'); }
    select.value = '';
    currentHeaderSetId = null;
    currentHeaders = [];
    renderHeaders();
    document.getElementById('editSetForm').style.display = 'none';
    loadHeaderSetList();
    updateHeaderStats();
    alert('删除成功');
}

function exportHeaderSets() {
    const headerSets = JSON.parse(localStorage.getItem('headerSets') || '[]');
    const exportData = { version: '1.0', exportTime: new Date().toISOString(), headerSets };
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `swagger-headers-${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}

function importHeaderSets(event) {
    const file = event.target.files[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = function (e) {
        try {
            const importData = JSON.parse(e.target.result);
            if (importData.headerSets && Array.isArray(importData.headerSets)) {
                const action = confirm('是否覆盖现有配置？\n点击"确定"覆盖，点击"取消"合并');
                let finalSets;
                if (action) { finalSets = importData.headerSets; }
                else {
                    const existingSets = JSON.parse(localStorage.getItem('headerSets') || '[]');
                    finalSets = [...existingSets, ...importData.headerSets];
                }
                localStorage.setItem('headerSets', JSON.stringify(finalSets));
                loadHeaderSetList();
                if (finalSets.length > 0) {
                    document.getElementById('headerSetSelect').value = finalSets[0].id;
                    loadHeaderSet();
                }
                alert(`成功导入 ${importData.headerSets.length} 个 Header 组`);
            } else {
                alert('无效的配置文件格式');
            }
        } catch (error) {
            alert('配置文件解析失败: ' + error.message);
        }
    };
    reader.readAsText(file);
    event.target.value = '';
}

function escapeHtml(text) {
    const map = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;' };
    return text.replace(/[&<>"']/g, m => map[m]);
}

window.addEventListener('DOMContentLoaded', () => {
    const keyInput = document.getElementById('headerKey');
    const valueInput = document.getElementById('headerValue');
    if (keyInput && valueInput) {
        valueInput.addEventListener('keypress', (e) => { if (e.key === 'Enter') { e.preventDefault(); addHeader(); } });
        keyInput.addEventListener('keypress', (e) => { if (e.key === 'Enter') { e.preventDefault(); valueInput.focus(); } });
    }
});
