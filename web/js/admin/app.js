let currentView = 'users';
let originalData = [];
let currentData = [];

const searchArea = document.getElementById('searchInput');

async function fetchUsers() {
    try {
        const response = await fetch('http://127.0.0.1:8080/admin/show_users');
        if (!response.ok) throw new Error(`HTTP错误 ${response.status}`);
        const { users } = await response.json();
        return users;
    } catch (error) {
        showError('用户数据加载失败');
        return [];
    }
}

async function fetchGroups() {
    try {
        const response = await fetch('http://127.0.0.1:8080/admin/show_groups');
        if (!response.ok) throw new Error(`HTTP错误 ${response.status}`);
        const { groups } = await response.json();
        return groups;
    } catch (error) {
        showError('群组数据加载失败');
        return [];
    }
}

async function switchView(viewType) {
    currentView = viewType;
    document.querySelectorAll('.view-tab').forEach(tab =>
        tab.classList.toggle('active', tab.textContent === `${viewType === 'users' ? '用户' : '群组'}列表`)
    );

    document.getElementById('searchInput').value = '';

    try {
        originalData = await (viewType === 'users' ? fetchUsers() : fetchGroups());
        currentData = [...originalData];
        renderTable();
    } catch (error) {
        console.error(error);
        showError(`数据加载失败: ${error.message}`);
    }
}

function renderTable() {
    const header = document.getElementById('tableHeader');
    const tbody = document.getElementById('dataList');

    header.innerHTML = currentView === 'users' ? `
        <tr>
            <th>用户ID</th>
            <th>用户名称</th>
            <th>操作</th>
        </tr>
    ` : `
        <tr>
            <th>群组ID</th>
            <th>群组名称</th>
            <th>群主ID</th>
            <th>操作</th>
        </tr>
    `;

    if (currentData.length === 0) {
        tbody.innerHTML = `<tr><td colspan="4" style="text-align:center;color:#64748b;">暂无数据</td></tr>`;
        return;
    }

    tbody.innerHTML = currentData.map(item => {
        if (currentView === 'users') {
            return `
        <tr>
            <td class="user-id">${item.id}</td>
            <td>${item.name}</td>
            <td>
                <div class="action-buttons">
                    <button class="btn btn-danger" onclick="handleDelete('${item.id}')">删除</button>
                    <button class="btn btn-primary" onclick="handleResetPassword('${item.id}')">重置密码</button>
                </div>
            </td>
        </tr>
    `;
        }
        return `
    <tr>
        <td class="user-id">${item.id}</td>
        <td>${item.name}</td>
        <td>${item.owner_id}</td>
        <td>
            <div class="action-buttons" style="display: inline-block; margin-left: 1rem;">
                <button class="btn btn-danger" onclick="handleGroupDelete('${item.id}')">删除群组</button>
            </div>
        </td>
    </tr>
`;
    }).join('');
}

async function handleDelete(userId) {
    if (confirm(`确定要删除用户 ${userId} 吗？`)) {
        try {
            const response = await fetch('http://127.0.0.1:8080/admin/delete_user', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ id: userId })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || '删除失败');
            }

            alert(`用户 ${userId} 已删除`);
            await switchView('users');
        } catch (error) {
            alert(error.message);
        }
    }
}

async function handleGroupDelete(groupId) {
    if (confirm(`确定要删除群组 ${groupId} 吗？该操作不可恢复！`)) {
        try {
            const response = await fetch('http://127.0.0.1:8080/admin/delete_group', {
                method: 'POST',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ id: groupId })
            });

            if (!response.ok) {
                const data = await response.json();
                throw new Error(data.error || '删除群组失败');
            }

            alert(`群组 ${groupId} 已删除`);
            await switchView('groups');
        } catch (error) {
            alert(error.message);
        }
    }
}

function handleResetPassword(userId) {
    if (confirm(`确定要重置用户 ${userId} 的密码吗？`)) {
        fetch('http://127.0.0.1:8080/admin/forget_password', {
            method: 'POST',
            credentials: 'include',
            body: JSON.stringify({ id: userId })
        })
            .then(response => {
                hideLoading();
                if (!response.ok) {
                    response.json().then(data => {
                        alert(data.error);
                    });
                }
            });

        alert(`用户 ${userId} 的密码已重置`);

        switchView('users');
    }
}

function initSearch() {
    searchArea.addEventListener('input', function (e) {
        const keyword = e.target.value.trim().toLowerCase();
        currentData = originalData.filter(item => {
            const fields = currentView === 'users'
                ? [item.id.toString(), item.name]
                : [item.id.toString(), item.name, item.owner_id?.toString()];
            return fields.some(field => field?.toLowerCase().includes(keyword));
        });
        renderTable();
    });
}

function showError(message) {
    const tbody = document.getElementById('dataList');
    tbody.innerHTML = `<tr><td colspan="3" style="text-align:center;color:#dc2626;">${message}</td></tr>`;
}

(async function init() {
    await switchView('users');
    initSearch();
})();