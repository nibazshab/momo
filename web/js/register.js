const nextStep = document.getElementById('next-step');
const phoneInput = document.getElementById('user-phone');
const passwordInput = document.getElementById('user-password');
const nameInput = document.getElementById('user-name');

nextStep.addEventListener('click', function () {
    const phone = phoneInput.value;
    const password = passwordInput.value;
    const name = nameInput.value;

    const phoneRegex = /^1[3456789]\d{9}$/;
    if (!phoneRegex.test(phone)) {
        phoneInput.style.borderColor = '#f00';
        return;
    } else {
        phoneInput.style.borderColor = '#5682a3';
    }

    if (password === "") {
        passwordInput.style.borderColor = '#f00';
        return;
    } else {
        passwordInput.style.borderColor = '#5682a3';
    }

    if (name === "") {
        nameInput.style.borderColor = '#f00';
        return;
    } else {
        nameInput.style.borderColor = '#5682a3';
    }

    showLoading();
    fetch('http://127.0.0.1:8080/register', {
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({
            id: parseInt(phone),
            password: password,
            name: name
        })
    })
        .then(response => {
            hideLoading();
            if (response.ok) {
                window.location.href = "/";
                return response.json();
            } else {

                response.json().then(data => {
                    alert(data.error);
                }).catch(() => {
                    alert(`Login failed with status code: ${response.status}`);
                });
                throw new Error(`Login failed with status code: ${response.status}`);
            }
        })
        .then(data => {
            console.log("Successfully logged in!", data);
        })
        .catch(error => {
            hideLoading();
            console.error('Error fetching data:', error);

        });
});
