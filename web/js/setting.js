function genWhoamiPage() {
    const form = document.createElement('div');
    form.className = 'setting-whoami';
    form.innerHTML = `
        <h2>个人信息</h2>
        <p>昵称：${userName}</p>
        <p>账号：${userId}</p>
    `;
    return form;
}

function genRenamePage() {
    const form = document.createElement('div');
    form.className = 'setting-rename';
    form.innerHTML = `
        <h2>修改用户名</h2>
        <p>当前名称：${userName}</p>
        <input type="text" id="user-name" placeholder="请输入新的用户名">
        <button id="setting-rename-button">确认</button>
    `;
    return form;
}

function renameListener() {
    const renameButton = document.getElementById('setting-rename-button');
    renameButton.addEventListener('click', async () => {
        const user_name = document.getElementById('user-name').value;
        showLoading();
        try {
            const response = await fetch('http://127.0.0.1:8080/api/v1/user/rename', {
                method: 'POST',
                credentials: 'include',
                body: JSON.stringify({
                    name: user_name,
                }),
            });
            const data = await response.json();
            hideLoading();
            if (response.ok) {
                alert(data.msg);
                window.location.reload();
            } else {
                alert(data.error);
            }
        } catch (e) { }
    });
}

function genRepasswordPage() {
    const form = document.createElement('div');
    form.className = 'setting-repassword';
    form.innerHTML = `
        <h2>修改密码</h2>
        <input type="password" id="user-passowrd-old" placeholder="请输入旧密码">
        <input type="password" id="user-password-new" placeholder="请输入新密码">
        <input type="password" id="user-password-new2" placeholder="请再输入新密码">
        <button id="setting-repassword-button">确认</button>
    `;
    return form;
}

function repasswordListener() {
    const repasswordButton = document.getElementById('setting-repassword-button');
    repasswordButton.addEventListener('click', async () => {
        const old_password = document.getElementById('user-passowrd-old').value;
        const new_password = document.getElementById('user-password-new').value;
        if (new_password !== document.getElementById('user-password-new2').value) {
            alert("两次密码不同");
            return;
        }
        showLoading();
        try {
            const response = await fetch('http://127.0.0.1:8080/api/v1/user/repassword', {
                method: 'POST',
                credentials: 'include',
                body: JSON.stringify({
                    old_password: old_password,
                    new_password: new_password,
                })
            });
            const data = await response.json();
            hideLoading();
            if (response.ok) {
                alert(data.msg);
                window.location.href = "/login";
            } else {
                alert(data.error);
            }
        } catch (e) { }
    });
}

function genCreateGroupPage() {
    const form = document.createElement('div');
    form.className = 'setting-createGroup';
    form.innerHTML = `
        <h2>创建聊天群组</h2>
        <input type="text" id="group-id" placeholder="请输入群组号">
        <input type="text" id="group-name" placeholder="请输入群组名称">
        <button id="setting-createGroup-button">确认</button>
    `;
    return form;
}

function createGroupListener() {
    const createGroupButton = document.getElementById('setting-createGroup-button');
    createGroupButton.addEventListener('click', async () => {
        const group_id = document.getElementById('group-id').value;
        const group_name = document.getElementById('group-name').value;
        showLoading();
        try {
            const response = await fetch('http://127.0.0.1:8080/api/v1/group/create', {
                method: 'POST',
                credentials: 'include',
                body: JSON.stringify({
                    id: parseInt(group_id),
                    name: group_name,
                })
            });
            const data = await response.json();
            hideLoading();
            if (response.ok) {
                alert(data.msg);
                window.location.reload();
            } else {
                alert(data.error);
            }
        } catch (e) { }
    });
}

function genJoinGroupPage() {
    const form = document.createElement('div');
    form.className = 'setting-joinGroup';
    form.innerHTML = `
        <h2>加入聊天群组</h2>
        <input type="text" id="group-id" placeholder="请输入群组号">
        <button id="setting-joinGroup-find-button">查找</button>
        <p id="group-name"></p>
        <p id="stop-text"></p>
        <button id="setting-joinGroup-button">加入群组</button>
    `;
    return form;
}

function joinGroupListener() {
    const findGroupButton = document.getElementById('setting-joinGroup-find-button');
    const joinGroupButton = document.getElementById('setting-joinGroup-button');
    const groupInfo = document.getElementById('group-name');
    const stopText = document.getElementById('stop-text');

    findGroupButton.addEventListener('click', async () => {
        const group_id = parseInt(document.getElementById('group-id').value);
        showLoading();

        try {

            const response = await fetch(`http://127.0.0.1:8080/api/v1/group/info/${group_id}`, {
                method: 'GET',
                credentials: 'include',
            });

            const data = await response.json();
            hideLoading();

            if (response.ok) {
                findGroupButton.style.display = 'none';
                console.log(data);
                groupInfo.innerText = `群组名称：${data.name}`;

                if (data.is_member) {
                    stopText.innerText = `已是该群组成员，无需再次加入`;
                    return;
                }

                joinGroupButton.style.display = 'block';
                joinGroupButton.onclick = () => handleJoinGroup(group_id);
            } else {
                alert(data.error);
            }
        } catch (error) {
            hideLoading();
            console.error('请求出错:', error);
            alert('请求失败，请稍后重试');
        }
    });

    async function handleJoinGroup(group_id) {
        showLoading();

        try {
            const response = await fetch(`http://127.0.0.1:8080/api/v1/group/join/${group_id}`, {
                method: 'GET',
                credentials: 'include',
            });

            const data = await response.json();
            hideLoading();

            if (response.ok) {
                alert(data.msg);
                window.location.reload();
            } else {
                alert(data.error);
            }
        } catch (error) {
            hideLoading();
            console.error('请求出错:', error);
            alert('请求失败，请稍后重试');
        }
    }
}

function genLogoutPage() {
    const form = document.createElement('div');
    form.className = 'setting-logout';
    form.innerHTML = `<button id="setting-logout-button">账号登出</button>`;
    return form;
}

function logoutListener() {
    const logoutButton = document.getElementById('setting-logout-button');
    logoutButton.addEventListener('click', async () => {
        showLoading();
        try {
            const response = await fetch('http://127.0.0.1:8080/api/v1/user/logout', {
                method: 'GET',
                credentials: 'include'
            });
            const data = await response.json();
            hideLoading();
            if (data.msg == "退出登陆成功") {
                window.location.href = '/login';
            } else {
                alert(data.error);
            }
        } catch (error) { }
    });
}

function genAboutPage() {
    const form = document.createElement('div');
    form.className = 'setting-about';
    form.innerHTML = `
        <img src="image/logo.png">
        <h1>Momo Web</h1>
        <p>Copyright <i class="fa-regular fa-copyright"></i> 2024-2025 Momo University.
            <br>
            All Rights Reserved.
        </p>
        <p>Maintainer Momo</p>
    `;
    return form;
}
