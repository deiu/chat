/*
Copyright 2025 Andrei Sambra

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

let ws = null;
let currentUsername = null;
let selectedRecipient = null;
let conversations = {};
let unreadMessages = new Set();
let currentOnlineUsers = [];

function getWebSocketUrl(path) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    return `${protocol}//${host}${path}`;
}

window.login = function() {
    const username = document.getElementById('usernameInput').value.trim();
    
    if (!username) {
        alert('Please enter a username');
        return;
    }

    currentUsername = username;
    
    ws = new WebSocket(getWebSocketUrl(`/ws?username=${encodeURIComponent(username)}`));

    ws.onmessage = function(event) {
        const data = JSON.parse(event.data);
        
        if (Array.isArray(data)) {
            updateOnlineUsers(data);
            return;
        }
    
        if (data.type === 'logout') {
            if (selectedRecipient === data.username) {
                selectedRecipient = null;
                document.getElementById('currentRecipient').textContent = '';
                document.getElementById('messageInput').disabled = true;
                document.getElementById('sendButton').disabled = true;
            }
            unreadMessages.delete(data.username);
            return;
        }
    
        handleIncomingMessage(data);
    };

    ws.onopen = function() {
        document.title = `${username}`;
        document.getElementById('currentUserDisplay').textContent = username;
        document.getElementById('login').style.display = 'none';
        document.getElementById('chat').style.display = 'flex';
    };

    ws.onclose = function(event) {
        if (!event.wasClean) {
            const usernameInput = document.getElementById('usernameInput');
            usernameInput.classList.add('error');
            usernameInput.value = '';
            usernameInput.placeholder = 'Username already taken - try another';
            setTimeout(() => {
                usernameInput.classList.remove('error');
                usernameInput.placeholder = 'Enter your username';
            }, 3000);
        }
        resetUI();
    };

    ws.onerror = function(error) {
        console.error('WebSocket error:', error);
    };
}

function logout() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        // Send logout message to server
        ws.send(JSON.stringify({
            type: 'logout',
            username: currentUsername
        }));
        ws.close();
    }
    resetUI();
}

function getConversationId(otherUser) {
    const users = [currentUsername, otherUser].sort();
    return users.join('-');
}

function getOrCreateConversation(conversationId) {
    if (!conversations[conversationId]) {
        const conversationsDiv = document.getElementById('conversations');
        const newConversation = document.createElement('div');
        newConversation.className = 'conversation';
        newConversation.id = `conversation-${conversationId}`;
        conversationsDiv.appendChild(newConversation);
        conversations[conversationId] = newConversation;
    }
    return conversations[conversationId];
}

function handleIncomingMessage(data) {
    const conversationId = getConversationId(data.from);
    let conversation = getOrCreateConversation(conversationId);
    
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message';
    messageDiv.innerHTML = `
        <div class="user-avatar" style="background-color: ${getAvatarColor(data.from)}">
            ${data.from.charAt(0).toUpperCase()}
        </div>
        <div class="message-content-wrapper">
            <div class="message-header">
                <span class="message-sender">${escapeHtml(data.from)}</span>
                <span class="message-time">${formatTime(new Date())}</span>
            </div>
            <div class="message-text">${escapeHtml(data.content)}</div>
        </div>
    `;
    conversation.appendChild(messageDiv);
    conversation.scrollTop = conversation.scrollHeight;

    if (selectedRecipient !== data.from) {
        unreadMessages.add(data.from);
        if (!document.hasFocus()) {
            document.title = `(New) Chat - ${currentUsername}`;
        }
        updateUsersList(currentOnlineUsers);
    }
}

function updateOnlineUsers(users) {
    currentOnlineUsers = users;
    const usersList = document.getElementById('usersList');
    usersList.innerHTML = '';
    
    const filteredUsers = users.filter(user => user.username !== currentUsername);
    
    filteredUsers.forEach(user => {
        const userDiv = document.createElement('div');
        userDiv.className = 'user-item';
        if (user.username === selectedRecipient) {
            userDiv.className += ' active';
        }
        if (unreadMessages.has(user.username)) {
            userDiv.className += ' has-unread';
        }
        userDiv.textContent = user.username;
        userDiv.onclick = () => selectRecipient(user.username);
        usersList.appendChild(userDiv);
    });
}

function updateUsersList(users) {
    const usersList = document.getElementById('usersList');
    usersList.innerHTML = '';
    
    const filteredUsers = users.filter(user => user.username !== currentUsername);
    
    filteredUsers.forEach(user => {
        const userDiv = document.createElement('div');
        userDiv.className = 'user-item';
        if (user.username === selectedRecipient) {
            userDiv.className += ' active';
        }
        if (unreadMessages.has(user.username)) {
            userDiv.className += ' has-unread';
        }
        userDiv.textContent = user.username;
        userDiv.onclick = () => selectRecipient(user.username);
        usersList.appendChild(userDiv);
    });
}

function selectRecipient(username) {
    selectedRecipient = username;
    document.getElementById('currentRecipient').textContent = username;
    document.getElementById('messageInput').disabled = false;
    document.getElementById('sendButton').disabled = false;
    
    // Show chat content, hide placeholder
    document.querySelector('.placeholder-screen').style.display = 'none';
    document.querySelector('.chat-content').style.display = 'flex';
    document.getElementById('chat-with-text').style.display = 'inline';

    unreadMessages.delete(username);
    document.title = `${currentUsername}`;
    
    Object.values(conversations).forEach(conv => {
        conv.style.display = 'none';
    });
    const currentConversation = getOrCreateConversation(getConversationId(username));
    currentConversation.style.display = 'block';

    const messageInput = document.getElementById('messageInput');
    messageInput.disabled = false;
    messageInput.focus();
    
    updateUsersList(currentOnlineUsers);

    hideSidebarOnMobile();
}

function sendMessage() {
    if (!ws || !selectedRecipient) return;

    const messageInput = document.getElementById('messageInput');
    const content = messageInput.value.trim();
    
    if (content) {
        const message = {
            to: selectedRecipient,
            content: content
        };
        ws.send(JSON.stringify(message));
        
        const conversationId = getConversationId(selectedRecipient);
        const conversation = getOrCreateConversation(conversationId);
        const messageDiv = document.createElement('div');
        messageDiv.className = 'message';
        messageDiv.innerHTML = `
            <div class="user-avatar" style="background-color: ${getAvatarColor(currentUsername)}">
                ${currentUsername.charAt(0).toUpperCase()}
            </div>
            <div class="message-content-wrapper">
                <div class="message-header">
                    <span class="message-sender">You</span>
                    <span class="message-time">${formatTime(new Date())}</span>
                </div>
                <div class="message-text">${escapeHtml(content)}</div>
            </div>
        `;
        conversation.appendChild(messageDiv);
        conversation.scrollTop = conversation.scrollHeight;
        
        messageInput.value = '';
    }
}

function resetUI() {
    document.getElementById('login').style.display = 'block';
    document.getElementById('chat').style.display = 'none';
    document.getElementById('conversations').innerHTML = '';
    document.getElementById('usersList').innerHTML = '';
    document.getElementById('usernameInput').value = '';
    document.getElementById('currentRecipient').textContent = '';
    document.getElementById('currentUserDisplay').textContent = '';
    document.querySelector('.placeholder-screen').style.display = 'flex';
    document.querySelector('.chat-content').style.display = 'none';
    document.getElementById('chat-with-text').style.display = 'none';
    document.title = 'Direct Chat';
    currentUsername = null;
    selectedRecipient = null;
    conversations = {};
    unreadMessages.clear();
    currentOnlineUsers = [];
    ws = null;
}

function toggleSidebar() {
    const sidebar = document.querySelector('.sidebar');
    sidebar.classList.toggle('show');
}

function hideSidebarOnMobile() {
    if (window.innerWidth <= 768) {
        const sidebar = document.querySelector('.sidebar');
        sidebar.classList.remove('show');
    }
}

function getAvatarColor(username) {
    // Generate consistent color based on username
    let hash = 0;
    for (let i = 0; i < username.length; i++) {
        hash = username.charCodeAt(i) + ((hash << 5) - hash);
    }
    const hue = hash % 360;
    return `hsl(${hue}, 70%, 45%)`; // Consistent, vibrant colors
}

function formatTime(date) {
    return date.toLocaleTimeString([], { 
        hour: '2-digit', 
        minute: '2-digit',
        hour12: false
    });
}

function escapeHtml(unsafe) {
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

window.onfocus = function() {
    if (currentUsername) {
        document.title = `Chat - ${currentUsername}`;
    }
};

document.getElementById('usernameInput').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        login();
    }
});

document.getElementById('messageInput').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        sendMessage();
    }
});

document.addEventListener('click', function(e) {
    if (window.innerWidth <= 768) {
        const sidebar = document.querySelector('.sidebar');
        const menuToggle = document.querySelector('.menu-toggle');
        
        if (!sidebar.contains(e.target) && !menuToggle.contains(e.target) && sidebar.classList.contains('show')) {
            sidebar.classList.remove('show');
        }
    }
});

window.addEventListener('resize', function() {
    if (window.innerWidth > 768) {
        const sidebar = document.querySelector('.sidebar');
        sidebar.classList.remove('show');
    }
});