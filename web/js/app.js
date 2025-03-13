let userId = null;
let userName = null;

const sidebar = document.querySelector('.sidebar');
const settingList = document.getElementById('setting-list');
const userList = document.getElementById('user-list');
const groupList = document.getElementById('group-list');
const mainContent = document.getElementById('main-content');

async function fetchUserInfo() {
    try {
        const response = await fetch('http://127.0.0.1:8080/api/v1/user/info/me', {
            method: 'GET',
            credentials: 'include'
        });
        const result = await response.json();

        if (response.ok) {

            userId = result.id;
            userName = result.name;
            console.log('User ID:', userId);
            console.log('User Name:', userName);
        } else {
            console.error('Failed to fetch user info:', result);
        }
    } catch (error) {
        console.error('Error fetching data:', error);
    }
}

async function fetchData(url, options = {}) {
    const response = await fetch(url, options);
    return response.json();
}

async function init() {
    try {
        await fetchUserInfo();

        const [userData, groupData] = await Promise.all([
            fetchData('http://127.0.0.1:8080/api/v1/user/lists', { credentials: 'include' }),
            fetchData('http://127.0.0.1:8080/api/v1/group/lists', { credentials: 'include' })
        ]);

        console.log("userData:", userData);
        console.log("groupData:", groupData);

        renderList('user-list', userData.users, 'user');
        renderList('group-list', groupData.groups, 'group');
    } catch (error) {
        console.error('Error fetching data:', error);
    }
}

window.addEventListener('load', init);

sidebar.addEventListener('click', (event) => {
    const target = event.target.closest('.top div');
    if (!target) return;
    event.preventDefault();
    document.querySelectorAll('.top div').forEach(link => link.classList.remove('active'));

    target.classList.add('active');

    const name = target.getAttribute('data-name');

    groupList.style.display = name === 'groups' ? 'block' : 'none';
    userList.style.display = name === 'users' ? 'block' : 'none';
    settingList.style.display = name === 'setting' ? 'block' : 'none';
});

const groupLink = document.querySelector('.top div[data-name="setting"]');
if (groupLink) {
    groupLink.click();
}

settingList.addEventListener('click', (event) => {
    const item = event.target.closest('.setting-item');
    if (!item) return;

    clearSelectedItems();
    item.classList.add('selected');

    mainContent.classList.add('center');
    mainContent.innerHTML = '';

    const action = item.getAttribute('data-action');
    switch (action) {
        case 'whoami':
            mainContent.appendChild(genWhoamiPage());
            break;
        case 'rename':
            mainContent.appendChild(genRenamePage());
            renameListener();
            break;
        case 'repassword':
            mainContent.appendChild(genRepasswordPage());
            repasswordListener();
            break;
        case 'createGroup':
            mainContent.appendChild(genCreateGroupPage());
            createGroupListener();
            break;
        case 'joinGroup':
            mainContent.appendChild(genJoinGroupPage());
            joinGroupListener();
            break;
        case 'logout':
            mainContent.appendChild(genLogoutPage());
            logoutListener();
            break;
        case 'about':
            mainContent.appendChild(genAboutPage());
            break;
    }
});

userList.addEventListener('click', (event) => {
    handleItemClick(event, '.user-item');
});

groupList.addEventListener('click', (event) => {
    handleItemClick(event, '.group-item');
});

function genChatInput() {
    const form = document.createElement('div');
    form.className = 'chat-input';
    form.innerHTML = `
            <input type="file" id="file-input"/>
            <label for="file-input" class="file-input-button">
                <i class="fa-solid fa-file-arrow-up"></i>
            </label>
        <div id="text-input" contenteditable="true" data-placeholder="输入内容..."></div>`;
    return form;
}

function bindGroupOptionEvents(uid) {
    const groupOption = document.getElementById('group-option');
    const exitGroup = document.getElementById('exit-group');

    groupOption.addEventListener('click', function (event) {
        event.stopPropagation();

        exitGroup.classList.add('show');
        groupOption.classList.add('move-left');
    });

    exitGroup.addEventListener('click', async () => {
        if (confirm("你确定要退出群组吗？") == true) {
            showLoading();
            try {
                const response = await fetch(`http://127.0.0.1:8080/api/v1/group/leave/${uid}`, {
                    method: 'GET',
                    credentials: 'include',
                });
                const data = await response.json();
                hideLoading();
                if (response.ok) {
                    alert("退出群组成功");
                    window.location.reload();
                } else {
                    alert(data.error);
                }
            } catch (error) { }
        }
    });

    document.addEventListener('click', function () {

        exitGroup.classList.remove('show');
        groupOption.classList.remove('move-left');
    });
}

