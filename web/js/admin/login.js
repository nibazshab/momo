const tokenArea = document.getElementById('admin-token');
const loginButton = document.getElementById('admin-login');
tokenArea.focus();

loginButton.addEventListener('click', function () {
    const token = tokenArea.value;

    showLoading();
    fetch('http://127.0.0.1:8080/admin/login', {
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({ token: token })
    })
        .then(response => {
            hideLoading();
            if (response.ok) {
                window.location.href = "/admin";
                return;
            }
            else {
                response.json().then(data => {
                    alert(data.error);
                });
            }
        });
});