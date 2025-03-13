function renderList(containerId, data, itemType) {
    const container = document.getElementById(containerId);

    if (!container) {
        console.error(`Container with ID "${containerId}" not found.`);
        return;
    }

    const items = data;

    if (!items || !Array.isArray(items)) {
        console.warn(`No data found for itemType "${itemType}".  Expected an array.`);
        return;
    }

    items.forEach(item => {
        const listItem = document.createElement('div');
        listItem.className = `${itemType}-item`;

        listItem.setAttribute('data-uid', item.id);
        listItem.setAttribute('data-name', item.name);

        const info = document.createElement('div');
        info.className = 'info';

        const name = document.createElement('p');
        name.textContent = item.name;

        info.appendChild(name);

        const icon = document.createElement('i');

        if (itemType === "user") {
            const id = document.createElement('p');
            id.textContent = item.id;
            id.className = 'user-id';
            info.appendChild(id);

            icon.className = 'fa-solid fa-user';
        } else {

            listItem.setAttribute('data-owner', item.owner_id);
            icon.className = 'fa-solid fa-user-group';
        }

        listItem.appendChild(icon);
        listItem.appendChild(info);

        container.appendChild(listItem);
    });
}

function clearSelectedItems() {
    const selectedItems = document.querySelectorAll('.group-item.selected, .user-item.selected, .setting-item.selected');
    selectedItems.forEach(item => item.classList.remove('selected'));
}
