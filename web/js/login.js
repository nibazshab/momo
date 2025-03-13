const nextStep = document.getElementById('next-step');
const phoneInput = document.getElementById('user-phone');
const passwordInput = document.getElementById('user-password');

nextStep.addEventListener('click', function () {
    const phone = phoneInput.value;
    const password = passwordInput.value;

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

    showLoading();
    fetch('http://127.0.0.1:8080/login', {
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({
            id: parseInt(phone),
            Password: password,
        })
    })
        .then(response => {
            hideLoading();
            if (response.ok) {
                window.location.href = "/";
                return response.json();
            } else {

                response.json().then(data => {
                    alert(data.error || `Login failed with status code: ${response.status}`);
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
