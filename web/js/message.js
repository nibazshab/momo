let chatSocket = null;

function handleItemClick(event, itemClass) {
    const item = event.target.closest(itemClass);
    if (!item) return;

    clearSelectedItems();
    item.classList.add('selected');
    mainContent.classList.remove('center');
    mainContent.innerHTML = '';

    const uid = item.getAttribute('data-uid');
    const name = item.getAttribute('data-name');
    const isGroupItem = itemClass === '.group-item';
    const owner = parseInt(item.getAttribute('data-owner'));
    const isUserList = itemClass === '.user-item';

    const chatHeader = genChatHeader(uid, name, isGroupItem, owner, isUserList);
    mainContent.appendChild(chatHeader);

    const historyMessages = [];
    mainContent.appendChild(genChatMessage(historyMessages));

    const chatInput = genChatInput();
    mainContent.appendChild(chatInput);

    const fileInput = chatInput.querySelector("#file-input");
    fileInput.addEventListener("change", handleFileUpload);

    chatInputListener();

    if (isGroupItem) {
        const option = document.createElement('span');
        option.id = 'group-option';
        option.className = 'right';
        option.innerHTML = `<i class="fa-solid fa-bars"></i>`;
        const exitGroup = document.createElement('span');
        exitGroup.id = "exit-group";
        exitGroup.innerHTML = `<i class="fa-solid fa-arrow-right-from-bracket"></i>`;
        chatHeader.appendChild(option);
        chatHeader.appendChild(exitGroup);
        bindGroupOptionEvents(uid);
    }

    if (isUserList) {

        requestUserConvId(parseInt(uid));
    } else if (isGroupItem) {

        setupChatWebSocket(uid);
    }
}


function requestUserConvId(targetId) {
    sendMessage({ target_id: targetId });
}

async function handleFileUpload(event) {
    const file = event.target.files[0];
    if (file) {
        console.log("Selected file:", file.name);
        const formData = new FormData();
        formData.append("file", file);
        try {
            const response = await fetch('http://127.0.0.1:8080/api/v1/upload', {
                method: 'POST',
                credentials: 'include',
                body: formData,
            });
            const data = await response.json();
            hideLoading();
            if (response.ok) {
                console.log(data.uuid, data.type);
                if (chatSocket && chatSocket.readyState === WebSocket.OPEN) {
                    chatSocket.send(JSON.stringify({
                        text: `${data.uuid}.${data.size}kb.${data.name}`,
                        type: data.type,
                    }));
                } else {
                    console.error('WebSocket 连接未建立或已关闭');
                }
            } else {
                alert(data.error);
            }
        } catch (error) {
            console.error("文件上传错误:", error);
        }
    }
}


function setupChatWebSocket(convId) {

    if (chatSocket) {
        chatSocket.close();
    }

    chatSocket = new WebSocket(`ws://127.0.0.1:8080/api/v1/ws/message?conv_id=${convId}`);

    chatSocket.onopen = function () {
        console.log('聊天 WebSocket 连接已打开, conv_id:', convId);
    };

    chatSocket.onmessage = function (event) {
        console.log('收到聊天消息:', event.data);
        const message = JSON.parse(event.data);
        displayChatMessage(message);
    };

    chatSocket.onclose = function () {
        console.log('聊天 WebSocket 连接已关闭, conv_id:', convId);
    };

    chatSocket.onerror = function (error) {
        console.error('聊天 WebSocket 错误:', error);
    };
}

function displayChatMessage(message) {
    const chatMessages = document.querySelector('.chat-messages');
    const messageElement = document.createElement('div');
    messageElement.className = message.user_id === userId ? 'message me' : 'message';

    if (message.type === 0) {
        messageElement.innerHTML = `
            <div class="header">
                <p class="user">${message.user_id === userId ? 'You' : message.user_name}</p>
                <span>${message.time}</span>
            </div>
            <div class="content">
                <p>${message.text}</p>
            </div>
        `;
    } else if (message.type === 1) {
        messageElement.innerHTML = `
            <div class="header">
                <p class="user">${message.user_id === userId ? 'You' : message.user_name}</p><span>${message.time}</span>
            </div>
            <div><img src="http://127.0.0.1:8080/api/v1/files/${message.text.slice(0, 36)}"></div>
        `;
    } else {
        messageElement.innerHTML = `
            <div class="header">
                <p class="user">${message.user_id === userId ? 'You' : message.user_name}</p><span>${message.time}</span>
            </div>
            <div>
                <a target="_blank" href="http://127.0.0.1:8080/api/v1/files/${message.text.slice(0, 36)}" download>${message.text.slice(37)}</a>
            </div>
        `;
    }

    chatMessages.appendChild(messageElement);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

const userConvIdSocket = new WebSocket(`ws://127.0.0.1:8080/api/v1/ws/convid`);
const messageQueue = [];

function sendMessage(data) {
    if (userConvIdSocket.readyState === WebSocket.OPEN) {
        userConvIdSocket.send(JSON.stringify(data));
    } else {
        console.log("WebSocket 连接未建立，消息已缓存");
        messageQueue.push(data);
    }
}

userConvIdSocket.onopen = function (event) {
    console.log("WebSocket 连接已打开");
    while (messageQueue.length > 0) {
        const message = messageQueue.shift();
        userConvIdSocket.send(JSON.stringify(message));
    }
};

userConvIdSocket.onmessage = function (event) {
    console.log("收到服务器消息:", event.data);
    const data = JSON.parse(event.data);
    if (data.conv_id) {

        setupChatWebSocket(data.conv_id);
    }
};

userConvIdSocket.onclose = function (event) {
    console.log("WebSocket 连接已关闭");
};

userConvIdSocket.onerror = function (error) {
    console.log("WebSocket 错误:", error);
};


function genChatMessage(messages = []) {
    const form = document.createElement('div');
    form.className = 'chat-messages';
    messages.forEach(message => {
        const messageElement = document.createElement('div');
        messageElement.className = message.user_id === userId ? 'message me' : 'message';
        messageElement.innerHTML = `<div class="header"><p class="user">${message.user_id === userId ? 'You' : message.user_name}</p><span>${message.time}</span></div><div class="content"><p>${message.text}</p></div>`;
        form.appendChild(messageElement);
    });
    return form;
}

function escapeHtml(unsafe) {
    return unsafe.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#039;');
}

function chatInputListener() {
    const textarea = document.getElementById('text-input');
    textarea.addEventListener('focus', function () {
        if (this.textContent.trim() === '') {
            this.innerHTML = '';
        }
    });
    textarea.addEventListener('blur', function () {
        if (this.textContent.trim() === '') {
            this.innerHTML = '';
        }
    });
    textarea.addEventListener('keydown', function (event) {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            const content = textarea.innerText.trim();
            if (content) {
                const escapedContent = escapeHtml(content);
                if (chatSocket && chatSocket.readyState === WebSocket.OPEN) {
                    chatSocket.send(JSON.stringify({ text: escapedContent }));
                } else {
                    console.error('WebSocket 连接未建立或已关闭');
                }
                textarea.innerText = '';
            }
        }
    });
}

function genChatHeader(uid, name, isGroupItem = false, owner) {
    const form = document.createElement('div');
    form.className = 'chat-header';
    form.innerHTML = `<span class="left"><h3>${name}</h3>${isGroupItem && owner === userId ? '<i class="fa-solid fa-crown"></i>' : ''}<p>${uid}</p></span>`;
    return form;
}
