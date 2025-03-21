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
    const overlay = document.createElement('div');
    overlay.className = 'sidebar-overlay';

    const buttonContainer = document.createElement('div');
    buttonContainer.className = 'group-actions';

    const exitButton = createActionButton('退出群组', () => handleExitGroup(uid));
    const membersButton = createActionButton('查看群组成员', () => handleViewMembers(uid));

    buttonContainer.appendChild(exitButton);
    buttonContainer.appendChild(membersButton);
    overlay.appendChild(buttonContainer);

    overlay.addEventListener('click', (e) => {
        e.stopPropagation();
    });

    groupOption.addEventListener('click', (e) => {
        e.stopPropagation();
        document.body.appendChild(overlay);
        requestAnimationFrame(() => {
            overlay.style.right = "0px";
        });
    });

    document.addEventListener('click', function () {
        if (overlay.parentNode) {
            overlay.style.right = "-300px";
            overlay.addEventListener('transitionend', () => {
                document.body.removeChild(overlay);
            }, { once: true });
        }
    });
}

function createActionButton(text, handler) {
    const button = document.createElement('button');
    button.className = 'group-action-btn';
    button.textContent = text;
    button.addEventListener('click', handler);
    return button;
}

async function handleExitGroup(uid) {
    if (!confirm("你确定要退出群组吗？")) return;

    showLoading();
    try {
        const response = await fetch(`http://127.0.0.1:8080/api/v1/group/leave/${uid}`, {
            method: 'GET',
            credentials: 'include',
        });
        const result = await response.json();
        if (response.ok) {
            alert("退出群组成功");
            window.location.reload();
        } else {
            console.error('Failed to leave group:', result);
        }
    } catch (error) {
        console.error('Error fetching data:', error);
    } finally {
        hideLoading();
    }
}

async function handleViewMembers(uid) {
    try {
        const response = await fetch(`http://127.0.0.1:8080/api/v1/group/member/${uid}`, {
            method: 'GET',
            credentials: 'include'
        });

        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);

        const data = await response.json();
        console.log('API response:', data);

        const overlay = document.querySelector('.sidebar-overlay');
        overlay.innerHTML = '';

        const groupOwner = parseInt(document.querySelector('.selected')?.dataset.owner);
        console.log('owner id:', groupOwner);

        const memberContainer = document.createElement('div');
        memberContainer.className = 'member-container';

        if (data?.users && data.users.length > 0) {
            data.users.forEach(user => {
                const span = document.createElement('span');
                span.setAttribute('uid', user.id);
                span.className = 'member-item';
                span.innerHTML = `${user.name}`;

                span.addEventListener('contextmenu', (e) => {
                    console.log('右键事件触发');
                    if (userId != groupOwner) {
                        console.log('user', userId);
                        console.log("非群主");
                        return;
                    }
                    e.preventDefault();
                    showMemberContextMenu(e, user.id, uid);
                });

                memberContainer.appendChild(span);
            });
        }

        overlay.appendChild(memberContainer);
    } catch (error) {
        console.error('Error:', error);
    }
}

let currentMenu = null;
function showMemberContextMenu(event, memberId, groupId) {

    if (currentMenu) currentMenu.remove();

    const menu = document.createElement('div');
    menu.className = 'member-context-menu';
    menu.style.cssText = `
        left: ${event.clientX}px;
        top: ${event.clientY}px;
    `;

    const removeItem = document.createElement('div');
    removeItem.className = 'menu-item';
    removeItem.innerHTML = '移出群组';
    removeItem.onclick = async () => {
        try {

            const isConfirm = confirm("确定要将该成员移出群组吗？");
            if (!isConfirm) return;
            const response = await fetch(`http://127.0.0.1:8080/api/v1/group/remove/${groupId}/${memberId}`, {
                credentials: 'include'
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error);
            }

            const removedElement = document.querySelector(`span[uid="${memberId}"]`);
            if (removedElement) {
                removedElement.remove();
                alert('已成功移出成员');
            }
        } catch (error) {
            alert(`${error}`);
        } finally {
            if (menu.parentNode) {
                menu.remove();
            }
            currentMenu = null;
        }
    };

    menu.appendChild(removeItem);
    document.body.appendChild(menu);
    currentMenu = menu;

    const clickHandler = (e) => {
        if (!menu.contains(e.target)) {
            menu.remove();
            document.removeEventListener('click', clickHandler);
            currentMenu = null;
        }
    };
    document.addEventListener('click', clickHandler);
}
