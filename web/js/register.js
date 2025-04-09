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
        .then(response => response.json())
        .then(data => {
            hideLoading();
            if (data.msg == "注册成功") {
                window.location.href = "/";
            } else {
                alert(data.error);
            }
        })
});
