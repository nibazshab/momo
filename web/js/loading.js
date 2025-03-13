document.body.insertAdjacentHTML('beforeend', `
    <div id="loading" class="loading-container">
        <div class="loading-spinner"></div>
    </div>
`);

function showLoading() {
    document.getElementById('loading').style.display = 'flex';
}

function hideLoading() {
    document.getElementById('loading').style.display = 'none';
}
